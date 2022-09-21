package request

import "net/http"

type option func(c *client)

func WithUserAgent(userAgent string) option {
	return func(c *client) {
		c.userAgent = userAgent
	}
}

func WithHTTPClient(httpClient *http.Client) option {
	return func(c *client) {
		c.httpClient = httpClient
	}
}

func WithBearerToken(bearerToken string) option {
	return func(c *client) {
		c.bearerToken = bearerToken
		c.authType = authTypeBearer
	}
}

func WithBasicAuth(user, pass string) option {
	return func(c *client) {
		c.basicUser = user
		c.basicPass = pass
		c.authType = authTypeBasic
	}
}

func WithErrChecker(errChecker HTTPErrChecker) option {
	return func(c *client) {
		c.errChecker = errChecker
	}
}
