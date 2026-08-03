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
	ID           string               `json:"id"`
	Email        string               `json:"email"`
	Username     string               `json:"username"`
	IsAdmin      bool                 `json:"is_admin"`
	PlatformRole string               `json:"platform_role"`
	AccountKind  string               `json:"account_kind"`
	Organization *OrganizationSummary `json:"organization,omitempty"`
}

type OrganizationSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	Role string `json:"role"`
}

// MessageResponse is a simple status message.
type MessageResponse struct {
	Message string `json:"message"`
}

// AuthTokenResponse is returned after sign-in or token refresh.
type AuthTokenResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
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
