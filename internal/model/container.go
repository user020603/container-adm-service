package model

import "time"

type Container struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	ContainerID   string    `json:"container_id" gorm:"unique;not null"`
	ContainerName string    `json:"container_name" gorm:"unique;not null"`
	ImageName     string    `json:"image_name" gorm:"not null"`
	Status        string    `json:"status" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
