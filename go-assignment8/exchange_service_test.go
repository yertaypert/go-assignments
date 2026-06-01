package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExchangeService_GetRate(t *testing.T) {
	t.Run("Successfull scenario", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(RateResponse{
				Base:   "USD",
				Target: "EUR",
				Rate:   0.85,
			})
		}))
		defer server.Close()

		service := NewExchangeService(server.URL)
		rate, err := service.GetRate("USD", "EUR")

		assert.NoError(t, err)
		assert.Equal(t, 0.85, rate)
	})

	t.Run("API Business Error - 400 Bad Request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "invalid currency pair"}`))
		}))
		defer server.Close()

		service := NewExchangeService(server.URL)
		_, err := service.GetRate("INVALID", "PAIR")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "api error: invalid currency pair")
	})

	t.Run("API Business Error - 404 Not Found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "invalid currency pair"}`))
		}))
		defer server.Close()

		service := NewExchangeService(server.URL)
		_, err := service.GetRate("USD", "UNKNOWN")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "api error: invalid currency pair")
	})

	t.Run("Malformed JSON - Truncated string", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"base": "USD", "rate":`))
		}))
		defer server.Close()

		service := NewExchangeService(server.URL)
		_, err := service.GetRate("USD", "EUR")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error")
	})

	t.Run("Malformed JSON - Internal Server Error message as text", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error")) // Not valid JSON
		}))
		defer server.Close()

		service := NewExchangeService(server.URL)
		_, err := service.GetRate("USD", "EUR")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error")
	})

	t.Run("Slow Response/Timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(200 * time.Millisecond) // Simulate slow response
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(RateResponse{Rate: 1.0})
		}))
		defer server.Close()

		service := NewExchangeService(server.URL)
		// Set a very short timeout to trigger it
		service.Client.Timeout = 50 * time.Millisecond

		_, err := service.GetRate("USD", "EUR")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "network error")
		assert.Contains(t, err.Error(), "Client.Timeout exceeded")
	})

	t.Run("Server Panic / 500 Internal Server Error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "unexpected crash")
		}))
		defer server.Close()

		service := NewExchangeService(server.URL)
		_, err := service.GetRate("USD", "EUR")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error")
	})

	t.Run("Empty Body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		service := NewExchangeService(server.URL)
		_, err := service.GetRate("USD", "EUR")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode error")
	})
}
