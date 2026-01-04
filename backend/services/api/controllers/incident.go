package controllers

import (
	"log"
	"net/http"

	"alerting-platform/common/config"
	magic_link "alerting-platform/common/magic_link"

	"github.com/gin-gonic/gin"
)

func (controller *Controller) ResolveIncident(c *gin.Context) {
	tokenString := c.Param("token")
	if tokenString == "" {
		c.String(http.StatusBadRequest, "Missing token")
		return
	}

	secretKey := []byte(config.GetConfig().Secret)

	claims, err := magic_link.ParseToken(tokenString, secretKey)
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid or expired token")
		return
	}
	log.Printf("[DEBUG] Resolving incident %s for service %d by on-caller %s", claims.IncidentID, claims.ServiceID, claims.OnCaller)

	err = controller.PubSubService.SendOncallerAcknowledgedMessage(c, claims.IncidentID, claims.OnCaller)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to send on-caller acknowledged message", "error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Incident resolved successfully"})
}
