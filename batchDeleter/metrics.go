// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package batchDeleter

import (
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/provider"
	"github.com/xmidt-org/webpa-common/v2/xmetrics"
)

const (
	DeletingQueueDepth = "deleting_queue_depth"
)

func Metrics() []xmetrics.Metric {
	return []xmetrics.Metric{
		{
			Name: DeletingQueueDepth,
			Help: "The depth of the delete queue",
			Type: "gauge",
		},
	}
}

type Measures struct {
	DeletingQueue metrics.Gauge
}

// NewMeasures constructs a Measures given a go-kit metrics Provider
func NewMeasures(p provider.Provider) *Measures {
	return &Measures{
		DeletingQueue: p.NewGauge(DeletingQueueDepth),
	}
}
