package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	loginURL     = "https://connect.tarc.edu.my/login"
	tarumtWiFiIP = "2.2.2.2"
	accountFile  = "./config.json"
	timeout      = 10 * time.Second
	maxRetries   = 3
)

type Account struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	fmt.Println("TAR UMT WiFi Auto Connect Program")

	account, err := getAccountData()
	if err != nil {
		fmt.Println("Error reading account data:", err)
		return
	}

	if !isConnectedToTARUMTWiFi() {
		fmt.Println("❌ You are not connected to TAR UMT WiFi")
		return
	}

	if err := login(account, 0); err != nil {
		fmt.Println("❌ Login failed:", err)
	} else {
		fmt.Println("✅ You are now connected to TAR UMT WiFi")
	}
}

func getAccountData() (*Account, error) {
	file, err := os.Open(accountFile)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
		}
	}(file)

	var account Account
	if err := json.NewDecoder(file).Decode(&account); err != nil {
		return nil, err
	}
	return &account, nil
}

func isConnectedToTARUMTWiFi() bool {
	addrs, err := net.LookupIP("connect.tarc.edu.my")
	if err != nil {
		return false
	}
	for _, addr := range addrs {
		if addr.String() == tarumtWiFiIP {
			return true
		}
	}
	return false
}

func login(account *Account, retries int) error {
	data := url.Values{
		"username": {account.Username},
		"password": {account.Password},
		"dst":      {"https://google.com"},
	}

	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest("POST", loginURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://wifi2.tarc.edu.my/")
	req.Header.Set("Origin", "https://wifi2.tarc.edu.my")

	resp, err := client.Do(req)
	if err != nil {
		if retries < maxRetries {
			fmt.Printf("⚠ Request timed out. Retrying %d/%d...\n", retries+1, maxRetries)
			return login(account, retries+1)
		}
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if bytes.Contains(body, []byte("<h1>You are logged in</h1>")) {
		fmt.Println("✅ Login successful!")
		return nil
	}
	return fmt.Errorf("login failed, please check your username and password")
}
