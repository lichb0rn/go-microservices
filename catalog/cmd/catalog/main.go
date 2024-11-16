package main

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/lichb0rn/go-microservices/catalog"
	"github.com/tinrab/kit/retry"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Fatal(err)
	}

	var repository catalog.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		repository, err = catalog.NewElasticRepository(cfg.DatabaseURL)
		if err != nil {
			log.Println(err)
		}
		return
	})

	defer repository.Close()
	log.Println("Listening on port 8080")
	s := catalog.NewService(repository)
	log.Fatal(catalog.ListendGRPC(s, 8080))
}
