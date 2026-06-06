package main

import (
	"fmt"
	"os"
	"path/filepath"

	"441hz/internal/cli"
	"441hz/internal/handler"
	"441hz/internal/presenter"
	apirouter "441hz/internal/router"
	"441hz/internal/service"
	"441hz/internal/worker"
)

func main() {
	app, err := buildApp()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		if err := app.RunArgs(os.Args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		if err := app.RunInteractive(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func buildApp() (*cli.App, error) {
	workspace := service.NewWorkspaceService("workspace")
	jobService := service.NewJobService(workspace)
	pythonRunner := worker.NewPythonRunner(defaultPythonPath(), "pyworker/worker.py")
	separationService := service.NewSeparationService(jobService, pythonRunner, "models/separation/demucs")
	inspectService := service.NewInspectService()
	output := presenter.NewTerminalPresenter(os.Stdout)

	handlers := handler.Registry{
		Analyze:  handler.NewAnalyzeHandler(separationService),
		Separate: handler.NewSeparateHandler(separationService),
		Inspect:  handler.NewInspectHandler(inspectService),
	}

	router := apirouter.New(handlers)
	parser := cli.NewParser()

	return cli.NewApp(parser, router, output), nil
}

func defaultPythonPath() string {
	venvPython := filepath.Join(".venv", "bin", "python")
	if _, err := os.Stat(venvPython); err == nil {
		return venvPython
	}
	return "python3"
}
