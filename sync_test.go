package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncPrinterConfigFetchesAndPersistsRemoteSecret(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "restaurant_user" || pass != "restaurant_password_2024" {
			http.Error(w, `{"success":false,"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/api/device/printer-config/101" {
			http.Error(w, `{"success":false,"error":"unexpected path"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"slug":"101","mqtt_secret":"fresh-secret-101"}`))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	cfg := Config{
		ConfigSyncBaseURL: server.URL,
		MQTTBroker:        "mqtt.serbia70.com",
		MQTTPort:          443,
		MQTTUsername:      "restaurant_user",
		MQTTPassword:      "restaurant_password_2024",
		ShopSlug:          "101",
		MQTTSecret:        "old-secret",
		PrinterVendor:     1155,
		PrinterProduct:    1803,
		MaxRetries:        10,
	}

	if err := SyncPrinterConfig(configPath, &cfg); err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	if cfg.MQTTSecret != "fresh-secret-101" {
		t.Fatalf("mqtt secret = %q, want %q", cfg.MQTTSecret, "fresh-secret-101")
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(raw), "fresh-secret-101") {
		t.Fatalf("persisted config missing secret: %s", string(raw))
	}
}

func TestSyncPrinterConfigFallsBackToLocalSecretWhenRemoteFails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"success":false,"error":"backend unavailable"}`, http.StatusServiceUnavailable)
	}))
	defer server.Close()

	cfg := Config{
		ConfigSyncBaseURL: server.URL,
		MQTTBroker:        "mqtt.serbia70.com",
		MQTTPort:          443,
		MQTTUsername:      "restaurant_user",
		MQTTPassword:      "restaurant_password_2024",
		ShopSlug:          "101",
		MQTTSecret:        "cached-secret-101",
		PrinterVendor:     1155,
		PrinterProduct:    1803,
		MaxRetries:        10,
	}

	if err := SyncPrinterConfig("", &cfg); err != nil {
		t.Fatalf("expected local fallback to succeed, got %v", err)
	}
	if cfg.MQTTSecret != "cached-secret-101" {
		t.Fatalf("mqtt secret = %q, want cached local secret", cfg.MQTTSecret)
	}
}
