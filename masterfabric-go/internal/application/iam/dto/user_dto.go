package dto

// RegisterRequest is the input for user registration.
type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginRequest is the input for user login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserResponse is the public user JSON shape expected by the frontend.
type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

// MessageResponse is a simple status message.
type MessageResponse struct {
	Message string `json:"message"`
}

// ValidateResponse confirms token validity.
type ValidateResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id"`
}

// UpdateMeRequest updates the authenticated user's profile.
type UpdateMeRequest struct {
	Username string `json:"username"`
}

// ChangePasswordRequest changes the authenticated user's password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}
