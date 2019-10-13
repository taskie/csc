package csc

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/k0kubun/pp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/taskie/csc"
	"github.com/taskie/csc/models"
	"github.com/taskie/osplus"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type Config struct {
	LogLevel string
	AbsMode  bool
}

var configFile string
var config Config
var (
	verbose, debug, version bool
)

func buildWalkFunc(ctx context.Context, db *sql.DB, basePath string) func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		if filepath.Base(path) == "csc.db" {
			return nil
		}
		var dbPath string
		if config.AbsMode {
			dbPath, err = filepath.Abs(path)
		} else {
			dbPath, err = filepath.Rel(basePath, path)
		}
		if err != nil {
			return err
		}
		mtime := info.ModTime()
		size := info.Size()

		qs := []qm.QueryMod{
			qm.Where(models.ObjectColumns.Path+" = ?", dbPath),
		}
		f, err := models.Objects(qs...).One(ctx, db)
		if err == nil {
			if f.Size == -1 {
				q := qm.WhereIn(models.ObjectColumns.ID+" = ?", f.ID)
				logrus.Debugf("Updating (size): %s", dbPath)
				n, err := models.Objects(q).UpdateAll(ctx, db, map[string]interface{}{
					models.ObjectColumns.Size: size,
				})
				if n != 1 {
					logrus.Warn("invalid number of updated records: " + string(n))
				}
				if err != nil {
					return err
				}
				logrus.Debugf("Updated (size): %s", dbPath)
			}
			if f.Mtime != mtime {
				sha256Hex, err := csc.CalcSha256HexString(path)
				if err != nil {
					return err
				}
				if f.Sha256 != sha256Hex {
					q := qm.WhereIn(models.ObjectColumns.ID+" = ?", f.ID)
					logrus.Debugf("Updating: %s", dbPath)
					n, err := models.Objects(q).UpdateAll(ctx, db, map[string]interface{}{
						models.ObjectColumns.Type:   "b",
						models.ObjectColumns.Mtime:  mtime,
						models.ObjectColumns.Size:   size,
						models.ObjectColumns.Sha256: sha256Hex,
					})
					if n != 1 {
						logrus.Warn("invalid number of updated records: " + string(n))
					}
					if err != nil {
						return err
					}
					logrus.Infof("Updated: %s", dbPath)
					return nil
				}
			}
		} else {
			sha256Hex, err := csc.CalcSha256HexString(path)
			if err != nil {
				return err
			}
			f = &models.Object{
				Path:      dbPath,
				Type:      "b",
				Mtime:     mtime,
				Size:      size,
				Sha256:    sha256Hex,
				Status:    "ok",
				UpdatedAt: time.Now(),
			}
			logrus.Debugf("Inserting: %s", dbPath)
			err = f.Insert(ctx, db, boil.Infer())
			if err != nil {
				// do nothing
				// return err
			}
			logrus.Infof("Inserted: %s", dbPath)
			return nil
		}
		return nil
	}
}

const initSQL = `CREATE TABLE objects (
	id INTEGER PRIMARY KEY,
	path TEXT UNIQUE NOT NULL,
	type TEXT NOT NULL,
	size INTEGER NOT NULL,
	mtime DATETIME NOT NULL,
	sha256 TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);
CREATE INDEX objects_path ON objects (path); 
CREATE INDEX objects_sha256_path ON objects (sha256, path);
CREATE INDEX objects_mtime ON objects (mtime);
CREATE INDEX objects_updated_at ON objects (updated_at);
`

func prepare() (context.Context, *sql.DB) {
	ctx := context.Background()
	dbName := "csc.db"
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		logrus.Fatal(err)
	}
	return ctx, db
}

func scan(cmd *cobra.Command, args []string) {
	ctx, db := prepare()
	defer db.Close()

	_, err := models.Objects().Count(ctx, db)
	if err != nil {
		_, err := db.Exec(initSQL)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	for _, arg := range args {
		err = filepath.Walk(arg, buildWalkFunc(ctx, db, arg))
		if err != nil {
			logrus.Fatal(err)
		}
	}
}

func sha256(cmd *cobra.Command, args []string) {
	ctx, db := prepare()
	defer db.Close()

	for _, arg := range args {
		fs, err := models.Objects(
			qm.Where(models.ObjectColumns.Sha256+" LIKE ?", arg+"%"),
			qm.OrderBy(models.ObjectColumns.Sha256+","+models.ObjectColumns.Path)).All(ctx, db)
		if err != nil {
			logrus.Fatal(err)
		}
		for _, f := range fs {
			fmt.Printf("%s\t%s\n", f.Sha256, f.Path)
		}
	}
}

func path(cmd *cobra.Command, args []string) {
	ctx, db := prepare()
	defer db.Close()

	for _, arg := range args {
		fs, err := models.Objects(
			qm.Where(models.ObjectColumns.Path+" LIKE ?", arg+"%"),
			qm.OrderBy(models.ObjectColumns.Path)).All(ctx, db)
		if err != nil {
			logrus.Fatal(err)
		}
		for _, f := range fs {
			fmt.Printf("%s\t%s\n", f.Sha256, f.Path)
		}
	}
}

func find(cmd *cobra.Command, args []string) {
	ctx, db := prepare()
	defer db.Close()

	sha256hexs := make([]interface{}, 0)
	for _, arg := range args {
		sha256hex, err := csc.CalcSha256HexString(arg)
		if err != nil {
			logrus.Error(err)
			continue
		}
		sha256hexs = append(sha256hexs, sha256hex)
	}
	fs, err := models.Objects(
		qm.WhereIn(models.ObjectColumns.Sha256+" IN ?", sha256hexs...),
		qm.OrderBy(models.ObjectColumns.Path)).All(ctx, db)
	if err != nil {
		logrus.Fatal(err)
	}
	for _, f := range fs {
		fmt.Printf("%s\t%s\n", f.Sha256, f.Path)
	}
}

const ScanCommandName = "scan"

var ScanCommand = &cobra.Command{
	Use:  ScanCommandName,
	Args: cobra.ArbitraryArgs,
	Run:  scan,
}

const Sha256CommandName = "sha256"

var Sha256Command = &cobra.Command{
	Use:  Sha256CommandName,
	Args: cobra.ArbitraryArgs,
	Run:  sha256,
}

const PathCommandName = "path"

var PathCommand = &cobra.Command{
	Use:  PathCommandName,
	Args: cobra.ArbitraryArgs,
	Run:  path,
}

const FindCommandName = "find"

var FindCommand = &cobra.Command{
	Use:  FindCommandName,
	Args: cobra.ArbitraryArgs,
	Run:  find,
}

const CommandName = "csc"

var Command = &cobra.Command{
	Use: CommandName,
}

func init() {
	Command.AddCommand(ScanCommand, Sha256Command, PathCommand, FindCommand)
	Command.PersistentFlags().StringVarP(&configFile, "config", "c", "", `config file (default "`+CommandName+`.yml")`)
	Command.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	Command.PersistentFlags().BoolVar(&debug, "debug", false, "debug output")
	Command.PersistentFlags().BoolVarP(&version, "version", "V", false, "show Version")
	Command.Flags().BoolP("abs-mode", "A", false, "absolute path mode")

	for _, s := range []string{"abs-mode"} {
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
