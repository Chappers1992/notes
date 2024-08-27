package main

import (
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
)

func main() {
    // API URL you want to test
    apiUrl := "https://jsonplaceholder.typicode.com/posts/1"

    // Create a new request
    req, err := http.NewRequest("GET", apiUrl, nil)
    if err != nil {
        log.Fatalf("Error creating request: %v", err)
    }

    // Add headers if needed
    req.Header.Add("Content-Type", "application/json")

    // Initialize the HTTP client and send the request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Fatalf("Error making the request: %v", err)
    }
    defer resp.Body.Close()

    // Check the response status
    if resp.StatusCode != http.StatusOK {
        log.Fatalf("Failed request: %s", resp.Status)
    }

    // Read and print the response body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatalf("Error reading response body: %v", err)
    }

    fmt.Println("Response Status:", resp.Status)
    fmt.Println("Response Body:", string(body))
}
