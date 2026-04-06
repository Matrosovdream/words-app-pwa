package usecase

import (
	"context"
	"errors"
	"time"

	"words-app/internal/entity"
	"words-app/internal/gateway/dictionary"
	"words-app/internal/gateway/translation"
	"words-app/internal/model"
	"words-app/internal/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type WordUseCase struct {
	DB                    *gorm.DB
	Log                   *logrus.Logger
	DictWordRepository    *repository.DictWordRepository
	TranslationRepository *repository.WordTranslationRepository
	RelationRepository    *repository.WordRelationRepository
	OccurrenceRepository  *repository.WordOccurrenceRepository
	PageRepository        *repository.ParsedPageRepository
	SettingRepository     *repository.AppSettingRepository
	Relations             dictionary.RelationsClient
}

func NewWordUseCase(db *gorm.DB, log *logrus.Logger,
	dictRepo *repository.DictWordRepository,
	transRepo *repository.WordTranslationRepository,
	relRepo *repository.WordRelationRepository,
	occRepo *repository.WordOccurrenceRepository,
	pageRepo *repository.ParsedPageRepository,
	settingRepo *repository.AppSettingRepository,
	relations dictionary.RelationsClient) *WordUseCase {
	return &WordUseCase{
		DB:                    db,
		Log:                   log,
		DictWordRepository:    dictRepo,
		TranslationRepository: transRepo,
		RelationRepository:    relRepo,
		OccurrenceRepository:  occRepo,
		PageRepository:        pageRepo,
		SettingRepository:     settingRepo,
		Relations:             relations,
	}
}

func (c *WordUseCase) getDefaultTargetLang(tx *gorm.DB) string {
	val, err := c.SettingRepository.Get(tx, entity.SettingDefaultTargetLang)
	if err != nil || val == "" {
		return "ru"
	}
	return val
}

func (c *WordUseCase) getDeepL(tx *gorm.DB) *translation.DeepLClient {
	apiKey, _ := c.SettingRepository.Get(tx, entity.SettingDeepLAPIKey)
	endpoint, _ := c.SettingRepository.Get(tx, entity.SettingDeepLEndpoint)
	return translation.NewDeepLClient(c.Log, apiKey, endpoint)
}

// Detail returns a fully-populated word view, lazy-fetching translations and
// synonyms/antonyms on first access and caching them in the DB.
func (c *WordUseCase) Detail(ctx context.Context, publicID string, targetLang string) (*model.WordDetailResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	word := new(entity.DictWord)
	if err := c.DictWordRepository.FindByPublicID(tx, word, publicID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.ErrNotFound
		}
		c.Log.Warnf("Failed find word : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	if targetLang == "" {
		targetLang = c.getDefaultTargetLang(tx)
	}

	// ensure translation exists for targetLang
	var translations []entity.WordTranslation
	if err := c.TranslationRepository.FindByWordAndLang(tx, &translations, word.ID, targetLang); err != nil {
		c.Log.Warnf("Failed load translations : %+v", err)
	}
	if len(translations) == 0 {
		// lazy fetch via DeepL
		dl := c.getDeepL(tx)
		if dl.APIKey != "" {
			text, err := dl.Translate(ctx, word.Lemma, word.Language, targetLang)
			if err != nil {
				c.Log.Warnf("DeepL translate failed for %s : %+v", word.Lemma, err)
			} else if text != "" {
				tr := &entity.WordTranslation{
					PublicID:       uuid.NewString(),
					WordID:         word.ID,
					TargetLanguage: targetLang,
					Translation:    text,
					Source:         "deepl",
					IsPrimary:      true,
					FetchedAt:      time.Now(),
				}
				if err := c.TranslationRepository.Create(tx, tr); err != nil {
					c.Log.Warnf("Failed to cache translation : %+v", err)
				} else {
					translations = append(translations, *tr)
				}
			}
		} else {
			c.Log.Info("DeepL api key not configured, skipping translation fetch")
		}
	}

	// ensure relations exist
	var relations []entity.WordRelation
	if err := c.RelationRepository.FindByWord(tx, &relations, word.ID); err != nil {
		c.Log.Warnf("Failed load relations : %+v", err)
	}
	if len(relations) == 0 && c.Relations != nil && word.Language == "en" {
		syns, _ := c.Relations.Synonyms(ctx, word.Lemma)
		ants, _ := c.Relations.Antonyms(ctx, word.Lemma)
		for _, s := range syns {
			rel := &entity.WordRelation{
				PublicID:     uuid.NewString(),
				WordID:       word.ID,
				RelatedText:  s,
				RelationType: entity.RelationTypeSynonym,
				Source:       c.Relations.Name(),
				FetchedAt:    time.Now(),
			}
			if err := c.RelationRepository.Create(tx, rel); err == nil {
				relations = append(relations, *rel)
			}
		}
		for _, a := range ants {
			rel := &entity.WordRelation{
				PublicID:     uuid.NewString(),
				WordID:       word.ID,
				RelatedText:  a,
				RelationType: entity.RelationTypeAntonym,
				Source:       c.Relations.Name(),
				FetchedAt:    time.Now(),
			}
			if err := c.RelationRepository.Create(tx, rel); err == nil {
				relations = append(relations, *rel)
			}
		}
	}

	// occurrences
	var occs []entity.WordOccurrence
	_ = c.OccurrenceRepository.FindByWord(tx, &occs, word.ID, 5)

	resp := &model.WordDetailResponse{
		ID:            word.PublicID,
		Lemma:         word.Lemma,
		Language:      word.Language,
		POS:           word.POS,
		FrequencyRank: word.FrequencyRank,
		IPA:           word.IPA,
		Definition:    word.Definition,
		Translations:  make([]model.WordTranslationDTO, 0, len(translations)),
		Synonyms:      []model.WordRelationDTO{},
		Antonyms:      []model.WordRelationDTO{},
		Occurrences:   make([]model.WordOccurrenceDTO, 0, len(occs)),
	}
	for _, t := range translations {
		resp.Translations = append(resp.Translations, model.WordTranslationDTO{
			Language:    t.TargetLanguage,
			Translation: t.Translation,
			Source:      t.Source,
			IsPrimary:   t.IsPrimary,
		})
	}
	for _, r := range relations {
		dto := model.WordRelationDTO{Type: r.RelationType, Text: r.RelatedText, Source: r.Source}
		if r.RelationType == entity.RelationTypeSynonym {
			resp.Synonyms = append(resp.Synonyms, dto)
		} else if r.RelationType == entity.RelationTypeAntonym {
			resp.Antonyms = append(resp.Antonyms, dto)
		}
	}
	for _, o := range occs {
		dto := model.WordOccurrenceDTO{Sentence: o.SampleSentence}
		page := new(entity.ParsedPage)
		if err := c.PageRepository.FindByID(tx, page, o.ParsedPageID); err == nil {
			dto.SourceURL = page.URL
			dto.PageTitle = page.Title
		}
		resp.Occurrences = append(resp.Occurrences, dto)
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return resp, nil
}
