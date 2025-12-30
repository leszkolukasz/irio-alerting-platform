package utils

import (
	"alerting-platform/api/db"
	"alerting-platform/api/dto"
)

func MapServiceToDTO(service db.MonitoredService, status string) dto.MonitoredServiceDTO {
	return dto.MonitoredServiceDTO{
		ID:                  service.ID,
		Name:                service.Name,
		URL:                 service.URL,
		Port:                service.Port,
		HealthCheckInterval: service.HealthCheckInterval,
		AlertWindow:         service.AlertWindow,
		AllowedResponseTime: service.AllowedResponseTime,
		FirstOncallerEmail:  service.FirstOncallerEmail,
		SecondOncallerEmail: service.SecondOncallerEmail,
		Status:              status,
	}
}
