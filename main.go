package main

import (
	"context"
	"strings"

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

	options, err := cli.NewProjectOptions(
		[]string{yamlFilePath},
		cli.WithResolvedPaths(true),
		cli.WithOsEnv,
		cli.WithDotEnv,
		cli.WithConfigFileEnv,
		cli.WithDefaultConfigPath,
	)
	if err != nil {
		panic(err)
	}

	// copied from https://github.com/docker/compose/blob/v2/cmd/compose/compose.go
	// see https://stackoverflow.com/questions/74830594/created-compose-project-not-listed-when-using-docker-compose-package-in-go

	project, err := cli.ProjectFromOptions(options)
	if err != nil {
		panic(err)
	}

	for i, s := range project.Services {
		s.CustomLabels = map[string]string{
			api.ProjectLabel:     project.Name,
			api.ServiceLabel:     s.Name,
			api.VersionLabel:     api.ComposeVersion,
			api.WorkingDirLabel:  project.WorkingDir,
			api.ConfigFilesLabel: strings.Join(project.ComposeFiles, ","),
			api.OneoffLabel:      "False", // default, will be overridden by `run` command
		}
		if options.EnvFile != "" {
			s.CustomLabels[api.EnvironmentFileLabel] = options.EnvFile
		}
		project.Services[i] = s
	}

	project.WithoutUnnecessaryResources()

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
