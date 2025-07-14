package handler

import (
	"net/http"

	"github.com/dropboks/sharedlib/utils"
	"github.com/dropboks/user-service/internal/domain/dto"
	"github.com/dropboks/user-service/internal/domain/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	UserHandler interface {
		GetProfile(ctx *gin.Context)
		UpdateUser(ctx *gin.Context)
		ChangeEmail(ctx *gin.Context)
		ChangePassword(ctx *gin.Context)
	}
	userHandler struct {
		userService service.UserService
		logger      zerolog.Logger
	}
)

func NewUserHandler(userService service.UserService, logger zerolog.Logger) UserHandler {
	return &userHandler{
		userService: userService,
		logger:      logger,
	}
}

func (u *userHandler) ChangePassword(ctx *gin.Context) {
	userId := utils.GetUserId(ctx)
	if userId == "" {
		u.logger.Error().Msg("unathorized. userId is not found")
		res := utils.ReturnResponseError(401, dto.Err_UNAUTHORIZED_USER_ID_NOTFOUND.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}
	var req dto.UpdatePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		u.logger.Error().Err(err).Msg("Bad Request")
		res := utils.ReturnResponseError(400, "invalid input")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	if err := u.userService.UpdatePassword(&req, userId); err != nil {
		u.logger.Error().Err(err).Msg("Failed to update password")
		switch err {
		case dto.Err_BAD_REQUEST_PASSWORD_CONFIRM_PASSWORD_DOESNT_MATCH:
			res := utils.ReturnResponseError(400, err.Error())
			ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
			return
		case dto.Err_UNAUTHORIZED_PASSWORD_WRONG:
			res := utils.ReturnResponseError(401, err.Error())
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
			return
		case dto.Err_NOTFOUND_USER_NOT_FOUND:
			res := utils.ReturnResponseError(404, err.Error())
			ctx.AbortWithStatusJSON(http.StatusNotFound, res)
			return
		}
		res := utils.ReturnResponseError(500, "failed to update password")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		return
	}
	res := utils.ReturnResponseSuccess(200, dto.SUCCESS_UPDATE_PASSWORD)
	ctx.JSON(http.StatusOK, res)
}

func (u *userHandler) ChangeEmail(ctx *gin.Context) {
	userId := utils.GetUserId(ctx)
	if userId == "" {
		u.logger.Error().Msg("unathorized. userId is not found")
		res := utils.ReturnResponseError(401, dto.Err_UNAUTHORIZED_USER_ID_NOTFOUND.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}
	var req dto.UpdateEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		u.logger.Error().Err(err).Msg("Bad Request")
		res := utils.ReturnResponseError(400, "invalid input")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	if err := u.userService.UpdateEmail(&req, userId); err != nil {
		u.logger.Error().Err(err).Msg("Failed to update email")
		res := utils.ReturnResponseError(500, "failed to update email")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		return
	}
	res := utils.ReturnResponseSuccess(200, dto.SUCCESS_UPDATE_EMAIL)
	ctx.JSON(http.StatusOK, res)
}

func (u *userHandler) UpdateUser(ctx *gin.Context) {
	userId := utils.GetUserId(ctx)
	if userId == "" {
		u.logger.Error().Msg("unathorized. userId is not found")
		res := utils.ReturnResponseError(401, dto.Err_UNAUTHORIZED_USER_ID_NOTFOUND.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}

	var req dto.UpdateUserRequest
	if err := ctx.ShouldBind(&req); err != nil {
		u.logger.Error().Err(err).Msg("Bad Request")
		res := utils.ReturnResponseError(400, "invalid input")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	if err := u.userService.UpdateUser(&req, userId); err != nil {
		switch err {
		case dto.Err_NOTFOUND_USER_NOT_FOUND:
			res := utils.ReturnResponseError(404, err.Error())
			ctx.AbortWithStatusJSON(http.StatusNotFound, res)
			return
		case dto.Err_BAD_REQUEST_WRONG_EXTENSION:
			res := utils.ReturnResponseError(400, err.Error())
			ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
			return
		case dto.Err_BAD_REQUEST_LIMIT_SIZE_EXCEEDED:
			res := utils.ReturnResponseError(400, err.Error())
			ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
			return
		}
		code := status.Code(err)
		message := status.Convert(err).Message()
		if code == codes.Internal {
			res := utils.ReturnResponseError(500, message)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
			return
		}
		res := utils.ReturnResponseError(500, err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		return
	}
	res := utils.ReturnResponseSuccess(200, dto.SUCCESS_UPDATE_PROFILE)
	ctx.JSON(http.StatusOK, res)
}

func (u *userHandler) GetProfile(ctx *gin.Context) {
	userId := utils.GetUserId(ctx)
	if userId == "" {
		u.logger.Error().Msg("unathorized")
		res := utils.ReturnResponseError(401, dto.Err_UNAUTHORIZED_USER_ID_NOTFOUND.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}
	user, err := u.userService.GetProfile(userId)
	if err != nil {
		switch err {
		case dto.Err_INTERNAL_FAILED_BUILD_QUERY, dto.Err_INTERNAL_FAILED_SCAN_USER:
			res := utils.ReturnResponseError(500, err.Error())
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
			return
		case dto.Err_NOTFOUND_USER_NOT_FOUND:
			res := utils.ReturnResponseError(404, err.Error())
			ctx.AbortWithStatusJSON(http.StatusNotFound, res)
			return
		}
	}
	res := utils.ReturnResponseSuccess(200, dto.SUCCESS_GET_PROFILE, user)
	ctx.JSON(http.StatusOK, res)
}
