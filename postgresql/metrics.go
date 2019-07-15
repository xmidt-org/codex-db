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

package postgresql

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/xmidt-org/codex-db"
	"github.com/xmidt-org/webpa-common/xmetrics"
)

const (
	RetryCounter                = "retry_count"
	PoolOpenConnectionsGauge    = "pool_open_connections"
	PoolInUseConnectionsGauge   = "pool_in_use_connections"
	PoolIdleConnectionsGauge    = "pool_idle_connections"
	SQLWaitCounter              = "sql_wait_count"
	SQLWaitDurationCounter      = "sql_wait_duration_seconds"
	SQLMaxIdleClosedCounter     = "sql_max_idle_closed"
	SQLMaxLifetimeClosedCounter = "sql_max_lifetime_closed"
	SQLQuerySuccessCounter      = "sql_query_success_count"
	SQLQueryFailureCounter      = "sql_query_failure_count"
	SQLInsertedRecordsCounter   = "sql_inserted_rows_count"
	SQLReadRecordsCounter       = "sql_read_rows_count"
	SQLDeletedRecordsCounter    = "sql_deleted_rows_count"
)

//Metrics returns the Metrics relevant to this package
func Metrics() []xmetrics.Metric {
	return []xmetrics.Metric{
		// TODO: Fix Retry Counter
		{
			Name: RetryCounter,
			Type: "counter",
			Help: "Indicates the number of retries for sql queries",
		},
		{
			Name: PoolOpenConnectionsGauge,
			Type: "gauge",
			Help: "The number of established connections both in use and idle",
		},
		{
			Name: PoolInUseConnectionsGauge,
			Type: "gauge",
			Help: " The number of connections currently in use",
		},
		{
			Name: PoolIdleConnectionsGauge,
			Type: "gauge",
			Help: "The number of idle connections",
		},
		{
			Name: SQLWaitCounter,
			Type: "counter",
			Help: "The total number of connections waited for",
		},
		{
			Name: SQLWaitDurationCounter,
			Type: "counter",
			Help: "The total time blocked waiting for a new connection (nano)",
		},
		{
			Name: SQLMaxIdleClosedCounter,
			Type: "counter",
			Help: "The total number of connections closed due to SetMaxIdleConns",
		},
		{
			Name: SQLMaxLifetimeClosedCounter,
			Type: "counter",
			Help: "The total number of connections closed due to SetConnMaxLifetime",
		},
		{
			Name:       SQLQuerySuccessCounter,
			Type:       "counter",
			Help:       "The total number of successful SQL queries",
			LabelNames: []string{db.TypeLabel},
		},
		{
			Name:       SQLQueryFailureCounter,
			Type:       "counter",
			Help:       "The total number of failed SQL queries",
			LabelNames: []string{db.TypeLabel},
		},
		{
			Name: SQLInsertedRecordsCounter,
			Type: "counter",
			Help: "The total number of rows inserted",
		},
		{
			Name: SQLReadRecordsCounter,
			Type: "counter",
			Help: "The total number of rows read",
		},
		{
			Name: SQLDeletedRecordsCounter,
			Type: "counter",
			Help: "The total number of rows deleted",
		},
	}
}

type Measures struct {
	Retry                xmetrics.Incrementer
	PoolOpenConnections  metrics.Gauge
	PoolInUseConnections metrics.Gauge
	PoolIdleConnections  metrics.Gauge

	SQLWaitCount         metrics.Counter
	SQLWaitDuration      metrics.Counter
	SQLMaxIdleClosed     metrics.Counter
	SQLMaxLifetimeClosed metrics.Counter
	SQLQuerySuccessCount metrics.Counter
	SQLQueryFailureCount metrics.Counter
	SQLInsertedRecords   metrics.Counter
	SQLReadRecords       metrics.Counter
	SQLDeletedRecords    metrics.Counter
}

func NewMeasures(p provider.Provider) Measures {
	return Measures{
		Retry:                xmetrics.NewIncrementer(p.NewCounter(RetryCounter)),
		PoolOpenConnections:  p.NewGauge(PoolOpenConnectionsGauge),
		PoolInUseConnections: p.NewGauge(PoolInUseConnectionsGauge),
		PoolIdleConnections:  p.NewGauge(PoolIdleConnectionsGauge),

		SQLWaitCount:         p.NewCounter(SQLWaitCounter),
		SQLWaitDuration:      p.NewCounter(SQLWaitDurationCounter),
		SQLMaxIdleClosed:     p.NewCounter(SQLMaxIdleClosedCounter),
		SQLMaxLifetimeClosed: p.NewCounter(SQLMaxLifetimeClosedCounter),
		SQLQuerySuccessCount: p.NewCounter(SQLQuerySuccessCounter),
		SQLQueryFailureCount: p.NewCounter(SQLQueryFailureCounter),
		SQLInsertedRecords:   p.NewCounter(SQLInsertedRecordsCounter),
		SQLReadRecords:       p.NewCounter(SQLReadRecordsCounter),
		SQLDeletedRecords:    p.NewCounter(SQLDeletedRecordsCounter),
	}
}
