package main

import (
	"log/slog"
	"math/rand/v2"
	"os"
	"strconv"
	"sync"
)

// CycBuffer is a cyclical shift register
type CycBuffer struct {
	MU      sync.Mutex
	MName   string   // Metric name
	Values  []string // Slice of whatever we need for responses
	MaxSize int      // How big this buffer can be
	Index   int      // We are at this index in the step buffer
}

// NewShiftCycBuffer creates cascading values to be shifted by one for each access
func NewShiftCycBuffer(maxSize, limit, tail int, mod float64, f string) *CycBuffer {
	values := make([]string, 0, maxSize)

	switch f {
	case "floatup":
		for i := 0; i < maxSize; i++ {
			values = append(values, strconv.FormatFloat(float64(limit-(limit-(tail*i)))*mod, 'f', 8, 64))
		}
	case "floatdown":
		for i := 0; i < maxSize; i++ {
			values = append(values, strconv.FormatFloat(float64(limit-(tail*i))*mod, 'f', 8, 64))
		}
	case "up":
		for i := 0; i < maxSize; i++ {
			values = append(values, strconv.Itoa(limit-(limit-(tail*i))))
		}
	case "down":
		for i := 0; i < maxSize; i++ {
			values = append(values, strconv.Itoa(limit-(tail*i)))
		}
	}

	return &CycBuffer{
		MName:   f,
		Values:  values,
		MaxSize: maxSize,
		Index:   0,
	}
}

// Shift returns the next value in the buffer
// First increase the Index, wrapping when reaching the full size
// then return the value at that spot
func (cb *CycBuffer) Shift() string {
	cb.MU.Lock()
	defer cb.MU.Unlock()

	cb.Index = (cb.Index + 1) % len(cb.Values)
	return cb.Values[cb.Index]
}

// NewRandCycBuffer creates slices of strings of configurable random numbers
func NewRandCycBuffer(maxSize, limit, tail int, mod float64, f string) *CycBuffer {
	values := make([]string, maxSize)
	for i := 0; i < maxSize; i++ {
		multiplier := mod * float64(rand.Int32N(int32(limit))) * rand.Float64()

		switch f {
		case "exp":
			values[i] = strconv.FormatFloat(multiplier, 'e', tail, 64)
		case "float":
			values[i] = strconv.FormatFloat(multiplier, 'f', tail, 64)
		case "int":
			values[i] = strconv.Itoa(int(multiplier))
		}
	}

	return &CycBuffer{
		MName:   f,
		Values:  values,
		MaxSize: maxSize,
		Index:   0,
	}
}

// RandBuffers is the engine for building random data buffers.
// Each of these can be queried by the endpoint to get well-defined random numbers.
// It grabs new ones every time to create better randomness.
func (eph *EPHandle) RandBuffers() {
	sizeR := FillEnvVarInt("RAND_SIZE", 4)
	limitR := FillEnvVarInt("RAND_LIMIT", 10000)
	tailR := FillEnvVarInt("RAND_TAIL", 8)
	modRenv := FillEnvVar("RAND_MOD")
	modR, err := strconv.ParseFloat(modRenv, 64)
	if err != nil {
		modR = 10000
	}

	for _, mt := range eph.MTypes {
		mt.MU.Lock()
		bufferExp := NewRandCycBuffer(sizeR, limitR, tailR, modR, mt.Name)

		bufferExp.MU.Lock()
		eph.MTypes[mt.Name].RandomBuffer = bufferExp.Values

		bufferExp.MU.Unlock()
		mt.MU.Unlock()
	}
}

// ShiftBuffers is the engine for advancing data (mtypes)
// to appear like it moves in a specific algorithmic shape (algos).
func (eph *EPHandle) ShiftBuffers() {
	// Run a Shift() on all CyclicBuffers
	// This advances buffer.Index along the algorithm
	for _, mt := range eph.MTypes {
		for _, buff := range mt.CyclicBuffer {
			buff.Shift()
		}
	}
}

// FillEnvVar returns the value of a runtime Environment Variable
func FillEnvVar(ev string) string {
	// If the EnvVar doesn't exist return a default string
	value := os.Getenv(ev)
	if value == "" {
		value = "ENOENT"
	}
	return value
}

// FillEnvVarInt returns a runtime Environment Variable as an int
// It takes the name of the ENV VAR and a default
// For non-default and string ENV VARs, use FillEnvVar()
func FillEnvVarInt(ev string, def int) int {
	fetch := os.Getenv(ev)
	if fetch == "" {
		return def
	}

	value, err := strconv.Atoi(fetch)
	if err != nil || value < 0 {
		slog.Info("Invalid environment variable " + ev)
		return def
	}
	return value
}
