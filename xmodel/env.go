package xmodel

import (
	"errors"
	"fmt"
	"github.com/xuperchain/xuperunion/pb"
	xmodel_pb "github.com/xuperchain/xuperunion/xmodel/pb"
)

// Env data structure for read/write sets environment
type Env struct {
	outputs    []*xmodel_pb.PureData
	modelCache *XMCache
}

// PrepareEnv get read/write sets with transaction
func (s *XModel) PrepareEnv(tx *pb.Transaction) (*Env, error) {
	inputs := []*xmodel_pb.VersionedData{}
	outputs := []*xmodel_pb.PureData{}
	env := &Env{}
	s.logger.Trace("PrepareEnv", "tx.TxInputsExt", tx.TxInputsExt, "tx.TxOutputsExt", tx.TxOutputsExt)
	for _, txIn := range tx.TxInputsExt {
		verData, err := s.GetUncommited(txIn.Bucket, txIn.Key)
		if err != nil {
			return nil, err
		}
		s.logger.Trace("prepareEnv", "verData", verData, "txIn", txIn)
		if GetVersion(verData) != txIn.GetVersion() {
			err := fmt.Errorf("prepareEnv fail, key:%s, inputs version is not valid: %s != %s", string(verData.PureData.Key), GetVersion(verData), txIn.GetVersion())
			return nil, err
		}
		inputs = append(inputs, verData)
	}
	for _, txOut := range tx.TxOutputsExt {
		outputs = append(outputs, &xmodel_pb.PureData{Bucket: txOut.Bucket, Key: txOut.Key, Value: txOut.Value})
	}
	utxoInputs, utxoOutputs, err := ParseContractUtxo(tx)
	if err != nil {
		return nil, err
	}
	if ok := IsContractUtxoEffective(utxoInputs, utxoOutputs, tx); !ok {
		s.logger.Warn("PrepareEnv CheckConUtxoEffective error")
		return nil, errors.New("PrepareEnv CheckConUtxoEffective error")
	}
	env.modelCache = NewXModelCacheWithInputs(inputs, utxoInputs)
	env.outputs = outputs
	s.logger.Trace("PrepareEnv done!", "env", env)
	return env, nil
}

// GetModelCache get instance of model cache
func (e *Env) GetModelCache() *XMCache {
	if e != nil {
		return e.modelCache
	}
	return nil
}

// GetOutputs get outputs
func (e *Env) GetOutputs() []*xmodel_pb.PureData {
	if e != nil {
		return e.outputs
	}
	return nil
}
