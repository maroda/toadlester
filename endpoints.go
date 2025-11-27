package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// EPHandle is called by main() and contains the mux
type EPHandle struct {
	Endpoints   []string
	Server      *http.Server
	Mux         *mux.Router
	Ticker      *time.Ticker
	StatsPerSec map[string][]string
}

// NewEPHandle initializes the endpoints
func NewEPHandle(endpoints []string) *EPHandle {
	return &EPHandle{
		Endpoints:   endpoints,
		Ticker:      time.NewTicker(1 * time.Second),
		StatsPerSec: make(map[string][]string, 3), // Three types: exp, float, int
	}
}

func (eph *EPHandle) SetupMux() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/ep/kv", KVdataHandler)
	r.HandleFunc("/ep/json", JSONdataHandler)
	r.HandleFunc("/rand/int", eph.RandDataIntHandler)
	r.HandleFunc("/rand/json", eph.RandDataJSONHandler)
	r.HandleFunc("/rand/all", eph.RandDataAllHandler)

	return r
}

func KVdataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/plaintext")
	w.Write([]byte("HelloWorld: 69"))
}

func JSONdataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]string{"HelloWorld": "69"})
}

func (eph *EPHandle) RandDataJSONHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(eph.StatsPerSec["exp"])
}

func (eph *EPHandle) RandDataIntHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/plaintext; charset=utf-8")
	for i, v := range eph.StatsPerSec["int"] {
		output := fmt.Sprintf("IntMetric_%d: %s\n", i, v)
		w.Write([]byte(output))
	}
}

func (eph *EPHandle) RandDataAllHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/plaintext; charset=utf-8")
	output := fmt.Sprintf("IntMetric: %s\nExpMetric: %s\nFloatMetric: %s\n", eph.StatsPerSec["int"][rand.N(len(eph.StatsPerSec["int"]))], eph.StatsPerSec["exp"][rand.N(len(eph.StatsPerSec["exp"]))], eph.StatsPerSec["float"][rand.N(len(eph.StatsPerSec["float"]))])
	w.Write([]byte(output))
}
