package db

import (
	"devflow/internal/config"
	"devflow/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Open(cfg config.Config) (*gorm.DB, error) {
	database, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := database.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	if cfg.AutoMigrate {
		if err := database.AutoMigrate(
			&model.User{},
			&model.Post{},
			&model.Follow{},
			&model.Like{},
			&model.Favorite{},
			&model.Comment{},
			&model.Notification{},
			&model.ProcessedEvent{},
		); err != nil {
			return nil, err
		}
	}

	return database, nil
}
