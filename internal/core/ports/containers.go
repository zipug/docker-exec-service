package ports

import (
	"context"
	"executor/internal/application/dto"
)

type ContainersRepository interface {
	GetContainerById(ctx context.Context, id int64) (*dto.ContainerDbo, error)
	GetContainerByContainerId(ctx context.Context, container_id string) (*dto.ContainerDbo, error)
	GetContainerByBotInfo(ctx context.Context, bot dto.ContainerDbo) (*dto.ContainerDbo, error)
	GetAllBots(ctx context.Context) ([]dto.ContainerDbo, error)
	CreateBot(ctx context.Context, bot dto.ContainerDbo) (int64, error)
	UpdateBotById(ctx context.Context, bot dto.ContainerDbo) (*dto.ContainerDbo, error)
	DeleteBotById(ctx context.Context, id int64) error
	DeleteBotByContainerId(ctx context.Context, container_id string) error
	DeleteBotByBotInfo(ctx context.Context, bot dto.ContainerDbo) error
	SetBotState(ctx context.Context, state string, id int64) error
	StopBotState(ctx context.Context, id, bot_id int64) error
}
