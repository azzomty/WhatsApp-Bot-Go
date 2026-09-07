import re

with open("internal/commands/api.go", "r") as f:
    c = f.read()

target = """type NominatimResponse struct {
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
}"""

new_target = """type OpenMeteoGeocodingResponse struct {
	Results []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

func getCoordinates(city string) (float64, float64, error) {
	apiURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=ar&format=json", url.QueryEscape(city))
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return 0, 0, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, 0, fmt.Errorf("Geocoding API error: %d", resp.StatusCode)
	}

	var res OpenMeteoGeocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, 0, err
	}

	if len(res.Results) == 0 {
		return 0, 0, fmt.Errorf("city not found")
	}

	return res.Results[0].Latitude, res.Results[0].Longitude, nil
}"""

c = c.replace(target, new_target)
with open("internal/commands/api.go", "w") as f:
    f.write(c)

