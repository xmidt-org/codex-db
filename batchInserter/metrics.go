// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package batchInserter

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/xmidt-org/webpa-common/v2/xmetrics"
)

const (
	InsertingQueueDepth            = "inserting_queue_depth"
	DroppedEventsFromDbFailCounter = "dropped_events_db_fail_count"
)

func Metrics() []xmetrics.Metric {
	return []xmetrics.Metric{
		{
			Name: InsertingQueueDepth,
			Help: "The depth of the insert queue",
			Type: "gauge",
		},
		{
			Name: DroppedEventsFromDbFailCounter,
			Help: "The total number of events dropped from the database query failing",
			Type: "counter",
		},
	}
}

type Measures struct {
	InsertingQueue               metrics.Gauge
	DroppedEventsFromDbFailCount metrics.Counter
}

// NewMeasures constructs a Measures given a go-kit metrics Provider
func NewMeasures(p provider.Provider) *Measures {
	return &Measures{
		InsertingQueue:               p.NewGauge(InsertingQueueDepth),
		DroppedEventsFromDbFailCount: p.NewCounter(DroppedEventsFromDbFailCounter),
	}
}
