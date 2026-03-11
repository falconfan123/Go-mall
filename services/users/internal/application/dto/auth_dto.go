package dto

// RegisterRequest 注册请求DTO
type RegisterRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Username        string `json:"username"`
	IP              string `json:"ip"`
}

// RegisterResponse 注册响应DTO
type RegisterResponse struct {
	StatusCode   uint32 `json:"status_code"`
	StatusMsg    string `json:"status_msg"`
	UserID       uint32 `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// LoginRequest 登录请求DTO
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	IP       string `json:"ip"`
}

// LoginResponse 登录响应DTO
type LoginResponse struct {
	StatusCode   uint32 `json:"status_code"`
	StatusMsg    string `json:"status_msg"`
	UserID       uint32 `json:"user_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest 登出请求DTO
type LogoutRequest struct {
	UserID uint32 `json:"user_id"`
	IP     string `json:"ip"`
}

// LogoutResponse 登出响应DTO
type LogoutResponse struct {
	StatusCode uint32 `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}
