package internal

import (
	redis_keys "alerting-plafform/incident-manager/redis"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	pubsub_internal "alerting-plafform/incident-manager/pubsub"
	"alerting-platform/common/db"
	pubsub_common "alerting-platform/common/pubsub"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestState(t *testing.T) (*miniredis.Miniredis, *redis.Client, *pubsub_internal.MockPubSubService, *ManagerState) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub redis connection", err)
	}
	rclient := db.MockRedis(s.Addr())

	mockPubSub := new(pubsub_internal.MockPubSubService)

	state := &ManagerState{
		mu:            sync.Mutex{},
		locks:         make(map[uint64]*sync.Mutex),
		pubSubService: mockPubSub,
		services:      make(map[uint64]ServiceInfo),
	}

	os.Setenv("REDIS_PREFIX", "test-prefix:")

	return s, rclient, mockPubSub, state
}

func TestHandleServiceUp(t *testing.T) {
	ctx := context.Background()
	serviceID := uint64(1)
	payload := pubsub_common.PubSubPayload{ServiceID: serviceID}

	t.Run("Success", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()

		serviceStatusKey := redis_keys.GetServiceStatusKey(serviceID)
		downSinceKey := redis_keys.GetDownSinceKey(serviceID)
		s.Set(serviceStatusKey, "DOWN")
		s.Set(downSinceKey, "12345")

		err := managerState.HandleServiceUp(ctx, payload, time.Now())
		assert.NoError(t, err)

		status, err := s.Get(serviceStatusKey)
		assert.NoError(t, err)
		assert.Equal(t, "UP", status)
		assert.False(t, s.Exists(downSinceKey))
	})

	t.Run("Error on Redis Exec", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()

		s.SetError("redis error")

		err := managerState.HandleServiceUp(ctx, payload, time.Now())
		assert.Error(t, err)
	})
}

func TestHandleServiceCreated(t *testing.T) {
	ctx := context.Background()
	payload := pubsub_common.PubSubPayload{
		ServiceID: 1,
		Data: pubsub_common.PubSubPayloadData{
			AlertWindow:         300,
			AllowedResponseTime: 5,
			Oncallers:           []string{"test@oncaller.com"},
		},
	}

	t.Run("Success", func(t *testing.T) {
		_, _, _, managerState := setupTestState(t)

		err := managerState.HandleServiceCreated(ctx, payload, time.Now())
		assert.NoError(t, err)

		service, exists := managerState.services[payload.ServiceID]
		assert.True(t, exists)
		assert.Equal(t, payload.Data.AlertWindow, service.AlertWindow)
		assert.Equal(t, payload.Data.AllowedResponseTime, service.AllowedResponseTime)
		assert.Equal(t, payload.Data.Oncallers, service.Oncallers)
	})
}

func TestHandleServiceModified(t *testing.T) {
	ctx := context.Background()
	serviceID := uint64(1)
	payload := pubsub_common.PubSubPayload{
		ServiceID: serviceID,
		Data: pubsub_common.PubSubPayloadData{
			AlertWindow:         600,
			AllowedResponseTime: 10,
			Oncallers:           []string{"new@oncaller.com"},
		},
	}

	t.Run("Success", func(t *testing.T) {
		_, _, _, managerState := setupTestState(t)
		managerState.services[serviceID] = ServiceInfo{
			ID:                  serviceID,
			AlertWindow:         300,
			AllowedResponseTime: 5,
			Oncallers:           []string{"old@oncaller.com"},
		}

		err := managerState.HandleServiceModified(ctx, payload, time.Now())
		assert.NoError(t, err)

		service, exists := managerState.services[serviceID]
		assert.True(t, exists)
		assert.Equal(t, payload.Data.AlertWindow, service.AlertWindow)
		assert.Equal(t, payload.Data.AllowedResponseTime, service.AllowedResponseTime)
		assert.Equal(t, payload.Data.Oncallers, service.Oncallers)
	})
}

