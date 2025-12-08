package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestSetupMux_Data(t *testing.T) {
	eph := NewEPHandle([]string{"exp", "float", "int"}, []string{"up", "down"})
	defer eph.Ticker.Stop()
	mux := eph.SetupMux()

	// Do not test actual values because they are randomized
	tests := []struct {
		name     string
		target   string
		wantCode int
		expect   string
	}{
		{name: "Randomizer", target: "/rand/all", wantCode: http.StatusOK, expect: "ExpMetric: "},
		{name: "Exponent Walk Up", target: "/series/exp/up", wantCode: http.StatusOK, expect: "Metric_exp_up: "},
		{name: "Exponent Walk Down", target: "/series/exp/down", wantCode: http.StatusOK, expect: "Metric_exp_down: "},
		{name: "Integer Walk Up", target: "/series/int/up", wantCode: http.StatusOK, expect: "Metric_int_up: "},
		{name: "Integer Walk Down", target: "/series/int/down", wantCode: http.StatusOK, expect: "Metric_int_down: "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.target, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			assertStatus(t, w.Code, tt.wantCode)
			assertStringContains(t, w.Body.String(), tt.expect)
		})
	}
}

func TestEPHandle_SeriesInternalDataHandler(t *testing.T) {
	eph := NewEPHandle([]string{"exp", "float", "int"}, []string{"up", "down"})
	defer eph.Ticker.Stop()
	mux := eph.SetupMux()

	tests := []struct {
		name     string
		target   string
		wantCode int
	}{
		{
			name:     "Logs Error for bad Type",
			target:   "/series/ex/up",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Logs Error for bad Algo",
			target:   "/series/exp/u",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Logs Error for too few elements",
			target:   "/series/exp/",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "Logs Error for too many elements",
			target:   "/series/exp/up/for/ever",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.target, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			assertStatus(t, w.Code, tt.wantCode)
		})
	}

}

func TestEPHandle_SeriesDataAllHandler(t *testing.T) {
	eph := NewEPHandle([]string{"exp", "float", "int"}, []string{"up", "down"})
	defer eph.Ticker.Stop()
	mux := eph.SetupMux()

	tests := []struct {
		name     string
		target   string
		wantCode int
	}{
		{
			name:     "Retrieves full metrics page",
			target:   "/metrics",
			wantCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.target, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			assertStatus(t, w.Code, tt.wantCode)
		})
	}
}

