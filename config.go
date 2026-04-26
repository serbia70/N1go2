package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	ConfigSyncBaseURL string `json:"config_sync_base_url"`
	MQTTBroker        string `json:"mqtt_broker"`
	MQTTPort          int    `json:"mqtt_port"`
	MQTTUsername      string `json:"mqtt_username"`
	MQTTPassword      string `json:"mqtt_password"`
	ShopSlug          string `json:"shop_slug"`
	MQTTSecret        string `json:"mqtt_secret"`
	MQTTOrderTopic    string `json:"mqtt_order_topic"`
	PrinterVendor     uint16 `json:"printer_vendor"`
	PrinterProduct    uint16 `json:"printer_product"`
	MaxRetries        int    `json:"max_retries"`
}

type fileConfig struct {
	ConfigSyncBaseURL *string `json:"config_sync_base_url"`
	MQTTBroker        *string `json:"mqtt_broker"`
	MQTTPort          *int    `json:"mqtt_port"`
	MQTTUsername      *string `json:"mqtt_username"`
	MQTTPassword      *string `json:"mqtt_password"`
	ShopSlug          *string `json:"shop_slug"`
	MQTTSecret        *string `json:"mqtt_secret"`
	MQTTOrderTopic    *string `json:"mqtt_order_topic"`
	PrinterVendor     *uint16 `json:"printer_vendor"`
	PrinterProduct    *uint16 `json:"printer_product"`
	MaxRetries        *int    `json:"max_retries"`
}

func DefaultConfig() Config {
	return Config{
		ConfigSyncBaseURL: getEnv("CONFIG_SYNC_BASE_URL", "https://food2api.serbia70.com"),
		MQTTBroker:        getEnv("MQTT_BROKER", "mqtt.serbia70.com"),
		MQTTPort:          getEnvInt("MQTT_PORT", 443),
		MQTTUsername:      getEnv("MQTT_USERNAME", ""),
		MQTTPassword:      getEnv("MQTT_PASSWORD", ""),
		ShopSlug:          getEnv("SHOP_SLUG", ""),
		MQTTSecret:        getEnv("MQTT_SECRET", ""),
		MQTTOrderTopic:    getEnv("MQTT_ORDER_TOPIC", ""),
		PrinterVendor:     uint16(getEnvInt("PRINTER_VENDOR", 0x0483)),
		PrinterProduct:    uint16(getEnvInt("PRINTER_PRODUCT", 0x070b)),
		MaxRetries:        getEnvInt("MAX_RETRIES", 10),
	}
}

