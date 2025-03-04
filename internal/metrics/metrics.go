package metrics

import (
	"log"
	"sync"
	"time"
	"github.com/cimem/azureadoexporter/internal/ado"
)

type Updater struct {
	metricsData []string
	mutex       sync.RWMutex
	client      *ado.Client
}

func NewUpdater() *Updater {
	return &Updater{
		client: ado.NewClient(),
	}
}

func (u *Updater) Start() {
	log.Println("Starting updateMetrics goroutine")
	for {
		log.Println("Fetching new metrics data...")
		u.updateMetrics()
		log.Println("Waiting for 5 minutes before next update...")
		time.Sleep(5 * time.Minute)
	}
}

func (u *Updater) updateMetrics() {
	for i := 0; i <= 5; i++ {
		startTime := time.Now()
		newData, err := u.client.FetchMetrics()
		if err != nil {
			log.Printf("Error retrieving data: %v\n, retrying in %d min...", err, i+1)
			time.Sleep(time.Duration(i+1) * time.Minute)
			continue
		}

		duration := time.Since(startTime)
		u.mutex.Lock()
		u.metricsData = newData
		u.mutex.Unlock()
		log.Printf("Metrics updated. Fetched %d metrics in %v\n", len(newData), duration)
		return
	}
	log.Println("Error processing request. Exiting.")
}

func (u *Updater) GetMetrics() []string {
	u.mutex.RLock()
	defer u.mutex.RUnlock()
	return u.metricsData
}
