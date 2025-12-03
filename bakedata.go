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
	NType   string   // Numeric Type
	MAlgo   string   // Metric Algorithm Name
	Values  []string // Slice of whatever we need for responses
	MaxSize int      // How big this buffer can be
	Index   int      // We are at this index in the step buffer
}

// NewShiftCycBuffer creates cascading values to be shifted by one for each access
func NewShiftCycBuffer(maxSize, limit, tail int, mod float64, f, a string) *CycBuffer {
	values := make([]string, 0, maxSize)
	saltF := mod * float64(rand.Int32N(int32(limit))) * rand.Float64()
	saltI := int(mod) * int(saltF)

	switch a {
	case "up":
		switch f {
		case "exp":
			for i := 0; i < maxSize; i++ {
				values = append(values, strconv.FormatFloat(saltF*float64(limit-(limit-(tail*i)))*mod, 'e', 8, 64))
			}
		case "float":
			for i := 0; i < maxSize; i++ {
				values = append(values, strconv.FormatFloat(saltF*float64(limit-(limit-(tail*i)))*mod, 'f', 8, 64))
			}
		case "int":
			for i := 0; i < maxSize; i++ {
				values = append(values, strconv.Itoa(saltI*(limit-(limit-(tail*i)))))
			}
		}
	case "down":
		switch f {
		case "exp":
			for i := 0; i < maxSize; i++ {
				values = append(values, strconv.FormatFloat(saltF*float64(limit-(tail*i))*mod, 'e', 8, 64))
			}
		case "float":
			for i := 0; i < maxSize; i++ {
				values = append(values, strconv.FormatFloat(saltF*float64(limit-(tail*i))*mod, 'f', 8, 64))
			}
		case "int":
			for i := 0; i < maxSize; i++ {
				values = append(values, strconv.Itoa(saltI*(limit-(tail*i))))
			}
		}
	case "random":
		switch f {
		case "exp":
			for i := 0; i < maxSize; i++ {
				values = append(values, strconv.FormatFloat(saltF, 'e', tail, 64))
			}
		case "float":
			for i := 0; i < maxSize; i++ {
				values = append(values, strconv.FormatFloat(saltF, 'f', tail, 64))
			}
		case "int":
			for i := 0; i < maxSize; i++ {
				values = append(values, strconv.Itoa(int(saltF)))
			}
		}
	}

	return &CycBuffer{
		NType:   f,
		MAlgo:   a,
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

// RandBuffers is the engine for building random data buffers.
// Each of these can be queried by the endpoint to get well-defined random numbers.
// It grabs new ones every time to create better randomness.
func (eph *EPHandle) RandBuffers() {
	for _, mt := range eph.MTypes {
		mt.MU.Lock()
		buffer := getRandomizedBuffer(mt.Name, "random")
		buffer.MU.Lock()
		eph.MTypes[mt.Name].RandomBuffer = buffer.Values
		buffer.MU.Unlock()
		mt.MU.Unlock()
	}
}

// ShiftBuffers is the engine for advancing data (mtypes)
// to appear like it moves in a specific algorithmic shape (algos).
func (eph *EPHandle) ShiftBuffers() {
	// Run a Shift() on all CyclicBuffers
	// This advances buffer.Index along the algorithm
	for _, mt := range eph.MTypes {
		for _, buff := range mt.ShiftRegisters {
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
		slog.Warn("Invalid environment variable " + ev)
		return def
	}
	return value
}
