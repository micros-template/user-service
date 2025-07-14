package service

import (
	"context"

	"10.1.20.130/dropping/user-service/internal/domain/repository"
	"github.com/dropboks/event-bus-client/pkg/event"
	"github.com/dropboks/proto-user/pkg/upb"
	"github.com/dropboks/sharedlib/model"
	"github.com/dropboks/sharedlib/utils"
	"github.com/rs/zerolog"
)

type (
	AuthService interface {
		CreateUser(*upb.User) (*upb.Status, error)
		UpdateUser(c context.Context, user *upb.User) error
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
