package dto

import (
	"database/sql"
	"executor/internal/core/models"
)

type ContainerDbo struct {
	Id            int64        `db:"id"`
	Port          int64        `db:"port"`
	ContainerName string       `db:"container_name"`
	ContainerID   string       `db:"container_id"`
	BotID         int64        `db:"bot_id"`
	ProjectID     int64        `db:"project_id"`
	UserID        int64        `db:"user_id"`
	Name          string       `db:"name"`
	Description   string       `db:"description"`
	Icon          string       `db:"icon"`
	State         string       `db:"state"`
	CreatedAt     sql.NullTime `db:"created_at"`
	UpdatedAt     sql.NullTime `db:"updated_at"`
	DeletedAt     sql.NullTime `db:"deleted_at"`
}

func (d *ContainerDbo) ToValue() models.Container {
	return models.Container{
		Id:            d.Id,
		Port:          d.Port,
		ContainerName: d.ContainerName,
		ContainerID:   d.ContainerID,
		BotID:         d.BotID,
		ProjectID:     d.ProjectID,
		UserID:        d.UserID,
		Name:          d.Name,
		Description:   d.Description,
		Icon:          d.Icon,
		State:         d.State,
	}
}

func ToContainerDbo(m models.Container) ContainerDbo {
	return ContainerDbo{
		Id:            m.Id,
		Port:          m.Port,
		ContainerName: m.ContainerName,
		ContainerID:   m.ContainerID,
		BotID:         m.BotID,
		ProjectID:     m.ProjectID,
		UserID:        m.UserID,
		Name:          m.Name,
		Description:   m.Description,
		Icon:          m.Icon,
		State:         m.State,
	}
}
