package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type OrderItem struct {
	Name      string      `json:"name"`
	ProductID int64       `json:"product_id"`
	Qty       int         `json:"quantity"`
	Price     int64       `json:"price"`
	Total     interface{} `json:"total"`
	Meta      string      `json:"meta"`
}

type Order struct {
	ID          interface{} `json:"id"`
	OrderNo     string      `json:"order_no"`
	TableInfo   string      `json:"table_info"`
	OrderType   string      `json:"order_type"`
	TotalAmount int64       `json:"total_amount"`
	Items       []OrderItem `json:"items"`
	Remarks     string      `json:"remarks"`
	UserPhone   string      `json:"user_phone"`
	UserName    string      `json:"user_name"`
	Shipping    string      `json:"shipping"`
	Date        string      `json:"date"`
	PickupNo    string      `json:"pickup_no"`
}

func (o *Order) IsDelivery() bool {
	rawTable := safeStr(o.TableInfo)
	if strings.Contains(o.OrderType, "delivery") {
		return true
	}
	if strings.Contains(strings.ToLower(rawTable), "waimai") {
		return true
	}
	if len(o.Shipping) > 2 {
		return true
	}
	if strings.Contains(rawTable, ",") && regexp.MustCompile(`\d{6,}`).MatchString(rawTable) {
		return true
	}
	return false
}

func (o *Order) GetTimeStr() string {
	if o.Date != "" && len(o.Date) >= 8 {
		return o.Date[len(o.Date)-8 : len(o.Date)-3]
	}
	return ""
}

func (o *Order) ExtractNote() string {
	if o.Remarks != "" {
		return o.Remarks
	}
	fullText := safeStr(o.TableInfo) + " " + safeStr(o.Shipping)
	re := regexp.MustCompile(`[\(\[（]备注[:：](.*?)[\)\]）]`)
	if match := re.FindStringSubmatch(fullText); len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func (o *Order) ParseShippingInfo() (name, phone, addr string) {
	rawAddr := o.Shipping
	if rawAddr == "" {
		return o.UserName, o.UserPhone, ""
	}
	clean := regexp.MustCompile(`\[.*?\]`).ReplaceAllString(rawAddr, "")
	clean = regexp.MustCompile(`\(.*?\)`).ReplaceAllString(clean, "")
	clean = regexp.MustCompile(`（.*?）`).ReplaceAllString(clean, "")
	parts := strings.Split(clean, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	name = o.UserName
	phone = o.UserPhone
	addr = clean
	if len(parts) >= 3 {
		if regexp.MustCompile(`\d{6,}`).MatchString(parts[1]) {
			name = parts[0]
			phone = parts[1]
			addr = strings.TrimSpace(strings.Trim(strings.Join(parts[2:], ", "), ", "))
		}
	}
	if len(o.UserName) > len(name) || o.UserName == "111" {
		name = o.UserName
	}
	return name, phone, addr
}

func safeStr(val interface{}) string {
	if val == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", val))
}

func safeParseMoney(value interface{}) int64 {
	if value == nil {
		return 0
	}
	s := strings.TrimSpace(fmt.Sprintf("%v", value))
	if s == "" {
		return 0
	}
	re := regexp.MustCompile(`[^\d.]`)
	cleanS := re.ReplaceAllString(s, "")
	if cleanS == "" {
		return 0
	}
	f, err := strconv.ParseFloat(cleanS, 64)
	if err != nil {
		return 0
	}
	return int64(f)
}
