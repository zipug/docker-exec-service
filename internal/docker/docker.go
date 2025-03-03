package docker

import (
	"context"
	"executor/internal/application/dto"
	"executor/internal/core/config"
	"executor/internal/core/models"
	"executor/internal/core/ports"
	freeport "executor/pkg/free-port"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerService struct {
	client *client.Client
	repo   ports.ContainersRepository
	ctx    context.Context
	cfg    *config.ExecutorConfig
}

func NewDockerService(ctx context.Context, repo ports.ContainersRepository, cfg *config.ExecutorConfig) *DockerService {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	return &DockerService{
		client: cli,
		repo:   repo,
		ctx:    ctx,
		cfg:    cfg,
	}
}

func (d *DockerService) PullImage(ctx context.Context, img string) error {
	reader, err := d.client.ImagePull(d.ctx, img, image.PullOptions{})
	d.client.ImageImport(d.ctx, image.ImportSource{SourceName: img}, img, image.ImportOptions{})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, reader)
	return nil
}

func (d *DockerService) CreateContainerConfig(ctx context.Context, bot models.Container) (*container.Config, error) {
	return &container.Config{
		Image: d.cfg.Docker.ImageName,
		Tty:   true,
		Env: []string{
			fmt.Sprintf("POSTGRES_HOST=%s", d.cfg.Postgres.Host),
			fmt.Sprintf("POSTGRES_PORT=%d", d.cfg.Postgres.Port),
			fmt.Sprintf("POSTGRES_USER=%s", d.cfg.Postgres.User),
			fmt.Sprintf("POSTGRES_PASSWORD=%s", d.cfg.Postgres.Password),
			fmt.Sprintf("POSTGRES_DB_NAME=%s", d.cfg.Postgres.DBName),
			fmt.Sprintf("POSTGRES_SSL_MODE=%s", d.cfg.Postgres.SSLMode),
			fmt.Sprintf("POSTGRES_MIGRATIONS_PATH=%s", d.cfg.Postgres.MigrationsPath),
			fmt.Sprintf("MINIO_HOST=%s", d.cfg.MiniO.Host),
			fmt.Sprintf("MINIO_PORT=%d", d.cfg.MiniO.Port),
			fmt.Sprintf("MINIO_ROOT_USER=%s", d.cfg.MiniO.User),
			fmt.Sprintf("MINIO_ROOT_PASSWORD=%s", d.cfg.MiniO.Password),
			fmt.Sprintf("MINIO_ARTICLES_BUCKET=%s", d.cfg.MiniO.BucketArticles),
			fmt.Sprintf("MINIO_ATTACHMENTS_BUCKET=%s", d.cfg.MiniO.BucketAttachments),
			fmt.Sprintf("MINIO_AVATARS_BUCKET=%s", d.cfg.MiniO.BucketAvatars),
			fmt.Sprintf("MINIO_USE_SSL=%t", d.cfg.MiniO.UseSsl),
			fmt.Sprintf("MINIO_URL_LIFETIME=%s", d.cfg.MiniO.UrlLifetime),
			fmt.Sprintf("TELEGRAM_BOT_TOKEN=%s", bot.ApiToken),
			fmt.Sprintf("SEARCH_URL=%s", d.cfg.SearchUrl),
			fmt.Sprintf("CONTAINER_BOT_ID=%d", bot.BotID),
			fmt.Sprintf("CONTAINER_PROJECT_ID=%d", bot.ProjectID),
			fmt.Sprintf("CONTAINER_USER_ID=%d", bot.UserID),
			fmt.Sprintf("CONTAINER_NAME=%s", bot.Name),
			fmt.Sprintf("CONTAINER_DESCRIPTION=%s", bot.Description),
			fmt.Sprintf("CONTAINER_ICON=%s", bot.Icon),
			fmt.Sprintf("OPEN_ROUTER_API_TOKEN=%s", d.cfg.OpenRouterAi.Token),
			fmt.Sprintf("OPEN_ROUTER_API_MODEL=%s", d.cfg.OpenRouterAi.Model),
			fmt.Sprintf("OPEN_ROUTER_API_URL=%s", d.cfg.OpenRouterAi.URL),
			fmt.Sprintf("GIGACHAT_GRPC_ADDRESS=%s", d.cfg.GigaChatAi.GRPCAddress),
			fmt.Sprintf("GIGACHAT_AUTH_URL=%s", d.cfg.GigaChatAi.AuthURL),
			fmt.Sprintf("GIGACHAT_AUTHORIZATION_KEY=%s", d.cfg.GigaChatAi.AuthorizationKey),
			fmt.Sprintf("GIGACHAT_SCOPE=%s", d.cfg.GigaChatAi.Scope),
			fmt.Sprintf("GIGACHAT_MODEL=%s", d.cfg.GigaChatAi.Model),
		},
		Labels: map[string]string{
			"co.elastic.logs/enabled":             "true",
			"co.elastic.logs/json.overwrite_keys": "true",
			"co.elastic.logs/json.add_error_key":  "true",
			"co.elastic.logs/json.expand_keys":    "true",
		},
	}, nil
}

