// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

// package batchInserter provides a wrapper around the db.Inserter to provide a
// way to group records together before inserting, in order to decrease
// database requests needed for inserting.
package batchInserter

import (
	"errors"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/goph/emperror"
	db "github.com/xmidt-org/codex-db"
	"github.com/xmidt-org/webpa-common/v2/logging"
	"github.com/xmidt-org/webpa-common/v2/semaphore"
)

const (
	minParseWorkers      = 1
	minInsertWorkers     = 1
	defaultInsertWorkers = 5
	minMaxBatchSize      = 0
	defaultMaxBatchSize  = 1
	minMaxBatchWaitTime  = time.Duration(1) * time.Millisecond
	defaultMinQueueSize  = 5
)

var (
	defaultLogger = log.NewNopLogger()
)

var (
	ErrBadBeginning = errors.New("invalid value for the beginning time of the record")
	ErrBadData      = errors.New("data nil or empty")
)

// defaultTicker is the production code that produces a ticker.  Note that we don't
// want to return *time.Ticker, as we want to be able to inject something for testing.
// We also need to return a closure to stop the ticker, so that we can call ticker.Stop() without
// being dependent on the *time.Ticker interface.
func defaultTicker(d time.Duration) (<-chan time.Time, func()) {
	ticker := time.NewTicker(d)
	return ticker.C, ticker.Stop
}

type TimeTracker interface {
	TrackTime(time.Duration)
}

// BatchInserter manages batching events that need to be inserted, ensuring
// that an event that needs to be inserted isn't waiting for longer than a set
// period of time and that each batch doesn't pass a specified size.
type BatchInserter struct {
	numBatchers   int
	insertQueue   chan RecordWithTime
	inserter      db.Inserter
	timeTracker   TimeTracker
	insertWorkers semaphore.Interface
	wg            sync.WaitGroup
	measures      *Measures
	logger        log.Logger
	config        Config
	ticker        func(time.Duration) (<-chan time.Time, func())
}

// Config holds the configuration values for a batch inserter.
type Config struct {
	ParseWorkers     int
	MaxInsertWorkers int
	MaxBatchSize     int
	MaxBatchWaitTime time.Duration
	QueueSize        int
}

// RecordWithTime provides the db record and the time this event was received by a service
type RecordWithTime struct {
	Record    db.Record
	Beginning time.Time
}

// NewBatchInserter creates a BatchInserter with the given values, ensuring
// that the configuration and other values given are valid.  If configuration
// values aren't valid, a default value is used.
func NewBatchInserter(config Config, logger log.Logger, metricsRegistry provider.Provider, inserter db.Inserter, timeTracker TimeTracker) (*BatchInserter, error) {
	if inserter == nil {
		return nil, errors.New("no inserter")
	}
	if config.ParseWorkers < minParseWorkers {
		config.ParseWorkers = minParseWorkers
	}
	if config.MaxInsertWorkers < minInsertWorkers {
		config.MaxInsertWorkers = defaultInsertWorkers
	}
	if config.MaxBatchSize < minMaxBatchSize {
		config.MaxBatchSize = defaultMaxBatchSize
	}
	if config.MaxBatchWaitTime < minMaxBatchWaitTime {
		config.MaxBatchWaitTime = minMaxBatchWaitTime
	}
	if config.QueueSize < defaultMinQueueSize {
		config.QueueSize = defaultMinQueueSize
	}
	if logger == nil {
		logger = defaultLogger
	}

	measures := NewMeasures(metricsRegistry)
	workers := semaphore.New(config.MaxInsertWorkers)
	queue := make(chan RecordWithTime, config.QueueSize)
	b := BatchInserter{
		config:        config,
		logger:        logger,
		measures:      measures,
		numBatchers:   config.ParseWorkers,
		insertWorkers: workers,
		inserter:      inserter,
		insertQueue:   queue,
		ticker:        defaultTicker,
		timeTracker:   timeTracker,
	}
	return &b, nil
}

