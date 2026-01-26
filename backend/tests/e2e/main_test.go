package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ApiUrl    = "http://localhost:8080/api/v1"
	TestEmail = "test_e2e_fix@example.com"
	TestPass  = "password123456"
)

func TestEndToEndFlow(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()
	fmt.Printf("[TEST] Mock server started at: %s\n", mockServer.URL)

	client := &http.Client{Timeout: 10 * time.Second}

	fmt.Println("[TEST] 1. Registering user...")
	regPayload := map[string]string{"email": TestEmail, "password": TestPass}
	resp, err := postJSON(client, ApiUrl+"/users", regPayload)
	require.NoError(t, err)

	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("[DEBUG] Registration Response (%d): %s\n", resp.StatusCode, string(body))

	fmt.Println("[TEST] 2. Logging in...")
	loginPayload := map[string]string{"email": TestEmail, "password": TestPass}
	resp, err = postJSON(client, ApiUrl+"/login", loginPayload)
	require.NoError(t, err)

	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("[DEBUG] Login Response (%d): %s\n", resp.StatusCode, string(body))

	require.Equal(t, 200, resp.StatusCode)

	var loginResp struct {
		Token string `json:"access_token"`
	}
	json.Unmarshal(body, &loginResp)
	token := loginResp.Token
	require.NotEmpty(t, token, "JWT Token is empty! Check JSON mapping.")

	fmt.Println("[TEST] 3. Creating service...")
	svcPayload := map[string]interface{}{
		"name":                fmt.Sprintf("Svc_%d", time.Now().Unix()),
		"url":                 mockServer.URL,
		"port":                80,
		"healthCheckInterval": 1,
		"alertWindow":         5,
		"allowedResponseTime": 5,
		"firstOncallerEmail":  TestEmail,
	}
	resp, err = postAuthenticatedJSON(client, "POST", ApiUrl+"/services/", svcPayload, token)
	require.NoError(t, err)

	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Printf("[DEBUG] Create Service Response (%d): %s\n", resp.StatusCode, string(body))

	require.Equal(t, 201, resp.StatusCode)

	var svcMap map[string]interface{}
	json.Unmarshal(body, &svcMap)

	var serviceID uint
	if id, ok := svcMap["serviceID"]; ok {
		serviceID = uint(id.(float64))
	} else if id, ok := svcMap["id"]; ok {
		serviceID = uint(id.(float64))
	} else if id, ok := svcMap["ID"]; ok {
		serviceID = uint(id.(float64))
	} else {
		t.Fatalf("ID field not found in response: %v", svcMap)
	}

	fmt.Printf("[TEST] Service created with ID: %d. Waiting for UP status...\n", serviceID)

	success := false
	for i := 0; i < 30; i++ {
		status := getServiceStatus(client, token, serviceID)
		if status == "UP" {
			success = true
			fmt.Printf("\n[TEST] SUCCESS! Status changed to UP.\n")
			break
		}
		fmt.Printf(".")
		time.Sleep(1 * time.Second)
	}

	assert.True(t, success, "\n[FAIL] Timeout: Service did not reach UP status")
}

func postJSON(client *http.Client, url string, data interface{}) (*http.Response, error) {
	b, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

func postAuthenticatedJSON(client *http.Client, method, url string, data interface{}, token string) (*http.Response, error) {
	b, _ := json.Marshal(data)
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	return client.Do(req)
}

func getServiceStatus(client *http.Client, token string, id uint) string {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/services/%d", ApiUrl, id), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := client.Do(req)
	if err != nil {
		return "ERR"
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "ERR_HTTP"
	}
	var res struct {
		Status string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&res)
	return res.Status
}
