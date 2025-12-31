package internal

import (
	redis_keys "alerting-plafform/incident-manager/redis"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"alerting-platform/common/db"
	pubsub_common "alerting-platform/common/pubsub"

	"github.com/redis/go-redis/v9"
)

func (managerState *ManagerState) HandleMessage(ctx context.Context, msg pubsub_common.PubSubMessage, eventType string) {
	payload, eventTime, err := pubsub_common.ExtractPayload(msg)
	if err != nil {
		log.Printf("[ERROR] Error extracting payload for topic %s: %v. Dropping message.", eventType, err)
		msg.Ack()
		return
	}

	switch eventType {
	case pubsub_common.ServiceUpTopic:
		err = managerState.HandleServiceUp(ctx, *payload, *eventTime)
	case pubsub_common.ServiceDownTopic:
		err = managerState.HandleServiceDown(ctx, *payload, *eventTime)
	case pubsub_common.ServiceCreatedTopic:
		err = managerState.HandleServiceCreated(ctx, *payload, *eventTime)
	case pubsub_common.ServiceModifiedTopic:
		err = managerState.HandleServiceModified(ctx, *payload, *eventTime)
	case pubsub_common.ServiceRemovedTopic:
		err = managerState.HandleServiceRemoved(ctx, *payload, *eventTime)
	case pubsub_common.OncallerAcknowledgedTopic:
		err = managerState.HandleOncallerAcknowledged(ctx, *payload, *eventTime)
	default:
		log.Printf("[WARNING] Unknown event type: %s", eventType)
	}

	if err != nil {
		log.Printf("[ERROR] Error handling message for topic %s: %v", eventType, err)
		msg.Nack()
	} else {
		msg.Ack()
	}
}

func (managerState *ManagerState) HandleServiceUp(ctx context.Context, payload pubsub_common.PubSubPayload, eventTime time.Time) error {

	redisClient := db.GetRedisClient()
	serviceStatusKey := redis_keys.GetServiceStatusKey(payload.ServiceID)
	downSinceKey := redis_keys.GetDownSinceKey(payload.ServiceID)

	log.Printf("[DEBUG] Service %d is UP", payload.ServiceID)

	pipe := redisClient.TxPipeline()

	pipe.Set(ctx, serviceStatusKey, "UP", 0).Err()
	pipe.Del(ctx, downSinceKey).Err()

	_, err := pipe.Exec(ctx)

	return err
}

func (managerState *ManagerState) HandleServiceDown(ctx context.Context, payload pubsub_common.PubSubPayload, eventTime time.Time) error {
	redisClient := db.GetRedisClient()
	serviceStatusKey := redis_keys.GetServiceStatusKey(payload.ServiceID)
	downSinceKey := redis_keys.GetDownSinceKey(payload.ServiceID)

	log.Printf("[DEBUG] Service %d is DOWN", payload.ServiceID)

	err := redisClient.Set(ctx, serviceStatusKey, "DOWN", 0).Err()

	if err != nil {
		return err
	}

	downSinceStr, err := redisClient.Get(ctx, downSinceKey).Result()

	if err == redis.Nil {
		err = redisClient.Set(ctx, downSinceKey, eventTime.Unix(), 0).Err()
		return err
	}

	downSince, err := strconv.ParseInt(downSinceStr, 10, 64)

	if err != nil {
		return err
	}

	currentTime := time.Now().UTC().Unix()

	managerState.mu.Lock()
	service, exists := managerState.Services[payload.ServiceID]
	managerState.mu.Unlock()

	if !exists {
		log.Printf("[WARNING] Service %d not found in configuration", payload.ServiceID)
		return nil
	}

	alertWindow := int64(service.AlertWindow)

	if currentTime-downSince >= alertWindow {
		incidentKey := redis_keys.GetIncidentKey(payload.ServiceID)
		exists := redisClient.Exists(ctx, incidentKey).Val()

		if exists == 0 {
			err = managerState.HandleNewIncident(ctx, payload.ServiceID, time.Unix(downSince, 0))
		}
	}

	return err
}

