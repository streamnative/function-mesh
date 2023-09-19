// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// MetricDescription is an exported struct that defines the metric description (Name, Help)
// as a new type named MetricDescription.
type MetricDescription struct {
	Name string
	Help string
	Type string
}

// metricsDescription is a map of string keys (metrics) to MetricDescription values (Name, Help).
var metricDescription = map[string]MetricDescription{
	"FunctionMeshControllerReconcileCount": {
		Name: "function_mesh_reconcile_count",
		Help: "Number of reconcile operations",
		Type: "Counter",
	},
	"FunctionMeshControllerReconcileLatency": {
		Name: "function_mesh_reconcile_latency",
		Help: "Latency of reconcile operations, bucket boundaries are 10ms, 100ms, 1s, 10s, 30s and 60s.",
		Type: "Histogram",
	},
}

var (
	// FunctionMeshControllerReconcileCount will count how many reconcile operations been done.
	FunctionMeshControllerReconcileCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: metricDescription["FunctionMeshControllerReconcileCount"].Name,
			Help: metricDescription["FunctionMeshControllerReconcileCount"].Help,
		}, []string{"type", "name", "namespace"},
	)

	// FunctionMeshControllerReconcileLatency will measure the latency of reconcile operations.
	FunctionMeshControllerReconcileLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: metricDescription["FunctionMeshControllerReconcileLatency"].Name,
			Help: metricDescription["FunctionMeshControllerReconcileLatency"].Help,
			// Bucket boundaries are 10ms, 100ms, 1s, 10s, 30s and 60s.
			Buckets: []float64{
				10, 100, 1000, 10000, 30000, 60000,
			},
		}, []string{"type", "name", "namespace"},
	)
)

// RegisterMetrics will register metrics with the global prometheus registry
func RegisterMetrics() {
	metrics.Registry.MustRegister(collectors.NewBuildInfoCollector())
	metrics.Registry.MustRegister(FunctionMeshControllerReconcileCount)
	metrics.Registry.MustRegister(FunctionMeshControllerReconcileLatency)
}

// ListMetrics will create a slice with the metrics available in metricDescription
func ListMetrics() []MetricDescription {
	v := make([]MetricDescription, 0, len(metricDescription))
	// Insert value (Name, Help) for each metric
	for _, value := range metricDescription {
		v = append(v, value)
	}

	return v
}
