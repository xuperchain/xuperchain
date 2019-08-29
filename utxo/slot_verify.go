package utxo

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/xuperchain/xuperunion/common"
	aclu "github.com/xuperchain/xuperunion/permission/acl/utils"
)

// step1: 判断待释放的槽位是否被占用
// step2: 判断交易发起者是否等于被释放的槽位对应的地址
func (uv *UtxoVM) verifyReleaseSlot(slotId, contendAddr, initiator string, bucket string) (bool, error) {
	// step1: 判断待释放的槽位是否被占用
	// 未被占用，则释放无效
	// TODO, @ToWorld 是否需要考虑unconfirmed情况?
	versionData, versionErr := uv.model3.Get(aclu.GetSlot2AddressBucket(), []byte(slotId))
	if common.NormalizedKVError(versionErr) == common.ErrKVNotFound {
		// 之前是空的，那么不允许释放
		return false, errors.New("can't release an empty slot")
	} else if versionErr != nil {
		return false, versionErr
	}
	// step2: 判断交易发起者是否等于被释放的槽位对应的地址
	// 是否有可能,preAddr和initiator都为空?
	preAddr := string(versionData.GetPureData().GetValue())
	if preAddr != initiator {
		return false, fmt.Errorf("couldn't release other's address-> %v->%v", preAddr, initiator)
	}

	return true, nil
}

// step1: 判断当前竞争者的balance是否超过1M
// step2: 判断当前槽位是否已经被占用
// step3: 判断当前槽位被占用的address与当前竞争者的xpower
func (uv *UtxoVM) verifyContendSlot(slotId, contendAddr string) (bool, error) {
	// step1: 判断当前竞争者的balance是否超过1M
	// 查询竞争者的当前确定余额是否超过100万
	// TODO, @ToWorld 根据XPower方式计算余额
	utxoLeftTotal, err := uv.GetBalance(contendAddr)
	if err != nil {
		return false, err
	}
	if utxoLeftTotal.Cmp(big.NewInt(1000000)) < 0 {
		return false, errors.New("contender's utxo is less than 1000000")
	}

	// step2: 判断当前槽位是否已经被占用
	// 如果已经被占用, 需要进一步比较xpower
	// 如果没有被占用, 不需要进一步比较xpower
	versionData, versionErr := uv.model3.Get(aclu.GetSlot2AddressBucket(), []byte(slotId))
	if common.NormalizedKVError(versionErr) == common.ErrKVNotFound {
		// 之前是空的
		return true, nil
	} else if versionErr != nil {
		// 其他错误，不能接受
		return false, versionErr
	}
	preAddr := string(versionData.GetPureData().GetValue())
	// 修改数据的人就是占用这个槽位的人
	if contendAddr == preAddr {
		return false, errors.New("self contend self")
	}

	/*
		// 如果当前的slot未被占用，不需要再计算XPower
		if preAddr == "" {
			return true, nil
		}
	*/

	// step3: 判断当前槽位被占用的address与当前竞争者的xpower
	// 比较两者的XPower
	contendAddrXPower, contendErr := uv.CalcXPower(contendAddr, uv.ledger.GetMeta().TrunkHeight)
	if contendErr != nil {
		return false, contendErr
	}
	preAddrXPower, preAddrErr := uv.CalcXPower(preAddr, uv.ledger.GetMeta().TrunkHeight)
	if preAddrErr != nil {
		return false, preAddrErr
	}
	if contendAddrXPower.Cmp(preAddrXPower) <= 0 {
		return false, errors.New("xpower is not enough")
	}

	return true, nil
}
