package cmd

import (
	"github.com/Graylog2/graylog-project-cli/git"
	"github.com/Graylog2/graylog-project-cli/runner"
	"github.com/spf13/cobra"
	"path/filepath"
)

var (
	runApiPort           int
	runWebPort           int
	runElasticsearchPort int
	runMongoDBPort       int
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

  # Start MongoDB and Elasticsearch services
  graylog-project run dev:services

  # Start Graylog DEV server without MongoDB and Elasticsearch
  graylog-project run dev:server
`,
		PersistentPreRunE: persistentPreRunCommand,
	}

	// Flags for all sub-commands
	runCmd.PersistentFlags().IntVarP(&runApiPort, "api-port", "g", 9000, "Graylog HTTP API port")
	runCmd.PersistentFlags().IntVarP(&runWebPort, "web-port", "w", 9000, "Graylog HTTP web port")
	runCmd.PersistentFlags().IntVarP(&runElasticsearchPort, "es-port", "e", 9220, "Elasticsearch port") // TODO: Use 9200 as default
	runCmd.PersistentFlags().IntVarP(&runMongoDBPort, "mongodb-port", "m", 27027, "MongoDB port")       // TODO: Use 27017 as default

	runDevCmd := &cobra.Command{
		Use:   runner.DevCommand,
		Short: "Starts a Graylog DEV server + MongoDB and Elasticsearch",
		RunE:  runCommand,
	}

	runDevServerCmd := &cobra.Command{
		Use:   runner.DevServerCommand,
		Short: "Starts a Graylog DEV server",
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
		Command:    cmd.Name(),
		RunnerRoot: filepath.Join(path, "runner"),
		Graylog: runner.GraylogConfig{
			HTTPPort: runApiPort,
			WebPort:  runWebPort,
		},
		MongoDB: runner.MongoDBConfig{
			Port: runMongoDBPort,
		},
		Elasticsearch: runner.ElasticsearchConfig{
			Port: runElasticsearchPort,
		},
	})
}
