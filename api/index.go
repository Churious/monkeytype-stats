package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Theme struct {
	Name, BgColor, MainColor, SubColor, TextColor string
}

type Record struct {
	Wpm, Acc float64
}

type MonkeyProfile struct {
	Data struct {
		Name          string `json:"name"`
		PersonalBests struct {
			Time  map[string][]Record `json:"time"`
			Words map[string][]Record `json:"words"`
		} `json:"personalBests"`
	} `json:"data"`
}

const svgTemplate = `<svg width="400" height="150" viewBox="0 0 400 150" fill="none" xmlns="http://www.w3.org/2000/svg">
    <style>
        .header { font: 800 20px 'Segoe UI', Ubuntu, Sans-Serif; fill: %s; }
        .stat-label { font: 600 14px 'Segoe UI', Ubuntu, Sans-Serif; fill: %s; }
        .stat-value { font: 700 28px 'Segoe UI', Ubuntu, Sans-Serif; fill: %s; }
        .bg { fill: %s; rx: 10px; }
        .sub-info { font: 600 12px 'Segoe UI', Ubuntu, Sans-Serif; fill: %s; opacity: 0.9; } 
    </style>
    <rect width="400" height="150" class="bg"/>
    <text x="25" y="35" class="header">Monkeytype Stats</text>
    <text x="375" y="35" text-anchor="end" class="sub-info">@%s (%s)</text>
    <text x="25" y="80" class="stat-label">Highest WPM</text>
    <text x="180" y="80" class="stat-value">%.0f</text>
    <text x="25" y="115" class="stat-label">Accuracy</text>
    <text x="180" y="115" class="stat-value">%.2f%%</text>
</svg>`

var (
	colorRegex   = regexp.MustCompile(`--([a-z-]+)-color:\s*(#[0-9a-fA-F]{3,8})`)
	defaultTheme = Theme{"dark", "#2c2e31", "#e2b714", "#646669", "#d1d0c5"}
	themeCache   = make(map[string]Theme)
	cacheMutex   sync.RWMutex
)

func getTheme(themeName string) Theme {
	themeName = strings.ReplaceAll(themeName, " ", "_")

	cacheMutex.RLock()
	if t, ok := themeCache[themeName]; ok {
		cacheMutex.RUnlock()
		return t
	}
	cacheMutex.RUnlock()

	url := fmt.Sprintf("https://raw.githubusercontent.com/monkeytypegame/monkeytype/master/frontend/static/themes/%s.css", themeName)
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)

	if err != nil || resp.StatusCode != 200 {
		return defaultTheme
	}
	defer resp.Body.Close()
	cssBytes, _ := io.ReadAll(resp.Body)
	cssContent := string(cssBytes)

	t := Theme{Name: themeName, BgColor: "#2c2e31", MainColor: "#e2b714", SubColor: "#646669", TextColor: "#d1d0c5"}
	matches := colorRegex.FindAllStringSubmatch(cssContent, -1)

	for _, match := range matches {
		if len(match) == 3 {
			switch match[1] {
			case "bg":
				t.BgColor = match[2]
			case "main":
				t.MainColor = match[2]
			case "sub":
				t.SubColor = match[2]
			case "text":
				t.TextColor = match[2]
			}
		}
	}

	cacheMutex.Lock()
	themeCache[themeName] = t
	cacheMutex.Unlock()
	return t
}

func fetchStats(username, mode, length string) (float64, float64, error) {
	url := fmt.Sprintf("https://api.monkeytype.com/users/%s/profile", username)
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var result MonkeyProfile
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, err
	}

	var records []Record
	if mode == "time" && result.Data.PersonalBests.Time != nil {
		records = result.Data.PersonalBests.Time[length]
	} else if mode == "words" && result.Data.PersonalBests.Words != nil {
		records = result.Data.PersonalBests.Words[length]
	}

	if len(records) == 0 {
		return 0, 0, fmt.Errorf("no data")
	}
	return records[0].Wpm, records[0].Acc, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		username = r.URL.Query().Get("user")
	}
	themeName := r.URL.Query().Get("theme")
	mode := r.URL.Query().Get("mode")
	length := r.URL.Query().Get("length")
	isTransparent := r.URL.Query().Get("transparent") == "true"

	if username == "" {
		username = "Guest"
	}
	if themeName == "" {
		themeName = "dark"
	}
	if mode == "" {
		mode = "time"
	}
	if length == "" {
		length = "60"
	}

	t := getTheme(themeName)
	wpm, acc, err := fetchStats(username, mode, length)

	if err != nil {
		wpm, acc = 0, 0
	}

	modeDisplay := fmt.Sprintf("%s %s", mode, length)

	svgBg := t.BgColor
	if isTransparent {
		svgBg = "none"
	}

	svgContent := fmt.Sprintf(svgTemplate,
		t.MainColor, t.SubColor, t.MainColor, svgBg, t.TextColor,
		username, modeDisplay, wpm, acc,
	)

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	w.Write([]byte(svgContent))
}
