package rpc

import (
	"alerting-platform/api/db"
	"alerting-platform/common/rpc"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type IncidentManagerServiceServer struct {
	rpc.UnimplementedIncidentManagerServiceServer
}

func (s *IncidentManagerServiceServer) GetAllServicesInfo(ctx context.Context, empty *emptypb.Empty) (*rpc.ServicesInfoForIncident, error) {
	conn := db.GetDBConnection()

	var services []db.MonitoredService
	result := conn.Find(&services)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, result.Error
	}

	rpcServices := make([]*rpc.ServiceInfoForIncident, 0, len(services))
	for _, service := range services {
		oncallers := []string{service.FirstOncallerEmail}
		if service.SecondOncallerEmail != nil {
			oncallers = append(oncallers, *service.SecondOncallerEmail)
		}

		rpcService := &rpc.ServiceInfoForIncident{
			ServiceId:           uint64(service.ID),
			AlertWindow:         int64(service.AlertWindow),
			AllowedResponseTime: int64(service.AllowedResponseTime),
			Oncallers:           oncallers,
		}
		rpcServices = append(rpcServices, rpcService)
	}

	return &rpc.ServicesInfoForIncident{
		Services: rpcServices,
	}, nil
}
