func (h *HttpClient) Request(info *RequestDetail, headers ...string) (*http.Response, error) {
	// If httpClient hasn't been initialized with custom redirect handling, do it now
	if h.httpClient.CheckRedirect == nil {
		h.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) > 0 {
				for key, values := range via[0].Header {
					req.Header[key] = values
				}
			}
			return nil
		}
	}

	req, err := http.NewRequest(info.Method, createRequestURL(h, info.Path), bytes.NewBuffer(info.Body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if h.token != "" {
		switch h.tokenType {
		case Bear:
			req.Header.Set("Authorization", "Bearer "+h.token)
		case Basic:
			auth := h.token
			if !strings.Contains(auth, ":") {
				auth += ":"
			}
			req.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
		case GitlabPrivate:
			req.Header.Set("PRIVATE-TOKEN", h.token)
		default:
			return nil, errors.Errorf("unknown token type: %v", h.tokenType)
		}
	}
	if h.userId != "" {
		req.Header.Set("X-Trp-User", h.userId)
	}

	for _, header := range headers {
		split := strings.SplitN(header, "=", 2)
		if len(split) == 2 {
			req.Header.Set(split[0], split[1])
		}
	}

	// Log request details
	fmt.Printf("Making request to: %s\n", req.URL.String())
	fmt.Printf("Request method: %s\n", req.Method)
	fmt.Printf("Request headers: %v\n", req.Header)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return nil, err
	}

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

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Headers: %v\n", resp.Header)

	return resp, nil
}
