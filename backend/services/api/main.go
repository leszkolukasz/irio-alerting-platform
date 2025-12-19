package main

import (
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"alerting-platform/api/controllers"
	"alerting-platform/api/db"
	"alerting-platform/api/middleware"
	"alerting-platform/api/rpc"
	"alerting-platform/common/config"
	alert_pb "alerting-platform/common/rpc"
)

func main() {
	if config.GetConfig().Env == config.PROD {
		gin.SetMode(gin.ReleaseMode)
	}

	dbConn := db.GetDBConnection()
	dbConn.AutoMigrate(&db.User{})

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		runRESTServer()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		runRPCServer()
	}()

	wg.Wait()
}

func runRESTServer() {
	router := gin.Default()

	router.Use(middleware.GetSecurityMiddleware())
	router.Use(middleware.GetCORSMiddleware())

	authMiddleware := middleware.GetJWTMiddleware()

	controllers.RegisterRoutes(router, authMiddleware)

	port := config.GetConfig().REST_APIPort

	log.Printf("Starting REST server on port %d", port)
	if err := http.ListenAndServe(":"+strconv.Itoa(port), router); err != nil {
		log.Fatal("Failed to start REST server: ", err)
	}
}

func runRPCServer() {
	port := config.GetConfig().RPCPort
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	alert_pb.RegisterAlertServiceServer(grpcServer, &rpc.AlertServiceServer{})

	reflection.Register(grpcServer)

	log.Printf("Starting gRPC server listening on port %d", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
