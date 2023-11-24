package crdprom

import (
	"github.com/geesugar/redis-tools/pkg/prom"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

type (
	Resource string
	Op       string
)

const (
	ResourceCluster        Resource = "cluster"
	ResourceProxy          Resource = "proxy"
	ResourceShard          Resource = "shard"
	ResourceShardReplica   Resource = "shardreplica"
	ResourceSingular       Resource = "singular"
	ResourceProxyPod       Resource = "proxy_pod"
	ResourceRedisPod       Resource = "redis_pod"
	ResourceDNS            Resource = "dns"
	ResourceProxyConfigmap Resource = "proxy_configmap"
	ResourceRedisConfigmap Resource = "redis_configmap"
	ResourceEntrypoint     Resource = "entrypoint_configmap"

	OpCreate         Op = "create"
	OpDelete         Op = "delete"
	OpUpdate         Op = "update"
	OpUpgrade        Op = "upgrade"
	OpScaleOut       Op = "scale_out"
	OpScaleIn        Op = "scale_in"
	OpScaleUP        Op = "scale_up"
	OpScaleDown      Op = "scale_down"
	OpInplaceUpgrade Op = "inplace_upgrade"
	OpRollingUpgrade Op = "rolling_upgrade"
)

var defaultMetrics = NewCRDCollector()

type CRDMetrics struct {
	opCounter *prometheus.CounterVec
}

func NewCRDCollector(opts ...prom.Option) *CRDMetrics {
	options := prom.DefaultOptions()
	options.Merge(opts...)

	cacheCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: options.Namespace,
			Subsystem: options.Subsystem,
			Name:      "resource_op_total",
			Help:      "resource operation total",
		},
		[]string{"cluster_id", "cluster_name", "resource", "op", "status"},
	)

	_ = metrics.Registry.Register(cacheCount)

	return &CRDMetrics{
		opCounter: cacheCount,
	}
}

func (cm *CRDMetrics) Inc(clusterID, cluster string, resource Resource, op Op, status string) {
	cm.opCounter.With(prometheus.Labels{
		"cluster_id":   clusterID,
		"cluster_name": cluster,
		"resource":     string(resource),
		"op":           string(op),
		"status":       status,
	}).Inc()
}

func Inc(clusterID, clusterName string, resource Resource, op Op, status string) {
	defaultMetrics.Inc(clusterID, clusterName, resource, op, status)
}
