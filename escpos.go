package main

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	ESC          = 0x1B
	GS           = 0x1D
	ALIGN_LEFT   = 0x00
	ALIGN_CENTER = 0x01
	ALIGN_RIGHT  = 0x02
	SIZE_NORMAL  = 0x00
	SIZE_DOUBLE  = 0x11
)

type PrintBuffer struct {
	buf bytes.Buffer
}

func NewPrintBuffer() *PrintBuffer {
	pb := &PrintBuffer{}
	pb.Init()
	return pb
}

func (pb *PrintBuffer) Init() {
	pb.buf.WriteByte(ESC)
	pb.buf.WriteByte('@')
}

func (pb *PrintBuffer) Align(align byte) {
	pb.buf.WriteByte(ESC)
	pb.buf.WriteByte('a')
	pb.buf.WriteByte(align)
}

func (pb *PrintBuffer) SetSize(size byte) {
	pb.buf.WriteByte(GS)
	pb.buf.WriteByte('!')
	pb.buf.WriteByte(size)
}

func (pb *PrintBuffer) Bold(on bool) {
	pb.buf.WriteByte(ESC)
	pb.buf.WriteByte('E')
	if on {
		pb.buf.WriteByte(1)
	} else {
		pb.buf.WriteByte(0)
	}
}

func (pb *PrintBuffer) Write(text string) {
	// 塞尔维亚字符转换
	text = transliterate(text)
	// 只对非ASCII字符编码为GB18030
	pb.buf.Write(toGB18030(text))
}

func (pb *PrintBuffer) WriteLine(text string) {
	pb.Write(text)
	pb.buf.WriteByte('\n')
}

func (pb *PrintBuffer) LineFeed(count int) {
	for i := 0; i < count; i++ {
		pb.buf.WriteByte('\n')
	}
}

func (pb *PrintBuffer) Separator(char string, count int) {
	for i := 0; i < count; i++ {
		pb.buf.WriteString(char)
	}
	pb.buf.WriteByte('\n')
}

func (pb *PrintBuffer) Bytes() []byte {
	return pb.buf.Bytes()
}

var SerbianMap = map[rune]string{
	'č': "c", 'ć': "c", 'ž': "z", 'š': "s", 'đ': "dj",
	'Č': "C", 'Ć': "C", 'Ž': "Z", 'Š': "S", 'Đ': "Dj",
}

func transliterate(text string) string {
	if text == "" {
		return ""
	}
	var result strings.Builder
	for _, c := range text {
		if replacement, ok := SerbianMap[c]; ok {
			result.WriteString(replacement)
		} else {
			result.WriteRune(c)
		}
	}
	return result.String()
}

func toGB18030(text string) []byte {
	result := bytes.Buffer{}
	encoder := simplifiedchinese.GB18030.NewEncoder()
	writer := transform.NewWriter(&result, encoder)

	for _, r := range text {
		if r < 128 {
			result.WriteByte(byte(r))
		} else {
			writer.Write([]byte(string(r)))
		}
	}
	writer.Close()
	return result.Bytes()
}

func FormatOrder(order Order) []byte {
	pb := NewPrintBuffer()

	pb.Align(ALIGN_CENTER)
	pb.SetSize(SIZE_DOUBLE)
	pb.Bold(true)

	if order.IsDelivery() {
		pb.WriteLine("外卖订单")
	} else {
		pb.WriteLine("堂食订单")
	}

	pb.Bold(false)
	pb.SetSize(SIZE_NORMAL)
	pb.Separator("=", 32)
	pb.Align(ALIGN_LEFT)

	pickupNo := order.PickupNo
	if pickupNo == "" {
		pickupNo = "-"
	}
	pb.WriteLine(fmt.Sprintf("订单: %s (取餐号: %s)", safeStr(order.ID), pickupNo))

	timeStr := order.GetTimeStr()
	if !order.IsDelivery() {
		pb.WriteLine(fmt.Sprintf("桌号: %s      时间: %s", safeStr(order.TableInfo), timeStr))
	} else {
		pb.WriteLine(fmt.Sprintf("时间: %s", timeStr))
	}

	pb.Separator("-", 32)

	if order.IsDelivery() {
		name, phone, addr := order.ParseShippingInfo()
		if name != "" || phone != "" {
			pb.WriteLine(fmt.Sprintf("%s  %s", name, phone))
		}
		if addr != "" {
			pb.WriteLine(fmt.Sprintf("地址: %s", addr))
		}
	}

	cNote := order.ExtractNote()
	if cNote != "" {
		pb.WriteLine(fmt.Sprintf("备注: %s", cNote))
	}

	if order.IsDelivery() || cNote != "" {
		pb.Separator("-", 32)
	}

	for _, item := range order.Items {
		name := safeStr(item.Name)
		if name == "" {
			name = fmt.Sprintf("菜品 #%d", item.ProductID)
		}
		qty := item.Qty
		if qty == 0 {
			qty = 1
		}
		uPrice := item.Price
		if uPrice == 0 {
			uPrice = safeParseMoney(item.Total) / int64(qty)
		}

		l1 := name
		l2 := ""
		if idx := strings.Index(name, "|"); idx != -1 {
			parts := strings.SplitN(name, "|", 2)
			l1 = strings.TrimSpace(parts[0])
			l2 = strings.TrimSpace(parts[1])
		}

		pb.Bold(true)
		pb.WriteLine(l1)
		pb.Bold(false)

		info := fmt.Sprintf("%d x %d", uPrice, qty)
		if l2 != "" {
			pb.WriteLine(fmt.Sprintf("%s   %s", l2, info))
		} else {
			pb.WriteLine(info)
		}

		if item.Meta != "" {
			pb.WriteLine(fmt.Sprintf("  [%s]", item.Meta))
		}
		pb.LineFeed(1)
	}

	pb.Separator("-", 32)

	pb.Align(ALIGN_CENTER)
	pb.SetSize(SIZE_DOUBLE)
	pb.Bold(true)
	pb.WriteLine(fmt.Sprintf("合计: RSD %d", order.TotalAmount))
	pb.Bold(false)
	pb.SetSize(SIZE_NORMAL)
	pb.Separator("=", 32)

	pb.LineFeed(3)
	pb.buf.WriteByte(ESC)
	pb.buf.WriteByte('V')
	pb.buf.WriteByte(0x00)

	return pb.Bytes()
}
