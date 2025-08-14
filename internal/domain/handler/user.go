package handler

import (
	"fmt"
	"net/http"

	"10.1.20.130/dropping/sharedlib/utils"
	"10.1.20.130/dropping/user-service/internal/domain/dto"
	"10.1.20.130/dropping/user-service/internal/domain/service"
	u "10.1.20.130/dropping/user-service/pkg/utils"
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
		DeleteUser(ctx *gin.Context)
	}
	userHandler struct {
		userService service.UserService
		logger      zerolog.Logger
		logEmitter  u.LoggerServiceUtil
	}
)

func NewUserHandler(userService service.UserService, logEmitter u.LoggerServiceUtil, logger zerolog.Logger) UserHandler {
	return &userHandler{
		userService: userService,
		logger:      logger,
		logEmitter:  logEmitter,
	}
}

// @Summary Delete User
// @Description Delete User based on its ID (from token)
// @Tags User-Service
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body dto.DeleteUserRequest true "Body Request"
// @Success 200 {object} dto.DeleteUserSuccessExample "Delete User Success"
// @Failure 400 {object} dto.GlobalInvalidInputExample "Bad request - invalid input, password and confirm_password doesn't match"
// @Failure 401 {object} dto.GlobalUnauthorizedErrorExample "Unauthorized - token invalid, wrong password"
// @Failure 404 {object} dto.GlobalUserNotFoundExample "User not found"
// @Failure 500 {object} dto.GlobalInternalServerErrorExample "Internal server error"
// @Router / [delete]
func (u *userHandler) DeleteUser(ctx *gin.Context) {
	userId := utils.GetUserId(ctx)
	if userId == "" {
		go func() {
			if err := u.logEmitter.EmitLog("ERR", fmt.Sprintf("unathorized. userId is not found. userId: %s", userId)); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		res := utils.ReturnResponseError(401, dto.Err_UNAUTHORIZED_USER_ID_NOTFOUND.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}
	var req dto.DeleteUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		go func() {
			if err := u.logEmitter.EmitLog("ERR", fmt.Sprintf("bad request: Err:%s", err.Error())); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		res := utils.ReturnResponseError(400, "invalid input")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	if err := u.userService.DeleteUser(&req, userId); err != nil {
		switch err {
		case dto.Err_UNAUTHORIZED_PASSWORD_WRONG:
			res := utils.ReturnResponseError(401, err.Error())
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
			return
		case dto.Err_NOTFOUND_USER_NOT_FOUND:
			res := utils.ReturnResponseError(404, err.Error())
			ctx.AbortWithStatusJSON(http.StatusNotFound, res)
			return
		}
		res := utils.ReturnResponseError(500, "internal server error")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		return
	}
	res := utils.ReturnResponseSuccess(200, dto.SUCCESS_DELETE_USER)
	ctx.JSON(http.StatusOK, res)
}

// @Summary Change Password
// @Description Change Password based on its ID (from token)
// @Tags User-Service
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body dto.UpdatePasswordRequest true "Body Request"
// @Success 200 {object} dto.ChangePasswordSuccessExample "Change Password Success"
// @Failure 400 {object} dto.GlobalInvalidInputExample "Bad request - invalid input, password and confirm_password doesn't match"
// @Failure 401 {object} dto.GlobalUnauthorizedErrorExample "Unauthorized - token invalid, wrong password"
// @Failure 404 {object} dto.GlobalUserNotFoundExample "User not found"
// @Failure 500 {object} dto.GlobalInternalServerErrorExample "Internal server error"
// @Router /password [patch]
func (u *userHandler) ChangePassword(ctx *gin.Context) {
	userId := utils.GetUserId(ctx)
	if userId == "" {
		go func() {
			if err := u.logEmitter.EmitLog("ERR", fmt.Sprintf("unathorized. userId is not found. userId: %s", userId)); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		res := utils.ReturnResponseError(401, dto.Err_UNAUTHORIZED_USER_ID_NOTFOUND.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}
	var req dto.UpdatePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		go func() {
			if err := u.logEmitter.EmitLog("ERR", fmt.Sprintf("bad request: Err:%s", err.Error())); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		res := utils.ReturnResponseError(400, "invalid input")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	if err := u.userService.UpdatePassword(&req, userId); err != nil {
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
		res := utils.ReturnResponseError(500, "internal server error")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		return
	}
	res := utils.ReturnResponseSuccess(200, dto.SUCCESS_UPDATE_PASSWORD)
	ctx.JSON(http.StatusOK, res)
}

// @Summary Change Email
// @Description Change Email based on its ID (from token)
// @Tags User-Service
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param request body dto.UpdateEmailRequest true "Body Request"
// @Success 200 {object} dto.ChangeEmailSuccessExample "Change Email Success - need verification in auth API"
// @Failure 400 {object} dto.GlobalInvalidInputExample "Bad request - invalid input"
// @Failure 401 {object} dto.GlobalUnauthorizedErrorExample "Unauthorized"
// @Failure 500 {object} dto.GlobalInternalServerErrorExample "Internal server error"
// @Router /email [patch]
func (u *userHandler) ChangeEmail(ctx *gin.Context) {
	userId := utils.GetUserId(ctx)
	if userId == "" {
		go func() {
			if err := u.logEmitter.EmitLog("ERR", fmt.Sprintf("unathorized. userId is not found. userId: %s", userId)); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		res := utils.ReturnResponseError(401, dto.Err_UNAUTHORIZED_USER_ID_NOTFOUND.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}
	var req dto.UpdateEmailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		go func() {
			if err := u.logEmitter.EmitLog("ERR", fmt.Sprintf("bad request: Err:%s", err.Error())); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		res := utils.ReturnResponseError(400, "invalid input")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}
	if err := u.userService.UpdateEmail(&req, userId); err != nil {
		res := utils.ReturnResponseError(500, "internal server error")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		return
	}
	res := utils.ReturnResponseSuccess(200, dto.SUCCESS_UPDATE_EMAIL)
	ctx.JSON(http.StatusOK, res)
}

// @Summary Update User
// @Description Update User based on its ID (from token)
// @Tags User-Service
// @Accept multipart/form-data
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param image formData file false "Profile image"
// @Param request formData dto.UpdateUserRequest true "Body Request"
// @Success 200 {object} dto.UpdateUserSuccessExample "Update user Success"
// @Failure 400 {object} dto.GlobalInvalidInputExample "Bad request - invalid input, wrong image extension, and limit image exceeded"
// @Failure 401 {object} dto.GlobalUnauthorizedErrorExample "Unauthorized"
// @Failure 404 {object} dto.GlobalUserNotFoundExample "User not found"
// @Failure 500 {object} dto.GlobalInternalServerErrorExample "Internal server error"
// @Router / [patch]
func (u *userHandler) UpdateUser(ctx *gin.Context) {
	userId := utils.GetUserId(ctx)
	if userId == "" {
		go func() {
			if err := u.logEmitter.EmitLog("ERR", fmt.Sprintf("unathorized. userId is not found. userId: %s", userId)); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		res := utils.ReturnResponseError(401, dto.Err_UNAUTHORIZED_USER_ID_NOTFOUND.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}

	var req dto.UpdateUserRequest
	if err := ctx.ShouldBind(&req); err != nil {
		go func() {
			if err := u.logEmitter.EmitLog("ERR", fmt.Sprintf("bad request: Err:%s", err.Error())); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
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
		if code == codes.Internal {
			res := utils.ReturnResponseError(500, "internal server error")
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
			return
		}
		res := utils.ReturnResponseError(500, "internal server error")
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, res)
		return
	}
	res := utils.ReturnResponseSuccess(200, dto.SUCCESS_UPDATE_PROFILE)
	ctx.JSON(http.StatusOK, res)
}

// @Summary Get User Profile
// @Description Get profile User based on its ID (from token)
// @Tags User-Service
// @Accept */*
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} dto.GetProfileSuccessExample "Get Profile Success"
// @Failure 401 {object} dto.GlobalUnauthorizedErrorExample "Unauthorized"
// @Failure 404 {object} dto.GlobalUserNotFoundExample "User not found
// @Failure 500 {object} dto.GlobalInternalServerErrorExample "Internal server error
// @Router /me [get]
func (u *userHandler) GetProfile(ctx *gin.Context) {
	userId := utils.GetUserId(ctx)
	if userId == "" {
		go func() {
			if err := u.logEmitter.EmitLog("ERR", fmt.Sprintf("unathorized. userId is not found. userId: %s", userId)); err != nil {
				u.logger.Error().Err(err).Msg("failed to emit log")
			}
		}()
		res := utils.ReturnResponseError(401, dto.Err_UNAUTHORIZED_USER_ID_NOTFOUND.Error())
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}
	user, err := u.userService.GetProfile(userId)
	if err != nil {
		switch err {
		case dto.Err_INTERNAL_FAILED_BUILD_QUERY, dto.Err_INTERNAL_FAILED_SCAN_USER:
			res := utils.ReturnResponseError(500, "internal server error")
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
