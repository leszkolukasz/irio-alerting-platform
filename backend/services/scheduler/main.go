package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"alerting-platform/common/config"
	"alerting-platform/common/rpc"

	"cloud.google.com/go/pubsub"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MonitoringTask struct {
	ServiceId uint64 `json:"service_id,omitempty"`
	URL       string `json:"url,omitempty"`
}

type Task struct {
	ID     uint64
	Cancel context.CancelFunc
}

type scheduler struct {
	activeTasks   map[uint64]*Task
	mu            sync.Mutex
	client        rpc.SchedulerServiceClient
	pubsubClient  *pubsub.Client
	incidentTopic *pubsub.Topic // the topic we write to
}

func (s *scheduler) addTask(id uint64, t *Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeTasks[id] = t
}

func (s *scheduler) scheduleMonitor(ctx context.Context, service *rpc.ServiceInfoForScheduler) {
	goRoutineCtx, cancel := context.WithCancel(ctx)
	serviceId := service.ServiceId
	healthCheckInterval := service.HealthCheckInterval
	url := service.Url
	task := &Task{
		ID:     serviceId,
		Cancel: cancel,
	}

	s.addTask(service.ServiceId, task)
	go func(ctx context.Context, interval int64, serviceId uint64, url string) {
		d := time.Duration(interval) * time.Second
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		for {
			select {
			case <-goRoutineCtx.Done():
				return
			case <-ticker.C:
				// here the message is sent to the broker
				monitoringTask := MonitoringTask{
					ServiceId: serviceId,
					URL:       url,
				}

				data, err := json.Marshal(monitoringTask)
				if err != nil {
					log.Printf("Error marshaling task: %v", err)
					continue
				}

				result := s.incidentTopic.Publish(ctx,
					&pubsub.Message{
						Data: data,
					},
				)
				_, err = result.Get(ctx)

				if err != nil {
					log.Printf("Error could not write monitoringTask to the broker: %v\n", err)
				}
			}
		}

	}(goRoutineCtx, healthCheckInterval, serviceId, url)
}

func main() {
	cfg := config.GetConfig()

	fmt.Println("Scheduler Service is starting...")
	fmt.Printf("Config loaded. Project ID: %s\n", cfg.ProjectID)
	addr := fmt.Sprintf("%s:%d", cfg.APIHost, cfg.RPCPort)
	fmt.Printf("Connection to API at %s\n", addr)

	// pubsub connection
	pubsubClient, err := pubsub.NewClient(context.Background(), cfg.ProjectID)

	if err != nil {
		log.Fatalf("Failed to connect ot pubsub: %v", err)
	}

	// grpc connection
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Failed to connect ot API: %v", err)
	}

	defer conn.Close()

	client := rpc.NewSchedulerServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.GetAllSchedulerConfigurations(ctx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("Error calling GetAllSchedulerConfigurations: %v", err)
	}

	fmt.Printf("Success, recived %d configurations\n", len(resp.Services))

	topicName := "execute-health-check"

	sched := &scheduler{
		activeTasks:   make(map[uint64]*Task),
		client:        client,
		pubsubClient:  pubsubClient,
		incidentTopic: pubsubClient.Topic(topicName),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	for _, service := range resp.Services {
		sched.scheduleMonitor(ctx, service)
	}

	// its waiting for a ctrl+c signal
	<-ctx.Done()
}
