package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
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
