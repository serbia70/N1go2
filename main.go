package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configPath := flag.String("config", "./config.json", "Path to config file")
	flag.Parse()

	log.Println("========================================")
	log.Println("  点餐打印服务 - N1 终端")
	log.Println("========================================")

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	if err := SyncPrinterConfig(*configPath, &cfg); err != nil {
		log.Fatalf("同步打印配置失败: %v", err)
	}

	log.Printf("配置: MQTT=%s:%d, Shop=%s, Printer=%04X:%04X, Retries=%d",
		cfg.MQTTBroker, cfg.MQTTPort, cfg.ShopSlug, cfg.PrinterVendor, cfg.PrinterProduct, cfg.MaxRetries)

	svc := NewService(cfg)
	if err := svc.Init(); err != nil {
		log.Fatalf("初始化失败: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("服务运行中，按 Ctrl+C 停止...")
	<-sigChan

	log.Println("停止服务...")
	svc.Stop()
	log.Println("已停止")
}
