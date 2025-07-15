// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package cassandra

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetrics(t *testing.T) {
	assert := assert.New(t)

	m := Metrics()

	assert.NotNil(m)
}
