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
	"github.com/stretchr/testify/mock"
	"github.com/xmidt-org/codex-db"
	"github.com/xmidt-org/codex-db/blacklist"
)

type mockInserter struct {
	mock.Mock
}

func (i *mockInserter) InsertRecords(records ...db.Record) error {
	args := i.Called(records)
	return args.Error(0)
}

type mockPruner struct {
	mock.Mock
}

func (p *mockPruner) GetRecordsToDelete(shard int, limit int, deathDate int64) ([]db.RecordToDelete, error) {
	args := p.Called(shard, limit, deathDate)
	return args.Get(0).([]db.RecordToDelete), args.Error(1)
}

func (p *mockPruner) DeleteRecord(shard int, deathdate int64, recordID int64) error {
	args := p.Called(shard, deathdate, recordID)
	return args.Error(0)
}

type mockRG struct {
	mock.Mock
}

func (rg *mockRG) GetRecords(deviceID string, limit int) ([]db.Record, error) {
	args := rg.Called(deviceID, limit)
	return args.Get(0).([]db.Record), args.Error(1)
}

func (rg *mockRG) GetRecordsOfType(deviceID string, limit int, eventType db.EventType) ([]db.Record, error) {
	args := rg.Called(deviceID, limit, eventType)
	return args.Get(0).([]db.Record), args.Error(1)
}

type mockLG struct {
	mock.Mock
}

func (rg *mockLG) GetBlacklist() ([]blacklist.BlackListedItem, error) {
	args := rg.Called()
	return args.Get(0).([]blacklist.BlackListedItem), args.Error(1)
}
