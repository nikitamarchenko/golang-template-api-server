package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/nikitamarchenko/golang-template-api-server/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	argHTTPServerPort                      = "http-port"
	argHTTPReadinessProbePeriodSeconds     = "http-readiness-probe-period-seconds"
	argHTTPReadinessProbePeriodSecondsDesc = `value from k8s pod readinessProbe.periodSeconds
link https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#Probe`
	argAllowRootUser = "allow-root-user"
)

// serverCmd represents the server command.
var serverCmd = &cobra.Command{ //nolint
	Use:   "server",
	Short: "",
	Long:  ``,
	RunE: func(_ *cobra.Command, _ []string) error {
		logLevel := viper.GetString("log-level")
		sLoglevel := slog.LevelInfo
		err := sLoglevel.UnmarshalText([]byte(logLevel))
		if err != nil {
			return fmt.Errorf("run: %w", err)
		}
		logOpt := slog.HandlerOptions{
			Level:       sLoglevel,
			AddSource:   viper.GetBool(argGlobalLogShowSource),
			ReplaceAttr: nil,
		}
		logger := slog.New(slog.NewJSONHandler(os.Stdout, &logOpt))
		config := server.Config{
			Port:                            viper.GetInt(argHTTPServerPort),
			HTTPReadinessProbePeriodSeconds: viper.GetInt(argHTTPReadinessProbePeriodSeconds),
			LogLevel:                        sLoglevel,
		}

		return server.Run(logger, config, viper.GetBool(argAllowRootUser))
	},
}

func init() { //nolint
	err := serverInitRun()
	if err != nil {
		fmt.Printf("Error: cmd.server.init: %v", err) //nolint
		os.Exit(1)
	}
}

func serverInitRun() error {
	flags := serverCmd.Flags()

	// Port
	flags.Int(argHTTPServerPort, 8080, "server port")

	err := viper.BindPFlag(argHTTPServerPort, flags.Lookup(argHTTPServerPort))
	if err != nil {
		return err //nolint:wrapcheck // don't need wrap here
	}

	// readinessPprobe.periodSeconds
	flags.Int(argHTTPReadinessProbePeriodSeconds, 10,
		argHTTPReadinessProbePeriodSecondsDesc)

	err = viper.BindPFlag(argHTTPReadinessProbePeriodSeconds,
		flags.Lookup(argHTTPReadinessProbePeriodSeconds))
	if err != nil {
		return err //nolint:wrapcheck // don't need wrap here
	}

	// Allow root user
	flags.Bool(argAllowRootUser, false, "allow run server as root user")

	err = viper.BindPFlag(argAllowRootUser, flags.Lookup(argAllowRootUser))
	if err != nil {
		return err //nolint:wrapcheck // don't need wrap here
	}

	rootCmd.AddCommand(serverCmd)

	return nil
}
