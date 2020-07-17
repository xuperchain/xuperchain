package xmodel

import (
	"fmt"

	"github.com/xuperchain/xuperchain/core/pb"
	xmodel_pb "github.com/xuperchain/xuperchain/core/xmodel/pb"
)

func parseVersion(version string) ([]byte, int, error) {
	txid := []byte{}
	offset := 0
	okNum, err := fmt.Sscanf(version, "%x_%d", &txid, &offset)
	if okNum != 2 && err != nil {
		return nil, 0, fmt.Errorf("parseVersion failed, invalid version: %s", version)
	}
	return txid, offset, nil
}

//GetTxidFromVersion parse version and fetch txid from version string
func GetTxidFromVersion(version string) []byte {
	txid, _, err := parseVersion(version)
	if err != nil {
		return []byte("")
	}
	return txid
}

// MakeVersion generate a version by txid and offset, version = txid_offset
func MakeVersion(txid []byte, offset int32) string {
	return fmt.Sprintf("%x_%d", txid, offset)
}

// GetVersion get VersionedData's version, if refTxid is nil, return ""
func GetVersion(vd *xmodel_pb.VersionedData) string {
	if vd.RefTxid == nil {
		return ""
	}
	return MakeVersion(vd.RefTxid, vd.RefOffset)
}

// GetVersionOfTxInput get version of TxInput
func GetVersionOfTxInput(txIn *pb.TxInputExt) string {
	if txIn.RefTxid == nil {
		return ""
	}
	return MakeVersion(txIn.RefTxid, txIn.RefOffset)
}

// GetTxOutputs get transaction outputs
func GetTxOutputs(pds []*xmodel_pb.PureData) []*pb.TxOutputExt {
	outputs := make([]*pb.TxOutputExt, 0, len(pds))
	for _, pd := range pds {
		outputs = append(outputs, &pb.TxOutputExt{
			Bucket: pd.Bucket,
			Key:    pd.Key,
			Value:  pd.Value,
		})
	}
	return outputs
}

// GetTxInputs get transaction inputs
func GetTxInputs(vds []*xmodel_pb.VersionedData) []*pb.TxInputExt {
	inputs := make([]*pb.TxInputExt, 0, len(vds))
	for _, vd := range vds {
		inputs = append(inputs, &pb.TxInputExt{
			Bucket:    vd.GetPureData().GetBucket(),
			Key:       vd.GetPureData().GetKey(),
			RefTxid:   vd.RefTxid,
			RefOffset: vd.RefOffset,
		})
	}
	return inputs
}

// IsEmptyVersionedData check if VersionedData is empty
func IsEmptyVersionedData(vd *xmodel_pb.VersionedData) bool {
	return vd.RefTxid == nil && vd.RefOffset == 0
}

func makeEmptyVersionedData(bucket string, key []byte) *xmodel_pb.VersionedData {
	verData := &xmodel_pb.VersionedData{PureData: &xmodel_pb.PureData{}}
	verData.PureData.Bucket = bucket
	verData.PureData.Key = key
	return verData
}
