package dto

// RegisterRequest 注册请求DTO
type RegisterRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Username        string `json:"username"`
	IP              string `json:"ip"`
	DeviceID        string `json:"device_id"`
}

// RegisterResponse 注册响应DTO
type RegisterResponse struct {
	StatusCode     uint32 `json:"status_code"`
	StatusMsg      string `json:"status_msg"`
	UserID         uint32 `json:"user_id"`
	ShortToken     string `json:"short_token"`
	LongToken      string `json:"long_token"`
	ShortExpiresIn int64  `json:"short_expires_in"`
	LongExpiresIn  int64  `json:"long_expires_in"`
}

// LoginRequest 登录请求DTO
type LoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	IP       string `json:"ip"`
	DeviceID string `json:"device_id"`
}

// LoginResponse 登录响应DTO
type LoginResponse struct {
	StatusCode     uint32 `json:"status_code"`
	StatusMsg      string `json:"status_msg"`
	UserID         uint32 `json:"user_id"`
	ShortToken     string `json:"short_token"`
	LongToken      string `json:"long_token"`
	ShortExpiresIn int64  `json:"short_expires_in"`
	LongExpiresIn  int64  `json:"long_expires_in"`
}

// LogoutRequest 登出请求DTO
type LogoutRequest struct {
	UserID    uint32 `json:"user_id"`
	LongToken string `json:"long_token"`
	IP        string `json:"ip"`
}

// LogoutResponse 登出响应DTO
type LogoutResponse struct {
	StatusCode uint32 `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}