func (d *DockerService) CreateContainer(ctx context.Context, bot models.Container) (*container.CreateResponse, int64, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		return nil, 0, err
	}
	hostBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: fmt.Sprintf("%d", port),
	}
	containerPort, err := nat.NewPort("tcp", fmt.Sprintf("%d", port))
	if err != nil {
		return nil, 0, err
	}
	portBinding := nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}
	cfg, err := d.CreateContainerConfig(ctx, bot)
	if err != nil {
		return nil, 0, err
	}
	resp, err := d.client.ContainerCreate(d.ctx, cfg, &container.HostConfig{
		PortBindings: portBinding,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
		NetworkMode: "host",
	}, nil, nil, bot.ContainerName)
	if err != nil {
		return nil, 0, err
	}
	bot.ContainerID = resp.ID
	bot.Port = int64(port)
	dbo := dto.ToContainerDbo(bot)
	fmt.Println(dbo)
	id, err := d.repo.CreateBot(ctx, dbo)
	if err != nil {
		return nil, 0, err
	}
	return &resp, id, nil
}

func (d *DockerService) GetContainerByBotInfo(ctx context.Context, bot models.Container) (*models.Container, error) {
	dbo, err := d.repo.GetContainerByBotInfo(ctx, dto.ToContainerDbo(bot))
	if err != nil {
		return nil, err
	}
	res := dbo.ToValue()
	return &res, nil
}

func (d *DockerService) RunContainer(ctx context.Context, container_id string, db_id int64) error {
	if err := d.repo.SetBotState(ctx, "running", db_id); err != nil {
		return err
	}
	return d.client.ContainerStart(d.ctx, container_id, container.StartOptions{})
}

func (d *DockerService) GetContainerLogs(ctx context.Context, id string) error {
	out, err := d.client.ContainerLogs(d.ctx, id, container.LogsOptions{ShowStdout: true})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, out)
	return nil
}

func (d *DockerService) StopContainer(ctx context.Context, bot models.Container) error {
	dbo := dto.ToContainerDbo(bot)
	cont, err := d.repo.GetContainerByBotInfo(ctx, dbo)
	if err != nil {
		return err
	}
	if err := d.client.ContainerStop(ctx, cont.ContainerID, container.StopOptions{Timeout: &d.cfg.Docker.Timeout}); err != nil {
		return err
	}
	return d.repo.StopBotState(ctx, cont.Id, cont.BotID)
}

func (d *DockerService) StopAllContainers(ctx context.Context) error {
	fmt.Println("stopping all containers...")
	containers, err := d.repo.GetAllBots(ctx)
	if err != nil {
		return err
	}
	/*TODO ERR GROUP*/
	for _, c := range containers {
		if err := d.client.ContainerStop(ctx, c.ContainerID, container.StopOptions{Timeout: &d.cfg.Docker.Timeout}); err != nil {
			return err
		}
		if err := d.repo.StopBotState(ctx, c.Id, c.BotID); err != nil {
			return err
		}
		fmt.Printf("[%s] container stopped\n", c.ContainerID)
	}
	fmt.Println("all containers stopped")
	return nil
}

