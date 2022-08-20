package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	home, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatal(err)
	}

	envDir := ".config/.sso_unila.env"
	envPath := path.Join(home, envDir)

	username, password := loadEnv(envPath)

	if err := doRequest(username, password); err != nil {
		logrus.Fatal("login process failed: ", err)
	}
}

func loadEnv(envPath string) (string, string) {
	if err := godotenv.Load(envPath); err != nil {
		logrus.Fatal("failed to load environment. reason: ", err)
	}

	username := os.Getenv("SSO_USERNAME")
	password := os.Getenv("SSO_PASSWORD")

	if username == "" || password == "" {
		logrus.Fatal("username and password are not found in environment variable")
	}

	return username, password
}

func doRequest(username, password string) error {
	host := "https://sso.unila.ac.id/login"

	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{Transport: tr}

	r, err := http.NewRequest(http.MethodPost, host, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	bodyByte, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	body := string(bodyByte)

	failedLoginKeyWord := "RADIUS server"
	alreadyLoginKeWord := "MAC Address"
	SuccessLoginKeyWord := "Anda sudah berhasil login"

	if strings.Contains(body, failedLoginKeyWord) {
		return fmt.Errorf("SSO server return an error. Probably caused by credentials mismatch")
	}

	if strings.Contains(body, alreadyLoginKeWord) {
		logrus.Info("You are already logged in")
		return nil
	}

	if strings.Contains(body, SuccessLoginKeyWord) {
		logrus.Info("Login success")
		return nil
	}

	return errors.New("unknown response from SSO server")
}
