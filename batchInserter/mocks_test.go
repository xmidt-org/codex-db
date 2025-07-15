// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package batchInserter

import (
	"time"

	"github.com/stretchr/testify/mock"
	db "github.com/xmidt-org/codex-db"
)

type mockInserter struct {
	mock.Mock
}

func (c *mockInserter) InsertRecords(records ...db.Record) error {
	args := c.Called(records)
	return args.Error(0)
}

type mockTracker struct {
	mock.Mock
}

func (t *mockTracker) TrackTime(d time.Duration) {
	t.Called(d)
}
