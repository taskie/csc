package cscman

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/taskie/csc"
	"github.com/taskie/csc/models"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

var absMode = true

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
		if absMode {
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

	_, err = models.Objects().Count(ctx, db)
	if err != nil {
		_, err := db.Exec(initSQL)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	for _, arg := range os.Args[1:] {
		err = filepath.Walk(arg, buildWalkFunc(ctx, db, arg))
		if err != nil {
			logrus.Fatal(err)
		}
	}
}