func LoadConfig(configPath string) (Config, error) {
	cfg := DefaultConfig()

	if strings.TrimSpace(configPath) != "" {
		raw, err := os.ReadFile(configPath)
		if err != nil {
			return Config{}, fmt.Errorf("读取配置文件失败: %w", err)
		}

		if strings.TrimSpace(string(raw)) != "" {
			var fc fileConfig
			if err := json.Unmarshal(raw, &fc); err != nil {
				return Config{}, fmt.Errorf("解析配置文件失败: %w", err)
			}
			mergeFileConfig(&cfg, fc)
		}
	}

	overrideFromEnv(&cfg)

	if strings.TrimSpace(cfg.MQTTOrderTopic) == "" && strings.TrimSpace(cfg.ShopSlug) != "" && strings.TrimSpace(cfg.MQTTSecret) != "" {
		cfg.MQTTOrderTopic = fmt.Sprintf("restaurant/%s/%s/order", cfg.ShopSlug, cfg.MQTTSecret)
	}

	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 10
	}

	if err := cfg.ValidateBase(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func mergeFileConfig(cfg *Config, fc fileConfig) {
	if fc.ConfigSyncBaseURL != nil {
		cfg.ConfigSyncBaseURL = strings.TrimSpace(*fc.ConfigSyncBaseURL)
	}
	if fc.MQTTBroker != nil {
		cfg.MQTTBroker = strings.TrimSpace(*fc.MQTTBroker)
	}
	if fc.MQTTPort != nil {
		cfg.MQTTPort = *fc.MQTTPort
	}
	if fc.MQTTUsername != nil {
		cfg.MQTTUsername = strings.TrimSpace(*fc.MQTTUsername)
	}
	if fc.MQTTPassword != nil {
		cfg.MQTTPassword = strings.TrimSpace(*fc.MQTTPassword)
	}
	if fc.ShopSlug != nil {
		cfg.ShopSlug = strings.TrimSpace(*fc.ShopSlug)
	}
	if fc.MQTTSecret != nil {
		cfg.MQTTSecret = strings.TrimSpace(*fc.MQTTSecret)
	}
	if fc.MQTTOrderTopic != nil {
		cfg.MQTTOrderTopic = strings.TrimSpace(*fc.MQTTOrderTopic)
	}
	if fc.PrinterVendor != nil {
		cfg.PrinterVendor = *fc.PrinterVendor
	}
	if fc.PrinterProduct != nil {
		cfg.PrinterProduct = *fc.PrinterProduct
	}
	if fc.MaxRetries != nil {
		cfg.MaxRetries = *fc.MaxRetries
	}
}

func overrideFromEnv(cfg *Config) {
	if value := os.Getenv("CONFIG_SYNC_BASE_URL"); strings.TrimSpace(value) != "" {
		cfg.ConfigSyncBaseURL = strings.TrimSpace(value)
	}
	if value := os.Getenv("MQTT_BROKER"); strings.TrimSpace(value) != "" {
		cfg.MQTTBroker = strings.TrimSpace(value)
	}
	if value := getEnvIntMaybe("MQTT_PORT"); value > 0 {
		cfg.MQTTPort = value
	}
	if value := os.Getenv("MQTT_USERNAME"); strings.TrimSpace(value) != "" {
		cfg.MQTTUsername = strings.TrimSpace(value)
	}
	if value := os.Getenv("MQTT_PASSWORD"); strings.TrimSpace(value) != "" {
		cfg.MQTTPassword = strings.TrimSpace(value)
	}
	if value := os.Getenv("SHOP_SLUG"); strings.TrimSpace(value) != "" {
		cfg.ShopSlug = strings.TrimSpace(value)
	}
	if value := os.Getenv("MQTT_SECRET"); strings.TrimSpace(value) != "" {
		cfg.MQTTSecret = strings.TrimSpace(value)
	}
	if value := os.Getenv("MQTT_ORDER_TOPIC"); strings.TrimSpace(value) != "" {
		cfg.MQTTOrderTopic = strings.TrimSpace(value)
	}
	if value := getEnvIntMaybe("PRINTER_VENDOR"); value > 0 {
		cfg.PrinterVendor = uint16(value)
	}
	if value := getEnvIntMaybe("PRINTER_PRODUCT"); value > 0 {
		cfg.PrinterProduct = uint16(value)
	}
	if value := getEnvIntMaybe("MAX_RETRIES"); value > 0 {
		cfg.MaxRetries = value
	}
}

func (c Config) ValidateBase() error {
	if strings.TrimSpace(c.MQTTBroker) == "" {
		return fmt.Errorf("MQTT_BROKER 未配置")
	}
	if c.MQTTPort <= 0 {
		return fmt.Errorf("MQTT_PORT 无效")
	}
	if missingOrPlaceholder(c.MQTTUsername) {
		return fmt.Errorf("MQTT_USERNAME 未配置")
	}
	if missingOrPlaceholder(c.MQTTPassword) {
		return fmt.Errorf("MQTT_PASSWORD 未配置")
	}
	if missingOrPlaceholder(c.ShopSlug) {
		return fmt.Errorf("SHOP_SLUG 未配置")
	}
	if c.PrinterVendor == 0 || c.PrinterProduct == 0 {
		return fmt.Errorf("打印机 VID/PID 未配置")
	}
	return nil
}

func (c Config) ValidateResolved() error {
	if err := c.ValidateBase(); err != nil {
		return err
	}
	if missingOrPlaceholder(c.MQTTSecret) {
		return fmt.Errorf("MQTT_SECRET 未配置")
	}
	return nil
}

func missingOrPlaceholder(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return true
	}
	return strings.EqualFold(trimmed, "CHANGE_ME")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := getEnvIntMaybe(key); value > 0 {
		return value
	}
	return defaultValue
}

func getEnvIntMaybe(key string) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return 0
	}
	var parsed int
	_, err := fmt.Sscanf(value, "%d", &parsed)
	if err != nil {
		return 0
	}
	return parsed
}
