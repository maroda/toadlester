package main

import (
	"math/rand/v2"
	"strconv"
)

// LoopMetrics is a collection that cycles around values to simulate change.
// The idea is that each time an endpoint is called,
// the internal shift register is advanced by one,
// giving the illusion that metrics are changing in
// a cyclical fashion.
type LoopMetrics struct {
	KVmetrics   CycBuffer
	JSONmetrics CycBuffer
}

func (lm *LoopMetrics) MarshalJSON() ([]byte, error) {
	panic("implement me")
}

// CycBuffer is a cyclical shift register
type CycBuffer struct {
	Values  []string // Slice of whatever we need for responses
	MaxSize int      // How big this buffer can be
	Index   int      // We are at this index in the step buffer
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
		Values:  values,
		MaxSize: maxSize,
		Index:   0,
	}
}

// DynaMetrics change based on internal time, when called by endpoint
// they give the illusion of more dynamically changing metrics.
type DynaMetrics struct {
	KVmetrics   CycBuffer
	JSONmetrics CycBuffer
}

func (dm *DynaMetrics) ShiftMetric(buff CycBuffer) error {
	panic("implement me")
}
