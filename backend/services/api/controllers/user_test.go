package controllers

import (
	"alerting-platform/api/db"
	"alerting-platform/api/dto"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegisterUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success 201", func(t *testing.T) {
		mockRepo := new(MockRepository)
		controller := &Controller{
			Repository: mockRepo,
		}

		mockRepo.On("CreateUser", mock.Anything, mock.MatchedBy(func(u *db.User) bool {
			return u.Email == "test@example.com" && u.PasswordHash != ""
		})).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := dto.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonValue, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonValue))
		c.Request.Header.Set("Content-Type", "application/json")

		controller.RegisterUser(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "User registered successfully")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid Input 400", func(t *testing.T) {
		mockRepo := new(MockRepository)
		controller := &Controller{
			Repository: mockRepo,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("invalid-json"))
		c.Request.Header.Set("Content-Type", "application/json")

		controller.RegisterUser(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid input")
		mockRepo.AssertNotCalled(t, "CreateUser")
	})

	t.Run("Database Error 500", func(t *testing.T) {
		mockRepo := new(MockRepository)
		controller := &Controller{
			Repository: mockRepo,
		}

		mockRepo.On("CreateUser", mock.Anything, mock.Anything).Return(errors.New("db error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := dto.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		jsonValue, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonValue))
		c.Request.Header.Set("Content-Type", "application/json")
		controller.RegisterUser(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Failed to create user")
		mockRepo.AssertExpectations(t)
	})
}