func TestHandleServiceDown(t *testing.T) {
	ctx := context.Background()
	serviceID := uint64(1)
	payload := pubsub_common.PubSubPayload{ServiceID: serviceID}

	t.Run("Success - First Time Down", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()

		now := time.Now()
		err := managerState.HandleServiceDown(ctx, payload, now)
		assert.NoError(t, err)

		status, err := s.Get(redis_keys.GetServiceStatusKey(serviceID))
		assert.NoError(t, err)
		assert.Equal(t, "DOWN", status)

		downSince, err := s.Get(redis_keys.GetDownSinceKey(serviceID))
		assert.NoError(t, err)
		downSinceInt, _ := strconv.ParseInt(downSince, 10, 64)
		assert.Equal(t, now.Unix(), downSinceInt)
	})

	t.Run("Success - Already Down, Create Incident", func(t *testing.T) {
		s, _, mockPubSub, managerState := setupTestState(t)
		defer s.Close()

		alertWindow := 5
		downSince := time.Now().UTC().Add(-time.Duration(alertWindow+1) * time.Second)
		managerState.services[serviceID] = ServiceInfo{
			ID:                  serviceID,
			AlertWindow:         alertWindow,
			AllowedResponseTime: 10,
			Oncallers:           []string{"test@oncaller.com"},
		}

		s.Set(redis_keys.GetDownSinceKey(serviceID), strconv.FormatInt(downSince.Unix(), 10))
		mockPubSub.On("SendIncidentStartMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		mockPubSub.On("SendNotifyOncallerMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := managerState.HandleServiceDown(ctx, payload, time.Now())
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		assert.True(t, s.Exists(redis_keys.GetIncidentKey(serviceID)))
		mockPubSub.AssertExpectations(t)
	})

	t.Run("Error on Set Status", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()

		s.SetError("redis error")

		err := managerState.HandleServiceDown(ctx, payload, time.Now())
		assert.Error(t, err)
	})

	t.Run("Error on Set DownSince", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()

		serviceStatusKey := redis_keys.GetServiceStatusKey(serviceID)
		downSinceKey := redis_keys.GetDownSinceKey(serviceID)

		s.Set(serviceStatusKey, "DOWN")
		s.Del(downSinceKey)

		s.SetError("redis error")

		err := managerState.HandleServiceDown(ctx, payload, time.Now())
		assert.Error(t, err)
		s.SetError("")
	})
}

func TestHandleServiceRemoved(t *testing.T) {
	ctx := context.Background()
	serviceID := uint64(1)
	payload := pubsub_common.PubSubPayload{ServiceID: serviceID}

	t.Run("Success", func(t *testing.T) {
		s, rclient, _, managerState := setupTestState(t)
		defer s.Close()

		managerState.services[serviceID] = ServiceInfo{ID: serviceID}
		s.Set(redis_keys.GetIncidentKey(serviceID), "incident-data")
		s.ZAdd(redis_keys.GetOncallerDeadlineSetKey(), 1, fmt.Sprintf("%d", serviceID))

		err := managerState.HandleServiceRemoved(ctx, payload, time.Now())
		assert.NoError(t, err)

		_, exists := managerState.services[serviceID]
		assert.False(t, exists)
		assert.False(t, s.Exists(redis_keys.GetIncidentKey(serviceID)))
		members, err := rclient.ZRange(t.Context(), redis_keys.GetOncallerDeadlineSetKey(), 0, -1).Result()
		assert.NoError(t, err)
		assert.Empty(t, members)
	})

	t.Run("Error on Deleting Incident", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()

		incidentKey := redis_keys.GetIncidentKey(serviceID)
		s.Set(incidentKey, "some-data")

		s.SetError("redis error")

		err := managerState.HandleServiceRemoved(ctx, payload, time.Now())
		assert.Error(t, err)
		s.SetError("")
	})

	t.Run("Error on ZRem", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()

		oncallerDeadlineSetKey := redis_keys.GetOncallerDeadlineSetKey()
		s.ZAdd(oncallerDeadlineSetKey, 1, fmt.Sprintf("%d", serviceID))

		s.SetError("redis error")

		err := managerState.HandleServiceRemoved(ctx, payload, time.Now())
		assert.Error(t, err)
		s.SetError("")
	})
}

