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
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
	db "github.com/xmidt-org/codex-db"
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

	SQLDurationSeconds        = "sql_duration_seconds"
	SQLQuerySuccessCounter    = "sql_query_success_count"
	SQLQueryFailureCounter    = "sql_query_failure_count"
	SQLInsertedRecordsCounter = "sql_inserted_rows_count"
	SQLReadRecordsCounter     = "sql_read_rows_count"
	SQLDeletedRecordsCounter  = "sql_deleted_rows_count"
)

//Metrics returns the Metrics relevant to this package
func Metrics() []xmetrics.Metric {
	return []xmetrics.Metric{
		{
			Name: PoolInUseConnectionsGauge,
			Type: "gauge",
			Help: " The number of connections currently in use",
		},
		{
			Name:    SQLDurationSeconds,
			Type:    "histogram",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{0.0625, 0.125, .25, .5, 1, 5, 10, 20, 40, 80, 160},
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
	PoolInUseConnections metrics.Gauge
	SQLDuration          metrics.Histogram
	SQLQuerySuccessCount metrics.Counter
	SQLQueryFailureCount metrics.Counter
	SQLInsertedRecords   metrics.Counter
	SQLReadRecords       metrics.Counter
	SQLDeletedRecords    metrics.Counter
}

func NewMeasures(p provider.Provider) Measures {
	return Measures{
		PoolInUseConnections: p.NewGauge(PoolInUseConnectionsGauge),
		SQLDuration:          p.NewHistogram(SQLDurationSeconds, 11),
		SQLQuerySuccessCount: p.NewCounter(SQLQuerySuccessCounter),
		SQLQueryFailureCount: p.NewCounter(SQLQueryFailureCounter),
		SQLInsertedRecords:   p.NewCounter(SQLInsertedRecordsCounter),
		SQLReadRecords:       p.NewCounter(SQLReadRecordsCounter),
		SQLDeletedRecords:    p.NewCounter(SQLDeletedRecordsCounter),
	}
}
