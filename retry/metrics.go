// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package dbretry

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/xmidt-org/codex-db"
	"github.com/xmidt-org/webpa-common/v2/xmetrics"
)

const (
	SQLQueryRetryCounter = "sql_query_retry_count"
	SQLQueryEndCounter   = "sql_query_end_counter"
)

// Metrics returns the Metrics relevant to this package
func Metrics() []xmetrics.Metric {
	return []xmetrics.Metric{
		{
			Name:       SQLQueryRetryCounter,
			Type:       "counter",
			Help:       "The total number of SQL queries retried",
			LabelNames: []string{db.TypeLabel},
		},
		{
			Name:       SQLQueryEndCounter,
			Type:       "counter",
			Help:       "the total number of SQL queries that are done, no more retrying",
			LabelNames: []string{db.TypeLabel},
		},
	}
}

type Measures struct {
	SQLQueryRetryCount metrics.Counter
	SQLQueryEndCount   metrics.Counter
}

func NewMeasures(p provider.Provider) Measures {
	return Measures{
		SQLQueryRetryCount: p.NewCounter(SQLQueryRetryCounter),
		SQLQueryEndCount:   p.NewCounter(SQLQueryEndCounter),
	}
}
