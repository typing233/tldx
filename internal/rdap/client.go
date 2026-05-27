package rdap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CheckResult struct {
	Domain    string
	Available bool
	Error     error
	Duration  time.Duration
}

type Client struct {
	httpClient *http.Client
	bootstrap  *Bootstrap
	retries    int
}

func NewClient(httpClient *http.Client, bootstrap *Bootstrap, retries int) *Client {
	return &Client{
		httpClient: httpClient,
		bootstrap:  bootstrap,
		retries:    retries,
	}
}

func (c *Client) Check(ctx context.Context, domain string) CheckResult {
	start := time.Now()

	tld := extractTLD(domain)
	server, err := c.bootstrap.ServerForTLD(tld)
	if err != nil {
		return CheckResult{Domain: domain, Error: err, Duration: time.Since(start)}
	}

	url := server + "/domain/" + domain

	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt*attempt) * 500 * time.Millisecond
			select {
			case <-ctx.Done():
				return CheckResult{Domain: domain, Error: ctx.Err(), Duration: time.Since(start)}
			case <-time.After(backoff):
			}
		}

		result, err := c.doCheck(ctx, url, domain)
		if err == nil {
			result.Duration = time.Since(start)
			return result
		}

		lastErr = err
		if !isRetryable(err) {
			break
		}
	}

	return CheckResult{Domain: domain, Error: lastErr, Duration: time.Since(start)}
}

func (c *Client) doCheck(ctx context.Context, url, domain string) (CheckResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return CheckResult{Domain: domain}, err
	}
	req.Header.Set("Accept", "application/rdap+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if isNetworkRetryable(err) {
			return CheckResult{Domain: domain}, &retryableError{msg: fmt.Sprintf("network error: %v", err)}
		}
		return CheckResult{Domain: domain}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
		return CheckResult{Domain: domain, Available: false}, nil
	case resp.StatusCode == http.StatusNotFound:
		return CheckResult{Domain: domain, Available: true}, nil
	case resp.StatusCode == http.StatusTooManyRequests:
		return CheckResult{Domain: domain}, &retryableError{msg: "rate limited (429)"}
	case resp.StatusCode >= 500:
		return CheckResult{Domain: domain}, &retryableError{msg: fmt.Sprintf("server error (%d)", resp.StatusCode)}
	default:
		return CheckResult{Domain: domain}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}

func extractTLD(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return domain
	}
	return parts[len(parts)-1]
}

type retryableError struct {
	msg string
}

func (e *retryableError) Error() string {
	return e.msg
}

func isRetryable(err error) bool {
	var re *retryableError
	return errors.As(err, &re)
}

func isNetworkRetryable(err error) bool {
	if errors.Is(err, context.Canceled) {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout()
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Timeout() || urlErr.Temporary()
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return dnsErr.IsTemporary
	}
	return false
}
