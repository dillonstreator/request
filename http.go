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
	defaultUserAgent = "github.com/dillonstreator/request@v" + version

	authTypeBasic authType = iota
	authTypeToken
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

type ResponseUnmarshaler func(bytes []byte, out any) error

type client struct {
	httpClient  *http.Client
	userAgent   string
	contentType string
	baseURL     string
	token       string
	basicUser   string
	basicPass   string
	authType    authType

	errChecker          HTTPErrChecker
	responseUnmarshaler ResponseUnmarshaler
}

var _ Client = (*client)(nil)

func NewClient(baseURL string, options ...option) *client {
	c := &client{
		httpClient:  &http.Client{},
		userAgent:   defaultUserAgent,
		contentType: "application/json",
		baseURL:     strings.TrimSuffix(baseURL, "/"),
	}

	for _, o := range options {
		o(c)
	}

	if c.errChecker == nil {
		c.errChecker = defaultErrChecker
	}

	if c.responseUnmarshaler == nil {
		c.responseUnmarshaler = json.Unmarshal
	}

	return c
}

func (c *client) Request(ctx context.Context, method, path string, headers http.Header, body io.Reader, out any) (*http.Response, error) {
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
	req.Header.Add("Content-Type", c.contentType)

	switch c.authType {
	case authTypeBasic:
		req.SetBasicAuth(c.basicUser, c.basicPass)
	case authTypeToken:
		req.Header.Set("Authorization", c.token)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return res, err
	}

	if err := c.errChecker(req, res); err != nil {
		return res, err
	}

	if out != nil {
		resBody, err := io.ReadAll(res.Body)
		if err != nil {
			return res, err
		}
		res.Body.Close()

		res.Body = io.NopCloser(bytes.NewBuffer(resBody))

		err = c.responseUnmarshaler(resBody, out)
		if err != nil {
			return res, fmt.Errorf("unmarshaling response body: %w", err)
		}
	}

	return res, nil
}

func (c *client) Get(ctx context.Context, path string, headers http.Header, query url.Values, out interface{}) (*http.Response, error) {
	if len(query) > 0 {
		path = fmt.Sprintf("%s?%s", path, query.Encode())
	}

	return c.Request(ctx, http.MethodGet, path, nil, nil, out)
}

func (c *client) Post(ctx context.Context, path string, headers http.Header, body io.Reader, out interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPost, path, headers, body, out)
}

func (c *client) Put(ctx context.Context, path string, headers http.Header, body io.Reader, out interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPut, path, headers, body, out)
}

func (c *client) Patch(ctx context.Context, path string, headers http.Header, body io.Reader, out interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodPatch, path, headers, body, out)
}

func (c *client) Delete(ctx context.Context, path string, headers http.Header, body io.Reader, out interface{}) (*http.Response, error) {
	return c.Request(ctx, http.MethodDelete, path, headers, body, out)
}
