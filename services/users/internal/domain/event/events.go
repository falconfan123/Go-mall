package event

import "time"

// BaseEvent 基础事件结构
type BaseEvent struct {
	EventID    string    `json:"event_id"`
	EventType  string    `json:"event_type"`
	OccurredAt time.Time `json:"occurred_at"`
}

// UserRegisteredEvent 用户注册成功事件
type UserRegisteredEvent struct {
	BaseEvent
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	IP       string `json:"ip"`
}

// UserLoggedInEvent 用户登录成功事件
type UserLoggedInEvent struct {
	BaseEvent
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	IP     string `json:"ip"`
}

// UserLoggedOutEvent 用户登出事件
type UserLoggedOutEvent struct {
	BaseEvent
	UserID int64  `json:"user_id"`
	IP     string `json:"ip"`
}

// UserInfoUpdatedEvent 用户信息更新事件
type UserInfoUpdatedEvent struct {
	BaseEvent
	UserID   int64  `json:"user_id"`
	Field    string `json:"field"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

// AddressAddedEvent 地址添加事件
type AddressAddedEvent struct {
	BaseEvent
	UserID    int64 `json:"user_id"`
	AddressID int64 `json:"address_id"`
}

// AddressUpdatedEvent 地址更新事件
type AddressUpdatedEvent struct {
	BaseEvent
	UserID    int64 `json:"user_id"`
	AddressID int64 `json:"address_id"`
}

// AddressDeletedEvent 地址删除事件
type AddressDeletedEvent struct {
	BaseEvent
	UserID    int64 `json:"user_id"`
	AddressID int64 `json:"address_id"`
}
