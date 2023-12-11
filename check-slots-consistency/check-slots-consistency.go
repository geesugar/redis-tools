package check_slots_consistency

import (
	"fmt"
	"log"

	rh "github.com/geesugar/redis-tools/pkg/redis-helper"
	"github.com/spf13/cobra"
)

var (
	addr string
)

func NewCheckSlotsConsistencyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "check-slots-consistency",
		Run: Run,
	}

	cmd.Flags().StringVarP(&addr, "addr", "", "", "redis addr")

	return cmd
}

func Run(cmd *cobra.Command, args []string) {
	ctx := cmd.Context()

	nodes, err := rh.GetClusterNodes(ctx, addr)
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
		nodes, err := rh.GetClusterNodes(ctx, node.Addr)
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
