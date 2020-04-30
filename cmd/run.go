package cmd

import (
	"github.com/Graylog2/graylog-project-cli/runner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run Graylog server, MongoDB , Elasticsearch and other services",
		Long: `This command offers several sub-commands to start Graylog servers, MongoDB, Elasticsearch and other services.

Examples:
  # Always change into the graylog-project directory first
  cd /path/to/graylog-project

  # Start a Graylog DEV server including MongoDB and Elasticsearch
  graylog-project run dev

  # Start MongoDB and Elasticsearch services
  graylog-project run services

  # Start Graylog DEV server without MongoDB and Elasticsearch
  graylog-project run dev-server
`,
		PersistentPreRunE: persistentPreRunCommand,
	}

	// Flags for all sub-commands
	runCmd.PersistentFlags().IntP("http-port", "g", 9000, "Graylog HTTP port")
	runCmd.PersistentFlags().IntP("es-port", "e", 9220, "Elasticsearch port") // TODO: Use 9200 as default
	runCmd.PersistentFlags().IntP("mongodb-port", "m", 27027, "MongoDB port") // TODO: Use 27017 as default

	runDevCmd := &cobra.Command{
		Use:   "dev",
		Short: "Starts a Graylog DEV server + MongoDB and Elasticsearch",
		RunE:  runCommand,
	}

	runDevServerCmd := &cobra.Command{
		Use:   "dev-server",
		Short: "Starts a Graylog DEV server",
		RunE:  runCommand,
	}

	runServicesCmd := &cobra.Command{
		Use:   "services",
		Short: "Starts services like MongoDB and Elasticsearch",
		RunE:  runCommand,
	}

	runCmd.AddCommand(runDevCmd)
	runCmd.AddCommand(runDevServerCmd)
	runCmd.AddCommand(runServicesCmd)

	RootCmd.AddCommand(runCmd)

	viper.BindPFlag("run.graylog.http-port", runCmd.PersistentFlags().Lookup("http-port"))
	viper.BindPFlag("run.elasticsearch.port", runCmd.PersistentFlags().Lookup("es-port"))
	viper.BindPFlag("run.mongodb.port", runCmd.PersistentFlags().Lookup("mongodb-port"))
}

func persistentPreRunCommand(cmd *cobra.Command, args []string) error {
	return runner.CheckSetup()
}

func runCommand(cmd *cobra.Command, args []string) error {
	return runner.DispatchCommand(runner.Config{
		Command: cmd.Name(),
		Graylog: runner.GraylogConfig{
			HTTPPort: viper.GetInt("run.graylog.http-port"),
		},
		MongoDB: runner.MongoDBConfig{
			Port: viper.GetInt("run.mongodb.port"),
		},
		Elasticsearch: runner.ElasticsearchConfig{
			Port: viper.GetInt("run.elasticsearch.port"),
		},
	})
}
