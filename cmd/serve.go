package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"pipe/internal/api"
	"pipe/internal/bot"
	"pipe/internal/config"
	"pipe/internal/repository"
	"pipe/internal/repository/cassandra"
	"pipe/internal/repository/redis"
	"pipe/internal/services"
	"time"
)

func Serve() {
	config.LoadConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cassandraSession, err := cassandra.NewCassandraSession(config.AppConfig.CassandraHost, config.AppConfig.CassandraKeyspace)
	if err != nil {
		log.Fatalf("failed connect to cassandra: %v", err)
	}

	redisClient, err := redis.NewRedisClient(config.AppConfig.RedisHost)
	if err != nil {
		log.Fatalf("failed connect to redis: %v", err)
	}

	accountRepository := repository.NewAccountCassandraRepository(cassandraSession)
	messageRepository := repository.NewMessageCassandraRepository(cassandraSession)
	redisRepository := repository.NewRedisRepository(redisClient)

	app := services.NewApp(
		services.NewAccountService(accountRepository),
		services.NewMessageService(messageRepository, redisRepository),
	)

	tg, err := bot.NewTelegram(ctx, app)
	if err != nil {
		log.Fatal("couldn't connect to the telegram server")
	}

	go tg.Start()

	wa := api.NewWebApp(config.AppConfig.ServerAddr, app, tg.Bot)

	go func() {
		log.Fatal(wa.Start())
	}()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	defer wa.Shutdown(shutdownCtx)
	defer tg.Shutdown()

	log.Println("server is up and running")
	<-ctx.Done()
	log.Println("shutting down the server...")
}
