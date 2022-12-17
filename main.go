package main

import (
	"context"
	"path/filepath"

	"github.com/compose-spec/compose-go/cli"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
)

type service struct {
	apiService api.Service
}

func (s *service) list(ctx context.Context) {
	stacks, err := s.apiService.List(ctx, api.ListOptions{
		All: true,
	})
	if err != nil {
		panic(err)
	}

	for _, stack := range stacks {
		println(stack.Name)
	}
}

func (s *service) create(ctx context.Context) {
	yamlFilePath := "/home/ubuntu/junk/wp/docker-compose.yml"
	workdir := filepath.Dir(yamlFilePath)

	project, err := cli.ProjectFromOptions(&cli.ProjectOptions{
		WorkingDir:  workdir,
		ConfigPaths: []string{yamlFilePath},
	})
	if err != nil {
		panic(err)
	}

	project.Name = "wp"

	if err := s.apiService.Create(ctx, project, api.CreateOptions{}); err != nil {
		panic(err)
	}
}

func main() {
	// setup
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		panic(err)
	}

	dockerCli.Initialize(&flags.ClientOptions{
		Common: &flags.CommonOptions{},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	composeService := &service{
		apiService: compose.NewComposeService(dockerCli),
	}

	composeService.list(ctx)

	composeService.create(ctx)

	composeService.list(ctx)
}
