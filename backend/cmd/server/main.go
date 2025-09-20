package main

import (
    "log"
    "os"

    "github.com/paulo-hortelan/fibra-ctrl/internal/api"
    "github.com/paulo-hortelan/fibra-ctrl/internal/connection"
    "github.com/paulo-hortelan/fibra-ctrl/internal/queue"
    "github.com/paulo-hortelan/fibra-ctrl/internal/worker"
    "github.com/paulo-hortelan/fibra-ctrl/internal/repository"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func main() {
    connPool := connection.New()

    cmdQueue := queue.New()

    workerPool := worker.NewPool(5, connPool, cmdQueue)

    workerPool.Start()

    dsn := "host=" + os.Getenv("DB_HOST") +
        " user=" + os.Getenv("DB_USER") +
        " password=" + os.Getenv("DB_PASSWORD") +
        " dbname=" + os.Getenv("DB_NAME") +
        " port=" + os.Getenv("DB_PORT") +
        " sslmode=disable TimeZone=UTC"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("failed to connect database:", err)
    }

    userRepo := repository.NewUserRepository(db)

    apiServer := api.NewServer(":8080", cmdQueue, nil, userRepo)
    log.Fatal(apiServer.Start())
}