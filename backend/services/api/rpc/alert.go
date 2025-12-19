package rpc

import (
	alert_pb "alerting-platform/common/rpc"
	"context"
)

type AlertServiceServer struct {
	alert_pb.UnimplementedAlertServiceServer
}

func (s *AlertServiceServer) GetAlertData(ctx context.Context, req *alert_pb.AlertRequest) (*alert_pb.AlertData, error) {
	return &alert_pb.AlertData{
		AlertName: "Sample Alert for ID: " + req.AlertId,
	}, nil
}
