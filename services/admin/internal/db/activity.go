package db

import (
	"time"

	"gorm.io/gorm"
)

type Activity struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string    `gorm:"type:varchar(255);not null" json:"name"`
	ProductID    int64     `gorm:"not null" json:"product_id"`
	SeckillPrice int64     `gorm:"not null" json:"seckill_price"`
	TotalStock   int64     `gorm:"not null" json:"total_stock"`
	LimitPerUser int64     `gorm:"default:1" json:"limit_per_user"`
	StartTime    time.Time `gorm:"not null" json:"start_time"`
	EndTime      time.Time `gorm:"not null" json:"end_time"`
	Status       string    `gorm:"type:varchar(20);default:'pending'" json:"status"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Activity) TableName() string {
	return "activities"
}

func (a *Activity) Create(db *gorm.DB) error {
	return db.Create(a).Error
}

func (a *Activity) Update(db *gorm.DB) error {
	return db.Save(a).Error
}

func (a *Activity) Delete(db *gorm.DB) error {
	return db.Delete(a).Error
}

func GetActivityByID(db *gorm.DB, id int64) (*Activity, error) {
	var activity Activity
	err := db.First(&activity, id).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

func ListActivities(db *gorm.DB, page, pageSize int64, status string) ([]*Activity, int64, error) {
	var activities []*Activity
	var total int64

	query := db.Model(&Activity{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(int(offset)).Limit(int(pageSize)).Find(&activities).Error; err != nil {
		return nil, 0, err
	}

	return activities, total, nil
}
