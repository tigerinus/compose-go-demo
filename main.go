package main

import (
	"context"

	"github.com/docker/cli/cli/command"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
)

func main() {
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		panic(err)
	}

	composeService := compose.NewComposeService(dockerCli)
	serviceProxy := api.NewServiceProxy().WithService(composeService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stacks, err := serviceProxy.List(ctx, api.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, stack := range stacks {
		println(stack.Name)
	}
}
