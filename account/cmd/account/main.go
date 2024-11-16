package main

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/lichb0rn/go-microservices/account"
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

	var repository account.Repository
	retry.ForeverSleep(2*time.Second, func(_ int) (err error) {
		repository, err = account.NewPostgresRepository(cfg.DatabaseURL)
		if err != nil {
			log.Println(err)
		}
		return
	})

	defer repository.Close()
	log.Println("Listening on port 8080")
	s := account.NewService(repository)
	log.Fatal(account.ListendGRPC(s, 8080))
}
