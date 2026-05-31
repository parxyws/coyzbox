package psql

import (
	"fmt"
	"time"

	"github.com/parxyws/cozybox/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// InitPostgres initializes the GORM DB with a Read/Write split.
func InitPostgres(cfg *config.Config) (*gorm.DB, error) {
	var err error

	// Connect to Write (Source) Database
	writeDsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.WriteDB.User, cfg.WriteDB.Password, cfg.WriteDB.Host, cfg.WriteDB.Port, cfg.WriteDB.NameDB)

	db, err := gorm.Open(postgres.Open(writeDsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to write database: %w", err)
	}

	// Connect to Read (Replica) Database via DBResolver
	readDsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", cfg.ReadDB.User, cfg.ReadDB.Password, cfg.ReadDB.Host, cfg.ReadDB.Port, cfg.ReadDB.NameDB)

	// Configure DBResolver Plugin
	err = db.Use(
		dbresolver.Register(dbresolver.Config{
			Replicas: []gorm.Dialector{postgres.Open(readDsn)},
			Policy:   dbresolver.RandomPolicy{}, // Use RandomPolicy for load balancing across multiple replicas
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
