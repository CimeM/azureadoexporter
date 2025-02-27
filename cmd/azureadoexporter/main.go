package main

import (
	"os"
	"time"

	"log"

	// "sync"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/cimem/azureadoexporter/internal/apihandlers"
	"github.com/cimem/azureadoexporter/internal/metrics"
	// "github.com/cimem/azureadoexporter/internal/azureadocomms"

)

func main() {
	metricsUpdater := metrics.NewUpdater()
	go metricsUpdater.Start()

	// Wait for initial data
	time.Sleep(10 * time.Second)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	apihandlers.RegisterHandlers(e, metricsUpdater)

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	log.Printf("Starting server on port %s\n", httpPort)
	e.Logger.Fatal(e.Start(":" + httpPort))
}

