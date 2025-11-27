package main

import (
	"strconv"
	"testing"
)

func TestNewRandCycBuffer(t *testing.T) {
	tests := []struct {
		name   string
		size   int
		limit  int
		tail   int
		mod    float64
		format string
	}{
		{name: "Returns parseable exponent", size: 1, limit: 10000, tail: 8, mod: 10000, format: "exp"},
		{name: "Returns parseable float", size: 1, limit: 1000, tail: 4, mod: 1000, format: "float"},
		{name: "Returns parseable integer", size: 1, limit: 1000, tail: 4, mod: 1000, format: "int"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			get := NewRandCycBuffer(tt.size, tt.limit, tt.tail, tt.mod, tt.format)
			got := get.Values
			for _, v := range got {
				// The value is random, it won't be useful to test,
				// so if ParseFloat errors, something is wrong with the number.
				_, err := strconv.ParseFloat(v, 64)
				assertError(t, err, nil)
			}
		})
	}
}
