package main

import (
	"errors"
	"log"
	"net/http"
)

func main() {
	eph := NewEPHandle([]string{"int"})
	defer eph.Ticker.Stop()

	// Run webserver in parallel to metric creation
	go func() {
		addr := ":4330"
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
			// Every second the ticker runs, create a new buffer with new values.
			// Currently just doing "exp" for testing, all three would do this
			bufferExp := NewRandCycBuffer(4, 10000, 8, 10000, "exp")
			eph.StatsPerSec["exp"] = bufferExp.Values
			bufferFloat := NewRandCycBuffer(4, 10000, 8, 10000, "float")
			eph.StatsPerSec["float"] = bufferFloat.Values
			bufferInt := NewRandCycBuffer(4, 10000, 8, 10000, "int")
			eph.StatsPerSec["int"] = bufferInt.Values
			/*
				for _, v := range eph.StatsPerSec {
					fmt.Println(v)
				}
			*/
		}
	}
}