func (managerState *ManagerState) HandleServiceCreated(ctx context.Context, payload pubsub_common.PubSubPayload, eventTime time.Time) error {
	log.Printf("[DEBUG] Service %d created", payload.ServiceID)

	managerState.mu.Lock()
	defer managerState.mu.Unlock()

	service := ServiceInfo{
		ID:                  payload.ServiceID,
		AlertWindow:         payload.Data.AlertWindow,
		AllowedResponseTime: payload.Data.AllowedResponseTime,
		Oncallers:           payload.Data.Oncallers,
	}

	managerState.Services[service.ID] = service
	return nil
}

func (managerState *ManagerState) HandleServiceModified(ctx context.Context, payload pubsub_common.PubSubPayload, eventTime time.Time) error {
	log.Printf("[DEBUG] Service %d modified", payload.ServiceID)

	managerState.mu.Lock()
	defer managerState.mu.Unlock()

	service, exists := managerState.Services[payload.ServiceID]

	if !exists {
		log.Printf("[WARNING] Service %d not found in configuration", payload.ServiceID)
		return nil
	}

	service.AlertWindow = payload.Data.AlertWindow
	service.AllowedResponseTime = payload.Data.AllowedResponseTime
	service.Oncallers = payload.Data.Oncallers

	managerState.Services[service.ID] = service
	return nil
}

func (managerState *ManagerState) HandleServiceRemoved(ctx context.Context, payload pubsub_common.PubSubPayload, eventTime time.Time) error {
	log.Printf("[DEBUG] Service %d removed", payload.ServiceID)

	managerState.mu.Lock()
	delete(managerState.Services, payload.ServiceID)
	managerState.mu.Unlock()

	redisClient := db.GetRedisClient()
	incidentKey := redis_keys.GetIncidentKey(payload.ServiceID)

	exists := redisClient.Exists(ctx, incidentKey).Val()

	if exists != 0 {
		err := redisClient.Del(ctx, incidentKey).Err()
		if err != nil {
			log.Printf("[ERROR] Failed to delete incident for removed service %d: %v", payload.ServiceID, err)
			return err
		}
		log.Printf("[DEBUG] Deleted ongoing incident for removed service %d", payload.ServiceID)
	}

	oncallerDeadlineSetKey := redis_keys.GetOncallerDeadlineSetKey()
	err := redisClient.ZRem(ctx, oncallerDeadlineSetKey, payload.ServiceID).Err()
	if err != nil {
		log.Printf("[ERROR] Failed to remove service %d from oncaller deadline set: %v", payload.ServiceID, err)
		return err
	}

	return nil
}

func (managerState *ManagerState) HandleOncallerAcknowledged(ctx context.Context, payload pubsub_common.PubSubPayload, eventTime time.Time) error {
	log.Printf("[DEBUG] Oncaller %s acknowledged incident for service %d", payload.OnCaller, payload.ServiceID)

	return managerState.handleIncidentResolved(ctx, payload.ServiceID, payload.OnCaller)
}

