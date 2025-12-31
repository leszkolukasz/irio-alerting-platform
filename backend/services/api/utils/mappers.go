package utils

import (
	"alerting-platform/api/db"
	"alerting-platform/api/dto"
	"alerting-platform/common/db/firestore"
	"time"
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

func MapIncidentToDTO(logs []firestore.IncidentLog) dto.IncidentDTO {
	events := make([]dto.IncidentEventDTO, len(logs))
	for i, log := range logs {
		events[i] = dto.IncidentEventDTO{
			Timestamp: log.Timestamp.Format(time.RFC3339),
			Type:      log.Type,
			Oncaller:  log.Oncaller,
		}
	}

	return dto.IncidentDTO{
		ID:        logs[0].IncidentID,
		ServiceID: uint(logs[0].ServiceID),
		Events:    events,
	}
}
