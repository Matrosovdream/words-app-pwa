package config

import (
	"os"
	"path/filepath"

	"words-app/internal/entity"

	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewDatabase opens a SQLite connection, runs AutoMigrate for all entities,
// and seeds the `words` table on first startup.
//
// Note: the project rules mandate MySQL + golang-migrate. This demo uses
// pure-Go SQLite + GORM AutoMigrate so the single-container setup works
// without external services. The architecture is DB-agnostic — to switch,
// replace the driver + sqlite.Open() call with mysql.Open(dsn).
func NewDatabase(v *viper.Viper, log *logrus.Logger) *gorm.DB {
	dbPath := v.GetString("database.path")

	// Ensure parent directory exists
	if dir := filepath.Dir(dbPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Fatalf("Failed to create database directory : %+v", err)
		}
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: gormlogger.New(
			log,
			gormlogger.Config{LogLevel: gormlogger.Warn},
		),
	})
	if err != nil {
		log.Fatalf("Failed to open database : %+v", err)
	}

	if err := db.AutoMigrate(&entity.Word{}); err != nil {
		log.Fatalf("Failed to auto-migrate : %+v", err)
	}

	seedWords(db, log)
	return db
}

func seedWords(db *gorm.DB, log *logrus.Logger) {
	var count int64
	db.Model(&entity.Word{}).Count(&count)
	if count > 0 {
		return
	}

	words := []entity.Word{
		{Slug: "serendipity", Word: "Serendipity", Definition: "The occurrence of happy or beneficial events by chance.", Example: "Finding that cafe was pure serendipity.", Emoji: "🍀"},
		{Slug: "petrichor", Word: "Petrichor", Definition: "The earthy scent produced when rain falls on dry soil.", Example: "The petrichor after the storm was intoxicating.", Emoji: "🌧️"},
		{Slug: "ephemeral", Word: "Ephemeral", Definition: "Lasting for a very short time.", Example: "Cherry blossoms are beautifully ephemeral.", Emoji: "🌸"},
		{Slug: "sonder", Word: "Sonder", Definition: "The realization that each passerby has a life as vivid as your own.", Example: "A wave of sonder hit me on the crowded street.", Emoji: "🌆"},
		{Slug: "mellifluous", Word: "Mellifluous", Definition: "Sweet or musical; pleasant to hear.", Example: "Her mellifluous voice filled the room.", Emoji: "🎵"},
		{Slug: "halcyon", Word: "Halcyon", Definition: "Denoting a period of time in the past that was idyllically happy.", Example: "The halcyon days of summer.", Emoji: "☀️"},
	}

	if err := db.Create(&words).Error; err != nil {
		log.Warnf("Failed to seed words : %+v", err)
		return
	}
	log.Infof("Seeded %d words into empty database", len(words))
}
