package request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type authType int8

const (
	defaultUserAgent = "github.com/dillonstreator/request/" + version

	authTypeBasic authType = iota
	authTypeBearer
)

type HTTPError struct {
	HTTPResponse *http.Response
	StatusCode   int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http status %d", e.StatusCode)
}

type HTTPErrChecker func(req *http.Request, res *http.Response) error

func defaultErrChecker(req *http.Request, res *http.Response) error {
	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return &HTTPError{
			HTTPResponse: res,
			StatusCode:   res.StatusCode,
		}
	}

	return nil
}

type client struct {
	httpClient  *http.Client
	userAgent   string
	baseURL     string
	bearerToken string
	basicUser   string
	basicPass   string
	authType    authType

	errChecker HTTPErrChecker
}

var _ Client = (*client)(nil)

func NewClient(baseURL string, options ...option) *client {
	c := &client{
		httpClient: &http.Client{},
		userAgent:  defaultUserAgent,
		baseURL:    strings.TrimSuffix(baseURL, "/"),
	}

	for _, o := range options {
		o(c)
	}

	if c.errChecker == nil {
		c.errChecker = defaultErrChecker
	}

	return c
}

func (c *client) Request(ctx context.Context, method, path string, body io.Reader, headers http.Header, out interface{}) (*http.Response, error) {
	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}
	url := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	if headers != nil {
		req.Header = headers
	}

	req.Header.Add("User-Agent", c.userAgent)

	switch c.authType {
	case authTypeBasic:
		req.SetBasicAuth(c.basicUser, c.basicPass)
	case authTypeBearer:
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return res, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return res, err
	}
	res.Body.Close()

	res.Body = io.NopCloser(bytes.NewBuffer(resBody))

	if err := c.errChecker(req, res); err != nil {
		return res, err
	}

	if out != nil {
		err = json.Unmarshal(resBody, out)
		if err != nil {
			return res, fmt.Errorf("unmarshaling response body: %w", err)
		}
	}

	return res, nil
}

func (c *client) Get(ctx context.Context, path string, query url.Values, out interface{}) (*http.Response, error) {
	if len(query) > 0 {
		path = fmt.Sprintf("%s?%s", path, query.Encode())
	}

	return c.Request(ctx, http.MethodGet, path, nil, nil, out)
}

func (c *client) Post(ctx context.Context, path string, body io.Reader, out interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPost, path, body, nil, out)
}

func (c *client) Put(ctx context.Context, path string, body io.Reader, out interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPut, path, body, nil, out)
}

func (c *client) Patch(ctx context.Context, path string, body io.Reader, out interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPatch, path, body, nil, out)
}

func (c *client) Delete(ctx context.Context, path string, body io.Reader, out interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodDelete, path, body, nil, out)
}
