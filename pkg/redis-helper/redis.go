package rh

import (
	"context"

	"github.com/geesugar/redis-tools/pkg/prom/redisprom"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// DefaultNamespace is the default metrics namespace.
	DefaultNamespace = "cache"

	// DefaultSubsystem is the default metrics subsystem.
	DefaultSubsystem = "redis"
)

type Client struct {
	redis.UniversalClient
}

func (c *Client) GetClusterNodes(ctx context.Context) (nodes []*ClusterNode, err error) {
	dstResult, err := c.ClusterNodes(ctx).Result()
	if err != nil {
		return
	}
	nodes, err = ParseClusterNodes(dstResult)
	if err != nil {
		return
	}
	return nodes, nil
}

func (c *Client) GetClusterInfo(ctx context.Context) (info *ClusterInfo, err error) {
	dstResult, err := c.ClusterInfo(ctx).Result()
	if err != nil {
		return
	}
	info, err = ParseClusterInfo(dstResult)
	if err != nil {
		return
	}
	return info, nil
}

func NewClient(ctx context.Context, addr, usr, passwd string) (cli *Client, err error) {
	uc, err := NewUniversalClient(ctx, addr, usr, passwd)
	if err != nil {
		return nil, err
	}

	return &Client{
		UniversalClient: uc,
	}, nil
}

func NewUniversalClient(ctx context.Context, addr, usr, pass string) (cli redis.UniversalClient, err error) {
	cli = redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: usr,
		Password: pass,
	})

	// register conn pool metrics collector
	collector := redisprom.NewCollector(DefaultNamespace, DefaultSubsystem, cli)
	prometheus.Register(collector)

	cli.AddHook(redisprom.Hook())

	err = cli.Ping(ctx).Err()

	if err != nil {
		_ = cli.Close()
	}

	return cli, err
}
