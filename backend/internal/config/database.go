package config

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"words-app/internal/entity"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewDatabase opens a database connection, runs AutoMigrate for all entities,
// seeds the admin user and frequency list on first startup.
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

	if err := db.AutoMigrate(
		&entity.User{},
		&entity.Site{},
		&entity.SiteCategory{},
		&entity.ParseJob{},
		&entity.ParsedPage{},
		&entity.DictWord{},
		&entity.WordOccurrence{},
		&entity.WordTranslation{},
		&entity.WordRelation{},
		&entity.ReviewItem{},
		&entity.LearnCategory{},
		&entity.LearnItem{},
		&entity.AppSetting{},
		&entity.Country{},
		&entity.State{},
		&entity.City{},
	); err != nil {
		log.Fatalf("Failed to auto-migrate : %+v", err)
	}

	seedAdminUser(db, v, log)
	seedFrequencyList(db, v, log)
	seedDefaultSettings(db, v, log)
	seedCountries(db, log)
	seedStates(db, log)
	seedCities(db, log)
	return db
}

func seedAdminUser(db *gorm.DB, v *viper.Viper, log *logrus.Logger) {
	email := strings.TrimSpace(v.GetString("auth.admin_email"))
	password := v.GetString("auth.admin_password")
	if email == "" || password == "" {
		log.Warn("Admin email/password not configured, skipping admin seed")
		return
	}

	var count int64
	if err := db.Model(&entity.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		log.Warnf("Failed to check admin user : %+v", err)
		return
	}
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Warnf("Failed to hash admin password : %+v", err)
		return
	}

	user := &entity.User{
		PublicID:     uuid.NewString(),
		Email:        email,
		PasswordHash: string(hash),
		Role:         "admin",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := db.Create(user).Error; err != nil {
		log.Warnf("Failed to seed admin user : %+v", err)
		return
	}
	log.Infof("Seeded admin user: %s", email)
}

// seedFrequencyList loads the frequency file and inserts each word with its rank.
// Line number (1-based) is the frequency rank — lower = more common.
// Already-existing lemmas are left untouched.
func seedFrequencyList(db *gorm.DB, v *viper.Viper, log *logrus.Logger) {
	path := v.GetString("frequency_seed.path")
	if path == "" {
		return
	}

	var existing int64
	if err := db.Model(&entity.DictWord{}).Where("frequency_rank IS NOT NULL").Count(&existing).Error; err != nil {
		log.Warnf("Failed to count existing dict words : %+v", err)
		return
	}
	if existing > 0 {
		return
	}

	f, err := os.Open(path)
	if err != nil {
		log.Warnf("Frequency file not found (%s) : %+v — skipping seed", path, err)
		return
	}
	defer f.Close()

	seen := make(map[string]bool)
	rank := 0
	batch := make([]entity.DictWord, 0, 500)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := db.Create(&batch).Error; err != nil {
			log.Warnf("Failed to insert frequency batch : %+v", err)
		}
		batch = batch[:0]
	}

	for scanner.Scan() {
		lemma := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if lemma == "" || seen[lemma] {
			continue
		}
		seen[lemma] = true
		rank++
		r := rank
		batch = append(batch, entity.DictWord{
			PublicID:      uuid.NewString(),
			Lemma:         lemma,
			Language:      "en",
			FrequencyRank: &r,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
		if len(batch) >= 500 {
			flush()
		}
	}
	flush()

	if err := scanner.Err(); err != nil {
		log.Warnf("Error reading frequency file : %+v", err)
		return
	}
	log.Infof("Seeded %d common words from frequency list", rank)
}

func openCSV(path string) (*csv.Reader, *os.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	r := csv.NewReader(f)
	r.LazyQuotes = true
	r.TrimLeadingSpace = true
	// skip header
	if _, err := r.Read(); err != nil {
		f.Close()
		return nil, nil, err
	}
	return r, f, nil
}

func seedCountries(db *gorm.DB, log *logrus.Logger) {
	var count int64
	db.Model(&entity.Country{}).Count(&count)
	if count > 0 {
		return
	}

	r, f, err := openCSV("./data/countries.csv")
	if err != nil {
		log.Warnf("Countries seed file not found : %+v — skipping", err)
		return
	}
	defer f.Close()

	n := 0
	batch := make([]entity.Country, 0, 250)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 2 {
			continue
		}
		batch = append(batch, entity.Country{
			Name: strings.TrimSpace(record[0]),
			Code: strings.TrimSpace(record[1]),
		})
		n++
		if len(batch) >= 250 {
			if err := db.Create(&batch).Error; err != nil {
				log.Warnf("Failed to insert countries batch : %+v", err)
			}
			batch = batch[:0]
		}
	}
	if len(batch) > 0 {
		if err := db.Create(&batch).Error; err != nil {
			log.Warnf("Failed to insert countries batch : %+v", err)
		}
	}
	log.Infof("Seeded %d countries", n)
}

func seedStates(db *gorm.DB, log *logrus.Logger) {
	var count int64
	db.Model(&entity.State{}).Count(&count)
	if count > 0 {
		return
	}

	r, f, err := openCSV("./data/us_states.csv")
	if err != nil {
		log.Warnf("States seed file not found : %+v — skipping", err)
		return
	}
	defer f.Close()

	n := 0
	batch := make([]entity.State, 0, 50)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 2 {
			continue
		}
		batch = append(batch, entity.State{
			Name: strings.TrimSpace(record[1]),
			Code: strings.TrimSpace(record[0]),
		})
		n++
	}
	if len(batch) > 0 {
		if err := db.Create(&batch).Error; err != nil {
			log.Warnf("Failed to insert states batch : %+v", err)
		}
	}
	log.Infof("Seeded %d states", n)
}

func seedCities(db *gorm.DB, log *logrus.Logger) {
	var count int64
	db.Model(&entity.City{}).Count(&count)
	if count > 0 {
		return
	}

	// CSV format: ID,STATE_CODE,STATE_NAME,CITY,COUNTY,LATITUDE,LONGITUDE
	r, f, err := openCSV("./data/us_cities.csv")
	if err != nil {
		log.Warnf("Cities seed file not found : %+v — skipping", err)
		return
	}
	defer f.Close()

	seen := make(map[string]bool)
	n := 0
	batch := make([]entity.City, 0, 500)

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := db.Create(&batch).Error; err != nil {
			log.Warnf("Failed to insert cities batch : %+v", err)
		}
		batch = batch[:0]
	}

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(record) < 4 {
			continue
		}
		name := strings.TrimSpace(record[3])
		key := strings.ToLower(name)
		if seen[key] {
			continue
		}
		seen[key] = true
		batch = append(batch, entity.City{
			Name: name,
		})
		n++
		if len(batch) >= 500 {
			flush()
		}
	}
	flush()
	log.Infof("Seeded %d unique cities", n)
}

func seedDefaultSettings(db *gorm.DB, v *viper.Viper, log *logrus.Logger) {
	defaults := map[string]string{
		entity.SettingDefaultTargetLang: v.GetString("parser.default_target_language"),
		entity.SettingDeepLEndpoint:     "https://api-free.deepl.com/v2/translate",
		entity.SettingRelationsSource:   "datamuse",
	}
	for key, val := range defaults {
		if val == "" {
			continue
		}
		var s entity.AppSetting
		err := db.Where("key = ?", key).First(&s).Error
		if err == nil {
			continue
		}
		setting := entity.AppSetting{Key: key, Value: val, UpdatedAt: time.Now()}
		if err := db.Create(&setting).Error; err != nil {
			log.Warnf("Failed to seed setting %s : %+v", key, err)
		}
	}
}
