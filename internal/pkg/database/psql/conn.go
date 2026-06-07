package psql

import (
	"fmt"
	"time"

	"github.com/parxyws/cozybox/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func InitPostgres(cfg *config.Config) (*gorm.DB, error) {
	writeDsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.WriteDB.User, cfg.WriteDB.Password, cfg.WriteDB.Host, cfg.WriteDB.Port, cfg.WriteDB.NameDB)

	db, err := gorm.Open(postgres.Open(writeDsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to write database: %w", err)
	}

	readDsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.ReadDB.User, cfg.ReadDB.Password, cfg.ReadDB.Host, cfg.ReadDB.Port, cfg.ReadDB.NameDB)

	err = db.Use(
		dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{postgres.Open(readDsn)},
			Policy:   dbresolver.RandomPolicy{},
		}).
			SetMaxIdleConns(10).
			SetMaxOpenConns(100).
			SetConnMaxLifetime(time.Hour),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize db resolver plugin: %w", err)
	}

	return db, nil
}
