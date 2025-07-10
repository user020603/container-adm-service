package infrastructure

import (
	"fmt"
	"thanhnt208/container-adm-service/config"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db  *gorm.DB
	cfg *config.Config
}

var NewDatabase = func(cfg *config.Config) IDatabase {
	return &Database{
		cfg: cfg,
	}
}

func (d *Database) ConnectDB() (*gorm.DB, error) {
	if d.db != nil {
		return d.db, nil
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		d.cfg.DBHost, d.cfg.DBUser, d.cfg.DBPassword, d.cfg.DBName, d.cfg.DBPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	d.db = db
	return d.db, nil
}

func (d *Database) Close() error {
	if d.db == nil {
		return nil
	}

	sqlDB, err := d.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
	}

	return sqlDB.Close()
}