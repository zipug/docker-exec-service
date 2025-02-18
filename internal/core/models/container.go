package models

type Container struct {
	Id            int64
	Port          int64
	ContainerName string
	ContainerID   string
	BotID         int64
	ProjectID     int64
	UserID        int64
	Name          string
	Description   string
	Icon          string
	State         string
	ApiToken      string
}
