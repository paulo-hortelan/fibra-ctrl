package main

import (
	"fmt"
	"log"
	"os"

	"github.com/paulo-hortelan/fibra-ctrl/internal/api"
	"github.com/paulo-hortelan/fibra-ctrl/internal/connection"
	"github.com/paulo-hortelan/fibra-ctrl/internal/queue"
	"github.com/paulo-hortelan/fibra-ctrl/internal/repository"
	"github.com/paulo-hortelan/fibra-ctrl/internal/worker"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_USER", "fibrauser"),
		getEnv("DB_PASSWORD", "fibrapassword"),
		getEnv("DB_NAME", "fibradb"),
		getEnv("DB_PORT", "5432"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("falha ao conectar ao banco de dados:", err)
	}

	oltRepo, err := repository.NewDBOLTRepository(db)
	if err != nil {
		log.Fatal("falha ao inicializar repositório:", err)
	}

	connPool := connection.NewPool()
	cmdQueue := queue.New()
	workerPool := worker.NewPool(5, connPool, cmdQueue, oltRepo)
	workerPool.Start()

	apiServer := api.NewServer(":8080", cmdQueue, oltRepo)
	log.Fatal(apiServer.Start())
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
