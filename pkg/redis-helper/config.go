package rh

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"
)

const redisConfig = `
save {{ .Save  }}
appendfilename "appendonly.aof"
protected-mode {{ .ProtectedMode }}
cluster-require-full-coverage {{ .ClusterRequireFullCoverage }}
maxmemory-policy {{ .MaxMemoryPolicy }}
bind 0.0.0.0
maxmemory {{ .MaxMemory }}
client-output-buffer-limit slave {{ .ClientOutputBufferLimit }} {{ .ClientOutputBufferLimit }} 0
client-output-buffer-limit normal {{ .ClientOutputBufferLimit }} {{ .ClientOutputBufferLimit }} 0

lazyfree-lazy-eviction {{ .LazyFreeLazyEviction  }}
lazyfree-lazy-expire {{ .LazyFreeLazyExpire }}
lazyfree-lazy-server-del {{ .LazyFreeLazyServerDel }}
replica-lazy-flush {{ .ReplicaLazyFlush }}

cluster-migration-barrier {{ .ClusterMigrationBarrier }}

repl-diskless-sync {{ .ReplDiskLessSync }}

repl-timeout {{ .ReplTimeout }}

repl-backlog-size {{ .ReplBacklogSize }}
repl-backlog-ttl {{ .ReplBacklogTTL }}
slowlog-log-slower-than {{ .SlowlogLogSlowerThan }}

{{- if .Persistent }}
appendonly yes
aof-use-rdb-preamble yes
{{- end }}

{{- if .MasterUser }}
masteruser {{ .MasterUser }}
masterauth {{ .MasterAuth }}
{{- end }}

{{- if .IsElastic }}
cluster-enabled yes
cluster-allow-replica-migration no
cluster-replica-validity-factor 40
{{- end }}

maxclients {{ .MaxClients }}
timeout {{ .Timeout }}
activedefrag {{ .Activedefrag }}

port {{ .Port }}
tcp-backlog {{ .TcpBacklog }}
cluster-node-timeout {{ .ClusterNodeTimeout }}

{{- if .IsElastic }}
cluster-config-file {{.DBPath}}/nodes.conf
{{- end }}

{{- range $user := .ACLUsers }}
{{$user}}
{{- end }}
loglevel debug
`

type Config struct {
	// Save default ""
	Save string `redisconfigkey:"save" default:""`
	// ProtectedMode default no
	ProtectedMode string `redisconfigkey:"protected-mode" default:"no"`
	// ClusterRequireFullCoverage default no
	ClusterRequireFullCoverage string `redisconfigkey:"cluster-require-full-coverage" default:"no"`
	// MaxMemoryPolicy default allkeys-lru
	MaxMemoryPolicy string `redisconfigkey:"maxmemory-policy" default:"allkeys-lru"`
	// MaxMemory
	MaxMemory int64 `redisconfigkey:"maxmemory" default:""`
	// ClientOutputBufferLimit default 4294967296
	ClientOutputBufferLimit uint64 `redisconfigkey:"client-output-buffer-limit" default:"4294967296"`
	// LazyFreeLazyEviction  default yes
	LazyFreeLazyEviction string `redisconfigkey:"lazyfree-lazy-eviction" default:"yes"`
	// LazyFreeLazyExpire default yes
	LazyFreeLazyExpire string `redisconfigkey:"lazyfree-lazy-expire" default:"yes"`
	// LazyFreeLazyServerDel default yes
	LazyFreeLazyServerDel string `redisconfigkey:"lazyfree-lazy-server-del" default:"yes"`
	// ReplicaLazyFlush default yes
	ReplicaLazyFlush string `redisconfigkey:"replica-lazy-flush" default:"yes"`
	// ClusterMigrationBarrier default 999
	ClusterMigrationBarrier string `redisconfigkey:"cluster-migration-barrier" default:"999"`
	// ReplDiskLessSync default yes
	ReplDiskLessSync string `redisconfigkey:"repl-diskless-sync" default:"yes"`
	// ReplTimeout default 300
	ReplTimeout string `redisconfigkey:"repl-timeout" default:"300"`
	// ReplBacklogSize default 256000000
	ReplBacklogSize uint64 `redisconfigkey:"repl-backlog-size" default:"256000000"`
	// ReplBacklogTTL default 86400
	ReplBacklogTTL uint64 `redisconfigkey:"repl-backlog-ttl" default:"86400"`
	// SlowlogLogSlowerThan default 100000
	SlowlogLogSlowerThan uint64 `redisconfigkey:"slowlog-log-slower-than" default:"100000"`
	// Persistent default false, if true, appendonly will be yes and aof-use-rdb-preamble will be yes
	Persistent bool `default:"false"`
	// MasterUser
	MasterUser string `redisconfigkey:"masteruser" default:""`
	// MasterAuth
	MasterAuth string `redisconfigkey:"masterauth" default:""`
	// IsElastic default true, if true, cluster-enabled will be yes and cluster-replica-validity-factor will be 40
	IsElastic bool `default:"true"`
	// MaxClients default 10000
	MaxClients uint64 `redisconfigkey:"maxclients" default:"10000"`
	// Timeout default 0
	Timeout uint64 `redisconfigkey:"timeout" default:"0"`
	// Activedefrag default no
	Activedefrag string `redisconfigkey:"activedefrag" default:"no"`
	// Port
	Port int `redisconfigkey:"port" default:"6379"`
	// TcpBacklog default 1024
	TcpBacklog string `redisconfigkey:"tcp-backlog" default:"1024"`
	// ClusterNodeTimeout default 15000
	ClusterNodeTimeout uint64 `redisconfigkey:"cluster-node-timeout" default:"15000"`

	ClusterAllowReplicaMigration string `redisconfigkey:"cluster-allow-replica-migration" default:"no"`
	// DBPath is config for redis to set config key: cluster-config-file
	DBPath string `default:""`
	// ACLUsers
	ACLUsers []string
}

