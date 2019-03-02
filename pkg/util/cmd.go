package util

import (
	"context"
	"reflect"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"arhat.dev/kube-host-pty/pkg/util/log"
	"arhat.dev/kube-host-pty/pkg/version"
)

type Command struct {
	Context context.Context
	Exit    context.CancelFunc
	cobra.Command
}

func DefaultCmd(name string, optFromConfigFile interface{}, onConfigChanged func(interface{}), run func(context.Context, context.CancelFunc) error) *Command {
	var (
		logLevel   string
		configFile string
	)

	ctx, exit := context.WithCancel(context.Background())

	cmd := &Command{
		Context: ctx,
		Exit:    exit,
		Command: cobra.Command{
			Use: name,
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				log.Setup(name, log.Level(logLevel))
				if configFile != "" && optFromConfigFile != nil {
					if err := Unmarshal(configFile, optFromConfigFile, yaml.Unmarshal); err != nil {
						log.E("unmarshal config file failed", log.Err(err), log.String("config_file", configFile))
						return err
					}
				}

				if onConfigChanged != nil {
					Workers.Add(func(func()) (interface{}, error) {
						newOptions := reflect.New(reflect.TypeOf(optFromConfigFile)).Elem().Interface()
						updateCh := NotifyWhenConfigChanged(configFile, newOptions, yaml.Unmarshal)
						for range updateCh {
							onConfigChanged(newOptions)
						}

						return nil, nil
					})
				}
				return nil
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				// the run func MUST NOT be null
				return run(ctx, exit)
			},
			PersistentPostRun: func(cmd *cobra.Command, args []string) {
				Workers.wait()
			},
			Version: version.Info(),
		},
	}

	cmd.SetVersionTemplate(`{{ printf "%s" .Version }}`)
	cmd.PersistentFlags().StringVar(&configFile, "config", "", "set path to config file")
	cmd.PersistentFlags().StringVar(&logLevel, "log", "error", "set log level, one of [debug, info, warning, error, fatal, panic]")

	return cmd
}
