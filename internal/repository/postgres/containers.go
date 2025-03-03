package postgres

import (
	"context"
	"errors"
	"executor/internal/application/dto"
	pu "executor/pkg/postgres_utils"
	"fmt"
)

var (
	ErrBotNotFound   = errors.New("could not find bot_container by id")
	ErrBotsNotFound  = errors.New("could not find any bot_containers")
	ErrBotNotCreated = errors.New("could not create bot_container")
	ErrBotNotUpdated = errors.New("could not update bot_container")
	ErrBotNotDeleted = errors.New("could not delete bot_container")
)

func (repo *PostgresRepository) GetContainerById(ctx context.Context, id int64) (*dto.ContainerDbo, error) {
	rows, err := pu.Dispatch[dto.ContainerDbo](
		ctx,
		repo.db,
		`
		SELECT b.id, b.container_name, b.port, b.container_id, b.bot_id, b.project_id, b.user_id, b.name, b.description, b.icon, b.state, b.api_token
		FROM bot_containers b
		WHERE b.id = $1::bigint
		  AND b.deleted_at IS NULL;
		`,
		id,
	)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrBotNotFound
	}
	return &rows[0], nil
}

func (repo *PostgresRepository) GetContainerByContainerId(ctx context.Context, container_id string) (*dto.ContainerDbo, error) {
	rows, err := pu.Dispatch[dto.ContainerDbo](
		ctx,
		repo.db,
		`
		SELECT b.id, b.container_name, b.port, b.container_id, b.bot_id, b.project_id, b.user_id, b.name, b.description, b.icon, b.state, b.api_token
		FROM bot_containers b
		WHERE b.container_id = $1::text
		  AND b.deleted_at IS NULL;
		`,
		container_id,
	)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrBotNotFound
	}
	return &rows[0], nil
}

func (repo *PostgresRepository) GetContainerByBotInfo(ctx context.Context, bot dto.ContainerDbo) (*dto.ContainerDbo, error) {
	rows, err := pu.Dispatch[dto.ContainerDbo](
		ctx,
		repo.db,
		`
		SELECT b.id, b.container_name, b.port, b.container_id, b.bot_id, b.project_id, b.user_id, b.name, b.description, b.icon, b.state, b.api_token
		FROM bot_containers b
		WHERE b.bot_id = $1::bigint
		  AND b.project_id = $2::bigint
		  AND b.user_id = $3::bigint
		  AND b.deleted_at IS NULL;
		`,
		bot.BotID,
		bot.ProjectID,
		bot.UserID,
	)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrBotNotFound
	}
	return &rows[0], nil
}

func (repo *PostgresRepository) GetAllBots(ctx context.Context) ([]dto.ContainerDbo, error) {
	rows, err := pu.Dispatch[dto.ContainerDbo](
		ctx,
		repo.db,
		`
		SELECT b.id, b.container_name, b.port, b.bot_id, b.container_id, b.project_id, b.user_id, b.name, b.description, b.icon, b.state, b.api_token
		FROM bot_containers b
		WHERE b.deleted_at IS NULL;
		`,
	)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrBotsNotFound
	}
	return rows, nil
}

func (repo *PostgresRepository) CreateBot(ctx context.Context, bot dto.ContainerDbo) (int64, error) {
	fmt.Println(bot)
	rows, err := pu.Dispatch[dto.ContainerDbo](
		ctx,
		repo.db,
		`
		INSERT INTO bot_containers (
			container_name,
			port,
			container_id,
			bot_id,
			project_id,
			user_id,
			name,
			description,
			icon,
			state,
			api_token
		)
		VALUES (
			$1::text,
			$2::integer,
			$3::text,
			$4::bigint,
			$5::bigint,
			$6::bigint,
			$7::text,
			$8::text,
			$9::text,
			$10::text,
			$11::text
		)
		RETURNING *;
		`,
		bot.ContainerName,
		bot.Port,
		bot.ContainerID,
		bot.BotID,
		bot.ProjectID,
		bot.UserID,
		bot.Name,
		bot.Description,
		bot.Icon,
		bot.State,
		bot.ApiToken,
	)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return -1, ErrBotNotCreated
	}
	return rows[0].Id, nil
}

func (repo *PostgresRepository) UpdateBotById(ctx context.Context, bot dto.ContainerDbo) (*dto.ContainerDbo, error) {
	rows, err := pu.Dispatch[dto.ContainerDbo](
		ctx,
		repo.db,
		`
    UPDATE bot_containers
		SET name = $1::text,
		    description = $2::text,
		    icon = $3::text
		    api_token = $4::text
		WHERE id = $5::bigint
		  AND deleted_at IS NULL
		RETURNING *;
		`,
		bot.Name,
		bot.Description,
		bot.Icon,
		bot.Id,
		bot.ApiToken,
	)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrBotNotUpdated
	}
	return &rows[0], nil
}

func (repo *PostgresRepository) DeleteBotById(ctx context.Context, id int64) error {
	_, err := pu.Dispatch[dto.ContainerDbo](
		ctx,
		repo.db,
		`
		DELETE FROM bot_containers
		WHERE id = $1::bigint;
		`,
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (repo *PostgresRepository) DeleteBotByContainerId(ctx context.Context, container_id string) error {
	_, err := pu.Dispatch[dto.ContainerDbo](
		ctx,
		repo.db,
		`
		DELETE FROM bot_containers
		WHERE container_id = $1::text
		  AND state <> 'running';
		`,
		container_id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (repo *PostgresRepository) DeleteBotByBotInfo(ctx context.Context, bot dto.ContainerDbo) error {
	_, err := pu.Dispatch[dto.ContainerDbo](
		ctx,
		repo.db,
		`
		DELETE FROM bot_containers
		WHERE b.bot_id = $1::bigint
		  AND b.project_id = $2::bigint
		  AND b.user_id = $3::bigint
		  AND state <> 'running';
		`,
		bot.BotID,
		bot.ProjectID,
		bot.UserID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (repo *PostgresRepository) SetBotState(ctx context.Context, state string, id int64) error {
	_, err := pu.Dispatch[dto.ContainerDbo](
		ctx,
		repo.db,
		`
		UPDATE bot_containers
		SET state = $1::text
		WHERE id = $2::bigint;
		`,
		state,
		id,
	)
	if err != nil {
		return err
	}
	return nil
}

func (repo *PostgresRepository) StopBotState(ctx context.Context, id, bot_id int64) error {
	tx := repo.db.MustBegin()
	_, err := pu.DispatchTx[dto.ContainerDbo](
		ctx,
		tx,
		`
		UPDATE bot_containers
		SET state = 'stopped'
		WHERE id = $1::bigint
		  AND state <> 'deleted';
		`,
		id,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	_, err = pu.DispatchTx[dto.ContainerDbo](
		ctx,
		tx,
		`
		UPDATE bots
		SET state = 'stopped'
		WHERE id = $1::bigint
		  AND state <> 'deleted';
		`,
		bot_id,
	)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
