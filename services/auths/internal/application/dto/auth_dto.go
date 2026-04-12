package dto

// AuthReq 认证请求
type AuthReq struct {
	Token    string `json:"token"`
	ClientIP string `json:"clientIp"`
}

// AuthRes 认证响应
type AuthRes struct {
	UserID     int64  `json:"userId"`
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}

// GenerateTokenReq 生成Token请求
type GenerateTokenReq struct {
	UserID    int64  `json:"userId"`
	ClientIP  string `json:"clientIp"`
	ExpiresIn int64  `json:"expiresIn"` // 过期时间（秒）
}

// GenerateTokenResp 生成Token响应
type GenerateTokenResp struct {
	Token      string `json:"token"`
	ExpiresIn  int64  `json:"expiresIn"`
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}

// RenewTokenReq 续期Token请求
type RenewTokenReq struct {
	Token    string `json:"token"`
	ClientIP string `json:"clientIp"`
}

// RenewTokenResp 续期Token响应
type RenewTokenResp struct {
	Token      string `json:"token"`
	ExpiresIn  int64  `json:"expiresIn"`
	StatusCode int64  `json:"statusCode"`
	StatusMsg  string `json:"statusMsg"`
}
