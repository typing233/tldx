package rdap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

var fallbackServers = map[string]string{
	"com":    "https://rdap.verisign.com/com/v1",
	"net":    "https://rdap.verisign.com/net/v1",
	"org":    "https://rdap.org",
	"io":     "https://rdap.nic.io",
	"ai":     "https://rdap.nic.ai",
	"dev":    "https://rdap.nic.google",
	"app":    "https://rdap.nic.google",
	"co":     "https://rdap.nic.co",
	"me":     "https://rdap.nic.me",
	"to":     "https://rdap.nic.to",
	"info":   "https://rdap.nic.info",
	"biz":    "https://rdap.nic.biz",
	"xyz":    "https://rdap.nic.xyz",
	"tech":   "https://rdap.nic.tech",
	"online": "https://rdap.nic.online",
	"site":   "https://rdap.nic.site",
	"store":  "https://rdap.nic.store",
	"fun":    "https://rdap.nic.fun",
	"us":     "https://rdap.nic.us",
	"uk":     "https://rdap.nic.uk",
	"ca":     "https://rdap.ca.fury.ca/rdap",
	"de":     "https://rdap.denic.de",
	"fr":     "https://rdap.nic.fr",
	"jp":     "https://rdap.jprs.jp/rdap",
}

type Bootstrap struct {
	mu      sync.RWMutex
	servers map[string]string
}

type bootstrapResponse struct {
	Services [][][]string `json:"services"`
}

func NewBootstrap(ctx context.Context, httpClient *http.Client) *Bootstrap {
	b := &Bootstrap{
		servers: make(map[string]string),
	}

	// Copy fallback servers
	for k, v := range fallbackServers {
		b.servers[k] = v
	}

	// Try to fetch live bootstrap data
	b.fetchIANA(ctx, httpClient)

	return b
}

func (b *Bootstrap) fetchIANA(ctx context.Context, httpClient *http.Client) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://data.iana.org/rdap/dns.json", nil)
	if err != nil {
		return
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var data bootstrapResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, service := range data.Services {
		if len(service) < 2 || len(service[1]) == 0 {
			continue
		}
		serverURL := strings.TrimRight(service[1][0], "/")
		for _, tld := range service[0] {
			b.servers[strings.ToLower(tld)] = serverURL
		}
	}
}

func (b *Bootstrap) ServerForTLD(tld string) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	tld = strings.ToLower(tld)
	server, ok := b.servers[tld]
	if !ok {
		return "", fmt.Errorf("no RDAP server known for TLD %q", tld)
	}
	return server, nil
}
