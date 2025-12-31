package controllers

import (
	"alerting-platform/api/db"
	"alerting-platform/api/dto"
	"alerting-platform/api/middleware"
	"alerting-platform/api/redis"
	db_common "alerting-platform/common/db"
	"alerting-platform/common/db/firestore"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func setupTestRouter() (*gin.Engine, *MockRepository, *MockPubSubService, *firestore.MockLogRepository, *Controller) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	mockRepo := new(MockRepository)
	mockPubSub := new(MockPubSubService)
	mockLogRepo := new(firestore.MockLogRepository)

	controller := &Controller{
		Repository:    mockRepo,
		PubSubService: mockPubSub,
		LogRepository: mockLogRepo,
	}

	return router, mockRepo, mockPubSub, mockLogRepo, controller
}

func setupRedis(t *testing.T) *miniredis.Miniredis {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub redis connection", err)
	}

	db_common.MockRedis(s.Addr())

	return s
}

func TestCreateMonitoredService(t *testing.T) {
	_, mockRepo, mockPubSub, _, controller := setupTestRouter()

	jwtUser := &middleware.JWTUser{ID: 1, Email: "test@user.com"}
	serviceInput := dto.MonitoredServiceRequest{
		Name:                "Test Service",
		URL:                 "http://example.com",
		Port:                80,
		HealthCheckInterval: 60,
		AlertWindow:         300,
		AllowedResponseTime: 1000,
		FirstOncallerEmail:  "oncaller1@example.com",
	}

	t.Run("Success 201", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonValue, _ := json.Marshal(serviceInput)
		c.Request, _ = http.NewRequest(http.MethodPost, "/services", bytes.NewBuffer(jsonValue))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.IdentityKey, jwtUser)

		mockRepo.On("GetServiceByName", mock.Anything, serviceInput.Name).Return(nil, errors.New("not found")).Once()
		mockRepo.On("CreateService", mock.Anything, mock.AnythingOfType("*db.MonitoredService")).Return(nil).Once()
		mockPubSub.On("SendServiceCreatedMessage", mock.Anything, mock.AnythingOfType("db.MonitoredService")).Return(nil).Once()

		controller.CreateMonitoredService(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Monitored service created successfully")
		mockRepo.AssertExpectations(t)
		mockPubSub.AssertExpectations(t)
	})

	t.Run("Invalid Input 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodPost, "/services", bytes.NewBufferString("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.IdentityKey, jwtUser)

		controller.CreateMonitoredService(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid input")
	})

	t.Run("Service Name Taken 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonValue, _ := json.Marshal(serviceInput)
		c.Request, _ = http.NewRequest(http.MethodPost, "/services", bytes.NewBuffer(jsonValue))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.IdentityKey, jwtUser)

		mockRepo.On("GetServiceByName", mock.Anything, serviceInput.Name).Return(&db.MonitoredService{}, nil).Once()

		controller.CreateMonitoredService(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Service name already taken")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed to create service 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonValue, _ := json.Marshal(serviceInput)
		c.Request, _ = http.NewRequest(http.MethodPost, "/services", bytes.NewBuffer(jsonValue))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.IdentityKey, jwtUser)

		mockRepo.On("GetServiceByName", mock.Anything, serviceInput.Name).Return(nil, errors.New("not found")).Once()
		mockRepo.On("CreateService", mock.Anything, mock.AnythingOfType("*db.MonitoredService")).Return(errors.New("db error")).Once()

		controller.CreateMonitoredService(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create monitored service")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed to send message 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonValue, _ := json.Marshal(serviceInput)
		c.Request, _ = http.NewRequest(http.MethodPost, "/services", bytes.NewBuffer(jsonValue))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.IdentityKey, jwtUser)

		mockRepo.On("GetServiceByName", mock.Anything, serviceInput.Name).Return(nil, errors.New("not found")).Once()
		mockRepo.On("CreateService", mock.Anything, mock.AnythingOfType("*db.MonitoredService")).Return(nil).Once()
		mockPubSub.On("SendServiceCreatedMessage", mock.Anything, mock.AnythingOfType("db.MonitoredService")).Return(errors.New("pubsub error")).Once()

		controller.CreateMonitoredService(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to send service created message")
		mockRepo.AssertExpectations(t)
		mockPubSub.AssertExpectations(t)
	})
}

