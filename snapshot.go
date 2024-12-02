package grafana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SnapshotRequest struct {
	Dashboard interface{} `json:"dashboard"`
	Name      string      `json:"name"`
	Expires   int         `json:"expires"`
}

type SnapshotResponse struct {
	URL string `json:"url"`
}

func ClickSnapshot(url, token, dashboardUID string, hours int) (string, error) {
	dashboardJSON, err := getDashboardJSON(url, token, dashboardUID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch dashboard JSON: %v", err)
	}

	setDynamicTimeRange(dashboardJSON, fmt.Sprintf("%dh", hours))

	payload := SnapshotRequest{
		Dashboard: dashboardJSON,
		Name:      fmt.Sprintf("%s Snapshot - Last %d Hours", dashboardUID, hours),
		Expires:   0, // 0 means the snapshot will not expire
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot payload: %v", err)
	}

	apiURL := fmt.Sprintf("%s/api/snapshots", url)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send snapshot request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected response from Grafana API: %d, body: %s", resp.StatusCode, string(body))
	}

	var snapshotResp SnapshotResponse
	err = json.NewDecoder(resp.Body).Decode(&snapshotResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode snapshot response: %v", err)
	}

	return snapshotResp.URL, nil
}

func getDashboardJSON(url, token, dashboardUID string) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("%s/api/dashboards/uid/%s", url, dashboardUID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected response from Grafana API: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var dashboardResponse map[string]interface{}
	err = json.Unmarshal(body, &dashboardResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse dashboard JSON: %v", err)
	}

	dashboardJSON, ok := dashboardResponse["dashboard"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to extract 'dashboard' key from response")
	}

	return dashboardJSON, nil
}

func setDynamicTimeRange(dashboardJSON map[string]interface{}, timeRange string) {
	dashboardJSON["time"] = map[string]string{
		"from": fmt.Sprintf("now - %s", timeRange),
		"to":   "now",
	}
}
