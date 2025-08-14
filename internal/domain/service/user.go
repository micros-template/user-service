package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"10.1.20.130/dropping/event-bus-client/pkg/event"
	"10.1.20.130/dropping/log-management/pkg"
	"10.1.20.130/dropping/proto-file/pkg/fpb"
	"10.1.20.130/dropping/proto-user/pkg/upb"
	_dto "10.1.20.130/dropping/sharedlib/dto"
	"10.1.20.130/dropping/sharedlib/utils"
	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/repository"
	_mq "10.1.20.130/dropping/user-service/internal/infrastructure/message-queue"
	"10.1.20.130/dropping/user-service/pkg/constant"
	u "10.1.20.130/dropping/user-service/pkg/utils"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type (
	UserService interface {
		GetProfile(userId string) (dto.GetProfileResponse, error)
		UpdateUser(req *dto.UpdateUserRequest, userId string) error
		UpdateEmail(req *dto.UpdateEmailRequest, userId string) error
		UpdatePassword(req *dto.UpdatePasswordRequest, userId string) error
		DeleteUser(req *dto.DeleteUserRequest, userId string) error
	}
	userService struct {
		userRepository     repository.UserRepository
		logger             zerolog.Logger
		fileServiceClient  fpb.FileServiceClient
		redisRepository    repository.RedisRepository
		notificationStream _mq.Nats
		eventEmitter       event.Emitter
		logEmitter         pkg.LogEmitter
		util               u.LoggerServiceUtil
	}
)

func NewUserService(userRepo repository.UserRepository,
	logger zerolog.Logger,
	fileServiceClient fpb.FileServiceClient,
	redisRepository repository.RedisRepository,
	notificationStream _mq.Nats,
	eventEmitter event.Emitter,
	logEmitter pkg.LogEmitter,
	util u.LoggerServiceUtil,
) UserService {
	return &userService{
		userRepository:     userRepo,
		logger:             logger,
		fileServiceClient:  fileServiceClient,
		redisRepository:    redisRepository,
		notificationStream: notificationStream,
		eventEmitter:       eventEmitter,
		logEmitter:         logEmitter,
		util:               util,
	}
}

func (u *userService) DeleteUser(req *dto.DeleteUserRequest, userId string) error {
	//  [IMPROVE] -> send email for delete verification
	user, err := u.userRepository.QueryUserByUserId(userId)
	if err != nil {
		return err
	}
	ok := utils.HashPasswordCompare(req.Password, user.Password)
	if !ok {
		go func() {
			if err := u.util.EmitLog("ERR", dto.Err_UNAUTHORIZED_PASSWORD_WRONG.Error()); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return dto.Err_UNAUTHORIZED_PASSWORD_WRONG
	}
	if err := u.userRepository.DeleteUser(userId); err != nil {
		return err
	}
	go func() {
		u.eventEmitter.DeleteUser(context.Background(), &upb.UserId{
			UserId: userId,
		})
	}()
	return nil
}

func (u *userService) UpdatePassword(req *dto.UpdatePasswordRequest, userId string) error {
	if req.NewPassword != req.ConfirmNewPassword {
		go func() {
			if err := u.util.EmitLog("ERR", dto.Err_BAD_REQUEST_PASSWORD_CONFIRM_PASSWORD_DOESNT_MATCH.Error()); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return dto.Err_BAD_REQUEST_PASSWORD_CONFIRM_PASSWORD_DOESNT_MATCH
	}
	user, err := u.userRepository.QueryUserByUserId(userId)
	if err != nil {
		return err
	}
	ok := utils.HashPasswordCompare(req.Password, user.Password)
	if !ok {
		go func() {
			if err := u.util.EmitLog("ERR", dto.Err_UNAUTHORIZED_PASSWORD_WRONG.Error()); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return dto.Err_UNAUTHORIZED_PASSWORD_WRONG
	}
	newPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		go func() {
			if err := u.util.EmitLog("ERR", fmt.Sprintf("Hash Password Error: %v", err.Error())); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
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
		go func() {
			if err := u.util.EmitLog("ERR", dto.Err_INTERNAL_GENERATE_TOKEN.Error()); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
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
		go func() {
			if err := u.util.EmitLog("ERR", "marshal data error"); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		return err
	}
	_, err = u.notificationStream.Publish(ctx, subject, []byte(marshalledMsg))
	if err != nil {
		go func() {
			if err := u.util.EmitLog("ERR", dto.Err_INTERNAL_PUBLISH_MESSAGE.Error()); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
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
			go func() {
				if err := u.util.EmitLog("ERR", dto.Err_BAD_REQUEST_WRONG_EXTENSION.Error()); err != nil {
					u.logger.Error().Err(err).Msg("failed to emit log")
				}
			}()
			return dto.Err_BAD_REQUEST_WRONG_EXTENSION
		}
		if req.Image.Size > constant.MAX_UPLOAD_SIZE {
			go func() {
				if err := u.util.EmitLog("ERR", dto.Err_BAD_REQUEST_LIMIT_SIZE_EXCEEDED.Error()); err != nil {
					u.logger.Error().Err(err).Msg("failed to emit log")
				}
			}()
			return dto.Err_BAD_REQUEST_LIMIT_SIZE_EXCEEDED
		}
		image, err := utils.FileToByte(req.Image)
		if err != nil {
			go func() {
				if err := u.util.EmitLog("ERR", "error converting image to byte"); err != nil {
					u.logger.Error().Err(err).Msg("failed to emit log")
				}
			}()
			return dto.Err_INTERNAL_CONVERT_IMAGE
		}
		imageReq := &fpb.Image{
			Image: image,
			Ext:   ext,
		}
		resp, err := u.fileServiceClient.SaveProfileImage(ctx, imageReq)
		if err != nil {
			go func() {
				if err := u.util.EmitLog("ERR", fmt.Sprintf("Error uploading image to file service. err: %v", err.Error())); err != nil {
					u.logger.Error().Err(err).Msg("failed to emit log")
				}
			}()
			return err
		}
		us.Image = utils.StringPtr(resp.GetName())
	}
	err = u.userRepository.UpdateUser(&us)
	if err == nil && req.Image != nil && req.Image.Filename != "" {
		if _, err := u.fileServiceClient.RemoveProfileImage(ctx, &fpb.ImageName{Name: *user.Image}); err != nil {
			go func() {
				if err := u.util.EmitLog("ERR", fmt.Sprintf("Error remove image via file service. err: %v", err.Error())); err != nil {
					u.logger.Error().Err(err).Msg("failed to emit log")
				}
			}()
		}
	} else if err != nil {
		go func() {
			if err := u.util.EmitLog("ERR", fmt.Sprintf("update user failed. err: %v", err.Error())); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		if req.Image != nil && req.Image.Filename != "" {
			if _, err := u.fileServiceClient.RemoveProfileImage(ctx, &fpb.ImageName{Name: *us.Image}); err != nil {
				go func() {
					if err := u.util.EmitLog("ERR", fmt.Sprintf("Error remove image via file service. err: %v", err.Error())); err != nil {
						u.logger.Error().Err(err).Msg("failed to emit log")
					}
				}()
			}
		}
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
