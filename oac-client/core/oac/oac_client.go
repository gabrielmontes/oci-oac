package oac

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type OacClient struct {
	AccessToken string
	TokenExpiry time.Time
}

var cacheDir = filepath.Join(os.Getenv("HOME"), ".cache", "oac-client")
var tokenFile = filepath.Join(cacheDir, "oac_token.json")

// NewOacClient loads config from dotenv
func NewOacClient() (*OacClient, error) {
	client := &OacClient{}
	client.loadTokenFromFile()
	return client, nil
}

// GetToken returns a valid access token, obtaining a new one if expired
func (oacClient *OacClient) GetToken() (string, error) {
	if oacClient.AccessToken != "" && time.Now().Before(oacClient.TokenExpiry) {
		return oacClient.AccessToken, nil
	}

	if err := oacClient.obtainToken(); err != nil {
		return "", err
	}

	return oacClient.AccessToken, nil
}

// obtainToken performs Resource Owner Password flow to get a new token
func (oacClient *OacClient) obtainToken() error {
	idcsURL := strings.TrimRight(os.Getenv("IDCS_TOKEN_URL"), "/")
	clientID := os.Getenv("IDCS_OAC_CLIENT_ID")
	clientSecret := os.Getenv("IDCS_OAC_CLIENT_SECRET")
	scope := os.Getenv("IDCS_OAC_SCOPE")
	username := os.Getenv("OAC_USERNAME")
	password := os.Getenv("OAC_PASSWORD")
	grantType := os.Getenv("IDCS_GRANT_TYPE")

	if clientID == "" || clientSecret == "" || scope == "" || grantType == "" {
		return fmt.Errorf("missing required environment variables")
	}

	ctx := context.Background()
	var token *oauth2.Token
	var err error

	switch grantType {
	case "client_credentials":
		cfg := clientcredentials.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			TokenURL:     idcsURL,
			Scopes:       []string{scope},
		}
		token, err = cfg.Token(ctx)

	case "resource_owner":
		if username == "" || password == "" {
			return fmt.Errorf("username/password must be set for password grant")
		}
		cfg := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{scope},
			Endpoint: oauth2.Endpoint{
				TokenURL: idcsURL,
			},
		}
		token, err = cfg.PasswordCredentialsToken(ctx, username, password)

	default:
		return fmt.Errorf("unsupported grant type: %s", grantType)
	}

	if err != nil {
		return fmt.Errorf("failed to obtain token: %w", err)
	}

	oacClient.AccessToken = token.AccessToken
	// fallback if expiry is not set
	if token.Expiry.IsZero() {
		oacClient.TokenExpiry = time.Now().Add(time.Hour - time.Minute)
	} else {
		oacClient.TokenExpiry = token.Expiry.Add(-time.Minute)
	}
	oacClient.saveTokenToFile()

	return nil
}

// RestCall executes a REST API call against the OAC instance
func (c *OacClient) RestCall(method, path, bodyFile string) (string, error) {
	token, err := c.GetToken()
	if err != nil {
		return "", err
	}

	var bodyBytes []byte
	if bodyFile != "" {
		if _, err := os.Stat(bodyFile); err == nil {
			bodyBytes, err = os.ReadFile(bodyFile)
			if err != nil {
				return "", err
			}
		} else {
			bodyBytes = []byte(bodyFile)
		}
	}


	instanceUrl := os.Getenv("OAC_INSTANCE")
	url := strings.TrimRight(instanceUrl, "/") + "/" + strings.TrimLeft(path, "/")
	req, err := http.NewRequest(strings.ToUpper(method), url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		// retry once with fresh token
		c.AccessToken = ""
		token, err = c.GetToken()
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("request failed: %d %s", resp.StatusCode, body)
	}

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return prettyPrintJSON(resBody)
}

// saveTokenToFile caches token on disk
func (oacClient *OacClient) saveTokenToFile() {
	os.MkdirAll(cacheDir, os.ModePerm)
	data := map[string]any{
		"access_token": oacClient.AccessToken,
		"expires_at":   oacClient.TokenExpiry.Unix(),
	}
	b, _ := json.Marshal(data)
	_ = os.WriteFile(tokenFile, b, 0600)
}

// loadTokenFromFile loads token cache if present
func (oacClient *OacClient) loadTokenFromFile() {
	file, err := os.ReadFile(tokenFile)
	if err != nil {
		return
	}

	var data map[string]any
	if err := json.Unmarshal(file, &data); err != nil {
		return
	}

	token, tokenError := data["access_token"].(string)
	exp, expError := data["expires_at"].(float64)
	if !tokenError || !expError {
		return
	}

	oacClient.AccessToken = token
	oacClient.TokenExpiry = time.Unix(int64(exp), 0)
	if time.Now().After(oacClient.TokenExpiry) {
		oacClient.AccessToken = ""
	}
}

// prettyPrintJSON formats JSON response for readability
func prettyPrintJSON(data []byte) (string, error) {
	dataStr := strings.TrimSpace(string(data))
	if len(dataStr) == 0 {
		return "Request succeeded (no content).", nil
	}

	if strings.HasPrefix(dataStr, "{") {
		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(obj, "", "  ")
		return string(b), nil
	} else if strings.HasPrefix(dataStr, "[") {
		var arr []interface{}
		if err := json.Unmarshal(data, &arr); err != nil {
			return "", err
		}
		b, _ := json.MarshalIndent(arr, "", "  ")
		return string(b), nil
	} else {
		return dataStr, nil
	}
}
