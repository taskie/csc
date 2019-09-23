package csc

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/taskie/csc"
	"github.com/taskie/csc/models"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

func Main() {
	// logrus.SetLevel(logrus.DebugLevel)
	// boil.DebugMode = true

	dbName := "csc.db"
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()

	if len(os.Args) <= 1 {
		logrus.Fatal("please specify subcommand and args")
	}
	mode := os.Args[1]
	args := os.Args[2:]

	switch mode {
	case "sha256":
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
	case "path":
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
	case "find":
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
	default:
		logrus.Fatalf("invalid mode: %s", mode)
	}
}
