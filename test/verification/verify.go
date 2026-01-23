package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func main() {
	// Test 1: Invalid URL
	fmt.Println("Test 1: Invalid URL")
	status1 := sendRequest(`{"name":"Invalid","url":"htps://bad.com"}`)
	if status1 == 400 {
		fmt.Println("PASS: Got 400 Bad Request")
	} else {
		fmt.Printf("FAIL: Expected 400, got %d\n", status1)
	}

	// Test 2: Valid URL
	fmt.Println("\nTest 2: Valid URL")
	status2 := sendRequest(`{"name":"Valid","url":"https://google.com"}`)
	if status2 == 201 {
		fmt.Println("PASS: Got 201 Created")
	} else {
		fmt.Printf("FAIL: Expected 201, got %d\n", status2)
	}
}

func sendRequest(jsonBody string) int {
	url := "http://localhost:8081/targets"
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonBody)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
		return 0
	}
	defer resp.Body.Close()
	return resp.StatusCode
}
