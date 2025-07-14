package dto

import "mime/multipart"

type (
	UserData struct {
		UserId string `json:"user_id"`
	}

	UpdateUserRequest struct {
		FullName         string                `form:"full_name" binding:"required,min=1,max=100"`
		Image            *multipart.FileHeader `form:"image"`
		TwoFactorEnabled bool                  `form:"two_factor_enabled"`
	}
	UpdateEmailRequest struct {
		Email string `json:"email" binding:"required,email"`
	}
	UpdatePasswordRequest struct {
		Password           string `json:"password" binding:"required,min=6"`
		NewPassword        string `json:"new_password" binding:"required,min=6"`
		ConfirmNewPassword string `json:"confirm_new_password" binding:"required,min=6"`
	}
)
