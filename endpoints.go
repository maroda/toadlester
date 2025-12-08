package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

const (
	defSize  = 10
	defLimit = 10
	defTail  = 1
	defMod   = 1
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

	// Init each cyclical shift register for every existing mtype and random
	for _, algo := range balgos { // algorithms belong to buffers
		for _, mt := range mtypes { // numeric types belong to mtypes
			// Series of monotonic values
			newBuff := getConfiguredBuffer(mt, algo)
			names[mt].ShiftRegisters[algo] = newBuff

			slog.Debug("GOT SHIFT REGISTER",
				slog.String("name", mt),
				slog.Any("buffer", names[mt].ShiftRegisters[algo]))

			// Static Random values
			newRandomizer := getRandomizedBuffer(mt, "random")
			names[mt].RandomBuffer = newRandomizer.Values

			slog.Debug("GOT RANDOMIZER",
				slog.String("name", mt),
				slog.Any("buffer", names[mt].RandomBuffer))
		}
	}

	return &EPHandle{
		MTypes: names,
		Ticker: time.NewTicker(1 * time.Second),
	}
}

// SetupMux provides a new Mux with its internal routing configured
// These are the control points for Toadlester
func (eph *EPHandle) SetupMux() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/rand/all", eph.RandDataAllHandler)
	r.HandleFunc("/metrics", eph.SeriesDataAllHandler)
	r.PathPrefix("/reset").HandlerFunc(eph.ResetHandler)
	r.PathPrefix("/series").HandlerFunc(eph.SeriesInternalDataHandler)

	return r
}

// ResetHandler sets new values for each shift register buffer
// It uses the final parameter of the API URI to set a new Env Var for that value.
// Then a new buffer is requested, which reads Env Vars to configure.
func (eph *EPHandle) ResetHandler(w http.ResponseWriter, r *http.Request) {
	var output string

	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		slog.Error("Invalid reset data path")
		http.Error(w, "Invalid reset data path", http.StatusBadRequest)
		return
	}

	envvar := parts[2] // numeric type (exp, float, int)
	value := parts[3]  // algorithm name (up, down)

	if !eph.findEnvVar(envvar) {
		slog.Error("Invalid reset variable: " + envvar)
		http.Error(w, "Invalid reset variable: "+envvar, http.StatusBadRequest)
		return
	}

	// Set the env var being changed
	// TODO: Value validation
	os.Setenv(strings.ToUpper(envvar), value)

	// Locate buffer with params and update with new Env Var set
	params := strings.Split(envvar, "_")

	mtype := strings.ToLower(params[0])
	mconf := strings.ToLower(params[1])
	slog.Debug("params", slog.String("mtype", mtype), slog.String("malgo", mconf))

	// Get new buffers for all algorithms of this mtype
	for _, buff := range eph.MTypes[mtype].ShiftRegisters {
		// Hold on to history for logging
		oldValues := buff.Values

		// Get a new buffer
		buff.MU.Lock()
		newBuff := getConfiguredBuffer(buff.NType, buff.MAlgo)
		buff.Values = newBuff.Values
		buff.MU.Unlock()

		output = output + fmt.Sprintf("Set new %s value %s for %s\n", buff.MAlgo, envvar, value)

		slog.Info("Reset complete",
			slog.String("method", r.Method),
			slog.String("request", r.RequestURI),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("buffer", buff.MAlgo),
			slog.String("old_values", strings.Join(oldValues, ", ")),
			slog.String("new_values", strings.Join(buff.Values, ", ")))
	}

	w.Header().Set("Content-Type", "application/plaintext")
	w.Write([]byte(output))
}

// Validates Env Var name against types and algorithms
func (eph *EPHandle) findEnvVar(find string) bool {
	var front, back bool
	parts := strings.Split(find, "_")

	for k := range eph.MTypes {
		if strings.ToUpper(k) == parts[0] {
			// it's valid, tag it as true
			front = true
		}

		algoparams := []string{
			"SIZE",
			"LIMIT",
			"TAIL",
			"MOD",
		}
		for _, p := range algoparams {
			if p == parts[1] {
				// it's valid, tag it as true
				back = true
			}
		}
	}

	// return truth table of front and back
	return front && back
}

