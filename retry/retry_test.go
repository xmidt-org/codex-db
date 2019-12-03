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

package dbretry

import (
	"errors"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	db "github.com/xmidt-org/codex-db"
	"github.com/xmidt-org/webpa-common/xmetrics/xmetricstest"
)

func TestRetryInsertRecords(t *testing.T) {
	initialErr := errors.New("test initial error")
	failureErr := errors.New("test final error")
	tests := []struct {
		description         string
		numCalls            int
		maxElapsedTime      time.Duration
		expectedRetryMetric float64
		finalError          error
		expectedErr         error
	}{
		{
			description:    "Initial Success",
			numCalls:       1,
			maxElapsedTime: 1,
			finalError:     nil,
			expectedErr:    nil,
		},
		{
			description:         "Eventual Success",
			numCalls:            3,
			maxElapsedTime:      1 * time.Minute,
			expectedRetryMetric: 2.0,
			finalError:          nil,
			expectedErr:         nil,
		},
		{
			description:         "Eventual Failure",
			numCalls:            4,
			maxElapsedTime:      10 * time.Millisecond,
			expectedRetryMetric: -1,
			finalError:          failureErr,
			expectedErr:         failureErr,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockInserter)
			if tc.numCalls > 1 {
				mockObj.On("InsertRecords", mock.Anything).Return(initialErr).Times(tc.numCalls - 1)
			}
			if tc.numCalls > 0 {
				if tc.finalError == nil {
					mockObj.On("InsertRecords", mock.Anything).Return(nil).Once()
				} else {
					mockObj.On("InsertRecords", mock.Anything).Return(tc.finalError)
				}
			}
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)

			retryInsertService := RetryInsertService{
				inserter: mockObj,
				config: retryConfig{
					backoffConfig: backoff.ExponentialBackOff{
						InitialInterval:     1,
						RandomizationFactor: 0,
						Multiplier:          10,
						MaxInterval:         2000,
						MaxElapsedTime:      tc.maxElapsedTime,
						Clock:               backoff.SystemClock,
					},
					measures: m,
				},
			}
			p.Assert(t, SQLQueryRetryCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLQueryEndCounter)(xmetricstest.Value(0.0))
			err := retryInsertService.InsertRecords(db.Record{})
			mockObj.AssertExpectations(t)
			if tc.expectedRetryMetric >= 0 {
				p.Assert(t, SQLQueryRetryCounter, db.TypeLabel, db.InsertType)(xmetricstest.Value(tc.expectedRetryMetric))
			}
			p.Assert(t, SQLQueryEndCounter, db.TypeLabel, db.InsertType)(xmetricstest.Value(1.0))
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

type constClock struct{}

func (c *constClock) Now() time.Time {
	return time.Unix(0, 0)
}

func TestCreateRetryInsertService(t *testing.T) {
	r := RetryInsertService{
		inserter: new(mockInserter),
		config: retryConfig{
			backoffConfig: backoff.ExponentialBackOff{
				InitialInterval:     30,
				RandomizationFactor: 0.01,
				Multiplier:          200,
				MaxInterval:         50000,
				MaxElapsedTime:      0,
				Clock:               &constClock{},
			},
		},
	}
	assert := assert.New(t)
	p := xmetricstest.NewProvider(nil, Metrics)
	newService := CreateRetryInsertService(r.inserter, WithBackoff(r.config.backoffConfig), WithMeasures(p))
	assert.Equal(r.inserter, newService.inserter)
	assert.Equal(r.config.backoffConfig, newService.config.backoffConfig)
}

func TestCreateRetryInsertServiceUseDefaults(t *testing.T) {
	r := RetryInsertService{
		inserter: new(mockInserter),
		config: retryConfig{
			backoffConfig: backoff.ExponentialBackOff{
				InitialInterval:     -10,
				RandomizationFactor: -1,
				Multiplier:          0,
				MaxInterval:         -10,
				MaxElapsedTime:      -1,
				Clock:               nil,
			},
		},
	}
	assert := assert.New(t)
	newService := CreateRetryInsertService(r.inserter, WithBackoff(r.config.backoffConfig), WithMeasures(nil))
	assert.Equal(r.inserter, newService.inserter)
	assert.Equal(backoff.NewExponentialBackOff().InitialInterval, newService.config.backoffConfig.InitialInterval)
	assert.Equal(backoff.NewExponentialBackOff().RandomizationFactor, newService.config.backoffConfig.RandomizationFactor)
	assert.Equal(backoff.NewExponentialBackOff().Multiplier, newService.config.backoffConfig.Multiplier)
	assert.Equal(backoff.NewExponentialBackOff().MaxInterval, newService.config.backoffConfig.MaxInterval)
	assert.Equal(backoff.NewExponentialBackOff().MaxElapsedTime, newService.config.backoffConfig.MaxElapsedTime)
	assert.Equal(backoff.NewExponentialBackOff().Clock, newService.config.backoffConfig.Clock)
}
