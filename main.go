package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// 데이터 구조들 (그대로 유지)
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

// SVG 템플릿
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

var colorRegex = regexp.MustCompile(`--([a-z-]+)-color:\s*(#[0-9a-fA-F]{3,8})`)
var defaultTheme = Theme{"dark", "#2c2e31", "#e2b714", "#646669", "#d1d0c5"}

// CSS 파싱 로직 (그대로 유지)
func getTheme(themeName string) Theme {
	themeName = strings.ReplaceAll(themeName, " ", "_")
	url := fmt.Sprintf("https://raw.githubusercontent.com/monkeytypegame/monkeytype/master/frontend/static/themes/%s.css", themeName)

	client := http.Client{Timeout: 5 * time.Second}
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
	return t
}

func fetchStats(username, mode, length string) (float64, float64, error) {
	url := fmt.Sprintf("https://api.monkeytype.com/users/%s/profile", username)
	client := http.Client{Timeout: 5 * time.Second}
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

func main() {
	// 1. 커맨드라인 인자(Flag) 파싱
	username := flag.String("username", "Guest", "Monkeytype Username")
	themeName := flag.String("theme", "dark", "Theme Name")
	mode := flag.String("mode", "time", "Mode (time/words)")
	length := flag.String("length", "60", "Length (15/60/10...)")
	flag.Parse()

	// 2. 데이터 수집
	fmt.Printf("Generating stats for %s (Theme: %s)...\n", *username, *themeName)
	t := getTheme(*themeName)
	wpm, acc, err := fetchStats(*username, *mode, *length)
	if err != nil {
		log.Printf("⚠️ Error fetching stats: %v. Setting to 0.", err)
		wpm, acc = 0, 0
	}

	modeDisplay := fmt.Sprintf("%s %s", *mode, *length)
	svgContent := fmt.Sprintf(svgTemplate,
		t.MainColor, t.SubColor, t.MainColor, t.BgColor, t.TextColor,
		*username, modeDisplay, wpm, acc,
	)

	// 3. 파일로 저장 (stats.svg)
	file, err := os.Create("stats.svg")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	file.WriteString(svgContent)

	fmt.Println("✅ stats.svg generated successfully!")
}
