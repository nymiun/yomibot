package main

import (
	"time"

	"github.com/nemphi/lavago"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func NewDBConnection(dsn string) (_ *DB, err error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &DB{DB: db}, nil
}

type cacheItem struct {
	ID        string           `json:"id,omitempty"`
	SpotifyID string           `json:"spotifyId,omitempty"`
	Track     string           `json:"track,omitempty"`
	Timestamp time.Time        `json:"timestamp,omitempty"`
	Info      lavago.TrackInfo `json:"info,omitempty" gorm:"embedded"`
}

func (cacheItem) TableName() string {
	return "cache"
}

type historyItem struct {
	ID        string    `json:"id,omitempty"`
	GuildID   string    `json:"guildId,omitempty"`
	Track     string    `json:"track,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

func (historyItem) TableName() string {
	return "history"
}
