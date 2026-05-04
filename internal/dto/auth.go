package dto

// REQUESTS
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=255"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Phone    string `json:"phone" binding:"required,len=10"`
	Role     string `json:"role" binding:"required,oneof=admin marketer analyst"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RESPONSE
type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
}

type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

type UpdateProfileRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Bio     string `json:"bio"`
	Picture string `json:"picture"`
}
