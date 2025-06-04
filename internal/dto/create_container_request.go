package dto

type CreateContainerRequest struct {
	ContainerName string `json:"container_name" binding:"required"`
	ImageName     string `json:"image_name" binding:"required"`
}
