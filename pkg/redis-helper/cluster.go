package rh

import (
	"fmt"
	"strconv"
	"strings"
)

type ClusterNode struct {
	ID        string // this is node ID
	Addr      string
	Role      Role
	State     State
	Myself    bool
	SlotsStr  string
	Slots     Slots
	MasterID  string
	Connected bool
	Epoch     int64
}

type Role int

const (
	RoleNone Role = iota
	RoleMaster
	RoleSlave
)

type State int

const (
	StateNormal State = 0
	StatePFail  State = 1 << iota
	StateFail
	StateNoAddr
	StateHandshake
)

func (p *ClusterNode) IsNoAddr() bool { return p.State&StateNoAddr > 0 }
func (p *ClusterNode) IsMaster() bool { return p.Role == RoleMaster }
func (p *ClusterNode) IsSlave() bool  { return p.Role == RoleSlave }
func (p *ClusterNode) IsOK() bool {
	return p.State == StateNormal && (p.Role == RoleSlave || p.Role == RoleMaster)
}

// IsHealthy returns whether the cluster is healthy
func (p *ClusterNode) IsHealthy() bool { return p.State == StateNormal }

func (p *ClusterNode) CheckEqual(other *ClusterNode) error {
	if other == nil {
		return fmt.Errorf("invalid cluster node")
	}

	if p.ID != other.ID {
		return fmt.Errorf("ID not the same(%v, %v)", p.ID, other.ID)
	}
	if p.Addr != other.Addr {
		return fmt.Errorf("addr not the same(%v, %v)", p.Addr, other.Addr)
	}
	if p.Role != other.Role {
		return fmt.Errorf("role not the same(%v, %v)", p.Role, other.Role)
	}
	if p.State != other.State {
		return fmt.Errorf("state not the same(%v, %v)", p.State, other.State)
	}
	if p.MasterID != other.MasterID {
		return fmt.Errorf("MasterID not the same(%v, %v)", p.MasterID, other.MasterID)
	}
	if p.Connected != other.Connected {
		return fmt.Errorf("connected not the same(%v, %v)", p.Connected, other.Connected)
	}
	// if it was slave, the config epoch came from its master
	if p.IsMaster() && p.Epoch != other.Epoch {
		return fmt.Errorf("epoch not the same(%v, %v)", p.Epoch, other.Epoch)
	}
	if p.SlotsStr != other.SlotsStr {
		return fmt.Errorf("SlotSlice not the same(%v, %v)", p.SlotsStr, other.SlotsStr)
	}

	return nil
}

const (
	ClusterSlots = 16384
)

func ParseClusterNodes(s string) (nodes []*ClusterNode, err error) {
	l := strings.Split(s, "\n")
	nodes = make([]*ClusterNode, 0, len(l))

	for _, line := range l {
		if len(line) == 0 {
			continue
		}

		rs := strings.SplitN(line, " ", 9)
		// slave didnt contain slot info
		// nolint
		if len(rs) < 8 {
			continue
		}

		addrs := strings.SplitN(rs[1], "@", 2)
		// nolint
		if len(addrs) != 2 {
			continue
		}

		stats := strings.Split(rs[2], ",")

		role := RoleNone
		myselfFlag := false
		state := StateNormal
		connected := false

		for _, stat := range stats {
			switch stat {
			case "master":
				role = RoleMaster
			case "slave":
				role = RoleSlave

			case "myself":
				myselfFlag = true
			case "fail?":
				state |= StatePFail
			case "fail":
				state |= StateFail
			case "noaddr":
				state |= StateNoAddr
			case "handshake":
				state |= StateHandshake
			}
		}

		epoch, err := strconv.ParseInt(rs[6], 10, 64)
		if err != nil {
			return nil, err
		}

		if rs[7] == "connected" {
			connected = true
		}

		slotsStr := ""
		// nolint
		if len(rs) >= 9 {
			l := strings.SplitN(rs[8], "[", 2)
			slotsStr = strings.Trim(l[0], " ")
		}

		slots := NewSlots()
		for _, slotSlice := range strings.Split(slotsStr, " ") {
			err = slots.SetSlotSlice(slotSlice)
			if err != nil {
				return nil, err
			}
		}

		nodes = append(nodes, &ClusterNode{
			ID:        rs[0],
			Addr:      addrs[0],
			Role:      role,
			Myself:    myselfFlag,
			State:     state,
			MasterID:  rs[3],
			Connected: connected,
			SlotsStr:  slotsStr,
			Slots:     slots,
			Epoch:     epoch,
		})
	}

	// to have a final check for the all slots was parsed right
	//if totalSlotsNum != ClusterSlots {
	//	return nodes, false, fmt.Errorf("slots is not ok")
	//}

	return nodes, nil
}

func ParseClusterInfo(s string) (info *ClusterInfo, err error) {
	l := strings.Split(s, "\n")

	info = &ClusterInfo{}
	for _, line := range l {
		if len(line) == 0 {
			continue
		}
		kv := strings.Split(line, ":")
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "cluster_my_epoch":
			val, err := strconv.Atoi(kv[1])
			if err != nil {
				return nil, fmt.Errorf("invalid cluster_my_epoch")
			}
			info.MyEpoch = val
		}
	}

	return info, nil
}

func ExtractMyself(nodes []*ClusterNode) *ClusterNode {
	for _, node := range nodes {
		if node.Myself {
			return node
		}
	}
	return nil
}

type ClusterInfo struct {
	ID      string // this is node ID
	MyEpoch int
}
