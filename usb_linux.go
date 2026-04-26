//go:build linux
// +build linux

package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/gousb"
)

type USBPrinter struct {
	vendorID  gousb.ID
	productID gousb.ID
	ctx       *gousb.Context
	dev       *gousb.Device
	intf      *gousb.Interface
	endpoint  *gousb.OutEndpoint
	connected bool
}

func NewUSBPrinter(vendorID, productID uint16) *USBPrinter {
	return &USBPrinter{
		vendorID:  gousb.ID(vendorID),
		productID: gousb.ID(productID),
		connected: false,
	}
}

func (p *USBPrinter) Connect() error {
	if p.connected {
		return nil
	}

	if p.ctx == nil {
		p.ctx = gousb.NewContext()
	}

	devs, err := p.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == p.vendorID && desc.Product == p.productID
	})
	if err != nil {
		return fmt.Errorf("查找设备失败: %w", err)
	}

	if len(devs) == 0 {
		return fmt.Errorf("未找到打印机 (VID:%04X PID:%04X)", p.vendorID, p.productID)
	}

	p.dev = devs[0]

	for i := 1; i < len(devs); i++ {
		devs[i].Close()
	}

	cfg, err := p.dev.Config(1)
	if err != nil {
		p.dev.Close()
		p.dev = nil
		return fmt.Errorf("获取配置失败: %w", err)
	}

	intf, err := cfg.Interface(0, 0)
	if err != nil {
		cfg.Close()
		p.dev.Close()
		p.dev = nil
		return fmt.Errorf("声明接口失败: %w", err)
	}

	endpoint, err := intf.OutEndpoint(0x01)
	if err != nil {
		intf.Close()
		cfg.Close()
		p.dev.Close()
		p.dev = nil
		return fmt.Errorf("获取端点失败: %w", err)
	}

	p.intf = intf
	p.endpoint = endpoint
	p.connected = true

	log.Printf("[USB] 打印机已连接 (VID:%04X PID:%04X)", p.vendorID, p.productID)
	return nil
}

func (p *USBPrinter) Disconnect() {
	if p.intf != nil {
		p.intf.Close()
		p.intf = nil
	}
	if p.dev != nil {
		p.dev.Close()
		p.dev = nil
	}
	p.endpoint = nil
	p.connected = false
	log.Println("[USB] 打印机已断开")
}

func (p *USBPrinter) Close() {
	p.Disconnect()
	if p.ctx != nil {
		p.ctx.Close()
		p.ctx = nil
	}
}

func (p *USBPrinter) Print(data []byte, maxRetries int) error {
	if !p.connected {
		if err := p.Connect(); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
		for i := 0; i < 10; i++ {
			p.endpoint.Write([]byte{0x0A})
		}
		time.Sleep(50 * time.Millisecond)
		p.endpoint.Write([]byte{0x1B, 0x40})
		time.Sleep(50 * time.Millisecond)
	}

	chunkSize := 64
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		chunk := data[i:end]

		var lastErr error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			_, err := p.endpoint.Write(chunk)
			if err == nil {
				break
			}
			lastErr = err
			if attempt < maxRetries {
				log.Printf("[USB] 写入失败，第%d次重试", attempt+1)
				time.Sleep(100 * time.Millisecond)
			}
		}

		if lastErr != nil {
			return fmt.Errorf("打印失败: %w", lastErr)
		}

		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

func (p *USBPrinter) IsConnected() bool {
	return p.connected
}

func (p *USBPrinter) WaitForConnection() {
	log.Println("[USB] 等待打印机连接...")
	for {
		if err := p.Connect(); err == nil {
			return
		}
		time.Sleep(5 * time.Second)
	}
}
