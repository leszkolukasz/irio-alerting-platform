package rpc

import (
	"alerting-platform/common/rpc"
	"context"
	"log"

	"google.golang.org/protobuf/types/known/emptypb"
)

func GetAllServicesInfo(ctx context.Context) *rpc.ServicesInfoForIncident {
	rpcClient := GetIncidentManagerServiceClient()
	servicesInfo, err := rpcClient.GetAllServicesInfo(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("[FATAL] Failed to get services info from Incident Manager: %v", err)
	}
	return servicesInfo
}
