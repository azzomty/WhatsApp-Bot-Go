package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type WeatherResponse struct {
	CurrentCondition []struct {
		FeelsLikeC string `json:"FeelsLikeC"`
		Humidity   string `json:"humidity"`
		TempC      string `json:"temp_C"`
		WeatherDesc []struct {
			Value string `json:"value"`
		} `json:"weatherDesc"`
		WindspeedKmph string `json:"windspeedKmph"`
		LangAr []struct {
			Value string `json:"value"`
		} `json:"lang_ar"`
	} `json:"current_condition"`
	Weather []struct {
		Hourly []struct {
			Chanceofrain string `json:"chanceofrain"`
		} `json:"hourly"`
	} `json:"weather"`
}

func GetWeather(ctx *BotContext) {
	query := strings.TrimSpace(strings.TrimPrefix(ctx.Text, strings.Split(ctx.Text, " ")[0]))
	if query == "" {
		sendMessage(ctx, "اكتب اسم المدينة مع الأمر! مثلاً:\n.طقس الرياض")
		return
	}

	encodedQuery := url.QueryEscape(query)
	req, _ := http.NewRequest("GET", "https://wttr.in/"+encodedQuery+"?format=j1", nil)
	req.Header.Set("Accept-Language", "ar")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		sendMessage(ctx, "ما قدرت ألقى طقس هذي المدينة، تأكد من الاسم!")
		return
	}
	defer resp.Body.Close()

	var w WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil || len(w.CurrentCondition) == 0 {
		sendMessage(ctx, "حدث خطأ في جلب بيانات الطقس.")
		return
	}

	current := w.CurrentCondition[0]
	desc := current.WeatherDesc[0].Value
	if len(current.LangAr) > 0 {
		desc = current.LangAr[0].Value
	}
	
	rainChance := "0"
	if len(w.Weather) > 0 && len(w.Weather[0].Hourly) > 0 {
		rainChance = w.Weather[0].Hourly[0].Chanceofrain
	}

	msg := fmt.Sprintf("🌤️ *حالة الطقس في: %s*\n\n", query)
	msg += fmt.Sprintf("🌡️ *درجة الحرارة:* %s°C (المحسوسة: %s°C)\n", current.TempC, current.FeelsLikeC)
	msg += fmt.Sprintf("☁️ *الوصف:* %s\n", desc)
	msg += fmt.Sprintf("💧 *نسبة الرطوبة:* %s%%\n", current.Humidity)
	msg += fmt.Sprintf("🌧️ *احتمالية هطول المطر:* %s%%\n", rainChance)
	msg += fmt.Sprintf("💨 *سرعة الرياح:* %s كم/س\n", current.WindspeedKmph)

	sendMessage(ctx, msg)
}
