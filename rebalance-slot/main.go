// Description: rebalance slot
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	rh "github.com/geesugar/redis-tools/pkg/redis-helper"
)

const (
	MigrateBatchKeys     = 1000
	MigrateTimeoutSecond = 300
)

func PrintUsage() {
	fmt.Printf("Usage: %s ADDR NODEID SLOTS\n", os.Args[0])
}

func main() {
	if len(os.Args) < 4 {
		PrintUsage()
		os.Exit(1)
	}

	addr := os.Args[1]
	nodeID := os.Args[2]
	slots := os.Args[3]

	fmt.Printf("addr:%s node:%s slots:%s\n", addr, nodeID, slots)

	ctx := context.Background()

	nodes, err := rh.GetClusterNodes(ctx, addr)
	if err != nil {
		log.Fatalf("GetClusterNodes error: %s", err)
	}

	node := GetNodeByID(nodes, nodeID)
	if node == nil {
		log.Fatalf("node not found. node_id:%s", nodeID)
	}

	if !node.IsMaster() {
		log.Fatalf("node is not master. node_id:%s", nodeID)
	}

	specSlots := rh.NewSlots()
	err = specSlots.SetSlotSlice(slots)
	if err != nil {
		log.Fatalf("parse slots slice error: %s", err)
	}

	equl, diff := node.Slots.Compare(specSlots)
	if equl {
		fmt.Printf("slots is equal\n")
		return
	}

	fmt.Printf("slots is not equal. diff:%s\n", diff)

	// press enter to continue
	var input string
	fmt.Printf("press enter to continue...")
	fmt.Scanln(&input)

	masterCliMap := make(map[string]*rh.Client, len(nodes))
	for _, node := range nodes {
		if !node.IsMaster() {
			continue
		}

		cli, err := rh.NewClient(ctx, node.Addr, "", "")
		if err != nil {
			log.Fatalf("new client. addr:%s, err:%s", node.Addr, err)
		}
		defer cli.Close()

		masterCliMap[node.Addr] = cli
	}

	for slot, set := range specSlots {
		if !set {
			continue
		}

		if node.Slots[slot] {
			continue
		}

		// find a node which has this slot
		var srcNode *rh.ClusterNode
		for _, node := range nodes {
			if !node.IsMaster() {
				continue
			}

			if node.Slots[slot] {
				srcNode = node
				break
			}
		}

		keyCount, err := MigrateSlot(ctx, masterCliMap, srcNode.Addr, node.Addr, srcNode.ID, node.ID, slot)
		if err != nil {
			log.Fatalf("migrate slot error: %s", err)
		}

		fmt.Printf("migrate slot success. slot:%d, key_count:%d src_node_id:%s, dst_node_id:%s\n", slot, keyCount, srcNode.ID, node.ID)
	}
}

func GetNodeByID(nodes []*rh.ClusterNode, id string) *rh.ClusterNode {
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
	}
	return nil
}

func MigrateSlot(ctx context.Context, masterCliMap map[string]*rh.Client, srcAddr, dstAddr string, srcNodeID, dstNodeID string, slot int) (int, error) {
	srcCli, ok := masterCliMap[srcAddr]
	if !ok {
		return 0, fmt.Errorf("src addr not found. addr:%s", srcAddr)
	}

	dstCli, ok := masterCliMap[dstAddr]
	if !ok {
		return 0, fmt.Errorf("dst addr not found. addr:%s", dstAddr)
	}

	err := dstCli.ClusterSetSlot(ctx, slot, "IMPORTING", srcNodeID)
	if err != nil {
		return 0, fmt.Errorf("cluster set slot IMPORTING. addr:%s, slot:%d, src_node_id:%s, err:%s", dstAddr, slot, srcNodeID, err)
	}

	err = srcCli.ClusterSetSlot(ctx, slot, "MIGRATING", dstNodeID)
	if err != nil {
		return 0, fmt.Errorf("cluster set slot MIGRATING. addr:%s, slot:%d, dst_node_id:%s, err:%s", srcAddr, slot, dstNodeID, err)
	}

	// wait for slot migration
	keyCount, err := MigrateKeys(ctx, srcCli, dstCli, slot)
	if err != nil {
		return 0, fmt.Errorf("migrate keys. slot:%d, err:%s", slot, err)
	}

	err = dstCli.ClusterSetSlot(ctx, slot, "NODE", dstNodeID)
	if err != nil {
		return 0, fmt.Errorf("cluster set slot NODE. addr:%s, slot:%d, dst_node_id:%s, err:%s", dstAddr, slot, dstNodeID, err)
	}

	err = srcCli.ClusterSetSlot(ctx, slot, "NODE", dstNodeID)
	if err != nil {
		return 0, fmt.Errorf("cluster set slot NODE. addr:%s, slot:%d, dst_node_id:%s, err:%s", srcAddr, slot, dstNodeID, err)
	}

	for _, cli := range masterCliMap {
		if cli.Addr == srcAddr || cli.Addr == dstAddr {
			continue
		}

		err = cli.ClusterSetSlot(ctx, slot, "NODE", dstNodeID)
		if err != nil {
			fmt.Printf("cluster set slot NODE. addr:%s, slot:%d, dst_node_id:%s, err:%s\n", cli.Addr, slot, dstNodeID, err)
		}
	}

	return keyCount, nil
}

func MigrateKeys(ctx context.Context, srcCli, dstCli *rh.Client, slot int) (int, error) {
	keyCount := 0
	cmds := make([]interface{}, MigrateBatchKeys+7)

	for {
		cmd := srcCli.ClusterGetKeysInSlot(ctx, slot, MigrateBatchKeys)
		if cmd.Err() != nil {
			return 0, fmt.Errorf("cluster get keys in slot. slot:%d, batch_keys:%d, err:%s", slot, MigrateBatchKeys, cmd.Err())
		}

		if len(cmd.Val()) == 0 {
			break
		}

		cmds = cmds[:0]
		cmds = append(cmds, "MIGRATE")
		cmds = append(cmds, dstCli.Host)
		cmds = append(cmds, dstCli.Port)
		cmds = append(cmds, "")
		cmds = append(cmds, "0")
		cmds = append(cmds, MigrateTimeoutSecond)

		cmds = append(cmds, "KEYS")
		for _, k := range cmd.Val() {
			cmds = append(cmds, k)
		}

		migrateCmd := srcCli.Do(ctx, cmds...)
		if migrateCmd.Err() != nil {
			return 0, fmt.Errorf("migrate keys. slot:%d, batch_keys:%d, keyCount:%d curKeys:%d, err:%s", slot, MigrateBatchKeys, keyCount, len(cmd.Val()), migrateCmd.Err())
		}

		keyCount += len(cmd.Val())
	}

	return keyCount, nil
}
