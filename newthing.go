package rest

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"strings"
	"time"
	"troweprice.io/ea/unity/unity-cli/cmd/logger"
	"fmt"
)

type Option func(*HttpClient) error

type RequestDetail struct {
	Method string
	Path   string
	Body   []byte
}

type HttpClient struct {
	httpClient   *http.Client
	baseURL      string
	userId       string
	token        string
	tokenType    TokenType
	useHttps     bool
	useOpaqueURL bool
}

type TokenType int8

const (
	Bear TokenType = iota
	Basic
	GitlabPrivate
)

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
			if !strings.Contains(h.token, ":") {
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

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Headers: %v\n", resp.Header)

	return resp, nil
}

func AuthTokenOption(tokenType TokenType, token string) Option {
	return func(httpClient *HttpClient) error {
		httpClient.tokenType = tokenType
		httpClient.token = token
		return nil
	}
}

func EnableHttpsOption() Option {
	return func(httpClient *HttpClient) error {
		httpClient.useHttps = true
		return nil
	}
}

func SkipTLSVerifyOption() Option {
	return func(httpClient *HttpClient) error {
		httpClient.httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		return nil
	}
}

func UserIdOption(userId string) Option {
	return func(httpClient *HttpClient) error {
		if userId != "" {
			httpClient.userId = userId
		}
		return nil
	}
}

func BaseUrlOption(baseUrl string) Option {
	return func(httpClient *HttpClient) error {
		if baseUrl == "" {
			return errors.New("base URL is required")
		}
		httpClient.baseURL = baseUrl
		return nil
	}
}

func UseOpaqueUrlOption() Option {
	return func(httpClient *HttpClient) error {
		httpClient.useOpaqueURL = true
		return nil
	}
}

func NewHttpClientWithOptions(options ...Option) (*HttpClient, error) {
	httpClient := &HttpClient{
		httpClient: &http.Client{
			Timeout: time.Second * 60,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) > 0 {
					for key, values := range via[0].Header {
						req.Header[key] = values
					}
				}
				return nil
			},
		},
	}

	for _, option := range options {
		err := option(httpClient)
		if err != nil {
			return nil, err
		}
	}

	logger.Debugf("Create a httpClient with baseURL '%s', userID '%s', and https %t",
		httpClient.baseURL, httpClient.userId, httpClient.useHttps)

	return httpClient, nil
}

func createRequestURL(p *HttpClient, urlPath string) string {
	scheme := "https"
	if !p.useHttps {
		scheme = "http"
	}

	query := ""
	if p.userId != "" {
		parse, _ := url.Parse("userId=" + p.userId)
		query = parse.String()
	}

	opaque := ""
	if p.useOpaqueURL {
		opaque = "//" + p.baseURL + urlPath
	}
	urlObj := url.URL{
		Scheme:   scheme,
		Host:     p.baseURL,
		Path:     urlPath,
		Opaque:   opaque,
		RawQuery: query,
	}

	return urlObj.String()
}
