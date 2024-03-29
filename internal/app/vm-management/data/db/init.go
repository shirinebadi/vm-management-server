package db

import (
	"errors"

	"github.com/shirinebadi/vm-management-server/internal/app/vm-management/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("HttpMDB.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func migrate(db *gorm.DB) error {
	err := db.AutoMigrate(&model.User{})
	return err
}

func Init() (*gorm.DB, error) {
	db, err := NewDB()
	if err != nil {
		return nil, errors.New("Error in DB Creation")
	}

	if err = migrate(db); err != nil {
		return nil, errors.New("Error in DB Creation" + err.Error())
	}
	return db, nil
}