func (d *DockerService) DockerFactory(message models.BotMessage) error {
	switch message.Type {
	case "run":
		fmt.Println("Running container...")
		/*if err := d.PullImage(d.ctx, d.cfg.Docker.ImageName); err != nil {
			return err
		}*/
		model := models.Container{
			ContainerName: d.PrepareContainerName(fmt.Sprintf("%s%s", transliterateRussian(message.Payload.Name), time.Now())),
			BotID:         message.Payload.BotID,
			ProjectID:     message.Payload.ProjectID,
			UserID:        message.Payload.UserID,
			Name:          message.Payload.Name,
			Description:   message.Payload.Description,
			Icon:          message.Payload.Icon,
			ApiToken:      message.Payload.ApiToken,
			State:         "created",
		}
		fmt.Println(model)
		bot, err := d.GetContainerByBotInfo(d.ctx, model)
		if err != nil {
			resp, db_id, err := d.CreateContainer(d.ctx, model)
			if err != nil {
				return err
			}
			if err := d.RunContainer(d.ctx, resp.ID, db_id); err != nil {
				return err
			}
			if err := d.GetContainerLogs(d.ctx, resp.ID); err != nil {
				return err
			}
			fmt.Printf("[%s] container running\n", resp.ID)
		} else {
			fmt.Println("container already exists")
			if err := d.RunContainer(d.ctx, bot.ContainerID, bot.Id); err != nil {
				if strings.Contains(err.Error(), "No such container") {
					if err := d.repo.StopBotState(d.ctx, bot.Id, bot.BotID); err != nil {
						return err
					}
					if err := d.repo.DeleteBotById(d.ctx, bot.Id); err != nil {
						return err
					}
					resp, db_id, err := d.CreateContainer(d.ctx, model)
					if err != nil {
						return err
					}
					if err := d.RunContainer(d.ctx, resp.ID, db_id); err != nil {
						return err
					}
					if err := d.GetContainerLogs(d.ctx, resp.ID); err != nil {
						return err
					}
				} else {
					return err
				}
			}
			if err := d.GetContainerLogs(d.ctx, bot.ContainerID); err != nil {
				return err
			}
			fmt.Printf("[%s] container running\n", bot.ContainerID)
		}
	case "stop":
		fmt.Println("Stopping container...")
		model := models.Container{
			ContainerID: d.PrepareContainerName(fmt.Sprintf("%s%s", transliterateRussian(message.Payload.Name), time.Now())),
			BotID:       message.Payload.BotID,
			ProjectID:   message.Payload.ProjectID,
			UserID:      message.Payload.UserID,
			Name:        message.Payload.Name,
			Description: message.Payload.Description,
			Icon:        message.Payload.Icon,
		}
		if err := d.StopContainer(d.ctx, model); err != nil {
			return err
		}
		fmt.Println("container stopped")
	}
	return nil
}

func (d *DockerService) PrepareContainerName(str string) string {
	r1 := strings.NewReplacer(
		" ", "",
		"-", "",
		"_", "",
		"!", "",
		"*", "",
		".", "",
		",", "",
		"/", "",
		"\\", "",
		"{", "",
		"}", "",
		"[", "",
		"]", "",
		"(", "",
		")", "",
		":", "",
		"^", "",
		"@", "",
		"#", "",
		"$", "",
		"%", "",
		"&", "",
		"+", "",
		"=", "",
		"`", "",
		"~", "",
		"<", "",
		">", "",
		"?", "",
	)
	clean_str := r1.Replace(str)
	clean_str = strings.ToLower(clean_str)
	return fmt.Sprintf("tg-%s", clean_str)
}

var translitMap = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "yo", 'ж': "zh",
	'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o",
	'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts",
	'ч': "ch", 'ш': "sh", 'щ': "shch", 'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu",
	'я': "ya",
	// Uppercase Cyrillic letters
	'А': "A", 'Б': "B", 'В': "V", 'Г': "G", 'Д': "D", 'Е': "E", 'Ё': "Yo", 'Ж': "Zh",
	'З': "Z", 'И': "I", 'Й': "Y", 'К': "K", 'Л': "L", 'М': "M", 'Н': "N", 'О': "O",
	'П': "P", 'Р': "R", 'С': "S", 'Т': "T", 'У': "U", 'Ф': "F", 'Х': "Kh", 'Ц': "Ts",
	'Ч': "Ch", 'Ш': "Sh", 'Щ': "Shch", 'Ъ': "", 'Ы': "Y", 'Ь': "", 'Э': "E", 'Ю': "Yu",
	'Я': "Ya",
}

func transliterateRussian(input string) string {
	var result strings.Builder
	for _, char := range input {
		if val, ok := translitMap[char]; ok {
			result.WriteString(val)
		} else {
			result.WriteRune(char) // Keep non-Cyrillic characters as-is
		}
	}
	return result.String()
}
