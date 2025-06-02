package dto

type ContainerFilter struct {
	ContainerID   string `json:"container_id"`
	ContainerName string `json:"container_name"`
	ImageName         string `json:"image_name"`
	Status        string `json:"status"`
}
