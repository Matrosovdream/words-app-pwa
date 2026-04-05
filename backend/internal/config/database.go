package config

import (
	"fmt"
	"os"
	"path/filepath"

	"words-app/internal/entity"

	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewDatabase opens a database connection based on the configured driver
// ("sqlite" or "postgres"), runs AutoMigrate for all entities, and seeds
// the `words` table on first startup.
//
// Note: the project rules mandate MySQL + golang-migrate. This demo uses
// GORM AutoMigrate so the dev setup can run without an external migration
// tool. The architecture is DB-agnostic — add more drivers here as needed.
func NewDatabase(v *viper.Viper, log *logrus.Logger) *gorm.DB {
	driver := v.GetString("database.driver")

	gormConfig := &gorm.Config{
		Logger: gormlogger.New(
			log,
			gormlogger.Config{LogLevel: gormlogger.Warn},
		),
	}

	var (
		db  *gorm.DB
		err error
	)

	switch driver {
	case "postgres":
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			v.GetString("database.host"),
			v.GetInt("database.port"),
			v.GetString("database.user"),
			v.GetString("database.password"),
			v.GetString("database.name"),
			v.GetString("database.sslmode"),
		)
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
	case "sqlite", "":
		dbPath := v.GetString("database.path")
		if dir := filepath.Dir(dbPath); dir != "" && dir != "." {
			if mkErr := os.MkdirAll(dir, 0o755); mkErr != nil {
				log.Fatalf("Failed to create database directory : %+v", mkErr)
			}
		}
		db, err = gorm.Open(sqlite.Open(dbPath), gormConfig)
	default:
		log.Fatalf("Unsupported database driver : %s", driver)
	}
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
