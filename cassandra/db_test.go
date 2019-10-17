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

package cassandra

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	db "github.com/xmidt-org/codex-db"
	"github.com/xmidt-org/webpa-common/xmetrics/xmetricstest"
	"github.com/xmidt-org/wrp-go/wrp"
	"testing"
	"time"
)

var (
	goodEvent = wrp.SimpleEvent{
		Type:        wrp.SimpleEventMessageType,
		Source:      "test source",
		Destination: "testdestination",
		Metadata:    map[string]string{"test key": "test value"},
	}
)

func TestGetRecords(t *testing.T) {
	tests := []struct {
		description           string
		deviceID              string
		expectedRecords       []db.Record
		expectedSuccessMetric float64
		expectedFailureMetric float64
		expectedErr           error
		expectedCalls         int
	}{
		{
			description: "Success",
			deviceID:    "1234",
			expectedRecords: []db.Record{
				{
					Type:     0,
					DeviceID: "1234",
				},
			},
			expectedSuccessMetric: 1.0,
			expectedErr:           nil,
			expectedCalls:         1,
		},
		{
			description:           "Get Error",
			deviceID:              "1234",
			expectedRecords:       []db.Record{},
			expectedFailureMetric: 1.0,
			expectedErr:           errors.New("test Get error"),
			expectedCalls:         1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockFinder)
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)
			dbConnection := Connection{
				measures: m,
				finder:   mockObj,
			}
			if tc.expectedCalls > 0 {
				marshaledRecords, err := json.Marshal(tc.expectedRecords)
				assert.Nil(err)
				mockObj.On("findRecords", mock.Anything, mock.Anything, mock.Anything).Return(marshaledRecords, tc.expectedErr).Times(tc.expectedCalls)
			}
			p.Assert(t, SQLQuerySuccessCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLQueryFailureCounter)(xmetricstest.Value(0.0))

			records, err := dbConnection.GetRecords(tc.deviceID, 5)
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQuerySuccessCounter, db.TypeLabel, db.ReadType)(xmetricstest.Value(tc.expectedSuccessMetric))
			p.Assert(t, SQLQueryFailureCounter, db.TypeLabel, db.ReadType)(xmetricstest.Value(tc.expectedFailureMetric))
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
			assert.Equal(tc.expectedRecords, records)
		})
	}
}

func TestGetRecordsOfType(t *testing.T) {
	tests := []struct {
		description           string
		deviceID              string
		eventType             db.EventType
		expectedRecords       []db.Record
		expectedSuccessMetric float64
		expectedFailureMetric float64
		expectedErr           error
		expectedCalls         int
	}{
		{
			description: "Success",
			deviceID:    "1234",
			eventType:   1,
			expectedRecords: []db.Record{
				{
					Type:     1,
					DeviceID: "1234",
				},
			},
			expectedSuccessMetric: 1.0,
			expectedErr:           nil,
			expectedCalls:         1,
		},
		{
			description:           "Get Error",
			deviceID:              "1234",
			expectedRecords:       []db.Record{},
			expectedFailureMetric: 1.0,
			expectedErr:           errors.New("test Get error"),
			expectedCalls:         1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockFinder)
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)
			dbConnection := Connection{
				measures: m,
				finder:   mockObj,
			}
			if tc.expectedCalls > 0 {
				marshaledRecords, err := json.Marshal(tc.expectedRecords)
				assert.Nil(err)
				mockObj.On("findRecords", mock.Anything, mock.Anything, mock.Anything).Return(marshaledRecords, tc.expectedErr).Times(tc.expectedCalls)
			}
			p.Assert(t, SQLQuerySuccessCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLQueryFailureCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLReadRecordsCounter)(xmetricstest.Value(0.0))

			records, err := dbConnection.GetRecordsOfType(tc.deviceID, 5, tc.eventType)
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQuerySuccessCounter, db.TypeLabel, db.ReadType)(xmetricstest.Value(tc.expectedSuccessMetric))
			p.Assert(t, SQLQueryFailureCounter, db.TypeLabel, db.ReadType)(xmetricstest.Value(tc.expectedFailureMetric))
			p.Assert(t, SQLReadRecordsCounter)(xmetricstest.Value(float64(len(tc.expectedRecords))))
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
			assert.Equal(tc.expectedRecords, records)
		})
	}
}