func (managerState *ManagerState) HandleNewIncident(ctx context.Context, serviceID uint64, incidentStartTime time.Time) error {
	redisClient := db.GetRedisClient()
	incidentKey := redis_keys.GetIncidentKey(serviceID)

	incidentID := fmt.Sprintf("%d-%d", serviceID, incidentStartTime.Unix())

	log.Printf("[DEBUG] Starting incident %s for service %d", incidentID, serviceID)

	managerState.mu.Lock()
	service, exists := managerState.Services[serviceID]
	managerState.mu.Unlock()

	if !exists {
		log.Printf("[WARNING] Service %d not found in configuration", serviceID)
		return nil
	}

	secondOncaller := ""
	if len(service.Oncallers) > 1 {
		secondOncaller = service.Oncallers[1]
	}

	incidentInfo := IncidentInfo{
		IncidentID:          incidentID,
		ServiceID:           serviceID,
		State:               IncidentStateWaitingForFirstAck,
		IncidentStartTime:   incidentStartTime.Unix(),
		AllowedResponseTime: service.AllowedResponseTime,
		FirstOncaller:       service.Oncallers[0],
		SecondOncaller:      secondOncaller,
	}

	err := redisClient.HSet(ctx, incidentKey, incidentInfo).Err()

	if err != nil {
		return err
	}

	oncallerDeadlineSetKey := redis_keys.GetOncallerDeadlineSetKey()
	oncallerResponseDeadline := incidentStartTime.Add(time.Duration(incidentInfo.AllowedResponseTime) * time.Minute).Unix()

	err = redisClient.ZAdd(ctx, oncallerDeadlineSetKey, redis.Z{
		Score:  float64(oncallerResponseDeadline),
		Member: incidentInfo.ServiceID,
	}).Err()

	if err != nil {
		return err
	}

	go func() {
		err := managerState.SendIncidentStartMessage(
			context.Background(),
			incidentInfo.IncidentID,
			incidentInfo.ServiceID,
			incidentStartTime,
		)

		if err != nil {
			log.Printf("[ERROR] Failed to send incident start message for service %d: %v", serviceID, err)
		}
	}()

	go func() {
		err := managerState.SendNotifyOncallerMessage(
			context.Background(),
			incidentInfo.IncidentID,
			incidentInfo.ServiceID,
			incidentInfo.FirstOncaller,
			time.Now().UTC(),
		)

		if err != nil {
			log.Printf("[ERROR] Failed to send notify oncaller message for service %d: %v", serviceID, err)
		}
	}()

	return nil
}

func (managerState *ManagerState) HandleExpiredDeadline(ctx context.Context, serviceID uint64) error {
	redisClient := db.GetRedisClient()
	incidentKey := redis_keys.GetIncidentKey(serviceID)
	oncallerDeadlineSetKey := redis_keys.GetOncallerDeadlineSetKey()

	err := redisClient.ZRem(ctx, oncallerDeadlineSetKey, serviceID).Err()

	if err != nil {
		return err
	}

	incident, err := redisClient.HGetAll(ctx, incidentKey).Result()

	if err != nil {
		return err
	}

	allowedResponseTime, err := strconv.Atoi(incident["allowed_response_time"])
	if err != nil {
		return err
	}

	incidentStartTime, err := strconv.ParseInt(incident["incident_start_time"], 10, 64)
	if err != nil {
		return err
	}

	incidentInfo := IncidentInfo{
		IncidentID:          incident["incident_id"],
		ServiceID:           serviceID,
		State:               incident["state"],
		IncidentStartTime:   incidentStartTime,
		AllowedResponseTime: allowedResponseTime,
		FirstOncaller:       incident["first_oncaller"],
		SecondOncaller:      incident["second_oncaller"],
	}

	var requestedOncaller string
	switch incidentInfo.State {
	case IncidentStateWaitingForFirstAck:
		requestedOncaller = incidentInfo.FirstOncaller
	case IncidentStateWaitingForSecondAck:
		requestedOncaller = incidentInfo.SecondOncaller
	default:
		log.Printf("[WARNING] Unknown incident state for service %d: %s", serviceID, incidentInfo.State)
		return nil
	}

	log.Printf("[DEBUG] Deadline expired for service %d, oncaller %s", serviceID, requestedOncaller)

	go func() {
		err := managerState.SendAcknowledgeTimeoutMessage(
			context.Background(),
			incidentInfo.IncidentID,
			serviceID,
			requestedOncaller,
			time.Now().UTC(),
		)

		if err != nil {
			log.Printf("[ERROR] Failed to send acknowledge timeout message for service %d: %v", serviceID, err)
		}
	}()

	switch incidentInfo.State {
	case IncidentStateWaitingForFirstAck:

		err = redisClient.HSet(ctx, incidentKey, "state", IncidentStateWaitingForSecondAck).Err()
		if err != nil {
			return err
		}

		if incidentInfo.SecondOncaller == "" {
			log.Printf("[DEBUG] No second oncaller. Marking incident as unresolved")
			return managerState.handleIncidentUnresolved(ctx, serviceID)
		}

		log.Printf("[DEBUG] Second oncaller configured. Notifying %s", incidentInfo.SecondOncaller)

		oncallerResponseDeadline := time.Now().UTC().Add(time.Duration(incidentInfo.AllowedResponseTime) * time.Minute).Unix()

		err = redisClient.ZAdd(ctx, oncallerDeadlineSetKey, redis.Z{
			Score:  float64(oncallerResponseDeadline),
			Member: incidentInfo.ServiceID,
		}).Err()

		go func() {
			err := managerState.SendNotifyOncallerMessage(
				context.Background(),
				incidentInfo.IncidentID,
				serviceID,
				incidentInfo.SecondOncaller,
				time.Now().UTC(),
			)

			if err != nil {
				log.Printf("[ERROR] Failed to send notify oncaller message for service %d: %v", serviceID, err)
			}
		}()

		return err
	case IncidentStateWaitingForSecondAck:
		log.Printf("[DEBUG] Second oncaller did not respond in time. Marking incident as unresolved")
		return managerState.handleIncidentUnresolved(ctx, serviceID)
	default:
		log.Printf("[WARNING] Unknown incident state for service %d: %s", serviceID, incidentInfo.State)
	}

	return nil

}

