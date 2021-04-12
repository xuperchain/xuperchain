package server

import (
	"context"
	"fmt"
	"time"

	"github.com/xuperchain/xuperchain/core/pb"

)


var FilterMap = make(map[string]*pb.EvmFilterBody)
type evmFilter struct{
	filterMap map[string]*pb.EvmFilterBody
}

func newFilterService() *evmFilter{
	m := map[string]*pb.EvmFilterBody{}
	return &evmFilter{
		filterMap:m,
	}
}

func (fs *evmFilter)NewFilter(ctx context.Context,filter *pb.EvmFilterBody) (*pb.EvmFilterResponse, error){
	filter.Time = time.Now().Unix()
	filterID := generateID()
	filter.FilterID = filterID
	FilterMap[filterID] = filter

	go filterRecycling(filterID)					// 启动gourutine,定期清除Filter，防止Filter过多，占用内存
	resp := &pb.EvmFilterResponse{
		FilterID:filterID,
		Status:"SUCCESS",
	}
	return resp,nil
}

func (fs *evmFilter)UninstallFilter(ctx context.Context,filter *pb.EvmFilterBody) (*pb.EvmFilterResponse, error){
	filterID := filter.FilterID
	if _,ok := FilterMap[filterID];ok{
		delete(FilterMap,filterID)
	}
	resp := &pb.EvmFilterResponse{}
	resp.Status = "delete SUCCESS"
	return resp,nil
}

func (fs evmFilter)GetFilter(ctx context.Context, in *pb.EvmFilterBody) (*pb.EvmFilterBody, error){
	id := in.FilterID

	if filter,ok := FilterMap[id];ok {
		filter.Time = time.Now().Unix()
		FilterMap[id] = filter					// 如果有，则更新filter的时间戳
		filterBody := &pb.EvmFilterBody{}
		*filterBody = *filter
		return filterBody,nil
	}else{
		return nil,fmt.Errorf("filter not found")
	}
}

func filterRecycling(filterID string){
	deadDuration := int64(deadline.Seconds())
	ticker := time.NewTicker(deadline)
	for {
		<- ticker.C
		filter := &pb.EvmFilterBody{}
		var ok bool
		if filter,ok = FilterMap[filterID];!ok{
			return
		}
		timeNow := time.Now().Unix()
		lastUpdateTime := filter.Time
		if (timeNow - lastUpdateTime )> deadDuration {
			delete(FilterMap, filterID)
		}
	}
}