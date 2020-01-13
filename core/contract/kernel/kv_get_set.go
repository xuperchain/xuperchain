package kernel

import (
	"fmt"

	"github.com/xuperchain/xuperunion/contract"
)

// GetMethod define Get type
type GetMethod struct {
}

// SetMethod define Set type
type SetMethod struct {
}

// Invoke Get method implementation
func (gm *GetMethod) Invoke(ctx *KContext, args map[string][]byte) (*contract.Response, error) {
	if args["Bucket"] == nil || args["Key"] == nil {
		return nil, fmt.Errorf("invoke Get failed, args are invalid: %v", args)
	}
	bucket := string(args["Bucket"])
	key := args["Key"]
	getResult, err := ctx.ModelCache.Get(bucket, key)
	if err != nil {
		return nil, err
	}
	return &contract.Response{
		Status: contract.StatusOK,
		Body:   getResult.PureData.Value,
	}, nil
}

// Invoke Set method implementation
func (sm *SetMethod) Invoke(ctx *KContext, args map[string][]byte) (*contract.Response, error) {
	if args["Bucket"] == nil || args["Key"] == nil || args["Value"] == nil {
		return nil, fmt.Errorf("invoke Set failed, args are invalid: %v", args)
	}
	bucket := string(args["Bucket"])
	key := args["Key"]
	value := args["Value"]
	_, err := ctx.ModelCache.Get(bucket, key)
	err = ctx.ModelCache.Put(bucket, key, value)
	if err != nil {
		return nil, err
	}
	return &contract.Response{
		Status: contract.StatusOK,
		Body:   value,
	}, nil
}
