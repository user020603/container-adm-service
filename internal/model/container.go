package model

import "time"

type Container struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement" db:"id"`
	ContainerID string    `json:"container_id" gorm:"unique;not null" db:"container_id"`
	Name        string    `json:"name" gorm:"not null" db:"name"`
	Image       string    `json:"image" gorm:"not null" db:"image"`
	Status      string    `json:"status" gorm:"not null" db:"status"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime" db:"updated_at"`
}
