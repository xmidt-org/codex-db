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
	"github.com/stretchr/testify/mock"
	db "github.com/xmidt-org/codex-db"
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

func (df *mockDeviceFinder) getList(offset string, limit int) ([]string, error) {
	args := df.Called(offset, limit)
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
