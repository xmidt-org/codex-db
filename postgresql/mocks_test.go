// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package postgresql

import (
	"encoding/json"

	"github.com/stretchr/testify/mock"
	"github.com/xmidt-org/codex-db"
)

type mockFinder struct {
	mock.Mock
}

func (f *mockFinder) findRecords(out *[]db.Record, limit int, where ...interface{}) error {
	args := f.Called(out, limit, where)
	err := json.Unmarshal(args.Get(1).([]byte), out)
	if err != nil {
		return err
	}
	return args.Error(0)
}

func (f *mockFinder) findRecordsToDelete(limit int, shard int, deathDate int64) ([]db.RecordToDelete, error) {
	args := f.Called(limit, shard, deathDate)
	return args.Get(0).([]db.RecordToDelete), args.Error(1)
}

type mockDeviceFinder struct {
	mock.Mock
}

func (df *mockDeviceFinder) getList(offset string, limit int, where ...interface{}) ([]string, error) {
	args := df.Called(offset, limit, where)
	return args.Get(0).([]string), args.Error(1)
}

type mockMultiInsert struct {
	mock.Mock
}

func (c *mockMultiInsert) insert(records []db.Record) (int64, error) {
	args := c.Called(records)
	return int64(args.Int(0)), args.Error(1)
}

type mockDeleter struct {
	mock.Mock
}

func (d *mockDeleter) delete(value *db.Record, limit int, where ...interface{}) (int64, error) {
	args := d.Called(value, limit, where)
	return int64(args.Int(0)), args.Error(1)
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
