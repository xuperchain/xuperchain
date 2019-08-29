package kernel

import (
	"errors"
	"fmt"

	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/permission/acl/utils"
)

const (
	contendSlotCost = 50000
)

// ContendSlotMethod ...
type ContendSlotMethod struct {
}

// ReleaseSlotMethod ...
type ReleaseSlotMethod struct {
}

// {
//    "module_name": "xkernel",
//    "method_name": "ContendSlot",
//    "args": {
//		  "slot": "1",
//		  "address": "ABCDEFG"
//	  }
// }
// Invoke ...
func (cs *ContendSlotMethod) Invoke(ctx *KContext, args map[string][]byte) (*contract.Response, error) {
	// TODO, @ToWorld 这部分费用应该销毁
	if ctx.ResourceLimit.XFee < contendSlotCost {
		return nil, fmt.Errorf("gas not enough, expect no less than %d", contendSlotCost)
	}
	slotId := args["slot"]
	address := args["address"]
	// check if param is valid
	if slotId == nil || address == nil {
		return nil, errors.New("invoke ContendSlot error, slot or address is nil")
	}

	/*
		// check if one slot has been occupied by the address
		// 不能被占用:
		versionData, err := ctx.ModelCache.Get(utils.GetAddress2SlotBucket(), address)
		if  err != xmodel.ErrNotFound {
			return nil, fmt.Errorf("invoke ContendSlot error->`%v`", versionData)
		}*/
	// key: slotId, value: address
	err := ctx.ModelCache.Put(utils.GetSlot2AddressBucket(), slotId, address)
	if err != nil {
		return nil, err
	}
	// key: address value: slotId
	err = ctx.ModelCache.Put(utils.GetAddress2SlotBucket(), address, slotId)
	if err != nil {
		return nil, err
	}
	ctx.AddXFeeUsed(contendSlotCost)

	return &contract.Response{
		Status: contract.StatusOK,
	}, nil
}

// {
//	  "module_name": "xkernel",
//	  "method_name": "ReleaseSlot",
//	  "args": {
//		  "slot": "1",
//	  }
// }
// Invoke ...
func (rs *ReleaseSlotMethod) Invoke(ctx *KContext, args map[string][]byte) (*contract.Response, error) {
	slotId := args["slot"]
	// check if the slotId to be released is valid
	if slotId == nil {
		return nil, errors.New("invoke ReleaseSlot error, slot is nil")
	}
	// check if the slotId to be released has been occupied already
	versionData, err := ctx.ModelCache.Get(utils.GetSlot2AddressBucket(), slotId)
	if err != nil {
		return nil, err
	}
	address := versionData.GetPureData().GetValue()
	// if the slotId to be released has never been occupied already, return error
	if address == nil || string(address) == "" {
		return nil, errors.New("invoke ReleaseSlot error, slot has already been released or never been occupied")
	}
	// empty the slotId -> address relations
	err = ctx.ModelCache.Put(utils.GetSlot2AddressBucket(), slotId, []byte("None"))
	if err != nil {
		return nil, err
	}
	// empty the address -> slotId relations
	err = ctx.ModelCache.Put(utils.GetAddress2SlotBucket(), address, []byte("None"))
	if err != nil {
		return nil, err
	}

	return &contract.Response{
		Status: contract.StatusOK,
	}, nil
}
