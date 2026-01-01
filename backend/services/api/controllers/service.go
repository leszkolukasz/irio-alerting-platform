package controllers

import (
	"alerting-platform/api/redis"
	db_common "alerting-platform/common/db"
	"alerting-platform/common/db/firestore"
	"strconv"
	"time"

	"alerting-platform/api/db"
	"alerting-platform/api/dto"
	"alerting-platform/api/middleware"
	"alerting-platform/api/utils"

	"github.com/gin-gonic/gin"
)

func (controller *Controller) CreateMonitoredService(c *gin.Context) {
	var serviceInput dto.MonitoredServiceRequest

	if err := c.ShouldBind(&serviceInput); err != nil {
		c.JSON(400, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	userIdentity, exists := c.Get(middleware.IdentityKey)
	if !exists {
		c.JSON(500, gin.H{"message": "Failed to get user from context"})
		return
	}

	jwtUser := userIdentity.(*middleware.JWTUser)

	ctx := c.Request.Context()

	_, err := controller.Repository.GetServiceByName(ctx, serviceInput.Name)
	if err == nil {
		c.JSON(400, gin.H{"message": "Service name already taken"})
		return
	}

	service := db.MonitoredService{
		UserID:              jwtUser.ID,
		Name:                serviceInput.Name,
		URL:                 serviceInput.URL,
		Port:                serviceInput.Port,
		HealthCheckInterval: serviceInput.HealthCheckInterval,
		AlertWindow:         serviceInput.AlertWindow,
		AllowedResponseTime: serviceInput.AllowedResponseTime,
		FirstOncallerEmail:  serviceInput.FirstOncallerEmail,
		SecondOncallerEmail: serviceInput.SecondOncallerEmail,
	}

	err = controller.Repository.CreateService(ctx, &service)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to create monitored service", "error": err.Error()})
		return
	}

	err = controller.PubSubService.SendServiceCreatedMessage(ctx, service)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to send service created message", "error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Monitored service created successfully", "serviceID": service.ID})
}

func (controller *Controller) GetMyMonitoredServices(c *gin.Context) {
	userIdentity, exists := c.Get(middleware.IdentityKey)
	if !exists {
		c.JSON(500, gin.H{"message": "Failed to get user from context"})
		return
	}

	jwtUser := userIdentity.(*middleware.JWTUser)

	ctx := c.Request.Context()

	services, err := controller.Repository.GetServicesForUser(ctx, uint64(jwtUser.ID))
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to retrieve monitored services", "error": err.Error()})
		return
	}

	dtos := make([]dto.MonitoredServiceDTO, 0, len(services))

	redisClient := db_common.GetRedisClient()

	for _, s := range services {
		status, err := redisClient.Get(ctx, redis.GetServiceStatusKey(uint64(s.ID))).Result()
		if err != nil {
			status = "UNKNOWN"
		}

		dto := utils.MapServiceToDTO(s, status)
		dtos = append(dtos, dto)
	}

	c.JSON(200, dtos)
}

func (controller *Controller) GetMonitoredServiceByID(c *gin.Context) {
	serviceID := c.Param("id")

	userIdentity, exists := c.Get(middleware.IdentityKey)
	if !exists {
		c.JSON(500, gin.H{"message": "Failed to get user from context"})
		return
	}

	jwtUser := userIdentity.(*middleware.JWTUser)
	ctx := c.Request.Context()

	serviceIDInt, err := strconv.ParseUint(serviceID, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid service ID", "error": err.Error()})
		return
	}

	service, err := controller.Repository.GetServiceByIDAndUserID(ctx, serviceIDInt, uint64(jwtUser.ID))
	if err != nil {
		c.JSON(404, gin.H{"message": "Monitored service not found", "error": err.Error()})
		return
	}

	redisClient := db_common.GetRedisClient()
	status, err := redisClient.Get(ctx, redis.GetServiceStatusKey(uint64(service.ID))).Result()

	if err != nil {
		status = "UNKNOWN"
	}

	dto := utils.MapServiceToDTO(*service, status)

	c.JSON(200, dto)
}

func (controller *Controller) UpdateMonitoredService(c *gin.Context) {
	serviceID := c.Param("id")

	var serviceInput dto.MonitoredServiceRequest
	if err := c.ShouldBind(&serviceInput); err != nil {
		c.JSON(400, gin.H{"message": "Invalid input", "error": err.Error()})
		return
	}

	userIdentity, exists := c.Get(middleware.IdentityKey)
	if !exists {
		c.JSON(500, gin.H{"message": "Failed to get user from context"})
		return
	}

	jwtUser := userIdentity.(*middleware.JWTUser)
	ctx := c.Request.Context()

	serviceIDInt, err := strconv.ParseUint(serviceID, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid service ID", "error": err.Error()})
		return
	}

	service, err := controller.Repository.GetServiceByIDAndUserID(ctx, serviceIDInt, uint64(jwtUser.ID))
	if err != nil {
		c.JSON(404, gin.H{"message": "Monitored service not found", "error": err.Error()})
		return
	}

	service.Name = serviceInput.Name
	service.URL = serviceInput.URL
	service.Port = serviceInput.Port
	service.HealthCheckInterval = serviceInput.HealthCheckInterval
	service.AlertWindow = serviceInput.AlertWindow
	service.AllowedResponseTime = serviceInput.AllowedResponseTime
	service.FirstOncallerEmail = serviceInput.FirstOncallerEmail
	service.SecondOncallerEmail = serviceInput.SecondOncallerEmail

	controller.Repository.SaveService(ctx, service)

	err = controller.PubSubService.SendServiceUpdatedMessage(ctx, *service)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to send service updated message", "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Monitored service updated successfully"})
}

