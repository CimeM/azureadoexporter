package main

import (
	"os"
	"time"
	"azureadocomms"

	"log"

	"sync"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {

	var metricsData []string
	// mutex protects the slice from writing while reading
	var metricsMutex sync.RWMutex
	// Start the update function in a goroutine your application starts
	
	// Start the update function in a goroutine
	go updateMetrics(&metricsMutex, &metricsData)
	// wait for the initial data
	time.Sleep(10 * time.Second)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())	
	e.GET("/metrics", func(c echo.Context) error {
		log.Println("Accessed /metrics endpoint")
		metricsMutex.RLock()
    	defer metricsMutex.RUnlock()
		
		log.Printf("Serving %d metrics\n", len(metricsData))
		// Use the Echo context's Response().Writer to write the metrics
		for _, metric := range metricsData {
			_, err := c.Response().Write([]byte(metric + "\n"))
			if err != nil {
				log.Printf("Error writing metric: %v\n", err)
				return err
			}
		}
		return nil
	})

	// test eps
	// e.GET("/"+os.Getenv("ADO_ORGANIZATION")+"/"+os.Getenv("ADO_PROJECT")+"/_apis/pipelines", func(c echo.Context) error {
	// 	type Pipeline struct {
	// 		ID   string `json:"id"`
	// 		Name string `json:"name"`
	// 	}
	
	// 	// Create some sample data
	// 	pipelines := []Pipeline{
	// 		{ID: "1", Name: "Build Pipeline"},
	// 		{ID: "2", Name: "Test Pipeline"},
	// 		{ID: "3", Name: "Deploy Pipeline"},
	// 	}

	// 	// Return the data as JSON
	// 	return c.JSON(http.StatusOK, pipelines)
	
	// })

	httpPort := os.Getenv("PORT")
	if httpPort == "" {
		httpPort = "8080"
	}
	log.Printf("Starting server on port %s\n", httpPort)
	e.Logger.Fatal(e.Start(":" + httpPort))

	
}

// function to update the metrics data periodically
func updateMetrics(metricsMutex *sync.RWMutex, metricsData *[]string) {
	
	ado_url := os.Getenv("ADO_URL")
    project := os.Getenv("ADO_PROJECT")
    pat := os.Getenv("ADO_PERSONAL_ACCESS_TOKEN")
	organization := os.Getenv("ADO_ORGANIZATION")
	ado_creds := &azureadocomms.ADOCredentials{ ado_url, project, organization, pat}

	log.Println("Starting updateMetrics goroutine")
	for {
		// this call take some time to get all the data
		log.Println("Fetching new metrics data...")

		// exponential backoff function - backoff increases exponentially
		for i := 0; i <= 5; i++ {
			startTime := time.Now()
			newData, err := azureadocomms.Call( *ado_creds )
			// throw error in case retry did not work
			if i == 5 {
				log.Println("Error processing request. Exiting.")
			}
			if err != nil {
				// handle error
				log.Println("Error retrieving data: %v\n, retrying in %dmin...", err, i * i)
				time.Sleep(i * i * time.Minute)
			}
			// sucessful attempt
			duration := time.Since(startTime)
			metricsMutex.Lock()
			// metricsData = append(metricsData, "exporter_details{requestduration=\"%d\",retries=\"%d\"} 1", duration, retries )
			*metricsData = newData
			metricsMutex.Unlock()
			log.Printf("Metrics updated. Fetched %d metrics in %v\n", len(*metricsData), duration)
			log.Println("Waiting for 5 minutes before next update...")
			time.Sleep(5 * time.Minute)

		}
    }
}