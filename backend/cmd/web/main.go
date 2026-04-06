package main

import (
	"fmt"

	"words-app/internal/config"

	"github.com/go-playground/validator/v10"
)

func main() {
	v := config.NewViper()
	log := config.NewLogger(v)
	db := config.NewDatabase(v, log)
	validate := validator.New()
	app := config.NewFiber(v, log)

	config.Bootstrap(&config.BootstrapConfig{
		DB:        db,
		App:       app,
		Log:       log,
		Validate:  validate,
		Viper:     v,
		StaticDir: v.GetString("web.static_dir"),
	})

	port := v.GetInt("web.port")
	log.Infof("Starting web server on :%d", port)
	if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("Failed to start server : %+v", err)
	}
}
