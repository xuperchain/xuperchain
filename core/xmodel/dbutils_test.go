package xmodel

import (
	"testing"

	xmodel_pb "github.com/xuperchain/xuperchain/core/xmodel/pb"
)

func TestEqual(t *testing.T) {
	testCases := map[string]struct {
		pd     []*xmodel_pb.PureData
		vpd    []*xmodel_pb.PureData
		expect bool
	}{
		"testEqual": {
			expect: true,
			pd: []*xmodel_pb.PureData{
				&xmodel_pb.PureData{
					Bucket: "bucket1",
					Key:    []byte("key1"),
					Value:  []byte("value1"),
				},
				&xmodel_pb.PureData{
					Bucket: "bucket2",
					Key:    []byte("key2"),
					Value:  []byte("value2"),
				},
			},
			vpd: []*xmodel_pb.PureData{
				&xmodel_pb.PureData{
					Bucket: "bucket1",
					Key:    []byte("key1"),
					Value:  []byte("value1"),
				},
				&xmodel_pb.PureData{
					Bucket: "bucket2",
					Key:    []byte("key2"),
					Value:  []byte("value2"),
				},
			},
		},
		"testNotEqual": {
			expect: false,
			pd: []*xmodel_pb.PureData{
				&xmodel_pb.PureData{
					Bucket: "bucket1",
					Key:    []byte("key1"),
					Value:  []byte("value1"),
				},
				&xmodel_pb.PureData{
					Bucket: "bucket2",
					Key:    []byte("key2"),
					Value:  []byte("value2"),
				},
			},
			vpd: []*xmodel_pb.PureData{
				&xmodel_pb.PureData{
					Bucket: "bucket1",
					Key:    []byte("key1"),
					Value:  []byte("value1"),
				},
				&xmodel_pb.PureData{
					Bucket: "bucket3",
					Key:    []byte("key2"),
					Value:  []byte("value2"),
				},
			},
		},
		"testNotEqual2": {
			expect: false,
			pd: []*xmodel_pb.PureData{
				&xmodel_pb.PureData{
					Bucket: "bucket1",
					Key:    []byte("key1"),
					Value:  []byte("value1"),
				},
				&xmodel_pb.PureData{
					Bucket: "bucket2",
					Key:    []byte("key2"),
					Value:  []byte("value2"),
				},
			},
			vpd: []*xmodel_pb.PureData{
				&xmodel_pb.PureData{
					Bucket: "bucket1",
					Key:    []byte("key1"),
					Value:  []byte("value1"),
				},
				&xmodel_pb.PureData{
					Bucket: "bucket2",
					Key:    []byte("key2"),
					Value:  []byte("value3"),
				},
			},
		},
	}

	for k, v := range testCases {
		res := Equal(v.pd, v.vpd)
		t.Log(res)
		if res != v.expect {
			t.Error(k, "error", "expect", v.expect, "actual", res)
		}
	}
}