func TestHandleOncallerAcknowledged(t *testing.T) {
	ctx := context.Background()
	payload := pubsub_common.PubSubPayload{
		ServiceID: 1,
		OnCaller:  "test@oncaller.com",
	}

	t.Run("Success", func(t *testing.T) {
		s, _, mockPubSub, managerState := setupTestState(t)
		defer s.Close()

		incidentKey := redis_keys.GetIncidentKey(payload.ServiceID)
		s.HSet(incidentKey, "incident_id", "test-incident")

		mockPubSub.On("SendIncidentResolvedMessage", mock.Anything, "test-incident", payload.ServiceID, payload.OnCaller, mock.Anything).Return(nil).Once()

		err := managerState.HandleOncallerAcknowledged(ctx, payload, time.Now())
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		assert.False(t, s.Exists(incidentKey))
		mockPubSub.AssertExpectations(t)
	})
}

func TestHandleNewIncident(t *testing.T) {
	ctx := context.Background()
	serviceID := uint64(1)
	incidentStartTime := time.Now().UTC()

	t.Run("Error on HSet", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()

		managerState.services[serviceID] = ServiceInfo{
			ID:        serviceID,
			Oncallers: []string{"test@oncaller.com"},
		}

		s.SetError("redis error")

		err := managerState.HandleNewIncident(ctx, serviceID, incidentStartTime)
		assert.Error(t, err)
		s.SetError("")
	})

	t.Run("Error on ZAdd", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()

		managerState.services[serviceID] = ServiceInfo{
			ID:        serviceID,
			Oncallers: []string{"test@oncaller.com"},
		}

		s.SetError("redis error")

		err := managerState.HandleNewIncident(ctx, serviceID, incidentStartTime)
		assert.Error(t, err)
		s.SetError("")
	})

	t.Run("Error sending incident start message", func(t *testing.T) {
		s, _, mockPubSub, managerState := setupTestState(t)
		defer s.Close()

		managerState.services[serviceID] = ServiceInfo{
			ID:                  serviceID,
			AllowedResponseTime: 5,
			Oncallers:           []string{"test@oncaller.com"},
		}

		mockPubSub.On("SendIncidentStartMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("pubsub error")).Once()
		mockPubSub.On("SendNotifyOncallerMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		err := managerState.HandleNewIncident(ctx, serviceID, incidentStartTime)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		mockPubSub.AssertExpectations(t)
	})

	t.Run("Error sending notify oncaller message", func(t *testing.T) {
		s, _, mockPubSub, managerState := setupTestState(t)
		defer s.Close()

		managerState.services[serviceID] = ServiceInfo{
			ID:                  serviceID,
			AllowedResponseTime: 5,
			Oncallers:           []string{"test@oncaller.com"},
		}

		mockPubSub.On("SendIncidentStartMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		mockPubSub.On("SendNotifyOncallerMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("pubsub error")).Once()

		err := managerState.HandleNewIncident(ctx, serviceID, incidentStartTime)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		mockPubSub.AssertExpectations(t)
	})
}

func TestHandleExpiredDeadline(t *testing.T) {
	ctx := context.Background()
	serviceID := uint64(1)

	t.Run("Success - Escalate to Second Oncaller", func(t *testing.T) {
		s, _, mockPubSub, managerState := setupTestState(t)
		defer s.Close()

		incidentKey := redis_keys.GetIncidentKey(serviceID)
		incidentInfo := IncidentInfo{
			IncidentID:          "test-incident",
			State:               IncidentStateWaitingForFirstAck,
			AllowedResponseTime: 5,
			FirstOncaller:       "first@oncaller.com",
			SecondOncaller:      "second@oncaller.com",
		}
		s.HSet(incidentKey, "incident_id", incidentInfo.IncidentID, "state", incidentInfo.State, "allowed_response_time", strconv.Itoa(incidentInfo.AllowedResponseTime), "first_oncaller", incidentInfo.FirstOncaller, "second_oncaller", incidentInfo.SecondOncaller, "incident_start_time", strconv.Itoa(int(time.Now().Unix())))
		// Set the incident start time to the current time
		mockPubSub.On("SendAcknowledgeTimeoutMessage", mock.Anything, incidentInfo.IncidentID, serviceID, incidentInfo.FirstOncaller, mock.Anything).Return(nil).Once()
		mockPubSub.On("SendNotifyOncallerMessage", mock.Anything, incidentInfo.IncidentID, serviceID, incidentInfo.SecondOncaller, mock.Anything).Return(nil).Once()

		err := managerState.HandleExpiredDeadline(ctx, serviceID)
		assert.NoError(t, err)

		state := s.HGet(incidentKey, "state")
		assert.Equal(t, IncidentStateWaitingForSecondAck, state)

		deadlineSetKey := redis_keys.GetOncallerDeadlineSetKey()
		assert.True(t, s.Exists(deadlineSetKey))

		time.Sleep(100 * time.Millisecond)
		mockPubSub.AssertExpectations(t)
	})

	t.Run("Success - Unresolved", func(t *testing.T) {
		s, _, mockPubSub, managerState := setupTestState(t)
		defer s.Close()

		incidentKey := redis_keys.GetIncidentKey(serviceID)
		incidentInfo := IncidentInfo{
			IncidentID:          "test-incident",
			State:               IncidentStateWaitingForSecondAck,
			AllowedResponseTime: 5,
			FirstOncaller:       "first@oncaller.com",
			SecondOncaller:      "second@oncaller.com",
		}
		s.HSet(incidentKey, "incident_id", incidentInfo.IncidentID, "state", incidentInfo.State, "allowed_response_time", strconv.Itoa(incidentInfo.AllowedResponseTime), "first_oncaller", incidentInfo.FirstOncaller, "second_oncaller", incidentInfo.SecondOncaller, "incident_start_time", strconv.Itoa(int(time.Now().Unix())))

		mockPubSub.On("SendAcknowledgeTimeoutMessage", mock.Anything, incidentInfo.IncidentID, serviceID, incidentInfo.SecondOncaller, mock.Anything).Return(nil).Once()
		mockPubSub.On("SendIncidentUnresolvedMessage", mock.Anything, incidentInfo.IncidentID, serviceID, mock.Anything).Return(nil).Once()

		err := managerState.HandleExpiredDeadline(ctx, serviceID)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		assert.False(t, s.Exists(incidentKey))
		mockPubSub.AssertExpectations(t)
	})

	t.Run("Error on ZRem", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()
		s.SetError("redis error")
		err := managerState.HandleExpiredDeadline(ctx, serviceID)
		assert.Error(t, err)
		s.SetError("")
	})

	t.Run("Error on HGetAll", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()
		s.SetError("redis error")
		err := managerState.HandleExpiredDeadline(ctx, serviceID)
		assert.Error(t, err)
		s.SetError("")
	})

	t.Run("Error sending acknowledge timeout message", func(t *testing.T) {
		s, _, mockPubSub, managerState := setupTestState(t)
		defer s.Close()

		incidentKey := redis_keys.GetIncidentKey(serviceID)
		incidentInfo := IncidentInfo{
			IncidentID:          "test-incident",
			ServiceID:           serviceID,
			State:               IncidentStateWaitingForFirstAck,
			IncidentStartTime:   time.Now().Unix(),
			AllowedResponseTime: 5,
			FirstOncaller:       "first@oncaller.com",
			SecondOncaller:      "",
		}
		s.HSet(incidentKey, "incident_id", incidentInfo.IncidentID, "state", incidentInfo.State, "allowed_response_time", strconv.Itoa(incidentInfo.AllowedResponseTime), "incident_start_time", strconv.FormatInt(incidentInfo.IncidentStartTime, 10), "first_oncaller", incidentInfo.FirstOncaller, "second_oncaller", incidentInfo.SecondOncaller)

		mockPubSub.On("SendAcknowledgeTimeoutMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("pubsub error")).Once()
		mockPubSub.On("SendIncidentUnresolvedMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		err := managerState.HandleExpiredDeadline(ctx, serviceID)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		mockPubSub.AssertExpectations(t)
	})
}

func TestHandleIncidentUnresolved(t *testing.T) {
	ctx := context.Background()
	serviceID := uint64(1)

	t.Run("Success", func(t *testing.T) {
		s, _, mockPubSub, managerState := setupTestState(t)
		defer s.Close()

		incidentKey := redis_keys.GetIncidentKey(serviceID)
		downSinceKey := redis_keys.GetDownSinceKey(serviceID)
		s.HSet(incidentKey, "incident_id", "test-incident")
		s.Set(downSinceKey, "12345")

		mockPubSub.On("SendIncidentUnresolvedMessage", mock.Anything, "test-incident", serviceID, mock.Anything).Return(nil).Once()

		err := managerState.handleIncidentUnresolved(ctx, serviceID)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		assert.False(t, s.Exists(incidentKey))
		assert.False(t, s.Exists(downSinceKey))
		mockPubSub.AssertExpectations(t)
	})

	t.Run("Error on HGet", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()
		s.SetError("redis error")
		err := managerState.handleIncidentUnresolved(ctx, serviceID)
		assert.Error(t, err)
		s.SetError("")
	})

	t.Run("Error on pipeline exec", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()
		incidentKey := redis_keys.GetIncidentKey(serviceID)
		s.HSet(incidentKey, "incident_id", "test-incident")
		s.SetError("redis error")
		err := managerState.handleIncidentUnresolved(ctx, serviceID)
		assert.Error(t, err)
		s.SetError("")
	})

	t.Run("Error sending unresolved message", func(t *testing.T) {
		s, _, mockPubSub, managerState := setupTestState(t)
		defer s.Close()
		incidentKey := redis_keys.GetIncidentKey(serviceID)
		s.HSet(incidentKey, "incident_id", "test-incident")

		mockPubSub.On("SendIncidentUnresolvedMessage", mock.Anything, "test-incident", serviceID, mock.Anything).Return(errors.New("pubsub error")).Once()

		err := managerState.handleIncidentUnresolved(ctx, serviceID)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		mockPubSub.AssertExpectations(t)
	})
}

