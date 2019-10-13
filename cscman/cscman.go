package cscman

import (
	"database/sql"
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
