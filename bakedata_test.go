package main

import (
	"os"
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
			get := NewShiftCycBuffer(tt.size, tt.limit, tt.tail, tt.mod, tt.format, "random")
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

func TestCycBuffer_ShiftRegister(t *testing.T) {
	tests := []struct {
		name   string
		size   int
		limit  int
		tail   int
		mod    float64
		format string
		algo   string
	}{
		{name: "Returns integers going up", size: 5, limit: 100, tail: 1, mod: 100, format: "int", algo: "up"},
		{name: "Returns integers going down", size: 5, limit: 100, tail: 1, mod: 100, format: "int", algo: "down"},
		{name: "Returns floats going up", size: 10, limit: 10, tail: 1, mod: 1.1, format: "float", algo: "up"},
		{name: "Returns floats going down", size: 10, limit: 10, tail: 1, mod: 1.1, format: "float", algo: "down"},
		{name: "Returns exponentials going up", size: 10, limit: 10000, tail: 1, mod: 100.1, format: "exp", algo: "up"},
		{name: "Returns exponentials going down", size: 10, limit: 10000, tail: 1, mod: 100.1, format: "exp", algo: "down"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shifter := NewShiftCycBuffer(tt.size, tt.limit, tt.tail, tt.mod, tt.format, tt.algo)
			for i := 0; i < 5; i++ {
				t.Log(shifter.Values[i])
				nextIdx := (shifter.Index + 1) % shifter.MaxSize
				before := shifter.Values[shifter.Index]
				check := shifter.Shift()
				after := shifter.Values[nextIdx]
				if check != after {
					t.Errorf("Expected %s to change to %s, got %s", before, after, check)
				}
			}
		})
	}
}

func TestCycBuffer_ShiftRandomizer(t *testing.T) {
	tests := []struct {
		name   string
		size   int
		limit  int
		tail   int
		mod    float64
		format string
		algo   string
	}{
		{name: "Returns random exp", size: 5, limit: 5000, tail: 1, mod: 5000, format: "exp", algo: "random"},
		{name: "Returns random float", size: 5, limit: 100, tail: 1, mod: 1, format: "float", algo: "random"},
		{name: "Returns random int", size: 5, limit: 100, tail: 1, mod: 1, format: "int", algo: "random"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			get := NewShiftCycBuffer(tt.size, tt.limit, tt.tail, tt.mod, tt.format, tt.algo)
			got := get.Values
			for _, v := range got {
				vf, err := strconv.ParseFloat(v, 64)
				assertError(t, err, nil)
				vi := int(vf)
				t.Log(vf, vi)

				if vi < 0 || vi > (tt.limit*int(tt.mod)) {
					t.Errorf("Expected %T to be under limit (%d * %T)", vi, tt.limit, tt.mod)
				}
			}
		})
	}

}

func TestEPHandle_RandBuffers(t *testing.T) {
	eph := NewEPHandle([]string{"exp", "float", "int"}, []string{"up", "down"})
	defer eph.Ticker.Stop()

	// Choose a non-default value
	os.Setenv("RAND_SIZE", "10")

	eph.RandBuffers()

	for _, mt := range eph.MTypes {
		if len(mt.RandomBuffer) != 10 {
			t.Errorf("Random buffer length should be 10, got %d", len(mt.RandomBuffer))
		}
		t.Log(mt.RandomBuffer)
	}

}

func TestFillEnvVarInt(t *testing.T) {

	t.Run("returns the set default", func(t *testing.T) {
		ev := "ANYTHING"
		evDefault := 100
		want := evDefault
		got := FillEnvVarInt(ev, evDefault)

		assertInt(t, got, want)
	})

	t.Run("returns a set value", func(t *testing.T) {
		ev := "MEASUREMENT"
		evDefault := 123123
		want := evDefault

		// Set an env var to check
		err := os.Setenv(ev, strconv.Itoa(evDefault))
		assertError(t, err, nil)

		got := FillEnvVarInt(ev, evDefault)
		assertInt(t, got, want)
	})

	t.Run("Returns set default when OS variable is invalid", func(t *testing.T) {
		ev2 := "LIMITER"
		ev2default := 123
		want := ev2default

		// Set an OS version of the Env Var to an invalid value
		ev2set := -1
		err := os.Setenv(ev2, strconv.Itoa(ev2set))
		assertError(t, err, nil)

		// This will also trigger a log entry like: "Invalid environment variable"
		got := FillEnvVarInt(ev2, ev2default)
		assertInt(t, got, want)
	})
}

func TestFillEnvVar(t *testing.T) {

	t.Run("returns a default value", func(t *testing.T) {
		ev := "ANYTHING"
		want := "ENOENT"
		got := FillEnvVar(ev)

		assertStringContains(t, got, want)
	})

	t.Run("returns a set value", func(t *testing.T) {
		ev := "TOKEN"
		want := "ghp_1q2w3e4r5t6y7u8i9o0p"

		// Set an env var to check
		err := os.Setenv(ev, want)
		if err != nil {
			t.Errorf("could not set env var: %s", ev)
		}

		got := FillEnvVar(ev)
		assertStringContains(t, got, want)
	})
}
