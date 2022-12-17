package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/compose/v2/pkg/progress"
)

type service struct {
	apiService api.Service
}

func (s *service) print(ctx context.Context) error {
	stacks, err := s.apiService.List(ctx, api.ListOptions{
		All: true,
	})
	if err != nil {
		return err
	}

	for _, stack := range stacks {
		fmt.Printf("Stack: %+v", stack)
	}

	return nil
}

func (s *service) create(ctx context.Context, project *types.Project) error {
	fmt.Println("Creating project:", project.Name)
	return s.apiService.Create(ctx, project, api.CreateOptions{})
}

func (s *service) start(ctx context.Context, projectName string) error {
	fmt.Println("Starting project:", projectName)
	return s.apiService.Start(ctx, projectName, api.StartOptions{Wait: true})
}

func (s *service) stop(ctx context.Context, projectName string) error {
	fmt.Println("Stopping project:", projectName)
	return s.apiService.Stop(ctx, projectName, api.StopOptions{})
}

func (s *service) remove(ctx context.Context, projectName string) error {
	fmt.Println("Removing project:", projectName)
	return s.apiService.Remove(ctx, projectName, api.RemoveOptions{Force: true})
}

func (s *service) event(ctx context.Context, projectName string) error {
	return s.apiService.Events(ctx, projectName, api.EventsOptions{
		Consumer: func(event api.Event) error {
			fmt.Printf("Event: %+v", event)
			return nil
		},
	})
}

// partially copied from https://github.com/docker/compose/blob/v2/cmd/compose/compose.go
// (see https://stackoverflow.com/questions/74830594/created-compose-project-not-listed-when-using-docker-compose-package-in-go)
func loadProject(yamlFilePath string) (*types.Project, error) {
	options, err := cli.NewProjectOptions(
		[]string{yamlFilePath},
		cli.WithResolvedPaths(true),
		cli.WithOsEnv,
		cli.WithDotEnv,
		cli.WithConfigFileEnv,
		cli.WithDefaultConfigPath,
	)
	if err != nil {
		return nil, err
	}

	project, err := cli.ProjectFromOptions(options)
	if err != nil {
		return nil, err
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

	return project, nil
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

	r, w, err := os.Pipe()

	writer, err := progress.NewWriter(w)
	if err != nil {
		panic(err)
	}

	progress.WithContextWriter(ctx, writer)

	go func() {
		scanner := bufio.NewScanner(r)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if !scanner.Scan() {
					return
				}
				fmt.Println("Scan:", scanner.Text())
			}
		}
	}()

	composeService := &service{
		apiService: compose.NewComposeService(dockerCli),
	}

	yamlFilePath := "/home/ubuntu/junk/wp/docker-compose.yml"

	project, err := loadProject(yamlFilePath)
	if err != nil {
		panic(err)
	}

	go composeService.event(ctx, project.Name)

	if err := composeService.print(ctx); err != nil {
		panic(err)
	}

	if err := composeService.create(ctx, project); err != nil {
		panic(err)
	}

	if err := composeService.print(ctx); err != nil {
		panic(err)
	}

	if err := composeService.start(ctx, project.Name); err != nil {
		panic(err)
	}

	if err := composeService.print(ctx); err != nil {
		panic(err)
	}

	if err := composeService.stop(ctx, project.Name); err != nil {
		panic(err)
	}

	if err := composeService.print(ctx); err != nil {
		panic(err)
	}

	if err := composeService.remove(ctx, project.Name); err != nil {
		panic(err)
	}

	if err := composeService.print(ctx); err != nil {
		panic(err)
	}
}
