package usecase

import (
	"context"
	"errors"
	"time"

	"words-app/internal/entity"
	"words-app/internal/model"
	"words-app/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthUseCase struct {
	DB             *gorm.DB
	Log            *logrus.Logger
	Validate       *validator.Validate
	UserRepository *repository.UserRepository
	JWTSecret      []byte
	JWTTTL         time.Duration
}

func NewAuthUseCase(db *gorm.DB, log *logrus.Logger, validate *validator.Validate,
	userRepo *repository.UserRepository, jwtSecret string, jwtTTL time.Duration) *AuthUseCase {
	return &AuthUseCase{
		DB:             db,
		Log:            log,
		Validate:       validate,
		UserRepository: userRepo,
		JWTSecret:      []byte(jwtSecret),
		JWTTTL:         jwtTTL,
	}
}

func (c *AuthUseCase) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	tx := c.DB.WithContext(ctx).Begin()
	defer tx.Rollback()

	if err := c.Validate.Struct(req); err != nil {
		c.Log.Warnf("Invalid login request : %+v", err)
		return nil, fiber.ErrBadRequest
	}

	user := new(entity.User)
	if err := c.UserRepository.FindByEmail(tx, user, req.Email); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Log.Warnf("Login: user not found : email=%s", req.Email)
			return nil, fiber.ErrUnauthorized
		}
		c.Log.Warnf("Failed find user by email : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.Log.Warnf("Login: bad password : email=%s", req.Email)
		return nil, fiber.ErrUnauthorized
	}

	expiresAt := time.Now().Add(c.JWTTTL)
	claims := jwt.MapClaims{
		"sub":   user.PublicID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   expiresAt.Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(c.JWTSecret)
	if err != nil {
		c.Log.Warnf("Failed to sign jwt : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	if err := tx.Commit().Error; err != nil {
		c.Log.Warnf("Failed commit transaction : %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	return &model.LoginResponse{
		Token:     signed,
		ExpiresAt: expiresAt.Unix(),
		User: model.UserResponse{
			ID:    user.PublicID,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}

func (c *AuthUseCase) Me(ctx context.Context, userPublicID string) (*model.UserResponse, error) {
	user := new(entity.User)
	if err := c.UserRepository.FindByPublicID(c.DB.WithContext(ctx), user, userPublicID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.ErrUnauthorized
		}
		c.Log.Warnf("Failed to fetch user : %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return &model.UserResponse{ID: user.PublicID, Email: user.Email, Role: user.Role}, nil
}
