package main

import (
	"errors"
	"log"
	"net/http"
)

func main() {
	eph := NewEPHandle([]string{"exp", "float", "int"}, []string{"up", "down", "floatup", "floatdown"})
	defer eph.Ticker.Stop()

	// Use config env vars to set up buffers

	// Run webserver in parallel to metric creation
	go func() {
		addr := ":8899"
		eph.Server = &http.Server{
			Addr:    addr,
			Handler: eph.SetupMux(),
		}

		if err := eph.Server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// Main loop that creates metrics for endpoint handlers
	for {
		select {
		case <-eph.Ticker.C:
			eph.RandBuffers()  // Creates a new buffer every time for random data
			eph.ShiftBuffers() // Creates or updates the cyclical algorithm buffer
		}
	}
}
