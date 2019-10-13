package cscman

import (
	"context"
	"database/sql"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/taskie/csc"
	"github.com/taskie/csc/cscman/models"
	cscModels "github.com/taskie/csc/models"
	"github.com/volatiletech/sqlboiler/boil"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

type CscManConfig struct {
	DBPath string `json:"db_path"`
}

type CscMan struct {
	config *CscManConfig
	db     *sql.DB
}

func NewCscMan(config *CscManConfig) (*CscMan, error) {
	db, err := sql.Open("mysql", config.DBPath)
	if err != nil {
		return nil, err
	}
	return &CscMan{
		config: config,
		db:     db,
	}, nil
}

func (cm *CscMan) Close() error {
	return cm.db.Close()
}

func (cm *CscMan) Rsync(ctx context.Context, remote string) (string, error) {
	f, err := ioutil.TempFile("", "csc.db")
	if err != nil {
		return "", err
	}
	err = f.Close()
	if err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, "rsync", "-vzP", remote, f.Name())
	logrus.Info(cmd)
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

func (cm *CscMan) RegisterNamespace(ctx context.Context, name string, url string) error {
	cscdbPath, err := cm.Rsync(ctx, url)
	if err != nil {
		return err
	}
	defer os.Remove(cscdbPath)

	fi, err := os.Stat(cscdbPath)
	if err != nil {
		return err
	}

	cscdbSize := fi.Size()
	cscdbMtime := fi.ModTime()
	namespace := models.Namespace{
		Name:        name,
		URL:         url,
		Type:        "",
		CSCDBSize:   cscdbSize,
		CSCDBMtime:  cscdbMtime,
		CSCDBSha256: "",
		Status:      "new",
		Description: "",
	}
	err = namespace.Insert(ctx, cm.db, boil.Infer())
	if err != nil {
		return err
	}
	err = cm.syncWithCSCDBImpl(ctx, &namespace, cscdbPath, fi)
	if err != nil {
		return err
	}
	return nil
}

func (cm *CscMan) FindNamespace(ctx context.Context, name string) (*models.Namespace, error) {
	return models.Namespaces(qm.Where(models.NamespaceColumns.Name+" = ?", name)).One(ctx, cm.db)
}

func (cm *CscMan) SyncWithCSCDB(ctx context.Context, namespace *models.Namespace) error {
	cscdbPath, err := cm.Rsync(ctx, namespace.URL)
	if err != nil {
		return err
	}
	defer os.Remove(cscdbPath)
	fi, err := os.Stat(cscdbPath)
	if err != nil {
		return err
	}
	return cm.syncWithCSCDBImpl(ctx, namespace, cscdbPath, fi)
}

func (cm *CscMan) syncWithCSCDBImpl(ctx context.Context, namespace *models.Namespace, cscdbPath string, fi os.FileInfo) error {
	cscdbSize := fi.Size()
	cscdbMtime := fi.ModTime()
	if namespace.Status != "new" && cscdbSize == namespace.CSCDBSize && cscdbMtime == namespace.CSCDBMtime {
		return nil
	}
	cscdbSha256, err := csc.CalcSha256HexString(cscdbPath)
	if err != nil {
		return err
	}
	if namespace.Status != "new" && cscdbSha256 == namespace.CSCDBSha256 {
		return nil
	}

	db, err := sql.Open("sqlite3", cscdbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	olds, err := models.Objects(qm.Where(models.ObjectColumns.Namespace+" = ?", namespace.Name)).All(ctx, cm.db)
	if err != nil {
		return err
	}
	oldMap := make(map[string]*models.Object)
	for _, old := range olds {
		oldMap[old.Path] = old
	}

	objs, err := cscModels.Objects().All(ctx, db)
	if err != nil {
		return err
	}

	for _, src := range objs {
		if old, ok := oldMap[src.Path]; ok {
			if old.Type != src.Type || old.Size != src.Size || old.Mtime != src.Mtime || old.Sha256 != src.Sha256 || old.Status != src.Status {
				old.Type = src.Type
				old.Size = src.Size
				old.Mtime = src.Mtime
				old.Sha256 = src.Sha256
				old.Status = src.Status
				_, err = old.Update(ctx, cm.db, boil.Infer())
				if err != nil {
					return err
				}
			}
		} else {
			dst := models.Object{
				Namespace: namespace.Name,
				Path:      src.Path,
				Type:      src.Type,
				Size:      src.Size,
				Mtime:     src.Mtime,
				Sha256:    src.Sha256,
				Status:    src.Status,
			}
			err = dst.Insert(ctx, cm.db, boil.Infer())
			if err != nil {
				return err
			}
		}
	}

	namespace.CSCDBSize = cscdbSize
	namespace.CSCDBMtime = cscdbMtime
	namespace.CSCDBSha256 = cscdbSha256
	namespace.Status = "ok"
	_, err = namespace.Update(ctx, cm.db, boil.Infer())
	if err != nil {
		return err
	}
	return nil
}
