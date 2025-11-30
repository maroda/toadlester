package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSetupMux_ResponseCodes(t *testing.T) {
	eph := NewEPHandle([]string{"exp", "float", "int"}, []string{"up", "down"})
	defer eph.Ticker.Stop()
	mux := eph.SetupMux()

	tests := []struct {
		name     string
		target   string
		wantCode int
	}{
		{name: "KV static data answers", target: "/ep/kv", wantCode: http.StatusOK},
		{name: "JSON static data answers", target: "/ep/json", wantCode: http.StatusOK},
		{name: "Randomized data answers", target: "/rand/all", wantCode: http.StatusOK},
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

func TestSetupMux_Data(t *testing.T) {
	eph := NewEPHandle([]string{"exp", "float", "int"}, []string{"up", "down", "floatup", "floatdown"})
	defer eph.Ticker.Stop()
	mux := eph.SetupMux()

	// Do not test actual values because they change based on access and timing
	tests := []struct {
		name     string
		target   string
		wantCode int
		expect   string
	}{
		{name: "Exponent Walk Up", target: "/series/exp/up", wantCode: http.StatusOK, expect: "Metric_exp: "},
		{name: "Exponent Walk Down", target: "/series/exp/down", wantCode: http.StatusOK, expect: "Metric_exp: "},
		{name: "Float Walk Up", target: "/series/float/floatup", wantCode: http.StatusOK, expect: "Metric_float: "},
		{name: "Float Walk Down", target: "/series/float/floatdown", wantCode: http.StatusOK, expect: "Metric_float: "},
		{name: "Integer Walk Up", target: "/series/int/up", wantCode: http.StatusOK, expect: "Metric_int: "},
		{name: "Integer Walk Down", target: "/series/int/down", wantCode: http.StatusOK, expect: "Metric_int: "},
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
