package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dropboks/event-bus-client/pkg/event"
	"github.com/dropboks/proto-file/pkg/fpb"
	"github.com/dropboks/proto-user/pkg/upb"
	_dto "github.com/dropboks/sharedlib/dto"
	"github.com/dropboks/sharedlib/utils"
	"github.com/dropboks/user-service/internal/domain/dto"
	"github.com/dropboks/user-service/internal/domain/repository"
	_mq "github.com/dropboks/user-service/internal/infrastructure/message-queue"
	"github.com/dropboks/user-service/pkg/constant"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type (
	UserService interface {
		GetProfile(userId string) (dto.GetProfileResponse, error)
		UpdateUser(req *dto.UpdateUserRequest, userId string) error
		UpdateEmail(req *dto.UpdateEmailRequest, userId string) error
		UpdatePassword(req *dto.UpdatePasswordRequest, userId string) error
	}
	userService struct {
		userRepository     repository.UserRepository
		logger             zerolog.Logger
		fileServiceClient  fpb.FileServiceClient
		redisRepository    repository.RedisRepository
		notificationStream _mq.Nats
		eventEmitter       event.Emitter
	}
)

func NewUserService(userRepo repository.UserRepository,
	logger zerolog.Logger,
	fileServiceClient fpb.FileServiceClient,
	redisRepository repository.RedisRepository,
	notificationStream _mq.Nats,
	eventEmitter event.Emitter,
) UserService {
	return &userService{
		userRepository:     userRepo,
		logger:             logger,
		fileServiceClient:  fileServiceClient,
		redisRepository:    redisRepository,
		notificationStream: notificationStream,
		eventEmitter:       eventEmitter,
	}
}

func (u *userService) UpdatePassword(req *dto.UpdatePasswordRequest, userId string) error {
	if req.NewPassword != req.ConfirmNewPassword {
		return dto.Err_BAD_REQUEST_PASSWORD_CONFIRM_PASSWORD_DOESNT_MATCH
	}
	user, err := u.userRepository.QueryUserByUserId(userId)
	if err != nil {
		return err
	}
	ok := utils.HashPasswordCompare(req.Password, user.Password)
	if !ok {
		return dto.Err_UNAUTHORIZED_PASSWORD_WRONG
	}
	newPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	us := *user
	us.Password = newPassword
	if err := u.userRepository.UpdateUser(&us); err != nil {
		return err
	}
	go func() {
		u.eventEmitter.UpdateUser(context.Background(), &upb.User{
			Id:               us.ID,
			FullName:         us.FullName,
			Image:            us.Image,
			Email:            us.Email,
			Password:         us.Password,
			Verified:         us.Verified,
			TwoFactorEnabled: us.TwoFactorEnabled,
		})
	}()
	return nil
}

func (u *userService) UpdateEmail(req *dto.UpdateEmailRequest, userId string) error {
	ctx := context.Background()

	verificationToken, err := utils.RandomString64()
	if err != nil {
		u.logger.Error().Err(err).Msg("error generate verification token")
		return dto.Err_INTERNAL_GENERATE_TOKEN
	}

	savedEmail := fmt.Sprintf("newEmail:%s", userId)
	if err := u.redisRepository.SetResource(ctx, savedEmail, req.Email, 30*time.Minute); err != nil {
		return err
	}

	key := fmt.Sprintf("changeEmailToken:%s", userId)
	if err := u.redisRepository.SetResource(ctx, key, verificationToken, 30*time.Minute); err != nil {
		return err
	}

	link := fmt.Sprintf("%s/%suserid=%s&changeEmailToken=%s", viper.GetString("app.auth_url"), viper.GetString("app.verification_url"), userId, verificationToken)
	subject := fmt.Sprintf("%s.%s", viper.GetString("jetstream.notification.subject.mail"), userId)
	msg := &_dto.MailNotificationMessage{
		Receiver: []string{req.Email},
		MsgType:  "changeEmail",
		Message:  link,
	}
	marshalledMsg, err := json.Marshal(msg)
	if err != nil {
		u.logger.Error().Err(err).Msg("marshal data error")
		return err
	}
	_, err = u.notificationStream.Publish(ctx, subject, []byte(marshalledMsg))
	if err != nil {
		u.logger.Error().Err(err).Msg("publish notification error")
		return dto.Err_INTERNAL_PUBLISH_MESSAGE
	}
	return nil
}

func (u *userService) UpdateUser(req *dto.UpdateUserRequest, userId string) error {
	user, err := u.userRepository.QueryUserByUserId(userId)
	if err != nil {
		return err
	}
	us := *user
	trimmedName := strings.TrimSpace(req.FullName)
	if trimmedName != user.FullName {
		us.FullName = trimmedName
	}
	if req.TwoFactorEnabled != user.TwoFactorEnabled {
		us.TwoFactorEnabled = req.TwoFactorEnabled
	}
	ctx := context.Background()
	if req.Image != nil && req.Image.Filename != "" {
		ext := utils.GetFileNameExtension(req.Image.Filename)
		if ext != "jpg" && ext != "jpeg" && ext != "png" {
			return dto.Err_BAD_REQUEST_WRONG_EXTENSION
		}
		if req.Image.Size > constant.MAX_UPLOAD_SIZE {
			return dto.Err_BAD_REQUEST_LIMIT_SIZE_EXCEEDED
		}
		image, err := utils.FileToByte(req.Image)
		if err != nil {
			u.logger.Error().Err(err).Msg("error converting image")
			return dto.Err_INTERNAL_CONVERT_IMAGE
		}
		imageReq := &fpb.Image{
			Image: image,
			Ext:   ext,
		}
		resp, err := u.fileServiceClient.SaveProfileImage(ctx, imageReq)
		if err != nil {
			u.logger.Error().Err(err).Msg("Error uploading image to file service")
			return err
		}
		us.Image = utils.StringPtr(resp.GetName())
	}
	err = u.userRepository.UpdateUser(&us)
	if err == nil && req.Image != nil && req.Image.Filename != "" {
		_, err := u.fileServiceClient.RemoveProfileImage(ctx, &fpb.ImageName{Name: *user.Image})
		return err
	} else if err != nil && req.Image != nil && req.Image.Filename != "" {
		_, err := u.fileServiceClient.RemoveProfileImage(ctx, &fpb.ImageName{Name: *us.Image})
		return err
	}
	// push event bus in goroutine
	go func() {
		u.eventEmitter.UpdateUser(context.Background(), &upb.User{
			Id:               us.ID,
			FullName:         us.FullName,
			Image:            us.Image,
			Email:            us.Email,
			Password:         us.Password,
			Verified:         us.Verified,
			TwoFactorEnabled: us.TwoFactorEnabled,
		})
	}()
	return nil
}

func (u *userService) GetProfile(userId string) (dto.GetProfileResponse, error) {
	user, err := u.userRepository.QueryUserByUserId(userId)
	if err != nil {
		return dto.GetProfileResponse{}, err
	}
	profile := dto.GetProfileResponse{
		FullName:         user.FullName,
		Image:            user.Image,
		Email:            user.Email,
		Verified:         user.Verified,
		TwoFactorEnabled: user.TwoFactorEnabled,
	}
	return profile, nil
}