func TestEPHandle_ResetHandler(t *testing.T) {
	eph := NewEPHandle([]string{"exp", "float", "int"}, []string{"up", "down"})
	defer eph.Ticker.Stop()
	mux := eph.SetupMux()

	tests := []struct {
		name     string
		target   string
		wantCode int
		mtype    string
		typecnf  string
		envvar   string
		oldval   string
		newval   string
	}{
		{
			name:     "Path is too short",
			target:   "/reset/INT_SIZE",
			wantCode: http.StatusBadRequest,
			mtype:    "int",
			typecnf:  "size",
			envvar:   "INT_SIZE",
			oldval:   strconv.Itoa(defSize),
			newval:   "10",
		},
		{
			name:     "Path is too long",
			target:   "/reset/INT_SIZE/11/ok",
			wantCode: http.StatusBadRequest,
			mtype:    "int",
			typecnf:  "size",
			envvar:   "INT_SIZE",
			oldval:   strconv.Itoa(defSize),
			newval:   "10",
		},
		{
			name:     "No such env var",
			target:   "/reset/ONT_SIZE/11",
			wantCode: http.StatusBadRequest,
			mtype:    "int",
			typecnf:  "size",
			envvar:   "ONT_SIZE",
			oldval:   strconv.Itoa(defSize),
			newval:   "10",
		},
		{
			name:     "Reset Int Size",
			target:   "/reset/INT_SIZE/11",
			wantCode: http.StatusOK,
			mtype:    "int",
			typecnf:  "size",
			envvar:   "INT_SIZE",
			oldval:   strconv.Itoa(defSize),
			newval:   "11",
		},
		{
			name:     "Reset Int Limit",
			target:   "/reset/INT_LIMIT/22",
			wantCode: http.StatusOK,
			mtype:    "int",
			typecnf:  "limit",
			envvar:   "INT_LIMIT",
			oldval:   strconv.Itoa(defLimit),
			newval:   "22",
		},
		{
			name:     "Reset Float Tail",
			target:   "/reset/FLOAT_TAIL/8",
			wantCode: http.StatusOK,
			mtype:    "float",
			typecnf:  "tail",
			envvar:   "FLOAT_TAIL",
			oldval:   strconv.Itoa(defTail),
			newval:   "8",
		},
		{
			name:     "Reset Exp Tail",
			target:   "/reset/EXP_TAIL/8",
			wantCode: http.StatusOK,
			mtype:    "exp",
			typecnf:  "tail",
			envvar:   "EXP_TAIL",
			oldval:   strconv.Itoa(defTail),
			newval:   "8",
		},
		{
			name:     "Reset Exp Mod",
			target:   "/reset/EXP_MOD/111.11",
			wantCode: http.StatusOK,
			mtype:    "exp",
			typecnf:  "mod",
			envvar:   "EXP_MOD",
			oldval:   strconv.Itoa(defMod),
			newval:   "111.11",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.target, nil)
			w := httptest.NewRecorder()

			// Set the Env Var to the old value first to ensure it is changed
			os.Setenv(tt.envvar, tt.oldval)

			// Serving and fetching this endpoint will automatically reset the Env Var
			mux.ServeHTTP(w, r)

			// Now check if it was changed
			assertStatus(t, w.Code, tt.wantCode)
			if os.Getenv(tt.envvar) != tt.newval {
				t.Errorf("Expected variable to change to %s, got %s", tt.newval, os.Getenv(tt.envvar))
			}

			// Check data for each type
			switch tt.typecnf {
			case "size":
				// Size is a 1-to-1 relationship
				newL := strconv.Itoa(len(eph.MTypes[tt.mtype].ShiftRegisters["up"].Values))
				if newL != tt.newval {
					t.Errorf("Expected values length to be %s, got %s", tt.newval, newL)
				}
			case "limit":
				// LIMIT is not a literal limit, but part of a function
				oldI, _ := strconv.Atoi(tt.oldval) // Lower bounds
				newI, _ := strconv.Atoi(tt.newval) // New config
				topI := newI << 4                  // Bitwise change to match algorithm output

				ep := eph.MTypes[tt.mtype].ShiftRegisters["up"]
				lastval := ep.Values[len(ep.Values)-1]
				intval, err := strconv.Atoi(lastval)
				assertError(t, err, nil)

				// Value should lie within bounds
				if intval < oldI || intval > topI {
					t.Errorf("Expected value to be in range %d-%d, got %d", oldI, topI, intval)
				}
			case "tail":
				// in a float is precision after the decimal,
				// in an exponent it is the size of the mantissa.
				ep := eph.MTypes[tt.mtype].ShiftRegisters["up"]
				switch tt.mtype {
				case "exp":
					newI, _ := strconv.Atoi(tt.newval)
					lastval := ep.Values[len(ep.Values)-1]
					parts := strings.Split(lastval, "e")
					mantissa := parts[0]
					dotIndex := strings.Index(mantissa, ".")
					floatPrecision := len(mantissa) - dotIndex - 1 // Chars after the mantissa dot
					if floatPrecision != newI {
						t.Errorf("Expected exponent to have mantissa %d, got %d", newI, floatPrecision)
					}
				case "float":
					newI, _ := strconv.Atoi(tt.newval)
					lastval := ep.Values[len(ep.Values)-1]
					dotIndex := strings.Index(lastval, ".")
					floatPrecision := len(lastval) - dotIndex - 1 // Chars after the decimal point
					if floatPrecision != newI {
						t.Errorf("Expected float to have precision %d, got %d", newI, floatPrecision)
					}
				}
			case "mod":
				// MOD is similar to LIMIT, but it is a float.
				// They are used together to make big numbers.
				oldF, _ := strconv.ParseFloat(tt.oldval, 64) // Lower bounds
				newF, _ := strconv.ParseFloat(tt.newval, 64) // New config
				topF := newF * newF * newF                   // Match algorithm output size

				ep := eph.MTypes[tt.mtype].ShiftRegisters["up"]
				lastval := ep.Values[len(ep.Values)-1]
				floatval, err := strconv.ParseFloat(lastval, 64)
				assertError(t, err, nil)

				// Value should lie within bounds
				if floatval < oldF || floatval > topF {
					t.Errorf("Expected value to be in range %f-%f, got %f", oldF, topF, floatval)
				}
			}
		})
	}
}

// Helpers //

func assertError(t testing.TB, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("got error %q want %q", got, want)
	}
}

func assertGotError(t testing.TB, got error) {
	t.Helper()
	if got == nil {
		t.Errorf("Expected an error but got %q", got)
	}
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}

func assertInt(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct value, got %d, want %d", got, want)
	}
}

func assertInt64(t *testing.T, got, want int64) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct value, got %d, want %d", got, want)
	}
}

func assertStringContains(t *testing.T, full, want string) {
	t.Helper()
	if !strings.Contains(full, want) {
		t.Errorf("Did not find %q, expected string contains %q", want, full)
	}
}
