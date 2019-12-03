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

type retryConfig struct {
	backoffConfig backoff.ExponentialBackOff
	measures      Measures
}

// Option is the function used to configure the retry object.
type Option func(r *retryConfig)

// WithBackoff sets the exponential backoff to use when retrying.  If this
// isn't called, we use the backoff package's default ExponentialBackoff
// configuration.  If any values are considered invalid, they are replaced with
// those defaults.
func WithBackoff(b backoff.ExponentialBackOff) Option {
	return func(r *retryConfig) {
		r.backoffConfig = b
		if r.backoffConfig.InitialInterval < 0 {
			r.backoffConfig.InitialInterval = backoff.DefaultInitialInterval
		}
		if r.backoffConfig.RandomizationFactor < 0 {
			r.backoffConfig.RandomizationFactor = backoff.DefaultRandomizationFactor
		}
		if r.backoffConfig.Multiplier < 1 {
			r.backoffConfig.Multiplier = backoff.DefaultMultiplier
		}
		if r.backoffConfig.MaxInterval < 0 {
			r.backoffConfig.MaxInterval = backoff.DefaultMaxInterval
		}
		if r.backoffConfig.MaxElapsedTime < 0 {
			r.backoffConfig.MaxElapsedTime = backoff.DefaultMaxElapsedTime
		}
		if r.backoffConfig.Clock == nil {
			r.backoffConfig.Clock = backoff.SystemClock
		}
	}
}

// WithMeasures sets a provider to use for metrics.
func WithMeasures(p provider.Provider) Option {
	return func(r *retryConfig) {
		if p != nil {
			r.measures = NewMeasures(p)
		}
	}
}

// RetryInsertService is a wrapper for a db.Inserter.  If inserting fails, the
// retry service will continue to try until the configurable max elapsed time
// is reached.  The retries will exponentially backoff in the manner configured.
// To read more about this, see the backoff package GoDoc:
// https://godoc.org/gopkg.in/cenkalti/backoff.v3
type RetryInsertService struct {
	inserter db.Inserter
	config   retryConfig
}

// AddRetryMetric is a function to add to our metrics when we retry.  The
// function is passed to the backoff package and is called when we are retrying.
func (ri RetryInsertService) AddRetryMetric(_ error, _ time.Duration) {
	ri.config.measures.SQLQueryRetryCount.With(db.TypeLabel, db.InsertType).Add(1.0)
}

// InsertRecords uses the inserter to insert the records and uses the
// ExponentialBackoff to try again if inserting fails.
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
			backoffConfig: *backoff.NewExponentialBackOff(),
		},
	}
	for _, o := range options {
		o(&ris.config)
	}
	return ris
}
