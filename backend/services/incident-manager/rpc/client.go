package rpc

import (
	"alerting-platform/common/config"
	"alerting-platform/common/rpc"
	"log"
	"strconv"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	rpcClient *grpc.ClientConn
	once      sync.Once
)

func GetClient() *grpc.ClientConn {
	once.Do(func() {
		var url string
		cfg := config.GetConfig()

		if cfg.Env == config.PROD {
			url = "alerting-platform-api:" + strconv.Itoa(cfg.RPCPort)
		} else {
			url = "localhost:" + strconv.Itoa(cfg.RPCPort)
		}

		conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))

		if err != nil {
			log.Fatalf("Failed to connect to Incident Manager RPC server: %v", err)
		}
		rpcClient = conn
	})

	return rpcClient
}

func GetIncidentManagerServiceClient() rpc.IncidentManagerServiceClient {
	conn := GetClient()
	return rpc.NewIncidentManagerServiceClient(conn)
}
