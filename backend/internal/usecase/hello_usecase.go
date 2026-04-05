package usecase

import (
	"context"
	"time"

	"words-app/internal/model"

	"github.com/sirupsen/logrus"
)

type HelloUseCase struct {
	Log *logrus.Logger
}

func NewHelloUseCase(logger *logrus.Logger) *HelloUseCase {
	return &HelloUseCase{Log: logger}
}

func (c *HelloUseCase) Greet(ctx context.Context) (*model.HelloResponse, error) {
	return &model.HelloResponse{
		Message: "A tour of beautiful words.",
		Time:    time.Now().Format(time.RFC1123),
	}, nil
}
