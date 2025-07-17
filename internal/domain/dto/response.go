package dto

import "errors"

var (
	SUCCESS_GET_PROFILE     = "success get profile data"
	SUCCESS_UPDATE_PROFILE  = "success update profile data"
	SUCCESS_UPDATE_EMAIL    = "verify to change email"
	SUCCESS_UPDATE_PASSWORD = "success update password"
)

var (
	Err_INTERNAL_FAILED_BUILD_QUERY = errors.New("failed to build query")
	Err_INTERNAL_FAILED_SCAN_USER   = errors.New("failed to scan user")
	Err_INTERNAL_FAILED_INSERT_USER = errors.New("failed to insert user")
	Err_INTERNAL_FAILED_UPDATE_USER = errors.New("failed to update user")
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
		FullName         string  `json:"full_name"`
		Image            *string `json:"image"`
		Email            string  `json:"email"`
		Verified         bool    `json:"verified"`
		TwoFactorEnabled bool    `json:"two_factor_enabled"`
	}
)
