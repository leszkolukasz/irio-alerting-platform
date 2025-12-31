package internal

import (
	"alerting-plafform/incident-manager/rpc"
	"context"
	"sync"

	pubsub_internal "alerting-plafform/incident-manager/pubsub"

	"cloud.google.com/go/pubsub"
)

type ServiceInfo struct {
	ID uint64
	// DownSince           int64 // stored in Redis
	AlertWindow         int // in seconds
	AllowedResponseTime int // in minutes
	Oncallers           []string
}

const (
	IncidentStateStarted             = "STARTED"
	IncidentStateWaitingForFirstAck  = "WAITING_FOR_FIRST_ACK"
	IncidentStateWaitingForSecondAck = "WAITING_FOR_SECOND_ACK"
)

type IncidentInfo struct {
	IncidentID        string `redis:"incident_id"`
	ServiceID         uint64 `redis:"service_id"`
	State             string `redis:"state"`
	IncidentStartTime int64  `redis:"incident_start_time"`

	// Copied in case service info is changed through API
	AllowedResponseTime int    `redis:"allowed_response_time"`
	FirstOncaller       string `redis:"first_oncaller"`
	SecondOncaller      string `redis:"second_oncaller"`
}

type ManagerState struct {
	mu            sync.Mutex             // locks read/writes access to locks/Services
	locks         map[uint64]*sync.Mutex // lock per service
	pubSubService pubsub_internal.PubSubServiceI
	services      map[uint64]ServiceInfo
}

func NewManagerState(ctx context.Context, psClient *pubsub.Client) *ManagerState {
	state := &ManagerState{
		pubSubService: pubsub_internal.NewPubSubService(psClient),
		services:      make(map[uint64]ServiceInfo),
	}

	servicesInfo := rpc.GetAllServicesInfo(ctx)

	for _, svc := range servicesInfo.Services {
		service := ServiceInfo{
			ID:                  svc.ServiceId,
			AlertWindow:         int(svc.AlertWindow),
			AllowedResponseTime: int(svc.AllowedResponseTime),
			Oncallers:           svc.Oncallers,
		}

		state.services[service.ID] = service
	}

	return state
}

func (managerState *ManagerState) LockService(serviceID uint64) *sync.Mutex {
	managerState.mu.Lock()

	lock, exists := managerState.locks[serviceID]
	if !exists {
		lock = &sync.Mutex{}
		managerState.locks[serviceID] = lock
	}

	managerState.mu.Unlock()

	lock.Lock()

	return lock
}
