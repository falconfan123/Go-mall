package valueobject

import (
	"errors"
	"time"
)

// ValidPeriod 有效期值对象
type ValidPeriod struct {
	startTime time.Time
	endTime   time.Time
}

var (
	ErrInvalidPeriod = errors.New("end time must be after start time")
	ErrCouponExpired = errors.New("coupon has expired")
	ErrCouponNotActive = errors.New("coupon is not active yet")
)

// NewValidPeriod 创建有效期
func NewValidPeriod(startTime, endTime time.Time) (*ValidPeriod, error) {
	if endTime.Before(startTime) {
		return nil, ErrInvalidPeriod
	}

	return &ValidPeriod{
		startTime: startTime,
		endTime:   endTime,
	}, nil
}

// IsActive 判断是否在有效期内
func (p *ValidPeriod) IsActive() bool {
	now := time.Now()
	return now.After(p.startTime) && now.Before(p.endTime)
}

// Validate 验证是否可用
func (p *ValidPeriod) Validate() error {
	now := time.Now()
	if now.Before(p.startTime) {
		return ErrCouponNotActive
	}
	if now.After(p.endTime) {
		return ErrCouponExpired
	}
	return nil
}

// StartTime 获取开始时间
func (p *ValidPeriod) StartTime() time.Time {
	return p.startTime
}

// EndTime 获取结束时间
func (p *ValidPeriod) EndTime() time.Time {
	return p.endTime
}

// IsExpired 是否已过期
func (p *ValidPeriod) IsExpired() bool {
	return time.Now().After(p.endTime)
}

// IsNotStarted 是否已开始
func (p *ValidPeriod) IsNotStarted() bool {
	return time.Now().Before(p.startTime)
}
