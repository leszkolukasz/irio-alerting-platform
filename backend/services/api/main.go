package main

import (
	"context"
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
	"alerting-platform/common/db/firestore"
	pubsub_common "alerting-platform/common/pubsub"
	pb "alerting-platform/common/rpc"
)

func main() {
	config.Intro("API")

	if config.GetConfig().Env == config.PROD {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()

	dbConn := db.GetDBConnection()
	dbConn.AutoMigrate(&db.User{}, &db.MonitoredService{})

	psClient := pubsub_common.Init(ctx)
	defer psClient.Close()

	firestoreRepo := firestore.GetLogRepository(ctx)
	defer firestoreRepo.Close()

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
	pb.RegisterIncidentManagerServiceServer(grpcServer, rpc.NewIncidentManagerServiceServer(
		db.NewRepository(db.GetDBConnection()),
	))

	pb.RegisterSchedulerServiceServer(grpcServer, &rpc.SchedulerServiceServer{})
	reflection.Register(grpcServer)

	log.Printf("Starting gRPC server listening on port %d", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
