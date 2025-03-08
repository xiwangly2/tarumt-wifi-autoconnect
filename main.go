package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
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

type InterfaceState struct {
	Name string
	Up   bool
}

func main() {
	configFile := flag.String("config", "./config.json", "Path to the configuration file")
	flag.Parse()

	config, err := loadConfig(*configFile)
	if err != nil {
		logMessage(fmt.Sprintf("Failed to read configuration file: %v", err))
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	var lastStates []InterfaceState

	for {
		currentStates := getInterfaceStates()
		if hasStateChanged(lastStates, currentStates) {
			logMessage("Network status changed")
			loginIfNeeded(config, client)
			if config.OnlyOnce {
				break
			}
		}
		lastStates = currentStates
		time.Sleep(time.Duration(config.AttemptDelay) * time.Second)
	}
}

func getInterfaceStates() []InterfaceState {
	var states []InterfaceState
	ifaces, err := net.Interfaces()
	if err != nil {
		logMessage(fmt.Sprintf("Error getting network interfaces: %v", err))
		return states
	}

	for _, iface := range ifaces {
		state := InterfaceState{
			Name: iface.Name,
			Up:   iface.Flags&net.FlagUp != 0,
		}
		states = append(states, state)
	}
	return states
}

func hasStateChanged(lastStates, currentStates []InterfaceState) bool {
	if len(lastStates) != len(currentStates) {
		return true
	}

	for i, state := range lastStates {
		if state.Name != currentStates[i].Name || state.Up != currentStates[i].Up {
			return true
		}
	}
	return false
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