func (managerState *ManagerState) handleIncidentUnresolved(ctx context.Context, serviceID uint64) error {
	redisClient := db.GetRedisClient()
	incidentKey := redis_keys.GetIncidentKey(serviceID)
	downSinceKey := redis_keys.GetDownSinceKey(serviceID)

	incidentID, err := redisClient.HGet(ctx, incidentKey, "incident_id").Result()

	log.Printf("[DEBUG] Incident %s for service %d was not resolved in time", incidentID, serviceID)

	if err != nil {
		return err
	}

	pipe := redisClient.TxPipeline()

	pipe.Del(ctx, incidentKey).Err()
	pipe.Del(ctx, downSinceKey).Err()

	_, err = pipe.Exec(ctx)

	if err != nil {
		return err
	}

	go func() {
		err := managerState.SendIncidentUnresolvedMessage(
			context.Background(),
			incidentID,
			serviceID,
			time.Now().UTC(),
		)

		if err != nil {
			log.Printf("[ERROR] Failed to send incident unresolved message for service %d: %v", serviceID, err)
		}
	}()

	return nil
}

func (managerState *ManagerState) handleIncidentResolved(ctx context.Context, serviceID uint64, oncaller string) error {
	redisClient := db.GetRedisClient()
	incidentKey := redis_keys.GetIncidentKey(serviceID)
	downSinceKey := redis_keys.GetDownSinceKey(serviceID)

	incidentID, err := redisClient.HGet(ctx, incidentKey, "incident_id").Result()

	log.Printf("[DEBUG] Incident %s for service %d was resolved in time by %s", incidentID, serviceID, oncaller)

	if err != nil {
		return err
	}

	pipe := redisClient.TxPipeline()

	pipe.Del(ctx, incidentKey).Err()
	pipe.Del(ctx, downSinceKey).Err()

	_, err = pipe.Exec(ctx)

	if err != nil {
		return err
	}

	go func() {
		err := managerState.SendIncidentResolvedMessage(
			context.Background(),
			incidentID,
			serviceID,
			oncaller,
			time.Now().UTC(),
		)

		if err != nil {
			log.Printf("[ERROR] Failed to send incident resolved message for service %d: %v", serviceID, err)
		}
	}()

	return nil
}
