package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// EPHandle is called by main() and contains the mux
// It handles and routes all Endpoints (type EP)
type EPHandle struct {
	MTypes map[string]*MType
	Server *http.Server
	Mux    *mux.Router
	Ticker *time.Ticker
}

type MType struct {
	Name         string       // Metric name
	RandomBuffer []string     // Randomized metrics, key: name
	CyclicBuffer []*CycBuffer // Cyclical metric series, key: name
}

// NewEPHandle initializes MetricTypes, Buffers, and the Ticker.
// Server and Mux are done by calling func.
//
//	mtype = "exp", "float", "int"
//	buffers = "up", "down", "floatup", "floatdown"
func NewEPHandle(mtypes, buffers []string) *EPHandle {
	names := make(map[string]*MType)

	// Init each type with its name
	for _, mt := range mtypes {
		names[mt] = &MType{
			Name:         mt,
			RandomBuffer: make([]string, len(buffers)),
			CyclicBuffer: make([]*CycBuffer, 0),
		}
	}

	// Init each cyclical shift register for every existing mtype
	for _, buff := range buffers { // algorithms belong to buffers
		for _, mt := range mtypes { // numeric types belong to mtypes

			// Values in series that result from functions that fire when called
			size := FillEnvVarInt(strings.ToUpper(mt)+"_SIZE", 10)
			limit := FillEnvVarInt(strings.ToUpper(mt)+"_LIMIT", 10)
			tail := FillEnvVarInt(strings.ToUpper(mt)+"_TAIL", 1)
			modenv := FillEnvVar(strings.ToUpper(mt) + "_MOD")
			mod, err := strconv.ParseFloat(modenv, 64)
			if err != nil {
				slog.Warn("Default chosen instead of: " + modenv)
				mod = 10000
			}

			newCbuff := NewShiftCycBuffer(size, limit, tail, mod, buff)
			names[mt].CyclicBuffer = append(names[mt].CyclicBuffer, newCbuff)

			// Static Random values that can be changed by external processes
			sizeR := FillEnvVarInt("RAND_SIZE", 10)
			limitR := FillEnvVarInt("RAND_LIMIT", 10000)
			tailR := FillEnvVarInt("RAND_TAIL", 1)
			modRenv := FillEnvVar("RAND_MOD")
			modR, err := strconv.ParseFloat(modRenv, 64)
			if err != nil {
				slog.Warn("Default chosen instead of: " + modRenv)
				modR = 10000
			}

			newRbuff := NewRandCycBuffer(sizeR, limitR, tailR, modR, buff)
			names[mt].RandomBuffer = newRbuff.Values
		}
	}

	return &EPHandle{
		MTypes: names,
		Ticker: time.NewTicker(1 * time.Second),
	}
}

func (eph *EPHandle) SetupMux() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/ep/kv", KVdataHandler)
	r.HandleFunc("/ep/json", JSONdataHandler)
	r.HandleFunc("/rand/all", eph.RandDataAllHandler)

	r.PathPrefix("/series").HandlerFunc(eph.SeriesInternalDataHandler)

	return r
}

func (eph *EPHandle) SeriesInternalDataHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		slog.Error("Invalid series data path")
		http.Error(w, "Invalid series data path", http.StatusBadRequest)
		return
	}

	// align algorithm with buffer contents
	algotype := parts[2]      // numeric type (exp, float, int)
	algo := parts[3]          // algorithm name (up, down, upfloat, downfloat)
	useCBAlgo := &CycBuffer{} // buffer to hold match
	for _, mt := range eph.MTypes {
		for _, buff := range mt.CyclicBuffer {
			if buff.MName == algo {
				useCBAlgo = buff
				break
			}
		}
	}

	slog.Info("algorithm match",
		slog.String("requested", algotype),
		slog.String("algo", useCBAlgo.MName),
		slog.Any("values", useCBAlgo.Values),
	)

	w.Header().Set("Content-Type", "application/plaintext")
	output := fmt.Sprintf("Metric_%s: %s\n", algotype, useCBAlgo.Values[useCBAlgo.Index])
	w.Write([]byte(output))
}

func KVdataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/plaintext")
	w.Write([]byte("HelloWorld: 69"))
}

func JSONdataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]string{"HelloWorld": "69"})
}

// RandDataAllHandler returns randomly changing values in all supported types
func (eph *EPHandle) RandDataAllHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/plaintext; charset=utf-8")

	randexp := eph.MTypes["exp"].RandomBuffer[0]
	randfloat := eph.MTypes["float"].RandomBuffer[0]
	randint := eph.MTypes["int"].RandomBuffer[0]

	output := fmt.Sprintf("ExpMetric: %s\nFloatMetric: %s\nIntMetric: %s\n", randexp, randfloat, randint)
	w.Write([]byte(output))
}
