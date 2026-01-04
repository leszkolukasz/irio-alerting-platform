package controllers

import (
	"alerting-platform/common/config"
	"alerting-platform/common/magic_link"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	os.Setenv("REDIS_PREFIX", "test")
}

func TestResolveIncident(t *testing.T) {
	_, _, mockPubSub, _, controller := setupTestRouter()

	testSecret := "test-secret-key-123"
	config.GetConfig().Secret = testSecret

	incidentID := "INC-123"
	serviceID := uint64(99)
	email := "oncaller@example.com"

	t.Run("Success 200", func(t *testing.T) {
		validToken, _ := magic_link.GenerateToken(incidentID, serviceID, email, []byte(testSecret))

		mockPubSub.On("SendIncidentResolvedMessage", mock.Anything, incidentID, email).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, "/public/incidents/resolve/"+validToken, nil)
		c.Params = gin.Params{gin.Param{Key: "token", Value: validToken}}

		controller.ResolveIncident(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "Incident resolved successfully")
		mockPubSub.AssertExpectations(t)
	})

	t.Run("Missing Token 400", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, "/public/incidents/resolve/", nil)
		c.Params = gin.Params{gin.Param{Key: "token", Value: ""}}

		controller.ResolveIncident(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Missing token")
		mockPubSub.AssertNotCalled(t, "SendIncidentResolvedMessage")
	})

	t.Run("Invalid Token 401", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		invalidToken := "invalid.token.string"
		c.Request, _ = http.NewRequest(http.MethodGet, "/public/incidents/resolve/"+invalidToken, nil)
		c.Params = gin.Params{gin.Param{Key: "token", Value: invalidToken}}

		controller.ResolveIncident(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Link jest nieprawidłowy lub wygasł")
		mockPubSub.AssertNotCalled(t, "SendIncidentResolvedMessage")
	})

	t.Run("Wrong Secret 401", func(t *testing.T) {
		wrongSecret := []byte("wrong-secret")
		forgedToken, _ := magic_link.GenerateToken(incidentID, serviceID, email, wrongSecret)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, "/public/incidents/resolve/"+forgedToken, nil)
		c.Params = gin.Params{gin.Param{Key: "token", Value: forgedToken}}

		controller.ResolveIncident(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockPubSub.AssertNotCalled(t, "SendIncidentResolvedMessage")
	})

	t.Run("PubSub Error 500", func(t *testing.T) {
		validToken, _ := magic_link.GenerateToken(incidentID, serviceID, email, []byte(testSecret))

		mockPubSub.On("SendIncidentResolvedMessage", mock.Anything, incidentID, email).Return(errors.New("pubsub connection failed")).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, "/public/incidents/resolve/"+validToken, nil)
		c.Params = gin.Params{gin.Param{Key: "token", Value: validToken}}

		controller.ResolveIncident(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to send service deleted message")
		mockPubSub.AssertExpectations(t)
	})
}