func TestHandleIncidentResolved(t *testing.T) {
	ctx := context.Background()
	serviceID := uint64(1)
	oncaller := "test@oncaller.com"

	t.Run("Success", func(t *testing.T) {
		s, _, mockPubSub, managerState := setupTestState(t)
		defer s.Close()

		incidentKey := redis_keys.GetIncidentKey(serviceID)
		downSinceKey := redis_keys.GetDownSinceKey(serviceID)
		s.HSet(incidentKey, "incident_id", "test-incident")
		s.Set(downSinceKey, "12345")

		mockPubSub.On("SendIncidentResolvedMessage", mock.Anything, "test-incident", serviceID, oncaller, mock.Anything).Return(nil).Once()

		err := managerState.handleIncidentResolved(ctx, serviceID, oncaller)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		assert.False(t, s.Exists(incidentKey))
		assert.False(t, s.Exists(downSinceKey))
		mockPubSub.AssertExpectations(t)
	})

	t.Run("Error on HGet", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()
		s.SetError("redis error")
		err := managerState.handleIncidentResolved(ctx, serviceID, oncaller)
		assert.Error(t, err)
		s.SetError("")
	})

	t.Run("Error on pipeline exec", func(t *testing.T) {
		s, _, _, managerState := setupTestState(t)
		defer s.Close()
		incidentKey := redis_keys.GetIncidentKey(serviceID)
		s.HSet(incidentKey, "incident_id", "test-incident")
		s.SetError("redis error")
		err := managerState.handleIncidentResolved(ctx, serviceID, oncaller)
		assert.Error(t, err)
		s.SetError("")
	})

	t.Run("Error sending resolved message", func(t *testing.T) {
		s, _, mockPubSub, managerState := setupTestState(t)
		defer s.Close()
		incidentKey := redis_keys.GetIncidentKey(serviceID)
		s.HSet(incidentKey, "incident_id", "test-incident")

		mockPubSub.On("SendIncidentResolvedMessage", mock.Anything, "test-incident", serviceID, oncaller, mock.Anything).Return(errors.New("pubsub error")).Once()

		err := managerState.handleIncidentResolved(ctx, serviceID, oncaller)
		assert.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		mockPubSub.AssertExpectations(t)
	})
}
