package cscman

import (
	"context"
	"path/filepath"

	"github.com/iancoleman/strcase"
	"github.com/k0kubun/pp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/taskie/csc/cscman"
	"github.com/taskie/osplus"
	"github.com/volatiletech/sqlboiler/boil"
)

type Config struct {
	User, Password, Host, Database, LogLevel string
}

var configFile string
var config Config
var (
	verbose, debug, version bool
)

func (c *Config) DBPath() string {
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ")/" + c.Database + "?charset=utf8mb4&collation=utf8mb4_bin&parseTime=true"
}

func prepare() (context.Context, *cscman.CscMan) {
	ctx := context.Background()

	config := cscman.CscManConfig{
		DBPath: config.DBPath(),
	}
	logrus.Info(config.DBPath)
	cm, err := cscman.NewCscMan(&config)
	if err != nil {
		logrus.Fatal(err)
	}
	return ctx, cm
}

func register(cmd *cobra.Command, args []string) {
	ctx, cm := prepare()
	defer cm.Close()
	name := args[0]
	url := args[1]
	err := cm.RegisterNamespace(ctx, name, url)
	if err != nil {
		logrus.Fatal(err)
	}
}

const RegisterCommandName = "register"

var RegisterCommand = &cobra.Command{
	Use:  RegisterCommandName + " NAME URL",
	Args: cobra.ExactArgs(2),
	Run:  register,
}

func sync(cmd *cobra.Command, args []string) {
	ctx, cm := prepare()
	defer cm.Close()
	name := args[0]
	namespace, err := cm.FindNamespace(ctx, name)
	if err != nil {
		logrus.Fatal(err)
	}
	err = cm.SyncWithCSCDB(ctx, namespace)
	if err != nil {
		logrus.Fatal(err)
	}
}

const SyncCommandName = "sync"

var SyncCommand = &cobra.Command{
	Use:  SyncCommandName + " NAME",
	Args: cobra.ExactArgs(1),
	Run:  sync,
}

const CommandName = "cscman"

var Command = &cobra.Command{
	Use: CommandName,
}

func init() {
	Command.AddCommand(RegisterCommand, SyncCommand)
	Command.PersistentFlags().StringVarP(&configFile, "config", "c", "", `config file (default "`+CommandName+`.yml")`)
	Command.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	Command.PersistentFlags().BoolVar(&debug, "debug", false, "debug output")
	Command.PersistentFlags().BoolVarP(&version, "version", "V", false, "show Version")
	Command.Flags().StringP("user", "u", "", "user name")
	Command.Flags().StringP("password", "p", "", "password")
	Command.Flags().StringP("host", "H", "localhost", "database host")
	Command.Flags().StringP("database", "d", "cscman", "database name")

	for _, s := range []string{"user", "password", "host", "database"} {
		envKey := strcase.ToSnake(s)
		structKey := strcase.ToCamel(s)
		viper.BindPFlag(envKey, Command.Flags().Lookup(s))
		viper.RegisterAlias(structKey, envKey)
	}

	cobra.OnInitialize(initConfig)
}

func initConfig() {
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else if verbose {
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName(CommandName)
		conf, err := osplus.GetXdgConfigHome()
		if err != nil {
			logrus.Info(err)
		} else {
			viper.AddConfigPath(filepath.Join(conf, CommandName))
		}
		viper.AddConfigPath(".")
	}
	viper.SetEnvPrefix(CommandName)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		logrus.Debug(err)
	}
	err = viper.Unmarshal(&config)
	if err != nil {
		logrus.Warn(err)
	}
}

func Main() {
	if config.LogLevel != "" {
		lv, err := logrus.ParseLevel(config.LogLevel)
		if err != nil {
			logrus.Warn(err)
		} else {
			logrus.SetLevel(lv)
		}
	}
	if debug {
		if viper.ConfigFileUsed() != "" {
			logrus.Debugf("Using config file: %s", viper.ConfigFileUsed())
		}
		logrus.Debug(pp.Sprint(config))
		boil.DebugMode = true
	}
	Command.Execute()
}
