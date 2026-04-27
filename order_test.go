package main

import "testing"

func TestParseShippingInfoFallsBackToTableInfo(t *testing.T) {
	order := Order{
		TableInfo: "hui, 0613083888, ruma1 [货到付款/Cash] (备注:)",
	}

	name, phone, addr := order.ParseShippingInfo()
	if name != "hui" {
		t.Fatalf("name = %q, want %q", name, "hui")
	}
	if phone != "0613083888" {
		t.Fatalf("phone = %q, want %q", phone, "0613083888")
	}
	if addr != "ruma1" {
		t.Fatalf("addr = %q, want %q", addr, "ruma1")
	}
}

func TestExtractNoteScansLegacyFallbackFields(t *testing.T) {
	order := Order{
		TableInfo: "hui, 0613083888, ruma1 [货到付款/Cash] (备注: 不要辣)",
	}

	note := order.ExtractNote()
	if note != "不要辣" {
		t.Fatalf("note = %q, want %q", note, "不要辣")
	}
}
