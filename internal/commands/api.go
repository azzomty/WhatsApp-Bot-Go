package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type NominatimResponse struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

func getCoordinates(city string) (float64, float64, error) {
	apiURL := fmt.Sprintf("https://nominatim.openstreetmap.org/search?q=%s&format=json&limit=1", url.QueryEscape(city))
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("User-Agent", "WhatsAppBot/1.0 (https://github.com/azzomty/WhatsApp-Bot-Go)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, 0, fmt.Errorf("Nominatim error: %d", resp.StatusCode)
	}

	var res []NominatimResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, 0, err
	}

	if len(res) == 0 {
		return 0, 0, fmt.Errorf("city not found")
	}

	lat, _ := strconv.ParseFloat(res[0].Lat, 64)
	lon, _ := strconv.ParseFloat(res[0].Lon, 64)
	return lat, lon, nil
}

type AladhanResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Data   struct {
		Timings struct {
			Fajr    string `json:"Fajr"`
			Sunrise string `json:"Sunrise"`
			Dhuhr   string `json:"Dhuhr"`
			Asr     string `json:"Asr"`
			Maghrib string `json:"Maghrib"`
			Isha    string `json:"Isha"`
		} `json:"timings"`
		Meta struct {
			Timezone string `json:"timezone"`
		} `json:"meta"`
	} `json:"data"`
}

func fetchAladhanAPI(lat float64, lon float64) (*AladhanResponse, error) {
	// method=4 is Umm Al-Qura University, Makkah, highly accurate for Saudi/Gulf
	apiURL := fmt.Sprintf("https://api.aladhan.com/v1/timings?latitude=%f&longitude=%f&method=4", lat, lon)
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result AladhanResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API Error: %s", result.Status)
	}

	return &result, nil
}

func HandlePrayerTimes(ctx *BotContext, address string) {
	if address == "" {
		sendMessage(ctx, "يرجى كتابة اسم الدولة أو المدينة، مثال: .مواعيد صلاة السعودية")
		return
	}

	sendMessage(ctx, "جاري جلب المواقيت")

	lat, lon, err := getCoordinates(address)
	if err != nil {
		sendMessage(ctx, "لم أتمكن من العثور على هذه المدينة. تأكد من الاسم!")
		return
	}

	res, err := fetchAladhanAPI(lat, lon)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء جلب المواقيت")
		return
	}

	t := res.Data.Timings
	msg := fmt.Sprintf("*مواقيت الصلاة في (%s):*\n\n", address)
	msg += fmt.Sprintf("الفجر: %s\n", formatPrayer(t.Fajr, address))
	msg += fmt.Sprintf("الشروق: %s\n", formatPrayer(t.Sunrise, address))
	msg += fmt.Sprintf("الظهر: %s\n", formatPrayer(t.Dhuhr, address))
	msg += fmt.Sprintf("العصر: %s\n", formatPrayer(t.Asr, address))
	msg += fmt.Sprintf("المغرب: %s\n", formatPrayer(t.Maghrib, address))
	msg += fmt.Sprintf("العشاء: %s\n", formatPrayer(t.Isha, address))

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(msg)},
	})
}

func HandleCurrentTime(ctx *BotContext, address string) {
	if address == "" {
		sendMessage(ctx, "يرجى كتابة اسم الدولة أو المدينة، مثال: .توقيت السعودية")
		return
	}

	lat, lon, err := getCoordinates(address)
	if err != nil {
		sendMessage(ctx, "لم أتمكن من العثور على هذه المدينة. تأكد من الاسم!")
		return
	}

	res, err := fetchAladhanAPI(lat, lon)
	if err != nil {
		sendMessage(ctx, "حدث خطأ أثناء جلب التوقيت")
		return
	}

	loc, err := time.LoadLocation(res.Data.Meta.Timezone)
	if err != nil {
		sendMessage(ctx, "لم أتمكن من تحديد المنطقة الزمنية")
		return
	}

	currentTime := time.Now().In(loc)

	// User requested 12-hour for Saudi, 24-hour for Morocco (and others by default)
	timeFormat := "15:04" // 24-hour
	if strings.Contains(address, "سعودي") || strings.Contains(address, "السعودي") {
		timeFormat = "03:04 PM" // 12-hour
	}

	formattedTime := currentTime.Format(timeFormat)

	// Translate AM/PM to Arabic if using 12-hour format
	formattedTime = strings.ReplaceAll(formattedTime, "AM", "ص")
	formattedTime = strings.ReplaceAll(formattedTime, "PM", "م")

	msg := fmt.Sprintf("التوقيت الحالي في (%s): *%s*", address, formattedTime)

	ctx.Client.SendMessage(context.Background(), ctx.ChatID, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{Text: proto.String(msg)},
	})
}

func formatPrayer(t24 string, address string) string {
	if !strings.Contains(address, "سعودي") && !strings.Contains(address, "السعودي") {
		return t24 // Return 24-hour format directly for non-Saudi (like Morocco)
	}
	t, err := time.Parse("15:04", t24)
	if err != nil {
		return t24
	}
	formatted := t.Format("03:04 PM")
	formatted = strings.ReplaceAll(formatted, "AM", "ص")
	formatted = strings.ReplaceAll(formatted, "PM", "م")
	return formatted
}
