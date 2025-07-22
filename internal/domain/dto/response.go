package dto

import "errors"

var (
	SUCCESS_GET_PROFILE     = "success get profile data"
	SUCCESS_UPDATE_PROFILE  = "success update profile data"
	SUCCESS_UPDATE_EMAIL    = "verify to change email"
	SUCCESS_UPDATE_PASSWORD = "success update password"
	SUCCESS_DELETE_USER     = "success delete user"
)

var (
	Err_INTERNAL_FAILED_BUILD_QUERY = errors.New("failed to build query")
	Err_INTERNAL_FAILED_SCAN_USER   = errors.New("failed to scan user")
	Err_INTERNAL_FAILED_INSERT_USER = errors.New("failed to insert user")
	Err_INTERNAL_FAILED_UPDATE_USER = errors.New("failed to update user")
	Err_INTERNAL_FAILED_DELETE_USER = errors.New("failed to delete user")
	Err_INTERNAL_CONVERT_IMAGE      = errors.New("error processing image")
	Err_INTERNAL_GENERATE_TOKEN     = errors.New("error generate verification token")
	Err_INTERNAL_GET_RESOURCE       = errors.New("failed to get resource")
	Err_INTERNAL_SET_RESOURCE       = errors.New("failed save resource")
	Err_INTERNAL_DELETE_RESOURCE    = errors.New("failed to delete resource")
	Err_INTERNAL_PUBLISH_MESSAGE    = errors.New("error publish email")

	Err_NOTFOUND_USER_NOT_FOUND = errors.New("user not found")
	Err_NOTFOUND_KEY_NOTFOUND   = errors.New("resource is not found")

	Err_UNAUTHORIZED_USER_ID_NOTFOUND = errors.New("invalid token")
	Err_UNAUTHORIZED_PASSWORD_WRONG   = errors.New("wrong password")

	Err_BAD_REQUEST_WRONG_EXTENSION                        = errors.New("error file extension, support jpg, jpeg, and png")
	Err_BAD_REQUEST_LIMIT_SIZE_EXCEEDED                    = errors.New("max size exceeded: 6mb")
	Err_BAD_REQUEST_PASSWORD_CONFIRM_PASSWORD_DOESNT_MATCH = errors.New("password doesn't match")
)

type (
	GetProfileResponse struct {
		FullName         string  `json:"full_name" example:"John Doe"`
		Image            *string `json:"image" example:"https://example.com/image.jpg"`
		Email            string  `json:"email" example:"john.doe@example.com"`
		Verified         bool    `json:"verified" example:"true"`
		TwoFactorEnabled bool    `json:"two_factor_enabled" example:"false"`
	}

	GlobalInternalServerErrorExample struct {
		StatusCode uint16 `json:"status_code" example:"500"`
		Message    string `json:"message" example:"internal server error"`
	}
	GlobalUserNotFoundExample struct {
		StatusCode uint16 `json:"status_code" example:"404"`
		Message    string `json:"message" example:"user not found"`
	}

	GlobalUnauthorizedErrorExample struct {
		StatusCode uint16 `json:"status_code" example:"401"`
		Message    string `json:"message" example:"unauthorized"`
	}
	GlobalInvalidInputExample struct {
		StatusCode uint16 `json:"status_code" example:"400"`
		Message    string `json:"message" example:"invalid input"`
	}
	GetProfileSuccessExample struct {
		StatusCode uint16             `json:"status_code" example:"200"`
		Message    string             `json:"message" example:"success get profile data"`
		Data       GetProfileResponse `json:"data"`
	}
	UpdateUserSuccessExample struct {
		StatusCode uint16 `json:"status_code" example:"200"`
		Message    string `json:"message" example:"success update profile data"`
		Data       string `json:"data" example:"null"`
	}

	ChangeEmailSuccessExample struct {
		StatusCode uint16 `json:"status_code" example:"200"`
		Message    string `json:"message" example:"verify to change email"`
		Data       string `json:"data" example:"null"`
	}
	ChangePasswordSuccessExample struct {
		StatusCode uint16 `json:"status_code" example:"200"`
		Message    string `json:"message" example:"success delete user"`
		Data       string `json:"data" example:"null"`
	}
	DeleteUserSuccessExample struct {
		StatusCode uint16 `json:"status_code" example:"200"`
		Message    string `json:"message" example:"success delete user"`
		Data       string `json:"data" example:"null"`
	}
)
