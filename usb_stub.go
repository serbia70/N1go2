//go:build !linux
// +build !linux

package main

import "fmt"

type USBPrinter struct {
	connected bool
}

func NewUSBPrinter(vendorID, productID uint16) *USBPrinter {
	return &USBPrinter{}
}

func (p *USBPrinter) Connect() error {
	return fmt.Errorf("USB printing is only supported on linux targets")
}

func (p *USBPrinter) Disconnect() {
	p.connected = false
}

func (p *USBPrinter) Close() {
	p.connected = false
}

func (p *USBPrinter) Print(data []byte, maxRetries int) error {
	return fmt.Errorf("USB printing is only supported on linux targets")
}

func (p *USBPrinter) IsConnected() bool {
	return p.connected
}

func (p *USBPrinter) WaitForConnection() {
}
