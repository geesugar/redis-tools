package rh

import (
	"context"
	"fmt"
	"strings"

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

	Host string
	Port uint64
	Addr string
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

func (c *Client) ClusterSetSlot(ctx context.Context, slot int, subCmd string, nodeID string) (err error) {
	if strings.ToLower(subCmd) == "stable" {
		cmd := c.Do(ctx, "cluster", "setslot", slot, subCmd)
		return cmd.Err()
	}

	cmd := c.Do(ctx, "cluster", "setslot", slot, subCmd, nodeID)
	return cmd.Err()
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

func ParsePort(port string) (uint64, error) {
	// string 2 uint64
	var p uint64
	for _, c := range port {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid port")
		}
		p = p*10 + uint64(c-'0')
	}
	return p, nil
}

func ParseAddr(addr string) (host string, port uint64, err error) {
	s := strings.Split(addr, ":")
	if len(s) != 2 {
		return "", 0, fmt.Errorf("invalid addr")
	}

	port, err = ParsePort(s[1])
	if err != nil {
		return "", 0, err
	}

	return s[0], port, nil
}

func NewClient(ctx context.Context, addr, usr, passwd string) (cli *Client, err error) {
	// parse addr
	host, port, err := ParseAddr(addr)
	if err != nil {
		return nil, err
	}

	uc, err := NewUniversalClient(ctx, addr, usr, passwd)
	if err != nil {
		return nil, err
	}

	return &Client{
		UniversalClient: uc,
		Host:            host,
		Port:            port,
		Addr:            addr,
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
