package rpc

import (
	"alerting-platform/api/db"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

func TestGetAllServicesInfo(t *testing.T) {
	ctx := context.Background()
	empty := &emptypb.Empty{}

	t.Run("Success", func(t *testing.T) {
		mockRepo := new(db.MockRepository)
		server := &IncidentManagerServiceServer{repo: mockRepo}

		secondOncaller := "second@oncaller.com"
		services := []db.MonitoredService{
			{
				Model:               gorm.Model{ID: 1},
				AlertWindow:         300,
				AllowedResponseTime: 1000,
				FirstOncallerEmail:  "first@oncaller.com",
			},
			{
				Model:               gorm.Model{ID: 2},
				AlertWindow:         600,
				AllowedResponseTime: 500,
				FirstOncallerEmail:  "another@oncaller.com",
				SecondOncallerEmail: &secondOncaller,
			},
		}

		mockRepo.On("GetAllServices", ctx).Return(services, nil).Once()

		response, err := server.GetAllServicesInfo(ctx, empty)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Services, 2)

		assert.Equal(t, uint64(1), response.Services[0].ServiceId)
		assert.Equal(t, int64(300), response.Services[0].AlertWindow)
		assert.Equal(t, int64(1000), response.Services[0].AllowedResponseTime)
		assert.Equal(t, []string{"first@oncaller.com"}, response.Services[0].Oncallers)

		assert.Equal(t, uint64(2), response.Services[1].ServiceId)
		assert.Equal(t, int64(600), response.Services[1].AlertWindow)
		assert.Equal(t, int64(500), response.Services[1].AllowedResponseTime)
		assert.Equal(t, []string{"another@oncaller.com", "second@oncaller.com"}, response.Services[1].Oncallers)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Error from repository", func(t *testing.T) {
		mockRepo := new(db.MockRepository)
		server := &IncidentManagerServiceServer{repo: mockRepo}

		dbError := errors.New("database error")
		mockRepo.On("GetAllServices", ctx).Return(nil, dbError).Once()

		response, err := server.GetAllServicesInfo(ctx, empty)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		assert.Nil(t, response)

		mockRepo.AssertExpectations(t)
	})
}
