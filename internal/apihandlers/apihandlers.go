package apihandlers

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/cimem/azureadoexporter/internal/metrics"
)

func RegisterHandlers(e *echo.Echo, updater *metrics.Updater) {
	e.GET("/metrics", func(c echo.Context) error {
		log.Println("Accessed /metrics endpoint")
		metricsData := updater.GetMetrics()
		
		log.Printf("Serving %d metrics\n", len(metricsData))
		for _, metric := range metricsData {
			if _, err := c.Response().Write([]byte(metric + "\n")); err != nil {
				log.Printf("Error writing metric: %v\n", err)
				return err
			}
		}
		return nil
	})
}