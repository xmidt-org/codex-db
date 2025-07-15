// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package batchDeleter

import (
	"github.com/stretchr/testify/mock"
	"github.com/xmidt-org/codex-db"
)

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