func TestDeviceList(t *testing.T) {
	tests := []struct {
		description           string
		expectedIDs           []string
		expectedSuccessMetric float64
		expectedFailureMetric float64
		expectedErr           error
	}{
		{
			description:           "Success",
			expectedIDs:           []string{"aaa", "bbb", "ccc"},
			expectedSuccessMetric: 1.0,
			expectedErr:           nil,
		},
		{
			description:           "Get Error",
			expectedIDs:           []string{},
			expectedFailureMetric: 1.0,
			expectedErr:           errors.New("test Get error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockDeviceFinder)
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)
			dbConnection := Connection{
				measures:     m,
				deviceFinder: mockObj,
			}
			mockObj.On("getList", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.expectedIDs, tc.expectedErr).Once()
			p.Assert(t, SQLQuerySuccessCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLQueryFailureCounter)(xmetricstest.Value(0.0))

			result, err := dbConnection.GetDeviceList(time.Now(), time.Now(), 0, 10)
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQuerySuccessCounter, db.TypeLabel, db.ReadType)(xmetricstest.Value(tc.expectedSuccessMetric))
			p.Assert(t, SQLQueryFailureCounter, db.TypeLabel, db.ReadType)(xmetricstest.Value(tc.expectedFailureMetric))
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
			assert.Equal(tc.expectedIDs, result)
		})
	}
}

func TestMultiInsertEvent(t *testing.T) {
	testCreateErr := errors.New("test create error")
	goodRecord := db.Record{
		DeviceID: "1234",
	}

	tests := []struct {
		description           string
		records               []db.Record
		expectedSuccessMetric float64
		expectedFailureMetric float64
		createErr             error
		expectedErr           error
		expectedCalls         int
	}{
		{
			description:           "Success",
			records:               []db.Record{goodRecord, {}},
			expectedSuccessMetric: 1.0,
			expectedErr:           nil,
			expectedCalls:         1,
		},
		{
			description:           "Create Error",
			records:               []db.Record{goodRecord, {DeviceID: "54321"}},
			expectedFailureMetric: 1.0,
			createErr:             testCreateErr,
			expectedErr:           testCreateErr,
			expectedCalls:         1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockMultiInsert)
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)
			dbConnection := Connection{
				measures:    m,
				mutliInsert: mockObj,
			}
			if tc.expectedCalls > 0 {
				mockObj.On("insert", mock.Anything).Return(3, tc.createErr).Times(tc.expectedCalls)
			}
			p.Assert(t, SQLQuerySuccessCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLQueryFailureCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLInsertedRecordsCounter)(xmetricstest.Value(0.0))

			err := dbConnection.InsertRecords(tc.records...)
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQuerySuccessCounter, db.TypeLabel, db.InsertType)(xmetricstest.Value(tc.expectedSuccessMetric))
			p.Assert(t, SQLQueryFailureCounter, db.TypeLabel, db.InsertType)(xmetricstest.Value(tc.expectedFailureMetric))
			p.Assert(t, SQLInsertedRecordsCounter)(xmetricstest.Value(3.0))
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		description string
		expectedErr error
	}{
		{
			description: "Success",
			expectedErr: nil,
		},
		{
			description: "Close Error",
			expectedErr: errors.New("test close error"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockCloser)
			dbConnection := Connection{
				closer: mockObj,
				stopThreads: []chan struct{}{
					make(chan struct{}, 10),
				},
			}
			mockObj.On("close").Return(tc.expectedErr).Once()
			err := dbConnection.Close()
			mockObj.AssertExpectations(t)
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestPing(t *testing.T) {
	tests := []struct {
		description           string
		expectedSuccessMetric float64
		expectedFailureMetric float64
		expectedErr           error
	}{
		{
			description:           "Success",
			expectedSuccessMetric: 1.0,
			expectedErr:           nil,
		},
		{
			description:           "Ping Error",
			expectedFailureMetric: 1.0,
			expectedErr:           errors.New("test ping error"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			assert := assert.New(t)
			mockObj := new(mockPing)
			p := xmetricstest.NewProvider(nil, Metrics)
			m := NewMeasures(p)
			dbConnection := Connection{
				measures: m,
				pinger:   mockObj,
			}
			mockObj.On("ping").Return(tc.expectedErr).Once()
			p.Assert(t, SQLQuerySuccessCounter)(xmetricstest.Value(0.0))
			p.Assert(t, SQLQueryFailureCounter)(xmetricstest.Value(0.0))

			err := dbConnection.Ping()
			mockObj.AssertExpectations(t)
			p.Assert(t, SQLQuerySuccessCounter, db.TypeLabel, db.PingType)(xmetricstest.Value(tc.expectedSuccessMetric))
			p.Assert(t, SQLQueryFailureCounter, db.TypeLabel, db.PingType)(xmetricstest.Value(tc.expectedFailureMetric))
			if tc.expectedErr == nil || err == nil {
				assert.Equal(tc.expectedErr, err)
			} else {
				assert.Contains(err.Error(), tc.expectedErr.Error())
			}
		})
	}
}

func TestImplementsInterfaces(t *testing.T) {
	var (
		dbConn interface{}
	)
	assert := assert.New(t)
	dbConn = &Connection{}
	_, ok := dbConn.(db.Inserter)
	assert.True(ok, "not an inserter")
	_, ok = dbConn.(db.RecordGetter)
	assert.True(ok, "not an record getter")
}
