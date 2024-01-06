package configer

import (
	"github.com/c1emon/gcommon/logx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mapstructure
// bootstrap.yaml
type Application struct {
	Name     string
	Debug    bool
	LogLevel logx.Level
}

func (Application) name() string {
	return "application"
}

type Command cobra.Command

type cmd struct {
	cobraCmd *cobra.Command
}

func (c *cmd) Add(command *cmd) *cmd {
	c.cobraCmd.AddCommand(nil)
	return command
}

func WithShorthand(shorthand string) {

}

func CreateFlag(cmd *cobra.Command, name string) {

	cmd.PersistentFlags().IntP("port", "p", 8080, "server port")
	// cmd.PersistentFlags().
	viper.BindPFlag("server.port", cmd.PersistentFlags().Lookup("port"))
}

// func AddFlag[T any](v *T, name string) {

// }