func TestUpdateMonitoredService(t *testing.T) {
	_, mockRepo, mockPubSub, _, controller := setupTestRouter()

	jwtUser := &middleware.JWTUser{ID: 1, Email: "test@user.com"}
	serviceID := "1"
	serviceInput := dto.MonitoredServiceRequest{
		Name:                "Updated Service",
		URL:                 "http://updated-example.com",
		Port:                8080,
		HealthCheckInterval: 120,
		AlertWindow:         600,
		AllowedResponseTime: 2000,
		FirstOncallerEmail:  "first@example.com",
	}
	existingService := &db.MonitoredService{Model: gorm.Model{ID: 1}, UserID: 1}

	t.Run("Success 200", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonValue, _ := json.Marshal(serviceInput)
		c.Request, _ = http.NewRequest(http.MethodPut, "/services/"+serviceID, bytes.NewBuffer(jsonValue))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: serviceID}}

		mockRepo.On("GetServiceByIDAndUserID", mock.Anything, uint64(1), uint64(jwtUser.ID)).Return(existingService, nil).Once()
		mockRepo.On("SaveService", mock.Anything, mock.AnythingOfType("*db.MonitoredService")).Return().Once()
		mockPubSub.On("SendServiceUpdatedMessage", mock.Anything, mock.AnythingOfType("db.MonitoredService")).Return(nil).Once()

		controller.UpdateMonitoredService(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Monitored service updated successfully")
		mockRepo.AssertExpectations(t)
		mockPubSub.AssertExpectations(t)
	})

	t.Run("Invalid Input 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodPut, "/services/"+serviceID, bytes.NewBufferString("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: serviceID}}

		controller.UpdateMonitoredService(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid input")
	})

	t.Run("Invalid Service ID 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonValue, _ := json.Marshal(serviceInput)
		c.Request, _ = http.NewRequest(http.MethodPut, "/services/abc", bytes.NewBuffer(jsonValue))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "abc"}}

		controller.UpdateMonitoredService(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid service ID")
	})

	t.Run("Service Not Found 404", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonValue, _ := json.Marshal(serviceInput)
		c.Request, _ = http.NewRequest(http.MethodPut, "/services/"+serviceID, bytes.NewBuffer(jsonValue))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: serviceID}}

		mockRepo.On("GetServiceByIDAndUserID", mock.Anything, uint64(1), uint64(jwtUser.ID)).Return(nil, errors.New("not found")).Once()

		controller.UpdateMonitoredService(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Monitored service not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed to send message 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonValue, _ := json.Marshal(serviceInput)
		c.Request, _ = http.NewRequest(http.MethodPost, "/services", bytes.NewBuffer(jsonValue))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set(middleware.IdentityKey, jwtUser)

		mockRepo.On("GetServiceByName", mock.Anything, serviceInput.Name).Return(nil, errors.New("not found")).Once()
		mockRepo.On("CreateService", mock.Anything, mock.AnythingOfType("*db.MonitoredService")).Return(nil).Once()
		mockPubSub.On("SendServiceCreatedMessage", mock.Anything, mock.AnythingOfType("db.MonitoredService")).Return(errors.New("pubsub error")).Once()

		controller.CreateMonitoredService(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to send service created message")
		mockRepo.AssertExpectations(t)
		mockPubSub.AssertExpectations(t)
	})
}

