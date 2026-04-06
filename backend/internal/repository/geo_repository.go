package repository

import (
	"words-app/internal/entity"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type GeoRepository struct {
	Log *logrus.Logger
}

func NewGeoRepository(log *logrus.Logger) *GeoRepository {
	return &GeoRepository{Log: log}
}

// IsGeoName returns true if the given lowercase word matches any country, state, or city name.
func (r *GeoRepository) IsGeoName(db *gorm.DB, word string) bool {
	var count int64
	db.Model(&entity.Country{}).Where("LOWER(name) = ?", word).Count(&count)
	if count > 0 {
		return true
	}
	db.Model(&entity.State{}).Where("LOWER(name) = ?", word).Count(&count)
	if count > 0 {
		return true
	}
	db.Model(&entity.City{}).Where("LOWER(name) = ?", word).Count(&count)
	return count > 0
}

// LoadAllGeoNames returns a set of all geographic names (lowercased) for in-memory filtering.
func (r *GeoRepository) LoadAllGeoNames(db *gorm.DB) map[string]bool {
	names := make(map[string]bool)

	var countries []entity.Country
	db.Find(&countries)
	for _, c := range countries {
		for _, w := range splitWords(c.Name) {
			names[w] = true
		}
	}

	var states []entity.State
	db.Find(&states)
	for _, s := range states {
		for _, w := range splitWords(s.Name) {
			names[w] = true
		}
	}

	var cities []entity.City
	db.Find(&cities)
	for _, c := range cities {
		for _, w := range splitWords(c.Name) {
			names[w] = true
		}
	}

	return names
}

// splitWords lowercases and splits a name into individual words.
func splitWords(name string) []string {
	words := make([]string, 0, 2)
	var current []byte
	for i := 0; i < len(name); i++ {
		b := name[i]
		if b >= 'A' && b <= 'Z' {
			current = append(current, b+32) // lowercase
		} else if b >= 'a' && b <= 'z' {
			current = append(current, b)
		} else {
			if len(current) > 0 {
				words = append(words, string(current))
				current = current[:0]
			}
		}
	}
	if len(current) > 0 {
		words = append(words, string(current))
	}
	return words
}
