package cmd

import (
	"github.com/Graylog2/graylog-project-cli/logger"
	"github.com/Graylog2/graylog-project-cli/maven"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var mavenRunCmd = &cobra.Command{
	Use:     "maven-run",
	Aliases: []string{"mr"},
	Short:   "Build and run the Graylog server",
	Long: `This runs the maven build and starts the server afterwards.

Example:
  cd /path/to/graylog-project-internal
  graylog-project maven-run`,
	Run: mavenRunCommand,
}

func init() {
	RootCmd.AddCommand(mavenRunCmd)

	mavenRunCmd.Flags().BoolP("skip-run", "R", false, "Skip running the server after the build")
	mavenRunCmd.Flags().BoolP("skip-web", "W", false, "Skip building the web interface")
	mavenRunCmd.Flags().BoolP("clean", "c", false, "Run maven clean before the build")
	mavenRunCmd.Flags().String("maven-bin", "", "Path to maven binary")
	mavenRunCmd.Flags().IntP("http-port", "H", 9001, "Graylog server HTTP port") // TODO: Use 9200 as default
	mavenRunCmd.Flags().IntP("es-port", "E", 9220, "Elasticsearch port")         // TODO: Use 9200 as default
	mavenRunCmd.Flags().IntP("mongodb-port", "m", 27027, "MongoDB port")         // TODO: Use 27017 as default

	viper.BindPFlag("maven.build.clean", mavenRunCmd.Flags().Lookup("clean"))
	viper.BindPFlag("maven.build.skip-web", mavenRunCmd.Flags().Lookup("skip-web"))
	viper.BindPFlag("maven.bin", mavenRunCmd.Flags().Lookup("maven-bin"))
	viper.BindPFlag("maven.run.skip", mavenRunCmd.Flags().Lookup("skip-run"))
	viper.BindPFlag("maven.run.http-port", mavenRunCmd.Flags().Lookup("http-port"))
	viper.BindPFlag("maven.run.es-port", mavenRunCmd.Flags().Lookup("es-port"))
	viper.BindPFlag("maven.run.mongodb-port", mavenRunCmd.Flags().Lookup("mongodb-port"))
}

func mavenRunCommand(cmd *cobra.Command, args []string) {
	// TODO:
	//  - Create a work directory with a data/ dir and generate a node-id file
	//  - Create a default graylog.conf in the work directory with sane defaults (e.g. no versionchecks, lb timeout 0, etc)
	//  - Make ES and MongoDB ports configurable
	//  - Check if MongoDB and ES are running before starting the server
	workDir := "cli-run"         // TODO: Needs to be an option
	configFile := "graylog.conf" // TODO: Needs to be an option

	buildConfig := maven.BuildConfig{
		Clean:     viper.GetBool("maven.build.clean"),
		SkipWeb:   viper.GetBool("maven.build.skip-web"),
		MavenPath: viper.GetString("maven.bin"),
	}

	classPathFile, err := maven.BuildForRun(buildConfig)
	if err != nil {
		logger.Error("ERROR: %v", err)
		os.Exit(1)
	}

	if viper.GetBool("maven.run.skip") {
		return
	}

	runConfig := maven.RunConfig{
		HTTPPort:      viper.GetInt("maven.run.http-port"),
		ESPort:        viper.GetInt("maven.run.es-port"),
		MongoDBPort:   viper.GetInt("maven.run.mongodb-port"),
		WorkDir:       workDir,
		ConfigFile:    configFile,
		ClassPathFile: classPathFile,
	}

	if err := maven.RunServer(runConfig); err != nil {
		logger.Error("ERROR: %v", err)
		os.Exit(1)
	}
}
