package main

import (
	"errors"
	"log"
	"net/http"
)

// Global vars for easy access to reset during operation.
// Currently, not meant to be user-configurable.
// Only use these to control NewEPHandle.
var (
	NTypes = []string{"exp", "float", "int"} // Numeric Types
	MAlgos = []string{"up", "down"}          // Display Algorithms
)

func main() {
	eph := NewEPHandle(NTypes, MAlgos)
	defer eph.Ticker.Stop()

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
