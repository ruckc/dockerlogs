package pkg

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	docker_client "github.com/docker/docker/client"
)

type DockerProvider struct {
	name   string
	client *docker_client.Client

	containers []DockerSource
	sources    []Source
}

func NewLocalDockerProvider() *DockerProvider {
	client, err := docker_client.NewClientWithOpts()
	if err != nil {
		panic(err)
	}

	dp := DockerProvider{
		name:   "local",
		client: client,
	}

	return &dp
}

func (dp *DockerProvider) Name() string {
	return dp.name
}

func (dp *DockerProvider) Refresh(ctx context.Context) {
	dp.refreshContainers(ctx)
}

func (dp *DockerProvider) refreshContainers(ctx context.Context) {
	list, err := dp.client.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}

	sources := make([]Source, len(list))
	containers := make([]DockerSource, len(list))
	for i, v := range list {
		src := DockerSource{
			dp:        dp,
			container: v,
		}
		containers[i] = src
		sources[i] = &src
	}

	dp.containers = containers
	dp.sources = sources
}

func (dp *DockerProvider) GetSources() []Source {
	return dp.sources
}

type DockerSource struct {
	dp        *DockerProvider
	container types.Container
}

func (ds *DockerSource) Name() string {
	return strings.ReplaceAll(ds.container.Names[0], "/", "")
}

func (ds *DockerSource) Tail(ctx context.Context, follow bool, tail string) (io.ReadCloser, io.ReadCloser) {
	stdout, err := ds.dp.client.ContainerLogs(ctx, ds.container.ID, types.ContainerLogsOptions{Follow: follow, Tail: tail, ShowStdout: true, ShowStderr: false})
	if err != nil {
		panic(err)
	}

	stderr, err := ds.dp.client.ContainerLogs(ctx, ds.container.ID, types.ContainerLogsOptions{Follow: follow, Tail: tail, ShowStdout: false, ShowStderr: true})
	if err != nil {
		panic(err)
	}

	return stdout, stderr
}
