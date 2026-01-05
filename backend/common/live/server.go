package live

import (
	"alerting-platform/common/config"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func StartLiveServer(wg *sync.WaitGroup) {
	mux := http.NewServeMux()
	cfg := config.GetConfig()

	mux.HandleFunc("/live", LiveHandler)

	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.LivePort),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	wg.Add(1)

	go func() {
		defer wg.Done()

		log.Printf("[INFO] Liveness probe running on port :%d/live", cfg.REST_APIPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ERROR] Failed to start live server: %v", err)
		}
	}()
}
