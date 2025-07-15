// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package healthlogger

import (
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

func TestImplementsInterface(t *testing.T) {
	logger := log.NewNopLogger()
	hlogger := NewHealthLogger(logger)
	assert := assert.New(t)
	assert.NotNil(hlogger)
}
