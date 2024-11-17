package main

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/lichb0rn/go-microservices/order"
	"github.com/tinrab/kit/retry"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
	AccountURL  string `envconfig:"ACCOUNT_SERVICE_URL"`
	CatalogURL  string `envconfig:"CATALOG_SERVICE_URL"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	var repository order.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		repository, err = order.NewPostgresRepository(cfg.DatabaseURL)
		if err != nil {
			log.Println(err)
		}
		return
	})

	log.Println("Listening on port 8080")
	s := order.NewService(repository)
	log.Fatal(order.ListendGRPC(s, cfg.AccountURL, cfg.CatalogURL, 8080))
}
