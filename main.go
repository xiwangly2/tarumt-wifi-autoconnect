package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	LoginURL     string `json:"loginURL"`
	StatusURL    string `json:"statusURL"`
	RefererURL   string `json:"refererURL"`
	OriginURL    string `json:"originURL"`
	WiFiIP       string `json:"WiFiIP"`
	UserAgent    string `json:"userAgent"`
	AttemptDelay int    `json:"attemptDelay"`
	OnlyOnce     bool   `json:"onlyOnce"`
}

func main() {
	configFile := flag.String("config", "./config.json", "Path to the configuration file")
	flag.Parse()

	config, err := loadConfig(*configFile)
	if err != nil {
		fmt.Printf("[%s] Failed to read configuration file: %v\n", currentTime(), err)
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for {
		loginIfNeeded(config, client)
		if config.OnlyOnce {
			break
		}
		time.Sleep(time.Duration(config.AttemptDelay) * time.Second)
	}
}

func loginIfNeeded(config *Config, client *http.Client) {
	resp, err := client.Get(config.StatusURL)
	if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if err == nil {
			responseBody := string(body)

			if strings.Contains(responseBody, "If you are not redirected in a few seconds") {
				data := url.Values{
					"username": {config.Username},
					"password": {config.Password},
					"dst":      {"https://google.com"},
				}

				req, err := http.NewRequest("POST", config.LoginURL, bytes.NewBufferString(data.Encode()))
				if err != nil {
					logMessage(fmt.Sprintf("Error creating request: %v", err))
					return
				}

				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set("User-Agent", config.UserAgent)
				req.Header.Set("Referer", config.RefererURL)
				req.Header.Set("Origin", config.OriginURL)

				resp, err := client.Do(req)
				if err != nil {
					logMessage(fmt.Sprintf("Error during login: %v", err))
				} else {
					body, err := io.ReadAll(resp.Body)
					_ = resp.Body.Close()
					if err == nil && bytes.Contains(body, []byte("You are logged in")) {
						logMessage("Login successful")
					} else {
						logMessage("Login failed, please check your username and password")
					}
				}
			}
		}
	}
}

func currentTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func logMessage(msg string) {
	fmt.Printf("[%s] %s\n", currentTime(), msg)
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
