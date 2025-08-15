package service

import (
	"context"

	"github.com/micros-template/user-service/internal/domain/repository"

	"github.com/micros-template/event-bus-client/pkg/event"

	"github.com/micros-template/proto-user/pkg/upb"
	"github.com/micros-template/sharedlib/model"
	"github.com/micros-template/sharedlib/utils"
	"github.com/rs/zerolog"
)

type (
	AuthService interface {
		CreateUser(*upb.User) (*upb.Status, error)
		UpdateUser(c context.Context, user *upb.User) error
		DeleteUser(c context.Context, userId *upb.UserId) error
	}
	authService struct {
		userRepository repository.UserRepository
		logger         zerolog.Logger
		eventEmitter   event.Emitter
	}
)

func NewAuthService(userRepository repository.UserRepository, emitter event.Emitter, logger zerolog.Logger) AuthService {
	return &authService{
		userRepository: userRepository,
		logger:         logger,
		eventEmitter:   emitter,
	}
}

func (a *authService) UpdateUser(c context.Context, user *upb.User) error {
	u := &model.User{
		ID:               user.GetId(),
		FullName:         user.GetFullName(),
		Image:            utils.StringPtr(user.GetImage()),
		Email:            user.GetEmail(),
		Password:         user.GetPassword(),
		Verified:         user.GetVerified(),
		TwoFactorEnabled: user.GetTwoFactorEnabled(),
	}
	if err := a.userRepository.UpdateUser(u); err != nil {
		return err
	}
	// push event bus in goroutine
	go func() {
		a.eventEmitter.UpdateUser(context.Background(), &upb.User{
			Id:               u.ID,
			FullName:         u.FullName,
			Image:            u.Image,
			Email:            u.Email,
			Password:         u.Password,
			Verified:         u.Verified,
			TwoFactorEnabled: u.TwoFactorEnabled,
		})
	}()
	return nil
}

func (a *authService) CreateUser(user *upb.User) (*upb.Status, error) {
	u := &model.User{
		ID:               user.GetId(),
		FullName:         user.GetFullName(),
		Image:            utils.StringPtr(user.GetImage()),
		Email:            user.GetEmail(),
		Password:         user.GetPassword(),
		Verified:         user.GetVerified(),
		TwoFactorEnabled: user.GetTwoFactorEnabled(),
	}
	err := a.userRepository.CreateNewUser(u)
	if err != nil {
		return nil, err
	}
	// push event bus in goroutine
	go func() {
		a.eventEmitter.InsertUser(context.Background(), &upb.User{
			Id:               u.ID,
			FullName:         u.FullName,
			Image:            u.Image,
			Email:            u.Email,
			Password:         u.Password,
			Verified:         u.Verified,
			TwoFactorEnabled: u.TwoFactorEnabled,
		})
	}()
	return &upb.Status{Success: true}, nil
}

func (a *authService) DeleteUser(c context.Context, userId *upb.UserId) error {
	if err := a.userRepository.DeleteUser(userId.GetUserId()); err != nil {
		return err
	}
	// push event bus in goroutine
	go func() {
		a.eventEmitter.DeleteUser(context.Background(), userId)
	}()
	return nil
}
