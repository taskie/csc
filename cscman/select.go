package cscman

import (
	"context"

	"github.com/taskie/csc/cscman/models"
	"github.com/volatiletech/sqlboiler/queries/qm"
)

func (cm *CscMan) FindObjectBySha256Prefix(ctx context.Context, sha256Prefix string) ([]*models.Object, error) {
	fs, err := models.Objects(
		qm.Where(models.ObjectColumns.Sha256+" LIKE ?", sha256Prefix+"%"),
		qm.OrderBy(models.ObjectColumns.Sha256+","+models.ObjectColumns.Path)).All(ctx, cm.db)
	if err != nil {
		return nil, err
	}
	return fs, nil
}
func (cm *CscMan) FindObjectBySha256s(ctx context.Context, sha256s []string) ([]*models.Object, error) {
	sha256Interfaces := make([]interface{}, len(sha256s))
	for i, sha256 := range sha256s {
		sha256Interfaces[i] = sha256
	}
	fs, err := models.Objects(
		qm.WhereIn(models.ObjectColumns.Sha256+" IN ?", sha256Interfaces...),
		qm.OrderBy(models.ObjectColumns.Sha256+","+models.ObjectColumns.Path)).All(ctx, cm.db)
	if err != nil {
		return nil, err
	}
	return fs, nil
}
