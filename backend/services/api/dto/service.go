package dto

type MonitoredServiceRequest struct {
	Name                string  `json:"name" binding:"required"`
	URL                 string  `json:"url" binding:"required,url"`
	Port                int     `json:"port" binding:"required,min=1,max=65535"`
	HealthCheckInterval int     `json:"healthCheckInterval" binding:"required,min=1"`
	AlertWindow         int     `json:"alertWindow" binding:"required,min=1"`
	AllowedResponseTime int     `json:"allowedResponseTime" binding:"required,min=1"`
	FirstOncallerEmail  string  `json:"firstOncallerEmail" binding:"required,email"`
	SecondOncallerEmail *string `json:"secondOncallerEmail" binding:"omitempty,email"`
}

type MonitoredServiceDTO struct {
	ID                  uint    `json:"id"`
	Name                string  `json:"name"`
	URL                 string  `json:"url"`
	Port                int     `json:"port"`
	HealthCheckInterval int     `json:"healthCheckInterval"`
	AlertWindow         int     `json:"alertWindow"`
	AllowedResponseTime int     `json:"allowedResponseTime"`
	FirstOncallerEmail  string  `json:"firstOncallerEmail"`
	SecondOncallerEmail *string `json:"secondOncallerEmail"`
	Status              string  `json:"status"`
}
