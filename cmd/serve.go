package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"pipe/internal/api"
	"pipe/internal/bot"
	"pipe/internal/cassandra"
	"pipe/internal/config"
	"pipe/internal/repositories"
	"pipe/internal/services"
	"time"
)

func Serve() {
	config.LoadConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	session, err := cassandra.NewCassandraSession()
	if err != nil {
		log.Fatalf("failed connect to cassandra: %v", err)
	}

	accountRepository := repositories.NewAccountCassandraRepository(session)
	messageRepository := repositories.NewMessageCassandraRepository(session)

	app := services.NewApp(
		services.NewAccountService(accountRepository),
		services.NewMessageService(messageRepository),
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