func TestDeleteMonitoredService(t *testing.T) {
	_, mockRepo, mockPubSub, _, controller := setupTestRouter()

	jwtUser := &middleware.JWTUser{ID: 1, Email: "test@user.com"}
	serviceID := "1"

	t.Run("Success 200", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodDelete, "/services/"+serviceID, nil)
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: serviceID}}

		mockRepo.On("DeleteServiceForUser", mock.Anything, uint64(1), uint64(jwtUser.ID)).Return(1, nil).Once()
		mockPubSub.On("SendServiceDeletedMessage", mock.Anything, uint64(1)).Return(nil).Once()

		controller.DeleteMonitoredService(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Monitored service deleted successfully")
		mockRepo.AssertExpectations(t)
		mockPubSub.AssertExpectations(t)
	})

	t.Run("Invalid Service ID 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodDelete, "/services/abc", nil)
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "abc"}}

		controller.DeleteMonitoredService(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid service ID")
	})

	t.Run("Service Not Found 404", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodDelete, "/services/"+serviceID, nil)
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: serviceID}}

		mockRepo.On("DeleteServiceForUser", mock.Anything, uint64(1), uint64(jwtUser.ID)).Return(0, nil).Once()

		controller.DeleteMonitoredService(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Monitored service not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("DB Error 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodDelete, "/services/"+serviceID, nil)
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: serviceID}}

		mockRepo.On("DeleteServiceForUser", mock.Anything, uint64(1), uint64(jwtUser.ID)).Return(0, errors.New("db error")).Once()

		controller.DeleteMonitoredService(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to delete monitored service")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed to send message 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodDelete, "/services/"+serviceID, nil)
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: serviceID}}

		mockRepo.On("DeleteServiceForUser", mock.Anything, uint64(1), uint64(jwtUser.ID)).Return(1, nil).Once()
		mockPubSub.On("SendServiceDeletedMessage", mock.Anything, uint64(1)).Return(errors.New("pubsub error")).Once()

		controller.DeleteMonitoredService(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to send service deleted message")
		mockRepo.AssertExpectations(t)
		mockPubSub.AssertExpectations(t)
	})
}

func TestGetServiceIncidents(t *testing.T) {
	_, mockRepo, _, mockLogRepo, controller := setupTestRouter()

	jwtUser := &middleware.JWTUser{ID: 1, Email: "test@user.com"}
	serviceID := "1"
	service := &db.MonitoredService{Model: gorm.Model{ID: 1}, UserID: 1}
	incidents := []firestore.IncidentLog{
		{IncidentID: "incident-1", ServiceID: 1, Type: "start", Timestamp: time.Now()},
		{IncidentID: "incident-1", ServiceID: 1, Type: "end", Timestamp: time.Now().Add(5 * time.Minute)},
	}

	t.Run("Success 200", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, "/services/"+serviceID+"/incidents", nil)
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: serviceID}}

		mockRepo.On("GetServiceByIDAndUserID", mock.Anything, uint64(1), uint64(jwtUser.ID)).Return(service, nil).Once()
		mockLogRepo.On("GetIncidentsByService", mock.Anything, service.ID).Return(incidents, nil).Once()

		controller.GetServiceIncidents(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var result []dto.IncidentDTO
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Len(t, result, 1)
		assert.Equal(t, "incident-1", result[0].ID)
		mockRepo.AssertExpectations(t)
		mockLogRepo.AssertExpectations(t)
	})

	t.Run("Invalid Service ID 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, "/services/abc/incidents", nil)
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "abc"}}

		controller.GetServiceIncidents(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid service ID")
	})

	t.Run("Service Not Found 404", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, "/services/"+serviceID+"/incidents", nil)
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: serviceID}}

		mockRepo.On("GetServiceByIDAndUserID", mock.Anything, uint64(1), uint64(jwtUser.ID)).Return(nil, errors.New("not found")).Once()

		controller.GetServiceIncidents(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Monitored service not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failed to retrieve incidents 500", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, "/services/"+serviceID+"/incidents", nil)
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: serviceID}}

		mockRepo.On("GetServiceByIDAndUserID", mock.Anything, uint64(1), uint64(jwtUser.ID)).Return(service, nil).Once()
		mockLogRepo.On("GetIncidentsByService", mock.Anything, service.ID).Return(nil, errors.New("firestore error")).Once()

		controller.GetServiceIncidents(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to retrieve incidents")
		mockRepo.AssertExpectations(t)
		mockLogRepo.AssertExpectations(t)
	})
}