func (controller *Controller) DeleteMonitoredService(c *gin.Context) {
	serviceID := c.Param("id")

	userIdentity, exists := c.Get(middleware.IdentityKey)
	if !exists {
		c.JSON(500, gin.H{"message": "Failed to get user from context"})
		return
	}

	jwtUser := userIdentity.(*middleware.JWTUser)
	ctx := c.Request.Context()

	serviceIDInt, err := strconv.ParseUint(serviceID, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid service ID", "error": err.Error()})
		return
	}

	rowsAffected, err := controller.Repository.DeleteServiceForUser(ctx, serviceIDInt, uint64(jwtUser.ID))
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to delete monitored service", "error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(404, gin.H{"message": "Monitored service not found"})
		return
	}

	err = controller.PubSubService.SendServiceDeletedMessage(ctx, serviceIDInt)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to send service deleted message", "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Monitored service deleted successfully"})
}

func (controller *Controller) GetServiceIncidents(c *gin.Context) {
	serviceID := c.Param("id")

	userIdentity, exists := c.Get(middleware.IdentityKey)
	if !exists {
		c.JSON(500, gin.H{"message": "Failed to get user from context"})
		return
	}

	jwtUser := userIdentity.(*middleware.JWTUser)
	ctx := c.Request.Context()

	serviceIDInt, err := strconv.ParseUint(serviceID, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid service ID", "error": err.Error()})
		return
	}

	service, err := controller.Repository.GetServiceByIDAndUserID(ctx, serviceIDInt, uint64(jwtUser.ID))
	if err != nil {
		c.JSON(404, gin.H{"message": "Monitored service not found", "error": err.Error()})
		return
	}

	incidents, err := controller.LogRepository.GetIncidentsByService(ctx, service.ID)

	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to retrieve incidents", "error": err.Error()})
		return
	}

	incidentMap := make(map[string][]firestore.IncidentLog)
	for _, incident := range incidents {
		incidentMap[incident.IncidentID] = append(incidentMap[incident.IncidentID], incident)
	}

	incidentDTOs := make([]dto.IncidentDTO, 0, len(incidentMap))
	for _, logs := range incidentMap {
		incidentDTO := utils.MapIncidentToDTO(logs)
		incidentDTOs = append(incidentDTOs, incidentDTO)
	}

	c.JSON(200, incidentDTOs)
}

var granularities = map[string]time.Duration{
	"hour":  time.Hour,
	"day":   24 * time.Hour,
	"week":  7 * 24 * time.Hour,
	"month": 30 * 24 * time.Hour,
}

func (controller *Controller) GetServiceStatusMetrics(c *gin.Context) {
	serviceID := c.Param("id")
	granularity := c.Query("granularity")

	userIdentity, exists := c.Get(middleware.IdentityKey)
	if !exists {
		c.JSON(500, gin.H{"message": "Failed to get user from context"})
		return
	}

	jwtUser := userIdentity.(*middleware.JWTUser)
	ctx := c.Request.Context()

	serviceIDInt, err := strconv.ParseUint(serviceID, 10, 64)
	if err != nil {
		c.JSON(400, gin.H{"message": "Invalid service ID", "error": err.Error()})
		return
	}

	service, err := controller.Repository.GetServiceByIDAndUserID(ctx, serviceIDInt, uint64(jwtUser.ID))
	if err != nil {
		c.JSON(404, gin.H{"message": "Monitored service not found", "error": err.Error()})
		return
	}

	var duration time.Duration
	if val, ok := granularities[granularity]; ok {
		duration = val
	} else {
		c.JSON(400, gin.H{"message": "Invalid granularity"})
		return
	}

	startTime := time.Now().UTC().Add(-duration)
	metrics, err := controller.LogRepository.GetMetricsByServiceAndAfterTime(ctx, service.ID, startTime)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to retrieve status metrics", "error": err.Error()})
		return
	}

	metricsDTO := aggregateMetrics(metrics, startTime)

	response := dto.StatusMetrics{
		Granularity: granularity,
		Data:        metricsDTO,
	}

	c.JSON(200, response)
}

func aggregateMetrics(metrics []firestore.MetricLog, startTime time.Time) []dto.StatusDataPoint {
	const binCount = 50

	duration := time.Since(startTime)
	binDuration := duration / binCount
	bins := make([]dto.StatusDataPoint, binCount)

	for i := 0; i < binCount; i++ {
		binTime := startTime.Add(time.Duration(i) * binDuration)
		bins[i] = dto.StatusDataPoint{
			Timestamp: binTime.Format(time.RFC3339),
			Success:   0,
			Total:     0,
		}
	}

	for _, metric := range metrics {
		binIndex := int(metric.Timestamp.Sub(startTime) / binDuration)
		if binIndex >= 0 && binIndex < binCount {
			bins[binIndex].Total++
			if metric.Type == "UP" {
				bins[binIndex].Success++
			}
		}
	}

	return bins
}
