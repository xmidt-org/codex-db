// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package dbretry

import (
	"github.com/stretchr/testify/mock"
	db "github.com/xmidt-org/codex-db"
)

type mockInserter struct {
	mock.Mock
}

func (i *mockInserter) InsertRecords(records ...db.Record) error {
	args := i.Called(records)
	return args.Error(0)
}
