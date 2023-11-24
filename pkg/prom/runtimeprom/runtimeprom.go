package runtimeprom

import (
	"reflect"

	"github.com/geesugar/redis-tools/pkg/prom"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// var defaultMetrics = NewRuntimeCollector()
type Resource string

const (
	ResourceCluster Resource = "cluster"
)

var defaultMetrics *RuntimeMetrics

type RuntimeMetrics struct {
	ConsistencyCheckGauge     *prometheus.GaugeVec
	BatchMigrationSlotsGauge  *prometheus.GaugeVec
	ConsistencyCheckHistogram *prometheus.HistogramVec
	SlotsIsBalanceGauge       *prometheus.GaugeVec
	OpenSlotsGauge            *prometheus.GaugeVec
	ResourceStatusGauge       *prometheus.GaugeVec
}

func init() {
	defaultMetrics = NewRuntimeCollector()
	_ = metrics.Registry.Register(defaultMetrics.BatchMigrationSlotsGauge)
	_ = metrics.Registry.Register(defaultMetrics.ConsistencyCheckHistogram)
	_ = metrics.Registry.Register(defaultMetrics.ConsistencyCheckGauge)
	_ = metrics.Registry.Register(defaultMetrics.SlotsIsBalanceGauge)
	_ = metrics.Registry.Register(defaultMetrics.OpenSlotsGauge)
	_ = metrics.Registry.Register(defaultMetrics.ResourceStatusGauge)
	//defaultMetrics.registerAll()
}

func (cm RuntimeMetrics) registerAll() {
	v := reflect.ValueOf(cm)
	monitors := make([]prometheus.Collector, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		monitors[i] = v.Field(i).Interface().(prometheus.Collector)
	}
	prometheus.MustRegister(monitors...)
}

func NewRuntimeCollector(opts ...prom.Option) *RuntimeMetrics {
	options := prom.DefaultOptions()
	options.Merge(opts...)

	consistencyCheckGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: options.Namespace,
			Subsystem: options.Subsystem,
			Name:      "redis_consistency_check_duration_gauge_seconds",
			Help:      "before do rebalance, first need to check consistency(gauge)",
		}, []string{"cluster_id", "cluster_name"},
	)

	consistencyCheckHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "redis_consistency_check_duration_histogram_seconds",
		Help:      "before do rebalance, first need to check consistency(histogram)",
		Buckets:   []float64{.001, .005, .01, .5, 1, 5, 30, 180},
	}, []string{"cluster_id", "cluster_name"},
	)

	batchMigrationSlotsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: options.Namespace,
			Subsystem: options.Subsystem,
			Name:      "batch_migration_slots_seconds",
			Help:      "batch migration slots for rebalance",
		}, []string{"cluster_id", "cluster_name"},
	)

	slotsIsBalanceGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: options.Namespace,
			Subsystem: options.Subsystem,
			Name:      "slots_is_balance",
			Help:      "the cluster slots whether balance",
		}, []string{"cluster_id", "cluster_name"},
	)

	openSlots := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: options.Namespace,
			Subsystem: options.Subsystem,
			Name:      "open_slots",
			Help:      "the cluster open slots",
		}, []string{"cluster_id", "cluster_name", "open_slots"},
	)

	resourceStatusGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: options.Namespace,
			Subsystem: options.Subsystem,
			Name:      "resource_status",
			Help:      "resource status",
		}, []string{"cluster_id", "cluster_name", "resource"},
	)

	return &RuntimeMetrics{
		ConsistencyCheckGauge:     consistencyCheckGauge,
		ConsistencyCheckHistogram: consistencyCheckHistogram,
		BatchMigrationSlotsGauge:  batchMigrationSlotsGauge,
		SlotsIsBalanceGauge:       slotsIsBalanceGauge,
		OpenSlotsGauge:            openSlots,
		ResourceStatusGauge:       resourceStatusGauge,
	}
}

func SetConsistencyCheckMetrics(clusterID, clusterName string, duration float64) {
	defaultMetrics.ConsistencyCheckGauge.WithLabelValues(clusterID, clusterName).Set(duration)
	defaultMetrics.ConsistencyCheckHistogram.WithLabelValues(clusterID, clusterName).Observe(duration)
}

func SetBatchMigrationSlotsMetrics(clusterID, clusterName string, duration float64) {
	defaultMetrics.BatchMigrationSlotsGauge.WithLabelValues(clusterID, clusterName).Set(duration)
}

func SetSlotsIsBalanceMetrics(clusterID, clusterName string, isBalance int) {
	defaultMetrics.SlotsIsBalanceGauge.WithLabelValues(clusterID, clusterName).Set(float64(isBalance))
}

func SetOpenSlotsMetrics(clusterID, clusterName string, openSlots string) {
	value := 0
	if openSlots != "" {
		value = 1
	}

	defaultMetrics.OpenSlotsGauge.WithLabelValues(clusterID, clusterName, openSlots).Set(float64(value))
}

func SetResourceGaugeMetrics(clusterID, clusterName string, resource Resource, value int64) {
	defaultMetrics.ResourceStatusGauge.WithLabelValues(clusterID, clusterName, string(resource)).Set(float64(value))
}
