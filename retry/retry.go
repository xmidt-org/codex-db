/**
 * Copyright 2019 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package dbretry contains structs that implement various db interfaces as
// well as consume them.  They allow consumers to easily try to interact with
// the database a configurable number of times, with configurable backoff
// options and metrics.
package dbretry

import (
	"time"

	"github.com/cenkalti/backoff/v3"
	"github.com/go-kit/kit/metrics/provider"
	db "github.com/xmidt-org/codex-db"
)

var (
	defaultBackoff = backoff.ExponentialBackOff{
		InitialInterval:     time.Second,
		RandomizationFactor: 0.1,
		Multiplier:          5,
		MaxInterval:         7 * time.Second,
		MaxElapsedTime:      10 * time.Second,
		Clock:               backoff.SystemClock,
	}
)

type retryConfig struct {
	backoffConfig backoff.ExponentialBackOff
	measures      Measures
}

// Option is the function used to configure the retry objects.
type Option func(r *retryConfig)

// WithBackoff sets the backoff to use when retrying.  By default, this is
// an exponential backoff.
func WithBackoff(b backoff.ExponentialBackOff) Option {
	return func(r *retryConfig) {
		r.backoffConfig = b
	}
}

// WithMeasures provides a provider to use for metrics.
func WithMeasures(p provider.Provider) Option {
	return func(r *retryConfig) {
		if p != nil {
			r.measures = NewMeasures(p)
		}
	}
}

// RetryInsertService is a wrapper for a db.Inserter that attempts to insert
// a configurable number of times if the inserts fail.
type RetryInsertService struct {
	inserter db.Inserter
	config   retryConfig
}

// AddRetryMetric is a function to add to our metrics when we retry.  The
// function is passed to the backoff package and is called when we are retrying.
func (ri RetryInsertService) AddRetryMetric(_ error, _ time.Duration) {
	ri.config.measures.SQLQueryRetryCount.With(db.TypeLabel, db.InsertType).Add(1.0)
}

// InsertRecords uses the inserter to insert the records and tries again if
// inserting fails.  Between each try, it calculates how long to wait and then
// waits for that period of time before trying again. Only the error from the
// last failure is returned.
func (ri RetryInsertService) InsertRecords(records ...db.Record) error {

	insertFunc := func() error {
		return ri.inserter.InsertRecords(records...)
	}

	// with every insert, we have to make a copy of the ExponentialBackoff
	// struct, as it is not thread safe, and each thread needs its own clock.
	b := ri.config.backoffConfig

	err := backoff.RetryNotify(insertFunc, &b, ri.AddRetryMetric)
	ri.config.measures.SQLQueryEndCount.With(db.TypeLabel, db.InsertType).Add(1.0)
	return err
}

// CreateRetryInsertService takes an inserter and the options provided and
// creates a RetryInsertService.
func CreateRetryInsertService(inserter db.Inserter, options ...Option) RetryInsertService {
	ris := RetryInsertService{
		inserter: inserter,
		config: retryConfig{
			backoffConfig: defaultBackoff,
		},
	}
	for _, o := range options {
		o(&ris.config)
	}
	return ris
}