func (c *Config) Content() ([]byte, error) {
	redisConfigTmpl, err := template.New("redisConfig").Parse(redisConfig)
	if err != nil {
		//log.Fatalf("failed to parse redis config template: %v", err)
		return nil, fmt.Errorf("failed to parse redis config template: %v", err)
	}

	config := bytes.Buffer{}
	if err := redisConfigTmpl.Execute(&config, c); err != nil {
		return nil, fmt.Errorf("failed to execute redis config template: %v", err)
	}
	return config.Bytes(), nil
}

func GenProxyConfig(items map[string]string, port int) []string {
	result := make([]string, 0)
	result = append(result, "/usr/bin/redis-proxy")

	for k, v := range items {
		result = append(result, fmt.Sprintf("--%s", k))
		result = append(result, v)
	}
	result = append(result, "--port")
	result = append(result, strconv.Itoa(port))
	return result
}

func GenRedisConfig(items map[string]string, port int, memPerRedis int) string {
	config := make(map[string]string)
	for k, v := range items {
		config[k] = v
	}
	config["port"] = strconv.Itoa(port)
	config["maxmemory"] = strconv.Itoa(memPerRedis * 1024 * 1024) // TODO overflow
	config["loglevel"] = "debug"

	var str string
	for k, v := range config {
		str += fmt.Sprintf("%s %s\n", k, v)
	}
	return str
}

func NewConfig(maxMemory int64, port int, isElastic bool) *Config {
	c := &Config{
		Save:                         "",
		ProtectedMode:                "no",
		ClusterRequireFullCoverage:   "no",
		MaxMemoryPolicy:              "allkeys-lru",
		MaxMemory:                    maxMemory,
		ClientOutputBufferLimit:      4294967296,
		LazyFreeLazyEviction:         "yes",
		LazyFreeLazyExpire:           "yes",
		LazyFreeLazyServerDel:        "yes",
		ReplicaLazyFlush:             "yes",
		ClusterMigrationBarrier:      "999",
		ReplDiskLessSync:             "yes",
		ReplTimeout:                  "300",
		ReplBacklogSize:              256000000,
		ReplBacklogTTL:               86400,
		SlowlogLogSlowerThan:         100000,
		Persistent:                   false,
		MasterUser:                   "",
		MasterAuth:                   "",
		IsElastic:                    isElastic,
		MaxClients:                   50000,
		Timeout:                      0,
		Activedefrag:                 "no",
		Port:                         port,
		TcpBacklog:                   "1024",
		ClusterNodeTimeout:           15000,
		DBPath:                       "/data",
		ClusterAllowReplicaMigration: "no",
		ACLUsers:                     nil,
	}

	return c
}
