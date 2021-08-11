package main

// import (
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// type DB struct {
// 	*gorm.DB
// }

// func NewDBConnection(dsn string) (_ *DB, err error) {
// 	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &DB{DB: db}, nil
// }
