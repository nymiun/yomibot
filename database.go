package main

import (
	"time"

	"github.com/nemphi/lavago"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type pgdb struct {
	*gorm.DB
}

func newDBConnection(dsn string) (_ *pgdb, err error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(cacheItem{}, historyItem{})
	return &pgdb{DB: db}, nil
}

type cacheItem struct {
	ID        string           `json:"id,omitempty" gorm:"default: uuid_generate_v4()"`
	SpotifyID string           `json:"spotifyId,omitempty"`
	Track     string           `json:"track,omitempty"`
	Timestamp time.Time        `json:"timestamp,omitempty"`
	Info      lavago.TrackInfo `json:"info,omitempty" gorm:"embedded"`
}

func (cacheItem) TableName() string {
	return "cache"
}

type historyItem struct {
	ID        string    `json:"id,omitempty" gorm:"default: uuid_generate_v4()"`
	GuildID   string    `json:"guildId,omitempty"`
	Track     string    `json:"track,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

func (historyItem) TableName() string {
	return "history"
}
