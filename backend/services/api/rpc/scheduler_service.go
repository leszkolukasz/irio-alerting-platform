package rpc

import (
	"alerting-platform/api/db"
	"alerting-platform/common/rpc"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type SchedulerServiceServer struct {
	rpc.UnimplementedSchedulerServiceServer
}

// service SchedulerService {
//   rpc GetAllSchedulerConfigurations (google.protobuf.Empty) returns (SchedulerConfigResponse);
// }

func (s *SchedulerServiceServer) GetAllSchedulerConfigurations(ctx context.Context, empty *emptypb.Empty) (*rpc.SchedulerConfigResponse, error) {
	conn := db.GetDBConnection()

	var services []db.MonitoredService
	result := conn.Find(&services)

	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}
	// rpcServices := make([]*rpc.ServiceInfoForIncident, 0, len(services))
	// schedulerConfigResponse := make(rpc.SchedulerConfigResponse)
	rpcServices := make([]*rpc.ServiceInfoForScheduler, 0, len(services))
	for _, service := range services {
		rpcService := &rpc.ServiceInfoForScheduler{
			ServiceId:           uint64(service.ID),
			Url:                 service.URL,
			HealthCheckInterval: int64(service.HealthCheckInterval),
		}
		rpcServices = append(rpcServices, rpcService)
	}

	return &rpc.SchedulerConfigResponse{
		Services: rpcServices,
	}, nil
}
