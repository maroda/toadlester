package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
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
	MU             sync.Mutex
	Name           string                // Metric name
	RandomBuffer   []string              // Randomized metrics
	ShiftRegisters map[string]*CycBuffer // Map of Cyclical Buffers
}

// NewEPHandle initializes MetricTypes, Buffers, and the Ticker.
// Server and Mux are done by calling func.
//
//	mtype = "exp", "float", "int"
//	buffer algos = "up", "down", "floatup", "floatdown"
func NewEPHandle(mtypes, balgos []string) *EPHandle {
	names := make(map[string]*MType)

	// Init each type with its name
	for _, mt := range mtypes {
		names[mt] = &MType{
			Name:           mt,
			RandomBuffer:   make([]string, len(balgos)),
			ShiftRegisters: make(map[string]*CycBuffer),
		}
	}

	// Init each cyclical shift register for every existing mtype
	for _, algo := range balgos { // algorithms belong to buffers
		for _, mt := range mtypes { // numeric types belong to mtypes
			// Values in series that result from functions that fire when called
			// Series of monotonic values
			size := FillEnvVarInt(strings.ToUpper(mt)+"_SIZE", 10)
			limit := FillEnvVarInt(strings.ToUpper(mt)+"_LIMIT", 10)
			tail := FillEnvVarInt(strings.ToUpper(mt)+"_TAIL", 1)
			modenv := FillEnvVar(strings.ToUpper(mt) + "_MOD")
			mod, err := strconv.ParseFloat(modenv, 64)
			if err != nil {
				slog.Warn("Default chosen instead of: " + modenv)
				mod = 10000
			}

			slog.Debug("INIT SHIFT REGISTER",
				slog.String("name", mt),
				slog.Int("size", size),
				slog.Int("limit", limit),
				slog.Int("tail", tail),
				slog.Any("mod", mod),
				slog.Any("algo", algo))

			newCbuff := NewShiftCycBuffer(size, limit, tail, mod, mt, algo)
			names[mt].ShiftRegisters[algo] = newCbuff

			slog.Debug("GOT SHIFT REGISTER",
				slog.String("name", mt),
				slog.Any("buffer", names[mt].ShiftRegisters[algo]))

			// Static Random values
			algoR := "random"
			sizeR := FillEnvVarInt("RAND_SIZE", 1)
			limitR := FillEnvVarInt("RAND_LIMIT", 10000)
			tailR := FillEnvVarInt("RAND_TAIL", 1)
			modRenv := FillEnvVar("RAND_MOD")
			modR, err := strconv.ParseFloat(modRenv, 64)
			if err != nil {
				slog.Warn("Default chosen instead of: " + modRenv)
				modR = 10000
			}

			slog.Debug("INIT RANDOMIZER",
				slog.String("name", mt),
				slog.Int("size", sizeR),
				slog.Int("limit", limitR),
				slog.Int("tail", tailR),
				slog.Any("mod", modR),
				slog.Any("algo", algoR))

			newRbuff := NewShiftCycBuffer(sizeR, limitR, tailR, modR, mt, algoR)
			names[mt].RandomBuffer = newRbuff.Values

			slog.Debug("GOT RANDOMIZER",
				slog.String("name", mt),
				slog.Any("buffer", newRbuff))
		}
	}

	return &EPHandle{
		MTypes: names,
		Ticker: time.NewTicker(1 * time.Second),
	}
}

func (eph *EPHandle) SetupMux() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/rand/all", eph.RandDataAllHandler)

	r.PathPrefix("/series").HandlerFunc(eph.SeriesInternalDataHandler)

	return r
}

func (eph *EPHandle) findTypeKey(find string) bool {
	for k, _ := range eph.MTypes {
		if k == find {
			return true
		}
	}
	return false
}

func (eph *EPHandle) findAlgoKey(find string) bool {
	for _, t := range eph.MTypes {
		for _, b := range t.ShiftRegisters {
			if b.MAlgo == find {
				return true
			}
		}
	}
	return false
}

func (eph *EPHandle) SeriesInternalDataHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		slog.Error("Invalid series data path")
		http.Error(w, "Invalid series data path", http.StatusBadRequest)
		return
	}

	algotype := parts[2] // numeric type (exp, float, int)
	algo := parts[3]     // algorithm name (up, down)

	if !eph.findTypeKey(algotype) {
		slog.Error("Invalid series data path: " + algotype)
		http.Error(w, "Invalid series data path: "+algotype, http.StatusBadRequest)
		return
	}

	if !eph.findAlgoKey(algo) {
		slog.Error("Invalid series data path:" + algo)
		http.Error(w, "Invalid series data path: "+algo, http.StatusBadRequest)
		return
	}

	// assign buffer as shift register
	shiftReg := eph.MTypes[algotype].ShiftRegisters[algo]
	shiftReg.MU.Lock()
	algoVal := shiftReg.Values[shiftReg.Index]
	shiftReg.MU.Unlock()

	slog.Info("algorithm match",
		slog.String("method", r.Method),
		slog.String("request", r.RequestURI),
		slog.String("requested.type", algotype),
		slog.String("algo.name", shiftReg.MAlgo),
		slog.String("algo.value", algoVal),
		slog.Any("full.values", shiftReg.Values),
	)

	w.Header().Set("Content-Type", "application/plaintext")
	output := fmt.Sprintf("Metric_%s_%s: %s\n", algotype, algo, algoVal)
	w.Write([]byte(output))
}

// RandDataAllHandler returns randomly changing values in all supported types
func (eph *EPHandle) RandDataAllHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/plaintext; charset=utf-8")

	for _, mt := range eph.MTypes {
		mt.MU.Lock()
	}
	randexp := eph.MTypes["exp"].RandomBuffer[0]
	randfloat := eph.MTypes["float"].RandomBuffer[0]
	randint := eph.MTypes["int"].RandomBuffer[0]
	for _, mt := range eph.MTypes {
		mt.MU.Unlock()
	}

	slog.Info("randomizer match",
		slog.String("method", r.Method),
		slog.String("request", r.RequestURI),
		slog.String("random.exponent", randexp),
		slog.String("random.float", randfloat),
		slog.String("random.integer", randint),
	)

	output := fmt.Sprintf("ExpMetric: %s\nFloatMetric: %s\nIntMetric: %s\n", randexp, randfloat, randint)
	w.Write([]byte(output))
}
