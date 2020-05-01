package cmd

import (
	"github.com/Graylog2/graylog-project-cli/git"
	"github.com/Graylog2/graylog-project-cli/runner"
	"github.com/spf13/cobra"
	"path/filepath"
)

var (
	runApiPort           string
	runWebPort           string
	runElasticsearchPort string
	runMongoDBPort       string
	runCleanupVolumes    bool
	runBuildImages       bool
	runBuildClean        bool
	runBuildWeb          bool
)

func init() {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run Graylog server, MongoDB , Elasticsearch and other services",
		Long:  "This command offers several sub-commands to start Graylog servers, MongoDB, Elasticsearch and other services.",
		Example: `  # Always change into the graylog-project directory first
  cd /path/to/graylog-project

  # Start a Graylog DEV server including MongoDB and Elasticsearch
  graylog-project run dev
  graylog-project run dev --web    # Include production web build
  graylog-project run dev --clean  # Run clean server build

  # Start MongoDB and Elasticsearch services
  graylog-project run dev:services

  # Start Graylog DEV server without MongoDB and Elasticsearch
  graylog-project run dev:server
  graylog-project run dev:server --web    # Include production web build
  graylog-project run dev:server --clean  # Run clean server build

  # Start Graylog DEV web server
  graylog-project run dev:web

  # Cleanup all containers
  graylog-project run dev:cleanup  # Use -V to remove all volumes as well`,
		PersistentPreRunE: persistentPreRunCommand,
	}

	// Flags for all sub-commands
	runCmd.PersistentFlags().StringVarP(&runApiPort, "api-port", "g", "9000", "Graylog HTTP API port")
	runCmd.PersistentFlags().StringVarP(&runWebPort, "web-port", "w", "8080", "Graylog HTTP web port")
	runCmd.PersistentFlags().StringVarP(&runElasticsearchPort, "es-port", "e", "9200", "Elasticsearch port")
	runCmd.PersistentFlags().StringVarP(&runMongoDBPort, "mongodb-port", "m", "27017", "MongoDB port")
	runCmd.PersistentFlags().BoolVarP(&runBuildImages, "build-images", "B", false, "Rebuild Docker images")
	runCmd.PersistentFlags().BoolVar(&runBuildClean, "clean", false, "Run clean server build")
	runCmd.PersistentFlags().BoolVar(&runBuildWeb, "web", false, "Run server build including the web interface")

	runDevCmd := &cobra.Command{
		Use:   runner.DevCommand,
		Short: "Starts a Graylog DEV server + MongoDB and Elasticsearch",
		RunE:  runCommand,
	}

	runDevServerCmd := &cobra.Command{
		Use:   runner.DevServerCommand,
		Short: "Starts a Graylog DEV server (without MongoDB and Elasticsearch)",
		RunE:  runCommand,
	}

	runDevWebCmd := &cobra.Command{
		Use:   runner.DevWebCommand,
		Short: "Starts a Graylog Web DEV server",
		RunE:  runCommand,
	}

	runDevServicesCmd := &cobra.Command{
		Use:   runner.DevServicesCommand,
		Short: "Starts MongoDB and Elasticsearch",
		RunE:  runCommand,
	}

	runDevCleanupCmd := &cobra.Command{
		Use:   runner.DevCleanupCommand,
		Short: "Removes all containers (keeps data volumes by default)",
		RunE:  runCommand,
	}
	runDevCleanupCmd.Flags().BoolVarP(&runCleanupVolumes, "volumes", "V", false, "Remove data volumes as well")

	// graylog-project run release 3.2.5
	runReleaseCmd := &cobra.Command{
		Use:    runner.ReleaseCommand,
		Hidden: true, // Not implemented yet
		Short:  "Starts a Graylog release build + MongoDB and Elasticsearch",
		RunE:   runCommand,
	}

	// graylog-project run snapshot latest
	runSnapshotCmd := &cobra.Command{
		Use:    runner.SnapshotCommand,
		Hidden: true, // Not implemented yet
		Short:  "Starts a Graylog snapshot build + MongoDB and Elasticsearch",
		RunE:   runCommand,
	}

	runCmd.AddCommand(runDevCmd)
	runCmd.AddCommand(runDevServerCmd)
	runCmd.AddCommand(runDevWebCmd)
	runCmd.AddCommand(runDevServicesCmd)
	runCmd.AddCommand(runDevCleanupCmd)
	runCmd.AddCommand(runReleaseCmd)
	runCmd.AddCommand(runSnapshotCmd)

	RootCmd.AddCommand(runCmd)
}

func persistentPreRunCommand(cmd *cobra.Command, args []string) error {
	return runner.CheckSetup()
}

func runCommand(cmd *cobra.Command, args []string) error {
	path, err := git.ToplevelPath()
	if err != nil {
		return err
	}

	return runner.DispatchCommand(runner.Config{
		Command:        cmd.Name(),
		RunnerRoot:     filepath.Join(path, "runner"),
		BuildImages:    runBuildImages,
		CleanupVolumes: runCleanupVolumes,
		Graylog: runner.GraylogConfig{
			APIPort:    runApiPort,
			WebPort:    runWebPort,
			BuildClean: runBuildClean,
			BuildWeb:   runBuildWeb,
		},
		MongoDB: runner.MongoDBConfig{
			Port: runMongoDBPort,
		},
		Elasticsearch: runner.ElasticsearchConfig{
			Port: runElasticsearchPort,
		},
	})
}
