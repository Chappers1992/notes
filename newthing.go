 // Handle manual redirects
    if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusFound ||
        resp.StatusCode == http.StatusSeeOther || resp.StatusCode == http.StatusTemporaryRedirect || resp.StatusCode == http.StatusPermanentRedirect {

        location, err := resp.Location()
        if err != nil {
            return nil, err
        }

        // Create a new request to the redirect location
        req.URL = location
        resp.Body.Close() // Close the previous response body

        fmt.Printf("Redirecting to: %s\n", location.String())

        // Issue the new request
        resp, err = h.httpClient.Do(req)
        if err != nil {
            fmt.Printf("Error making redirected request: %v\n", err)
            return nil, err
        }
    }

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("Error reading response body: %v\n", err)
        return nil, err
    }
    resp.Body.Close()

    fmt.Printf("Response Body: %s\n", string(body))
    fmt.Printf("Response Status: %s\n", resp.Status)
    fmt.Printf("Response Headers: %v\n", resp.Header)

    return resp, nil
}