// Start starts the batcher, which pulls from the queue inside of the BatchInserter.
func (b *BatchInserter) Start() {
	for i := 0; i < b.numBatchers; i++ {
		b.wg.Add(1)
		go b.batchRecords()
	}
}

// Insert adds the event to the queue inside of BatchInserter, preparing for it
// to be inserted.  This can block, if the queue is full.  If the record has
// certain fields empty, an error is returned.
func (b *BatchInserter) Insert(rwt RecordWithTime) error {
	if b.timeTracker != nil && rwt.Beginning.IsZero() {
		return ErrBadBeginning
	}
	if rwt.Record.Data == nil || len(rwt.Record.Data) == 0 {
		return ErrBadData
	}
	b.insertQueue <- rwt
	if b.measures != nil {
		b.measures.InsertingQueue.Add(1.0)
	}
	return nil
}

// Stop closes the internal queue and waits for the workers to finish
// processing what has already been added.  This can block as it waits for
// everything to stop.  After Stop() is called, Insert() should not be called
// again, or there will be a panic.
// TODO: ensure consumers can't cause a panic?
func (b *BatchInserter) Stop() {
	close(b.insertQueue)
	b.wg.Wait()

	// Grab all the workers to make sure they are done.
	for i := 0; i < b.config.MaxInsertWorkers; i++ {
		b.insertWorkers.Acquire()
	}
}

func (b *BatchInserter) batchRecords() {
	var (
		insertRecords bool
		ticker        <-chan time.Time
		stop          func()
	)
	defer b.wg.Done()
	for rwt := range b.insertQueue {
		if b.measures != nil {
			b.measures.InsertingQueue.Add(-1.0)
		}
		ticker, stop = b.ticker(b.config.MaxBatchWaitTime)
		records := []db.Record{rwt.Record}
		beginTimes := []time.Time{rwt.Beginning}
		for {
			select {
			case <-ticker:
				insertRecords = true
			case r, ok := <-b.insertQueue:
				// if ok is false, the queue is closed.
				if !ok {
					insertRecords = true
					break
				}
				if b.measures != nil {
					b.measures.InsertingQueue.Add(-1.0)
				}
				records = append(records, r.Record)
				beginTimes = append(beginTimes, r.Beginning)
				if b.config.MaxBatchSize != 0 && len(records) >= b.config.MaxBatchSize {
					insertRecords = true
					break
				}
			}
			if insertRecords {
				b.insertWorkers.Acquire()
				go b.insertRecords(records, beginTimes)
				insertRecords = false
				break
			}
		}
		stop()
	}
}

func (b *BatchInserter) insertRecords(records []db.Record, beginTimes []time.Time) {
	defer b.insertWorkers.Release()
	err := b.inserter.InsertRecords(records...)
	if err != nil {
		if b.measures != nil {
			b.measures.DroppedEventsFromDbFailCount.Add(float64(len(records)))
		}
		logging.Error(b.logger, emperror.Context(err)...).Log(logging.MessageKey(),
			"Failed to add records to the database", logging.ErrorKey(), err.Error())
		b.sendTimes(beginTimes, time.Now())
		return
	}
	b.sendTimes(beginTimes, time.Now())
	logging.Debug(b.logger).Log(logging.MessageKey(), "Successfully upserted device information", "records", records)
	logging.Info(b.logger).Log(logging.MessageKey(), "Successfully upserted device information", "records", len(records))
}

func (b *BatchInserter) sendTimes(beginTimes []time.Time, endTime time.Time) {
	if b.timeTracker != nil {
		for _, beginTime := range beginTimes {
			b.timeTracker.TrackTime(endTime.Sub(beginTime))
		}
		logging.Debug(b.logger).Log(logging.MessageKey(), "Successfully tracked time taken to insert records")
	}
}
