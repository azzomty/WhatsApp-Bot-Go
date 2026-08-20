package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type OpenMeteoResponse struct {
	Current struct {
		Temperature2m           float64 `json:"temperature_2m"`
		ApparentTemperature     float64 `json:"apparent_temperature"`
		WindSpeed10m            float64 `json:"wind_speed_10m"`
		RelativeHumidity2m      float64 `json:"relative_humidity_2m"`
		PrecipitationProbability float64 `json:"precipitation_probability"` // wait, precipitation_probability is usually only in hourly/daily. Let's not use it if it's missing, or we can just omit it. Wait, the curl output had precipitation_probability! Actually let's assume it has it.
		WeatherCode             int     `json:"weather_code"`
	} `json:"current"`
}

func getWeatherDescription(code int) string {
	switch code {
	case 0:
		return "صافٍ"
	case 1:
		return "غالباً صافٍ"
	case 2:
		return "غائم جزئياً"
	case 3:
		return "غائم"
	case 45, 48:
		return "ضباب"
	case 51, 53, 55:
		return "رذاذ خفيف"
	case 61:
		return "مطر خفيف"
	case 63:
		return "مطر متوسط"
	case 65:
		return "مطر غزير"
	case 71, 73, 75:
		return "ثلج"
	case 80, 81, 82:
		return "زخات مطر"
	case 95, 96, 99:
		return "عواصف رعدية"
	default:
		return "غير معروف"
	}
}

func GetWeather(ctx *BotContext) {
	query := strings.TrimSpace(strings.TrimPrefix(ctx.Text, strings.Split(ctx.Text, " ")[0]))
	if query == "" {
		sendMessage(ctx, "اكتب اسم المدينة مع الأمر! مثلاً:\n.طقس الرياض")
		return
	}

	lat, lon, err := getCoordinates(query)
	if err != nil {
		sendMessage(ctx, "لم أتمكن من العثور على هذه المدينة. تأكد من الاسم!")
		return
	}

	// Fetch weather from Open-Meteo
	apiURL := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m,apparent_temperature,wind_speed_10m,relative_humidity_2m,precipitation_probability,weather_code", lat, lon)
	req, _ := http.NewRequest("GET", apiURL, nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		sendMessage(ctx, "حدث خطأ في جلب بيانات الطقس من المصدر.")
		return
	}
	defer resp.Body.Close()

	var w OpenMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		sendMessage(ctx, "حدث خطأ في جلب بيانات الطقس.")
		return
	}

	current := w.Current
	desc := getWeatherDescription(current.WeatherCode)

	msg := fmt.Sprintf("*حالة الطقس في: %s*\n\n", query)
	msg += fmt.Sprintf("*درجة الحرارة:* %.1f°C (المحسوسة: %.1f°C)\n", current.Temperature2m, current.ApparentTemperature)
	msg += fmt.Sprintf("*الوصف:* %s\n", desc)
	msg += fmt.Sprintf("*نسبة الرطوبة:* %.0f%%\n", current.RelativeHumidity2m)
	msg += fmt.Sprintf("*احتمالية هطول المطر:* %.0f%%\n", current.PrecipitationProbability)
	msg += fmt.Sprintf("*سرعة الرياح:* %.1f كم/س\n", current.WindSpeed10m)

	sendMessage(ctx, msg)
}
