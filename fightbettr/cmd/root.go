package cmd

import (
	"fmt"
	"log"
	"time"

	"fightbettr.com/fightbettr/pkg/logger"
	"fightbettr.com/fightbettr/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
)

var (
	cfgPath string
)

// rootCmd is the main Cobra command representing the root of the Fightbettr service.
var rootCmd = &cobra.Command{
	Use:   "Fightbettr Service",
	Short: "This CLI works with data to manage and redirect it",
	RunE: func(cmd *cobra.Command, args []string) error {
		showVersion, _ := cmd.Flags().GetBool("version")

		if showVersion {
			fmt.Println("Dev version", version.DevVersion)
			return nil
		}

		return cmd.Usage()
	},
}

// Execute runs the root command for the Fightbettr service.
// It executes the necessary logic for the command-line interface,
// handling errors and logging them if they occur.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "Config file path (default is ./config.yaml)")
	rootCmd.PersistentFlags().String("name", version.Name, "Application name label")
	rootCmd.PersistentFlags().Bool("log_json", false, "Enable JSON formatted logs output")
	rootCmd.PersistentFlags().Int("log_level", int(zapcore.DebugLevel), "Log level")

	bindViperPersistentFlag(rootCmd, "config_path", "config")
	bindViperPersistentFlag(rootCmd, "app.name", "name")
	bindViperPersistentFlag(rootCmd, "log_json", "log_json")
	bindViperPersistentFlag(rootCmd, "log_level", "log_level")

	err := initZapLogger()
	if err != nil {
		log.Fatalf("error while logger initializing: %s", err)
	}
}

// initZapLogger initializes the zap logger.
func initZapLogger() error {
	return logger.Init(zapcore.DebugLevel, "logs/log.json")
}

// initConfig initializes the service configuration.
// It sets default values, reads from environment variables, and reads from a config file if present.
func initConfig() {
	setConfigDefaults()

	if cfgPath != "" {
		viper.SetConfigFile(cfgPath)
	} else {
		viper.AddConfigPath("./configs")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// setConfigDefaults sets default values for various configuration options.
func setConfigDefaults() {
	// app defaults
	viper.SetDefault("app.env", "dev")
	viper.SetDefault("app.name", version.Name)
	viper.SetDefault("app.version", version.DevVersion)
	viper.SetDefault("app.run_date", time.Unix(version.RunDate, 0).Format(time.RFC1123))

	// http server
	viper.SetDefault("http.addr", "127.0.0.1:9091")
	viper.SetDefault("http.port", "9091")
	viper.SetDefault("http.ssl.enabled", false)

	// auth config
	viper.SetDefault("auth.cookie_name", "fb_api_token")
	viper.SetDefault("auth.jwt.cert", "")
	viper.SetDefault("auth.jwt.key", "")
}

// bindViperPersistentFlag binds a Viper configuration flag to a persistent Cobra command flag.
func bindViperPersistentFlag(cmd *cobra.Command, viperVal, flagName string) {
	if err := viper.BindPFlag(viperVal, cmd.PersistentFlags().Lookup(flagName)); err != nil {
		log.Printf("Failed to bind viper flag: %s", err)
	}
}
