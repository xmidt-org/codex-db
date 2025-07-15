// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package cassandra

import (
	"encoding/json"
	"github.com/stretchr/testify/mock"
	db "github.com/xmidt-org/codex-db"
	"time"
)

type mockFinder struct {
	mock.Mock
}

func (f *mockFinder) findRecords(limit int, filter string, where ...interface{}) ([]db.Record, error) {
	args := f.Called(limit, filter, where)
	result := make([]db.Record, 0)
	err := json.Unmarshal(args.Get(0).([]byte), &result)
	if err != nil {
		return []db.Record{}, err
	}
	return result, args.Error(1)
}

type mockDeviceFinder struct {
	mock.Mock
}

func (df *mockDeviceFinder) getList(startDate time.Time, endDate time.Time, offset int, limit int) ([]string, error) {
	args := df.Called(startDate, endDate, offset, limit)
	return args.Get(0).([]string), args.Error(1)
}

type mockMultiInsert struct {
	mock.Mock
}

func (c *mockMultiInsert) insert(records []db.Record) (int, error) {
	args := c.Called(records)
	return args.Int(0), args.Error(1)
}

type mockCloser struct {
	mock.Mock
}

func (d *mockCloser) close() error {
	args := d.Called()
	return args.Error(0)
}

type mockPing struct {
	mock.Mock
}

func (d *mockPing) ping() error {
	args := d.Called()
	return args.Error(0)
}
