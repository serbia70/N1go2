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
	LegacyTable string      `json:"table"`
	OrderType   string      `json:"order_type"`
	LegacyType  string      `json:"type"`
	TotalAmount int64       `json:"total_amount"`
	Items       []OrderItem `json:"items"`
	Remarks     string      `json:"remarks"`
	LegacyNote  string      `json:"note"`
	UserPhone   string      `json:"user_phone"`
	LegacyPhone string      `json:"phone"`
	UserName    string      `json:"user_name"`
	LegacyName  string      `json:"cust_name"`
	Shipping    string      `json:"shipping"`
	LegacyInfo  string      `json:"info"`
	Date        string      `json:"date"`
	PickupNo    string      `json:"pickup_no"`
}

func (o *Order) IsDelivery() bool {
	rawTable := safeStr(o.primaryTableInfo())
	if strings.Contains(strings.ToLower(o.primaryOrderType()), "delivery") {
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
	if note := safeStr(o.Remarks); note != "" {
		return note
	}
	if note := safeStr(o.LegacyNote); note != "" {
		return note
	}
	fullText := safeStr(o.primaryTableInfo()) + " " + safeStr(o.Shipping) + " " + safeStr(o.LegacyInfo)
	re := regexp.MustCompile(`[\(\[（]备注[:：](.*?)[\)\]）]`)
	if match := re.FindStringSubmatch(fullText); len(match) > 1 {
		return strings.TrimSpace(match[1])
	}
	return ""
}

func (o *Order) ParseShippingInfo() (name, phone, addr string) {
	rawAddr := o.Shipping
	if rawAddr == "" {
		rawTable := safeStr(o.primaryTableInfo())
		if len(rawTable) > 10 {
			rawAddr = rawTable
		}
	}
	if rawAddr == "" {
		return safeStr(o.primaryUserName()), safeStr(o.primaryUserPhone()), ""
	}
	clean := regexp.MustCompile(`\[.*?\]`).ReplaceAllString(rawAddr, "")
	clean = regexp.MustCompile(`\(.*?\)`).ReplaceAllString(clean, "")
	clean = regexp.MustCompile(`（.*?）`).ReplaceAllString(clean, "")
	parts := strings.Split(clean, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	name = safeStr(o.primaryUserName())
	phone = safeStr(o.primaryUserPhone())
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
	if legacyName := safeStr(o.LegacyName); len(legacyName) > len(name) || name == "111" {
		name = legacyName
	}
	return name, phone, addr
}

func (o *Order) primaryTableInfo() string {
	if table := safeStr(o.TableInfo); table != "" {
		return table
	}
	return safeStr(o.LegacyTable)
}

func (o *Order) primaryOrderType() string {
	if orderType := safeStr(o.OrderType); orderType != "" {
		return orderType
	}
	return safeStr(o.LegacyType)
}

func (o *Order) primaryUserPhone() string {
	if phone := safeStr(o.UserPhone); phone != "" {
		return phone
	}
	return safeStr(o.LegacyPhone)
}

func (o *Order) primaryUserName() string {
	if name := safeStr(o.UserName); name != "" {
		return name
	}
	return safeStr(o.LegacyName)
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
