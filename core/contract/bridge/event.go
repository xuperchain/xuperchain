package bridge

import (
	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/pb"
)

func eventsResourceUsed(events []*pb.ContractEvent) contract.Limits {
	var size int64
	for _, event := range events {
		size += int64(len(event.Contract) + len(event.Name) + len(event.Body))
	}
	return contract.Limits{
		Disk: size,
	}
}
