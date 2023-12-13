package main

import (
	"github.com/GearFramework/urlshort/internal/app"
	"github.com/GearFramework/urlshort/internal/config"
	"github.com/GearFramework/urlshort/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err.Error())
	}
}

func run() error {
	shortener, err := app.NewShortener(config.GetConfig())
	if err != nil {
		return err
	}
	gracefulStop(shortener.StopApp)
	s, err := server.NewServer(shortener.Conf, shortener)
	if err != nil {
		return err
	}
	s.InitRoutes()
	return s.Up()
}

func gracefulStop(stopCallback func()) {
	gracefulStopChan := make(chan os.Signal, 1)
	signal.Notify(
		gracefulStopChan,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	go func() {
		sig := <-gracefulStopChan
		stopCallback()
		log.Printf("Caught sig: %+v\n", sig)
		log.Println("Application graceful stop!")
		os.Exit(0)
	}()
}
