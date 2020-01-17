package utxo

import (
	"github.com/xuperchain/xuperchain/core/pb"
	"github.com/xuperchain/xuperchain/core/permission/acl/utils"
)

// queryContractStatData query stat data about contract, such as total contract and total account
func (uv *UtxoVM) queryContractStatData(bucket string) (int64, error) {
	dataCount := int64(0)
	prefixKey := pb.ExtUtxoTablePrefix + bucket + "/"
	it := uv.ldb.NewIteratorWithPrefix([]byte(prefixKey))
	defer it.Release()

	for it.Next() {
		dataCount++
	}
	if it.Error() != nil {
		return int64(0), it.Error()
	}

	return dataCount, nil
}

func (uv *UtxoVM) QueryContractStatData() (*pb.ContractStatData, error) {

	accountCount, accountCountErr := uv.queryContractStatData(utils.GetAccountBucket())
	if accountCountErr != nil {
		return &pb.ContractStatData{}, accountCountErr
	}

	contractCount, contractCountErr := uv.queryContractStatData(utils.GetContract2AccountBucket())
	if contractCountErr != nil {
		return &pb.ContractStatData{}, contractCountErr
	}

	data := &pb.ContractStatData{
		AccountCount:  accountCount,
		ContractCount: contractCount,
	}

	return data, nil
}
