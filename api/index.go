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

// ---------------------------
// 1. 데이터 구조 및 설정
// ---------------------------
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

// 전역 변수 (Vercel 인스턴스가 살아있는 동안 캐시 유지)
var (
	colorRegex   = regexp.MustCompile(`--([a-z-]+)-color:\s*(#[0-9a-fA-F]{3,8})`)
	defaultTheme = Theme{"dark", "#2c2e31", "#e2b714", "#646669", "#d1d0c5"}
	themeCache   = make(map[string]Theme)
	cacheMutex   sync.RWMutex
)

// ---------------------------
// 2. 로직 (테마 파싱 & 스탯 가져오기)
// ---------------------------
func getTheme(themeName string) Theme {
	themeName = strings.ReplaceAll(themeName, " ", "_")

	// 캐시 확인
	cacheMutex.RLock()
	if t, ok := themeCache[themeName]; ok {
		cacheMutex.RUnlock()
		return t
	}
	cacheMutex.RUnlock()

	// 깃허브에서 CSS 다운로드 (최신 테마 지원)
	url := fmt.Sprintf("https://raw.githubusercontent.com/monkeytypegame/monkeytype/master/frontend/static/themes/%s.css", themeName)
	client := http.Client{Timeout: 2 * time.Second} // Vercel 타임아웃 고려하여 짧게 설정
	resp, err := client.Get(url)

	if err != nil || resp.StatusCode != 200 {
		return defaultTheme
	}
	defer resp.Body.Close()
	cssBytes, _ := io.ReadAll(resp.Body)
	cssContent := string(cssBytes)

	// 정규식으로 색상 추출
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

// ---------------------------
// 3. Vercel 핸들러 (메인 함수)
// ---------------------------
func Handler(w http.ResponseWriter, r *http.Request) {
	// 파라미터 받기 (user 또는 username 둘 다 허용)
	username := r.URL.Query().Get("username")
	if username == "" {
		username = r.URL.Query().Get("user")
	}
	themeName := r.URL.Query().Get("theme")
	mode := r.URL.Query().Get("mode")
	length := r.URL.Query().Get("length")

	// 기본값 설정
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

	// 데이터 가져오기
	t := getTheme(themeName)
	wpm, acc, err := fetchStats(username, mode, length)

	// 에러나면 0으로 표시 (이미지는 깨지면 안 되니까)
	if err != nil {
		wpm, acc = 0, 0
	}

	modeDisplay := fmt.Sprintf("%s %s", mode, length)

	// SVG 생성
	svgContent := fmt.Sprintf(svgTemplate,
		t.MainColor, t.SubColor, t.MainColor, t.BgColor, t.TextColor,
		username, modeDisplay, wpm, acc,
	)

	// 헤더 설정 (SVG 이미지로 인식하게 함)
	w.Header().Set("Content-Type", "image/svg+xml")
	// GitHub의 이미지 캐싱(camo)을 뚫고 갱신되게 하기 위해 Cache-Control 설정
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

	w.Write([]byte(svgContent))
}
