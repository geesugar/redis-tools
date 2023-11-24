package redisprom

import (
	"context"
	"reflect"
	"time"

	"github.com/geesugar/redis-tools/pkg/prom"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var defaultHook = newHook()

var labelNames = []string{"command", "error"}

// NewHook creates a new go-redis hook instance and registers Prometheus collectors.
func newHook(opts ...prom.Option) *redisHook {
	options := prom.DefaultOptions()
	options.Merge(opts...)

	cmds := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: options.Namespace,
		Subsystem: options.Subsystem,
		Name:      "commands_duration_seconds",
		Help:      "Histogram of Redis commands",
		Buckets:   options.DurationBuckets,
	}, labelNames)

	metrics.Registry.Register(cmds)

	return &redisHook{
		options:  options,
		commands: cmds,
	}
}

type (
	// redisHook represents a go-redis hook that exports metrics of commands and pipelines.
	redisHook struct {
		options  *prom.Options
		commands *prometheus.HistogramVec
	}

	startKey struct{}
)

func Hook() redis.Hook {
	return defaultHook
}

func (hook *redisHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, startKey{}, time.Now()), nil
}

func (hook *redisHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	duration := float64(0)
	if start, ok := ctx.Value(startKey{}).(time.Time); ok {
		duration = time.Since(start).Seconds()
	}

	errstr := ""
	if cmd.Err() != nil {
		errstr = reflect.TypeOf(cmd.Err()).String()
	}

	hook.commands.WithLabelValues(cmd.Name(), errstr).Observe(duration)

	return nil
}

func (hook *redisHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, startKey{}, time.Now()), nil
}

func (hook *redisHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	duration := float64(0)
	if start, ok := ctx.Value(startKey{}).(time.Time); ok {
		duration = time.Since(start).Seconds()
	}

	hook.commands.WithLabelValues("pipline", "").Observe(duration)

	for _, cmd := range cmds {
		cmderrstr := ""
		if cmd.Err() != nil {
			cmderrstr = reflect.TypeOf(cmd.Err()).String()
		}

		hook.commands.WithLabelValues(cmd.Name(), cmderrstr).Observe(duration)
	}

	return nil
}

func register(collector prometheus.Collector) prometheus.Collector {
	err := prometheus.Register(collector)
	if err == nil {
		return collector
	}

	if arErr, ok := err.(prometheus.AlreadyRegisteredError); ok {
		return arErr.ExistingCollector
	}

	panic(err)
}