// Series of monotonic values
func getConfiguredBuffer(mt, algo string) *CycBuffer {
	size := FillEnvVarInt(strings.ToUpper(mt)+"_SIZE", defSize)
	limit := FillEnvVarInt(strings.ToUpper(mt)+"_LIMIT", defLimit)
	tail := FillEnvVarInt(strings.ToUpper(mt)+"_TAIL", defTail)
	modenv := FillEnvVar(strings.ToUpper(mt) + "_MOD")
	mod, err := strconv.ParseFloat(modenv, 64)
	if err != nil {
		slog.Debug("Default chosen",
			slog.String("mod", modenv))
		mod = defMod
	}

	slog.Debug("INIT SHIFT REGISTER",
		slog.String("name", mt),
		slog.Int("size", size),
		slog.Int("limit", limit),
		slog.Int("tail", tail),
		slog.Any("mod", mod),
		slog.Any("algo", algo))

	return NewShiftCycBuffer(size, limit, tail, mod, mt, algo)
}

// Static Random values with different defaults
func getRandomizedBuffer(mt, algo string) *CycBuffer {
	size := FillEnvVarInt("RAND_SIZE", 1)
	limit := FillEnvVarInt("RAND_LIMIT", 10000)
	tail := FillEnvVarInt("RAND_TAIL", 4)
	modenv := FillEnvVar("RAND_MOD")
	mod, err := strconv.ParseFloat(modenv, 64)
	if err != nil {
		slog.Debug("Default chosen",
			slog.String("mod", modenv))
		mod = 10000
	}

	slog.Debug("INIT RANDOMIZER",
		slog.String("name", mt),
		slog.Int("size", size),
		slog.Int("limit", limit),
		slog.Int("tail", tail),
		slog.Any("mod", mod),
		slog.Any("algo", algo))

	return NewShiftCycBuffer(size, limit, tail, mod, mt, algo)
}

func (eph *EPHandle) findTypeKey(find string) bool {
	for k := range eph.MTypes {
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

// SeriesInternalDataHandler returns a metric from the series and algorithm requested
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

	slog.Info("Algorithm match",
		slog.String("method", r.Method),
		slog.String("request", r.RequestURI),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("requested.type", algotype),
		slog.String("algo.name", shiftReg.MAlgo),
		slog.String("algo.value", algoVal),
		slog.Any("full.values", shiftReg.Values),
	)

	w.Header().Set("Content-Type", "application/plaintext; charset=utf-8")
	output := fmt.Sprintf("Metric_%s_%s: %s\n", algotype, algo, algoVal)
	w.Write([]byte(output))
}

func (eph *EPHandle) SeriesDataAllHandler(w http.ResponseWriter, r *http.Request) {
	report := map[string]string{}
	var output string

	for _, mt := range eph.MTypes {
		mt.MU.Lock()
		for _, buff := range mt.ShiftRegisters {
			idx := buff.NType + buff.MAlgo
			report[idx] = buff.Values[buff.Index]
			output = output + fmt.Sprintf("Metric_%s_%s: %s\n",
				buff.NType, buff.MAlgo, report[idx])
		}
		mt.MU.Unlock()
	}

	slog.Info("Randomizer match",
		slog.String("method", r.Method),
		slog.String("request", r.RequestURI),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("expup", report["expup"]),
		slog.String("floatup", report["floatup"]),
		slog.String("intup", report["intup"]),
		slog.String("expdown", report["expdown"]),
		slog.String("floatdown", report["floatdown"]),
		slog.String("intdown", report["intdown"]))

	w.Header().Set("Content-Type", "application/plaintext; charset=utf-8")
	w.Write([]byte(output))
}

// RandDataAllHandler returns randomly changing values in all supported types
func (eph *EPHandle) RandDataAllHandler(w http.ResponseWriter, r *http.Request) {
	for _, mt := range eph.MTypes {
		mt.MU.Lock()
	}
	randexp := eph.MTypes["exp"].RandomBuffer[0]
	randfloat := eph.MTypes["float"].RandomBuffer[0]
	randint := eph.MTypes["int"].RandomBuffer[0]
	for _, mt := range eph.MTypes {
		mt.MU.Unlock()
	}

	slog.Info("Randomizer match",
		slog.String("method", r.Method),
		slog.String("request", r.RequestURI),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("random.exponent", randexp),
		slog.String("random.float", randfloat),
		slog.String("random.integer", randint),
	)

	w.Header().Set("Content-Type", "application/plaintext; charset=utf-8")
	output := fmt.Sprintf("ExpMetric: %s\nFloatMetric: %s\nIntMetric: %s\n", randexp, randfloat, randint)
	w.Write([]byte(output))
}
