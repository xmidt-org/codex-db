// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package batchInserter

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/xmidt-org/webpa-common/v2/semaphore"

	"github.com/go-kit/kit/log"

	"github.com/go-kit/kit/metrics/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	db "github.com/xmidt-org/codex-db"
	"github.com/xmidt-org/webpa-common/v2/xmetrics/xmetricstest"
)

func TestNewBatchInserter(t *testing.T) {
	goodInserter := new(mockInserter)
	goodRegistry := xmetricstest.NewProvider(nil, Metrics)
	goodMeasures := NewMeasures(goodRegistry)
	goodConfig := Config{
		QueueSize:        1000,
		ParseWorkers:     50,
		MaxInsertWorkers: 5000,
		MaxBatchSize:     100,
		MaxBatchWaitTime: 5 * time.Hour,
	}
	tests := []struct {
		description           string
		config                Config
		inserter              db.Inserter
		logger                log.Logger
		registry              provider.Provider
		expectedBatchInserter *BatchInserter
		expectedErr           error
	}{
		{
			description: "Success",
			config:      goodConfig,
			inserter:    goodInserter,
			logger:      log.NewJSONLogger(os.Stdout),
			registry:    goodRegistry,
			expectedBatchInserter: &BatchInserter{
				inserter: goodInserter,
				measures: goodMeasures,
				config:   goodConfig,
				logger:   log.NewJSONLogger(os.Stdout),
			},
		},
		{
			description: "Success With Defaults",
			config: Config{
				MaxBatchSize:     -5,
				MaxBatchWaitTime: -2 * time.Minute,
			},
			inserter: goodInserter,
			registry: goodRegistry,
			expectedBatchInserter: &BatchInserter{
				inserter: goodInserter,
				measures: goodMeasures,
				config: Config{
					MaxBatchSize:     defaultMaxBatchSize,
					MaxBatchWaitTime: minMaxBatchWaitTime,
					QueueSize:        defaultMinQueueSize,
					ParseWorkers:     minParseWorkers,
					MaxInsertWorkers: defaultInsertWorkers,
				},
				logger: defaultLogger,
			},
		},
		{
			description: "Nil Inserter Error",
			expectedErr: errors.New("no inserter"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			bi, err := NewBatchInserter(tc.config, tc.logger, tc.registry, tc.inserter, nil)
			if tc.expectedBatchInserter == nil || bi == nil {
				assert.Equal(tc.expectedBatchInserter, bi)
			} else {
				assert.Equal(tc.expectedBatchInserter.inserter, bi.inserter)
				assert.Equal(tc.expectedBatchInserter.measures, bi.measures)
				assert.Equal(tc.expectedBatchInserter.config, bi.config)
				assert.Equal(tc.expectedBatchInserter.logger, bi.logger)
			}
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestBatchInserter(t *testing.T) {
	beginTime := time.Now()
	records := []db.Record{
		{
			Type: db.State,
			Data: []byte("test1"),
		},
		{
			Type: db.State,
			Data: []byte("test2"),
		},
		{
			Type: db.State,
			Data: []byte("test3"),
		},
		{
			Type: db.State,
			Data: []byte("test4"),
		},
		{
			Type: db.State,
			Data: []byte("test5"),
		},
	}
	tests := []struct {
		description           string
		insertErr             error
		recordsToInsert       []db.Record
		badBeginning          bool
		recordsExpected       [][]db.Record
		waitBtwnRecords       time.Duration
		expectedDroppedEvents float64
		expectStopCalled      bool
		expectedErr           error
	}{
		{
			description:     "Success",
			waitBtwnRecords: 1 * time.Millisecond,
			recordsToInsert: records[:5],
			recordsExpected: [][]db.Record{
				records[:3],
				records[3:5],
			},
			expectStopCalled: true,
		},
		{
			description:     "Nil Record",
			recordsToInsert: []db.Record{{}},
			expectedErr:     ErrBadData,
		},
		{
			description:     "Missing Beginning for Record",
			recordsToInsert: []db.Record{records[0]},
			badBeginning:    true,
			expectedErr:     ErrBadBeginning,
		},
		{
			description:     "Insert Records Error",
			recordsToInsert: records[3:5],
			waitBtwnRecords: 1 * time.Millisecond,
			recordsExpected: [][]db.Record{
				records[3:5],
			},
			insertErr:             errors.New("test insert error"),
			expectedDroppedEvents: 2,
			expectStopCalled:      true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			inserter := new(mockInserter)
			tracker := new(mockTracker)
			for _, r := range tc.recordsExpected {
				inserter.On("InsertRecords", r).Return(tc.insertErr).Once()
				tracker.On("TrackTime", mock.Anything).Times(len(r))
			}
			queue := make(chan RecordWithTime, 5)
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)
			stopCalled := false
			stop := func() {
				stopCalled = true
			}
			tickerChan := make(chan time.Time, 1)
			b := BatchInserter{
				config: Config{
					MaxBatchWaitTime: 10 * time.Millisecond,
					MaxBatchSize:     3,
					ParseWorkers:     1,
					MaxInsertWorkers: 5,
				},
				inserter:      inserter,
				insertQueue:   queue,
				insertWorkers: semaphore.New(5),
				measures:      m,
				logger:        log.NewNopLogger(),
				ticker: func(d time.Duration) (<-chan time.Time, func()) {
					return tickerChan, stop
				},
				timeTracker: tracker,
			}
			p.Assert(t, DroppedEventsFromDbFailCounter)(xmetricstest.Value(0))
			b.wg.Add(1)
			go b.batchRecords()
			for i, r := range tc.recordsToInsert {
				if i > 0 {
					time.Sleep(tc.waitBtwnRecords)
				}
				rwt := RecordWithTime{
					Record: r,
				}
				if !tc.badBeginning {
					rwt.Beginning = beginTime
				}
				err := b.Insert(rwt)
				if tc.expectedErr == nil || err == nil {
					assert.Equal(tc.expectedErr, err)
				} else {
					assert.Contains(err.Error(), tc.expectedErr.Error())
				}
			}
			tickerChan <- time.Now()
			b.Stop()
			inserter.AssertExpectations(t)
			assert.Equal(tc.expectStopCalled, stopCalled)
			p.Assert(t, DroppedEventsFromDbFailCounter)(xmetricstest.Value(tc.expectedDroppedEvents))
		})
	}
}
