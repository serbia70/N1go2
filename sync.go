package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type remotePrinterConfigResponse struct {
	Success    bool   `json:"success"`
	Slug       string `json:"slug"`
	MQTTSecret string `json:"mqtt_secret"`
	Error      string `json:"error"`
}

func SyncPrinterConfig(configPath string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	if strings.TrimSpace(cfg.ConfigSyncBaseURL) == "" {
		if err := cfg.ValidateResolved(); err != nil {
			return err
		}
		return nil
	}

	if err := syncRemoteMQTTSecret(cfg); err != nil {
		if !missingOrPlaceholder(cfg.MQTTSecret) {
			log.Printf("[ConfigSync] 拉取远端 secret 失败，继续使用本地缓存: %v", err)
			return nil
		}
		return err
	}

	if err := persistConfig(configPath, *cfg); err != nil {
		log.Printf("[ConfigSync] 写回本地配置失败: %v", err)
	}

	return cfg.ValidateResolved()
}

func syncRemoteMQTTSecret(cfg *Config) error {
	syncURL := strings.TrimRight(strings.TrimSpace(cfg.ConfigSyncBaseURL), "/")
	if syncURL == "" {
		return fmt.Errorf("CONFIG_SYNC_BASE_URL 未配置")
	}

	req, err := http.NewRequest(http.MethodGet, syncURL+"/api/device/printer-config/"+url.PathEscape(cfg.ShopSlug), nil)
	if err != nil {
		return fmt.Errorf("创建同步请求失败: %w", err)
	}
	req.SetBasicAuth(cfg.MQTTUsername, cfg.MQTTPassword)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求远端配置失败: %w", err)
	}
	defer resp.Body.Close()

	var payload remotePrinterConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("解析远端配置失败: %w", err)
	}
	if resp.StatusCode != http.StatusOK || !payload.Success {
		if strings.TrimSpace(payload.Error) != "" {
			return fmt.Errorf("远端配置返回错误: %s", payload.Error)
		}
		return fmt.Errorf("远端配置返回状态码: %d", resp.StatusCode)
	}

	mqttSecret := strings.TrimSpace(payload.MQTTSecret)
	if missingOrPlaceholder(mqttSecret) {
		return fmt.Errorf("远端 mqtt_secret 为空")
	}

	cfg.MQTTSecret = mqttSecret
	log.Printf("[ConfigSync] 已同步最新 mqtt_secret for shop %s", cfg.ShopSlug)
	return nil
}

func persistConfig(configPath string, cfg Config) error {
	if strings.TrimSpace(configPath) == "" {
		return nil
	}

	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	if err := os.WriteFile(configPath, raw, 0644); err != nil {
		return fmt.Errorf("写入配置失败: %w", err)
	}
	return nil
}
