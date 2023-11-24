// http://www.apache.org/licenses/LICENSE-2

// 描述：通过对redis cluster各个节点调用cluster nodes命令，获取各个节点的slot分配情况，判断是否有节点slot分配不均匀的情况
// 入参： redis cluster的ip和端口
// 出参： 无

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	rh "github.com/geesugar/redis-tools/pkg/redis-helper"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s ip:port", os.Args[0])
	}

	ctx := context.Background()

	nodes, err := GetClusterNodes(ctx, os.Args[1])
	if err != nil {
		log.Fatalf("GetClusterNodes error: %s", err)
	}

	masterNodes := make(map[string]*rh.ClusterNode, len(nodes))
	for _, node := range nodes {
		if node.IsMaster() {
			masterNodes[node.ID] = node
		}
	}

	if len(masterNodes) == 0 {
		log.Fatalf("no master node")
	}

	var originNodes []*rh.ClusterNode
	var origiNodeAddr string
	for _, node := range masterNodes {
		nodes, err := GetClusterNodes(ctx, node.Addr)
		if err != nil {
			log.Fatalf("get cluster nodes of node:%s node_id:%s, err: %s", node.Addr, node.ID, err)
		}

		if originNodes == nil {
			originNodes = nodes
			origiNodeAddr = node.Addr
			continue
		}

		fmt.Printf("compare node_id:%s slots. cur:%s origin:%s\n", node.ID, node.Addr, origiNodeAddr)
		if err := CompareNodesSlots(nodes, originNodes); err != nil {
			fmt.Printf("compare nodes slots error: %s\n", err)
		}
	}
}

func GetClusterNodes(ctx context.Context, addr string) ([]*rh.ClusterNode, error) {
	cli, err := rh.NewClient(ctx, addr, "", "")
	if err != nil {
		return nil, fmt.Errorf("new client. addr:%s, err:%s", addr, err)
	}
	defer cli.Close()

	nodes, err := cli.GetClusterNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("get cluster nodes. addr:%s, err:%s", addr, err)
	}

	return nodes, nil
}

func GetNodeByID(id string, nodes []*rh.ClusterNode) *rh.ClusterNode {
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
	}

	return nil
}

func CompareNodesSlots(nodes []*rh.ClusterNode, comparedNodes []*rh.ClusterNode) error {
	if len(nodes) != len(comparedNodes) {
		return fmt.Errorf("nodes count not equal. nodes:%d, comparedNodes:%d", len(nodes), len(comparedNodes))
	}

	for _, node := range nodes {
		if !node.IsMaster() {
			continue
		}

		comparedNode := GetNodeByID(node.ID, comparedNodes)
		if comparedNode == nil {
			return fmt.Errorf("compared node not found. node_id:%s", node.ID)
		}

		equal, str := node.Slots.Compare(comparedNode.Slots)
		if !equal {
			fmt.Printf("node_id:%s, slots not equal. %s\n", node.ID, str)
		} else {
			fmt.Printf("node_id:%s, slots equal\n", node.ID)
		}
	}

	return nil
}
