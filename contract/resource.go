package contract

import "github.com/xuperchain/xuperunion/pb"

const (
	maxResourceLimit = 0xFFFFFFFF
)

// Limits describes the usage or limit of resources
type Limits struct {
	Cpu    int64
	Memory int64
	Disk   int64
	Gas    int64
}

// TotalGas converts resource to gas
func (l *Limits) TotalGas() int64 {
	// FIXME:
	return l.Cpu/1000 + l.Memory/1000000 + l.Disk + l.Gas
}

// MaxLimits describes the maximum limit of resources
var MaxLimits = Limits{
	Cpu:    maxResourceLimit,
	Memory: maxResourceLimit,
	Disk:   maxResourceLimit,
	Gas:    maxResourceLimit,
}

// FromPbLimits converts []*pb.ResourceLimit to Limits
func FromPbLimits(rlimits []*pb.ResourceLimit) Limits {
	limits := Limits{}
	for _, l := range rlimits {
		switch l.GetType() {
		case pb.ResourceType_CPU:
			limits.Cpu = l.GetLimit()
		case pb.ResourceType_MEMORY:
			limits.Memory = l.GetLimit()
		case pb.ResourceType_DISK:
			limits.Disk = l.GetLimit()
		case pb.ResourceType_GAS:
			limits.Gas = l.GetLimit()
		}
	}
	return limits
}

// FromPbLimits converts Limits to []*pb.ResourceLimit
func ToPbLimits(limits Limits) []*pb.ResourceLimit {
	return []*pb.ResourceLimit{
		{Type: pb.ResourceType_CPU, Limit: limits.Cpu},
		{Type: pb.ResourceType_MEMORY, Limit: limits.Memory},
		{Type: pb.ResourceType_DISK, Limit: limits.Disk},
		{Type: pb.ResourceType_GAS, Limit: limits.Gas},
	}
}
