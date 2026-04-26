package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Service struct {
	config        Config
	client        mqtt.Client
	printer       *USBPrinter
	stopChan      chan struct{}
	retryCount    int
	lastOrderID   string
	lastOrderTime int64
}

func NewService(cfg Config) *Service {
	return &Service{
		config:     cfg,
		retryCount: cfg.MaxRetries,
		stopChan:   make(chan struct{}),
	}
}

func (s *Service) Init() error {
	s.printer = NewUSBPrinter(s.config.PrinterVendor, s.config.PrinterProduct)

	if err := s.initMQTT(); err != nil {
		return err
	}

	log.Println("=== 打印服务启动成功 ===")
	return nil
}

func (s *Service) Stop() {
	close(s.stopChan)
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect(250)
	}
	if s.printer != nil {
		s.printer.Close()
	}
}

func (s *Service) initMQTT() error {
	orderTopic := fmt.Sprintf("restaurant/%s/%s/order", s.config.ShopSlug, s.config.MQTTSecret)
	printTopic := fmt.Sprintf("restaurant/%s/%s/print", s.config.ShopSlug, s.config.MQTTSecret)
	if s.config.MQTTOrderTopic != "" {
		orderTopic = s.config.MQTTOrderTopic
	}
	controlTopic := fmt.Sprintf("restaurant/%s/%s/control", s.config.ShopSlug, s.config.MQTTSecret)

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("wss://%s:%d/mqtt", s.config.MQTTBroker, s.config.MQTTPort))
	opts.SetClientID(fmt.Sprintf("printer-n1-%d", time.Now().Unix()))
	opts.SetUsername(s.config.MQTTUsername)
	opts.SetPassword(s.config.MQTTPassword)
	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetryInterval(5 * time.Second)

	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Println("[MQTT] connected")
		client.Subscribe(orderTopic, 1, s.handleOrderMessage)
		client.Subscribe(printTopic, 1, s.handleOrderMessage)
		client.Subscribe(controlTopic, 0, s.handleControlMessage)
		log.Printf("[MQTT] subscribed %s", orderTopic)
		log.Printf("[MQTT] subscribed %s", printTopic)
	})

	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("[MQTT] 连接断开: %v", err)
	})

	s.client = mqtt.NewClient(opts)

	log.Printf("连接 MQTT: %s...", s.config.MQTTBroker)
	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("MQTT 连接失败: %w", token.Error())
	}

	return nil
}

func (s *Service) handleOrderMessage(client mqtt.Client, msg mqtt.Message) {
	var order Order
	if err := json.Unmarshal(msg.Payload(), &order); err != nil {
		log.Printf("[Error] 解析订单失败: %v", err)
		return
	}

	// Ignore non-order events such as table checkout status pushes.
	if len(order.Items) == 0 {
		log.Printf("[Order] skip non-order payload on %s", msg.Topic())
		return
	}

	orderID := fmt.Sprintf("%v", order.ID)
	now := time.Now().Unix()

	if orderID == s.lastOrderID && now-s.lastOrderTime < 5 {
		log.Printf("[Order] 忽略重复订单(5秒内): %s", orderID)
		return
	}
	s.lastOrderID = orderID
	s.lastOrderTime = now

	log.Printf("[Order] 原始数据: %s", string(msg.Payload()))
	log.Printf("[Order] 收到订单 #%s", safeStr(order.ID))
	log.Printf("[Order] 订单金额: %d", order.TotalAmount)
	log.Printf("[Order] 菜品数量: %d", len(order.Items))
	for i, item := range order.Items {
		log.Printf("[Order] 菜品%d: %s x%d, 小计: %v", i+1, safeStr(item.Name), item.Qty, item.Total)
	}

	if !s.printer.IsConnected() {
		log.Println("[Order] 等待打印机...")
		s.printer.WaitForConnection()
	}

	data := FormatOrder(order)
	log.Printf("[Order] 打印数据长度: %d 字节", len(data))
	log.Printf("[Order] 前20字节: %v", data[:min(20, len(data))])

	printData := string(data)
	log.Printf("[Order] 打印内容预览: %s", printData[:min(100, len(printData))])

	if err := s.printer.Print(data, s.retryCount); err != nil {
		log.Printf("[Error] 打印失败: %v", err)
	} else {
		log.Printf("[OK] order #%s printed", safeStr(order.ID))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *Service) handleControlMessage(client mqtt.Client, msg mqtt.Message) {
	var data map[string]interface{}
	if err := json.Unmarshal(msg.Payload(), &data); err != nil {
		return
	}

	action := safeStr(data["action"])
	switch action {
	case "shutdown":
		log.Println("[Control] 关机指令")
		exec.Command("shutdown", "-h", "now").Run()
	case "reboot":
		log.Println("[Control] 重启指令")
		exec.Command("shutdown", "-r", "now").Run()
	}
}
