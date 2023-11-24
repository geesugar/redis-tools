package rh

import (
	"bytes"
	"fmt"
	"text/template"
)

const proxyConfig = `
# [proxy requirepass $pass] Proxy password
requirepass {{ .RequirePass }}
 
# [pauth setuser $user $pass] Redis user & pass, same with mt-proxy
redis-user  {{ .RedisUser }}
redis-pass  {{ .RedisPass }}
 
 
# [proxy protect $status] Enable protect: yes/no, same with mt-proxy
overload-protection  {{ .OverloadProtection }}
 
# [proxy protect $worker,$redis] 
# overload-protection-worker-qps: The max qps number for each worker
# overload-protection-redis-qps: The number of protect pqs for each worker, same with mt-proxy
overload-protection-worker-qps  {{ .OverloadProtectionWorkerQPS }}
overload-protection-redis-qps  {{ .OverloadProtectionRedisQPS }}
 
# [proxy conn_token $conns] Max number of each proxy worker connections
proxy-conn-token {{ .ProxyConnToken }}

maxclients {{ .MaxClients }}

# [proxy redis_conn $conns] The number of max connection between proxy and redis, same with mt-proxy
proxy-redis-conn {{ .ProxyRedisConn }}
 
 
# [proxy worker $workers] The number of worker process, like with mt-proxy:worker-threads
proxy-worker-numbers {{ .ProxyWorkerNumbers }}
 
# [proxy slavemode $mode] Enable slave read: off/master_writeonly/master_readwrite, same with mt-proxy
proxy-slavemode {{ .ProxySlaveMode }}
 
 
# [proxy hotkey qps $qps] The number of buckets for calculate hot keys
proxy-hotkey-qps  {{ .ProxyHotkeyQPS }}
 
# [proxy hotkey max $max] The number of buckets for calculate hot keys
proxy-hotkey-maxkey {{ .ProxyHotkeyMaxKey }}
 
 
# [proxy bootstrap $ip $port] Bootstrap ip:port, same with mt-proxy
bootstrap-addr {{ .BootstrapAddr }}
 
# [PBAN ADDNODES $addr ...] pban proxy nodes
pban-nodes  {{ .PBanNodes }}
 
# [config set slowlog-log-slower-than $time_us] show ther slower request>$time_us, same with mt-proxy
slowlog-log-slower-than {{ .SlowlogLogSlowerThan }}
# [config set slowlog-max-len $len] max number show logs, same with mt-proxy
 
slowlog-max-len {{ .SlowlogMaxLen }}
# [acl setuser ($user $pass)/(deluser $user ...)] Acl for users
user {{ .User }}
`

type ProxyConfig struct {
	RequirePass                 string
	RedisUser                   string
	RedisPass                   string
	OverloadProtection          string
	OverloadProtectionWorkerQPS int32
	OverloadProtectionRedisQPS  int32
	ProxyConnToken              int32
	MaxClients                  int32
	ProxyRedisConn              int32
	ProxyWorkerNumbers          int32
	ProxySlaveMode              string
	ProxyHotkeyQPS              int32
	ProxyHotkeyMaxKey           int32
	BootstrapAddr               string
	PBanNodes                   string
	SlowlogMaxLen               string
	User                        string
}

func (c *ProxyConfig) Content() ([]byte, error) {
	redisConfigTmpl, err := template.New("proxyConfig").Parse(proxyConfig)
	if err != nil {
		//log.Fatalf("failed to parse redis config template: %v", err)
		return nil, fmt.Errorf("failed to parse proxy config template: %v", err)
	}

	config := bytes.Buffer{}
	if err := redisConfigTmpl.Execute(&config, c); err != nil {
		return nil, fmt.Errorf("failed to execute proxy config template: %v", err)
	}

	return config.Bytes(), nil
}
