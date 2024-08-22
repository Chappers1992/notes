// Custom HttpClient with preserved headers on redirect
func (h *HttpClient) Request(info *RequestDetail, headers ...string) (*http.Response, error) {
    req, err := http.NewRequest(info.Method, createRequestURL(h, info.Path), bytes.NewBuffer(info.Body))
    if err != nil {
        return nil, err
    }

    // Add default headers here
    req.Header.Add("Content-Type", "application/json")
    if h.token != "" {
        switch h.tokenType {
        case Bear:
            req.Header.Add("Authorization", "Bearer "+h.token)
        case Basic:
            auth := h.token
            req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(auth)))
        case GitlabPrivate:
            req.Header.Add("PRIVATE-TOKEN", h.token)
        default:
            return nil, errors.Errorf("unknown token type: %v", h.tokenType)
        }
    }

    if h.userId != "" {
        req.Header.Add("X-Trp-User", h.userId)
    }

    // Add custom headers
    for _, header := range headers {
        split := strings.Split(header, "=")
        req.Header.Add(split[0], split[1])
    }

    // Create a custom HTTP client with CheckRedirect to preserve headers
    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            // Copy all headers from the initial request to the new one
            for key, values := range via[0].Header {
                for _, value := range values {
                    req.Header.Add(key, value)
                }
            }
            return nil
        },
    }

    return client.Do(req)
}