func TestGetMyMonitoredServices(t *testing.T) {
	jwtUser := &middleware.JWTUser{ID: 1, Email: "test@user.com"}
	services := []db.MonitoredService{
		{Model: gorm.Model{ID: 1}, UserID: 1, Name: "Service 1"},
		{Model: gorm.Model{ID: 2}, UserID: 1, Name: "Service 2"},
		{Model: gorm.Model{ID: 3}, UserID: 1, Name: "Service 3"},
	}

	t.Run("Success 200", func(t *testing.T) {
		_, mockRepo, _, _, controller := setupTestRouter()
		s := setupRedis(t)
		defer s.Close()

		s.Set(redis.GetServiceStatusKey(1), "UP")
		s.Set(redis.GetServiceStatusKey(2), "DOWN")

		mockRepo.On("GetServicesForUser", mock.Anything, uint64(jwtUser.ID)).Return(services, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/services/me", nil)
		c.Set(middleware.IdentityKey, jwtUser)

		controller.GetMyMonitoredServices(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var dtos []dto.MonitoredServiceDTO
		json.Unmarshal(w.Body.Bytes(), &dtos)

		assert.Len(t, dtos, 3)
		assert.Equal(t, "UP", dtos[0].Status)
		assert.Equal(t, "DOWN", dtos[1].Status)
		assert.Equal(t, "UNKNOWN", dtos[2].Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("DB Error 500", func(t *testing.T) {
		_, mockRepo, _, _, controller := setupTestRouter()
		s := setupRedis(t)
		defer s.Close()

		mockRepo.On("GetServicesForUser", mock.Anything, uint64(jwtUser.ID)).Return(nil, errors.New("db error")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/services/me", nil)
		c.Set(middleware.IdentityKey, jwtUser)

		controller.GetMyMonitoredServices(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to retrieve monitored services")
		mockRepo.AssertExpectations(t)
	})
}

func TestGetMonitoredServiceByID(t *testing.T) {
	jwtUser := &middleware.JWTUser{ID: 1, Email: "test@user.com"}
	serviceID := "1"
	service := &db.MonitoredService{Model: gorm.Model{ID: 1}, UserID: 1, Name: "Test Service"}

	t.Run("Success 200 with status", func(t *testing.T) {
		_, mockRepo, _, _, controller := setupTestRouter()
		s := setupRedis(t)
		defer s.Close()

		s.Set(redis.GetServiceStatusKey(1), "UP")

		mockRepo.On("GetServiceByIDAndUserID", mock.Anything, uint64(1), uint64(jwtUser.ID)).Return(service, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/services/"+serviceID, nil)
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: serviceID}}

		controller.GetMonitoredServiceByID(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var result dto.MonitoredServiceDTO
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, "UP", result.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Success 200 without status", func(t *testing.T) {
		_, mockRepo, _, _, controller := setupTestRouter()
		s := setupRedis(t)
		defer s.Close()

		mockRepo.On("GetServiceByIDAndUserID", mock.Anything, uint64(1), uint64(jwtUser.ID)).Return(service, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/services/"+serviceID, nil)
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: serviceID}}

		controller.GetMonitoredServiceByID(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var result dto.MonitoredServiceDTO
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, "UNKNOWN", result.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid Service ID 400", func(t *testing.T) {
		_, _, _, _, controller := setupTestRouter()
		s := setupRedis(t)
		defer s.Close()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/services/abc", nil)
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "abc"}}

		controller.GetMonitoredServiceByID(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid service ID")
	})

	t.Run("Service Not Found 404", func(t *testing.T) {
		_, mockRepo, _, _, controller := setupTestRouter()
		s := setupRedis(t)
		defer s.Close()

		mockRepo.On("GetServiceByIDAndUserID", mock.Anything, uint64(1), uint64(jwtUser.ID)).Return(nil, errors.New("not found")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest(http.MethodGet, "/services/"+serviceID, nil)
		c.Set(middleware.IdentityKey, jwtUser)
		c.Params = gin.Params{gin.Param{Key: "id", Value: serviceID}}

		controller.GetMonitoredServiceByID(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "Monitored service not found")
		mockRepo.AssertExpectations(t)
	})
}
