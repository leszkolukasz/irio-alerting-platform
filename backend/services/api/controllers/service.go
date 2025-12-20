package controllers

import (
	"alerting-platform/api/db"
	"alerting-platform/api/dto"
	"alerting-platform/api/middleware"
	"alerting-platform/api/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateMonitoredService(c *gin.Context) {
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
	conn := db.GetDBConnection()

	_, err := gorm.G[db.MonitoredService](conn).Where("name = ?", serviceInput.Name).First(ctx)
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

	err = gorm.G[db.MonitoredService](conn).Create(ctx, &service)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to create monitored service", "error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "Monitored service created successfully", "serviceID": service.ID})
}

func GetMyMonitoredServices(c *gin.Context) {
	userIdentity, exists := c.Get(middleware.IdentityKey)
	if !exists {
		c.JSON(500, gin.H{"message": "Failed to get user from context"})
		return
	}

	jwtUser := userIdentity.(*middleware.JWTUser)

	ctx := c.Request.Context()
	conn := db.GetDBConnection()

	services, err := gorm.G[db.MonitoredService](conn).Where("user_id = ?", jwtUser.ID).Find(ctx)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to retrieve monitored services", "error": err.Error()})
		return
	}

	dtos := make([]dto.MonitoredServiceDTO, 0, len(services))

	for _, s := range services {
		dto := utils.MapServiceToDTO(s)
		dtos = append(dtos, dto)
	}

	c.JSON(200, dtos)
}

func GetMonitoredServiceByID(c *gin.Context) {
	serviceID := c.Param("id")

	userIdentity, exists := c.Get(middleware.IdentityKey)
	if !exists {
		c.JSON(500, gin.H{"message": "Failed to get user from context"})
		return
	}

	jwtUser := userIdentity.(*middleware.JWTUser)

	ctx := c.Request.Context()
	conn := db.GetDBConnection()

	service, err := gorm.G[db.MonitoredService](conn).Where("id = ? AND user_id = ?", serviceID, jwtUser.ID).First(ctx)
	if err != nil {
		c.JSON(404, gin.H{"message": "Monitored service not found", "error": err.Error()})
		return
	}

	dto := utils.MapServiceToDTO(service)

	c.JSON(200, dto)
}

func UpdateMonitoredService(c *gin.Context) {
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
	conn := db.GetDBConnection()

	service, err := gorm.G[db.MonitoredService](conn).Where("id = ? AND user_id = ?", serviceID, jwtUser.ID).First(ctx)
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

	conn.Save(&service)

	c.JSON(200, gin.H{"message": "Monitored service updated successfully"})
}

func DeleteMonitoredService(c *gin.Context) {
	serviceID := c.Param("id")

	userIdentity, exists := c.Get(middleware.IdentityKey)
	if !exists {
		c.JSON(500, gin.H{"message": "Failed to get user from context"})
		return
	}

	jwtUser := userIdentity.(*middleware.JWTUser)

	ctx := c.Request.Context()
	conn := db.GetDBConnection()

	rowsAffected, err := gorm.G[db.MonitoredService](conn).Where("id = ? AND user_id = ?", serviceID, jwtUser.ID).Delete(ctx)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to delete monitored service", "error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(404, gin.H{"message": "Monitored service not found"})
		return
	}

	c.JSON(200, gin.H{"message": "Monitored service deleted successfully"})
}
