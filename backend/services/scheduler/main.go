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
	"alerting-platform/common/live"
	pubsub_common "alerting-platform/common/pubsub"
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

func (s *scheduler) removeTask(id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if task, exists := s.activeTasks[id]; exists {
		task.Cancel()
		delete(s.activeTasks, id)
	}
}

func (s *scheduler) updateTask(ctx context.Context, service *rpc.ServiceInfoForScheduler) {
	s.removeTask(service.ServiceId)
	s.scheduleMonitor(ctx, service)
}

func (s *scheduler) HandleMessage(ctx context.Context, msg pubsub_common.PubSubMessage, eventType string) {
	payload, _, err := pubsub_common.ExtractPayload(msg)

	if err != nil {
		log.Printf("[ERROR] Error extracting payload for topic %s: %v\n", eventType, err)
		log.Printf("Dropping message...")
		msg.Ack()
		return
	}

	switch eventType {
	case pubsub_common.ServiceCreatedTopic, pubsub_common.ServiceModifiedTopic:
		serviceInfo := &rpc.ServiceInfoForScheduler{
			ServiceId:           payload.ServiceID,
			Url:                 payload.Data.URL,
			HealthCheckInterval: int64(payload.Data.HealthCheckInterval),
		}

		log.Printf("[INFO] Updating/Adding task for service %d (url: %s)", payload.ServiceID, payload.Data.URL)
		s.updateTask(ctx, serviceInfo)
	case pubsub_common.ServiceRemovedTopic:
		log.Printf("[INFO] Removing task for service %d", payload.ServiceID)
		s.removeTask(payload.ServiceID)
	default:
		log.Printf("[WARNING] Unknown event type: %s", eventType)
	}

	msg.Ack()
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
				log.Printf("[INFO] Stopping monitor for service %d", service.ServiceId)
				return
			case <-ticker.C:
				monitoringTask := MonitoringTask{
					ServiceId: serviceId,
					URL:       url,
				}

				data, err := json.Marshal(monitoringTask)
				if err != nil {
					log.Printf("Error marshaling task: %v", err)
					continue
				}

				err = pubsub_common.SendMessage(goRoutineCtx, s.pubsubClient, pubsub_common.ExecuteHealthCheckTopic, data, fmt.Sprintf("%d", serviceId))

				if err != nil {
					log.Printf("Error could not write monitoringTask to the broker: %v\n", err)
				}
			}
		}

	}(goRoutineCtx, healthCheckInterval, serviceId, url)
}

func main() {
	config.Intro("Scheduler")
	cfg := config.GetConfig()

	fmt.Printf("Config loaded. Project ID: %s\n", cfg.ProjectID)
	addr := fmt.Sprintf("%s:%d", cfg.APIHost, cfg.RPCPort)
	fmt.Printf("Connection to API at %s\n", addr)

	subscriptions := map[string]string{
		"scheduler-service-created":  pubsub_common.ServiceCreatedTopic,
		"scheduler-service-modified": pubsub_common.ServiceModifiedTopic,
		"scheduler-service-removed":  pubsub_common.ServiceRemovedTopic,
	}

	pubsubClient, err := pubsub.NewClient(context.Background(), cfg.ProjectID)

	if err != nil {
		log.Fatalf("Failed to connect ot pubsub: %v", err)
	}

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

	sched := &scheduler{
		activeTasks:   make(map[uint64]*Task),
		client:        client,
		pubsubClient:  pubsubClient,
		incidentTopic: pubsubClient.Topic(pubsub_common.ExecuteHealthCheckTopic),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	if cfg.Env == config.PROD {
		live.StartLiveServer(&wg)
	}

	pubsub_common.CreateSubscriptionsAndTopics(pubsubClient, subscriptions, []string{pubsub_common.ExecuteHealthCheckTopic})
	pubsub_common.SetupSubscriptionListeners(ctx, pubsubClient, subscriptions, &wg, sched.HandleMessage)

	for _, service := range resp.Services {
		sched.scheduleMonitor(ctx, service)
	}

	log.Println("Scheduler service is running...")
	wg.Wait()
}
