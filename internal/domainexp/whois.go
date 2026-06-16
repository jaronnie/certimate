package domainexp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

const (
	defaultWhoisServer       = "whois.iana.org"
	defaultBootstrapEndpoint = "https://data.iana.org/rdap/dns.json"
)

var ErrNoRegisteredDomain = errors.New("no registered domain")
var errNoWhoisServer = errors.New("iana whois response does not contain whois server")

type LookupResult struct {
	Domain   string
	NotAfter time.Time
}

type Client struct {
	Dialer            *net.Dialer
	HTTPClient        *http.Client
	WhoisServer       string
	BootstrapEndpoint string
}

func Lookup(ctx context.Context, host string) (*LookupResult, error) {
	return (&Client{}).Lookup(ctx, host)
}

func (c *Client) Lookup(ctx context.Context, host string) (*LookupResult, error) {
	domain, err := RegisteredDomain(host)
	if err != nil {
		return nil, err
	}

	dialer := c.Dialer
	if dialer == nil {
		dialer = &net.Dialer{Timeout: 10 * time.Second}
	}

	whoisServer := c.WhoisServer
	if whoisServer == "" {
		whoisServer, err = c.lookupWhoisServer(ctx, dialer, domain)
		if err != nil {
			if errors.Is(err, errNoWhoisServer) {
				return c.lookupRDAP(ctx, domain)
			}
			return nil, err
		}
	}

	resp, err := whoisQuery(ctx, dialer, whoisServer, domain)
	if err != nil {
		return nil, err
	}

	notAfter, ok := domainExpirationTime(resp)
	if !ok {
		return nil, fmt.Errorf("whois response does not contain expiration date")
	}

	return &LookupResult{
		Domain:   domain,
		NotAfter: notAfter,
	}, nil
}

func (c *Client) lookupWhoisServer(ctx context.Context, dialer *net.Dialer, domain string) (string, error) {
	tld := domainTLD(domain)
	if tld == "" {
		return "", ErrNoRegisteredDomain
	}

	resp, err := whoisQuery(ctx, dialer, defaultWhoisServer, tld)
	if err != nil {
		return "", err
	}

	server, ok := whoisServerFromIANA(resp)
	if !ok {
		return "", fmt.Errorf("%w for domain %q", errNoWhoisServer, domain)
	}
	return server, nil
}

func (c *Client) lookupRDAP(ctx context.Context, domain string) (*LookupResult, error) {
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	rdapEndpoint, err := c.lookupRDAPEndpoint(ctx, httpClient, domain)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rdapDomainURL(rdapEndpoint, domain), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/rdap+json, application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("rdap request failed: status %d", resp.StatusCode)
	}

	var body rdapDomainResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	notAfter, ok := rdapExpirationTime(body.Events)
	if !ok {
		return nil, fmt.Errorf("rdap response does not contain expiration event")
	}

	return &LookupResult{
		Domain:   domain,
		NotAfter: notAfter,
	}, nil
}

func (c *Client) lookupRDAPEndpoint(ctx context.Context, httpClient *http.Client, domain string) (string, error) {
	bootstrapEndpoint := c.BootstrapEndpoint
	if bootstrapEndpoint == "" {
		bootstrapEndpoint = defaultBootstrapEndpoint
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, bootstrapEndpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("rdap bootstrap request failed: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	endpoint, ok := rdapEndpointFromBootstrap(body, domain)
	if !ok {
		return "", fmt.Errorf("rdap bootstrap response does not contain endpoint for domain %q", domain)
	}
	return endpoint, nil
}

func whoisQuery(ctx context.Context, dialer *net.Dialer, server string, query string) (string, error) {
	if _, _, err := net.SplitHostPort(server); err != nil {
		server = net.JoinHostPort(server, "43")
	}

	conn, err := dialer.DialContext(ctx, "tcp4", server)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	} else {
		_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	}

	if _, err := fmt.Fprintf(conn, "%s\r\n", query); err != nil {
		return "", err
	}

	var b strings.Builder
	if _, err := io.Copy(&b, conn); err != nil {
		return "", err
	}
	return b.String(), nil
}

func RegisteredDomain(host string) (string, error) {
	host = strings.TrimSpace(strings.ToLower(host))
	host = strings.TrimSuffix(host, ".")
	host = strings.TrimPrefix(host, "*.")
	if host == "" || net.ParseIP(host) != nil {
		return "", ErrNoRegisteredDomain
	}

	domain, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrNoRegisteredDomain, err)
	}
	return domain, nil
}

func whoisServerFromIANA(resp string) (string, bool) {
	for _, line := range strings.Split(resp, "\n") {
		key, value, ok := splitWhoisLine(line)
		if !ok {
			continue
		}
		if strings.EqualFold(key, "whois") && value != "" {
			return value, true
		}
	}
	return "", false
}

func rdapDomainURL(endpoint, domain string) string {
	endpoint = strings.TrimRight(endpoint, "/")
	return endpoint + "/domain/" + url.PathEscape(domain)
}

type rdapDomainResponse struct {
	Events []rdapEvent `json:"events"`
}

type rdapBootstrapResponse struct {
	Services []rdapBootstrapService `json:"services"`
}

type rdapBootstrapService struct {
	TLDs      []string
	Endpoints []string
}

func (s *rdapBootstrapService) UnmarshalJSON(data []byte) error {
	var raw [2][]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	s.TLDs = raw[0]
	s.Endpoints = raw[1]
	return nil
}

type rdapEvent struct {
	EventAction string `json:"eventAction"`
	EventDate   string `json:"eventDate"`
}

func rdapExpirationTime(events []rdapEvent) (time.Time, bool) {
	for _, event := range events {
		if !strings.Contains(strings.ToLower(event.EventAction), "expir") {
			continue
		}

		t, err := time.Parse(time.RFC3339, event.EventDate)
		if err != nil {
			continue
		}
		return t, true
	}
	return time.Time{}, false
}

func rdapEndpointFromBootstrap(data []byte, domain string) (string, bool) {
	tld := domainTLD(domain)
	if tld == "" {
		return "", false
	}

	var body rdapBootstrapResponse
	if err := json.Unmarshal(data, &body); err != nil {
		return "", false
	}

	for _, service := range body.Services {
		for _, candidate := range service.TLDs {
			if strings.EqualFold(candidate, tld) && len(service.Endpoints) > 0 {
				return service.Endpoints[0], true
			}
		}
	}
	return "", false
}

func domainExpirationTime(resp string) (time.Time, bool) {
	for _, line := range strings.Split(resp, "\n") {
		key, value, ok := splitWhoisLine(line)
		if !ok || value == "" || !isExpirationField(key) {
			continue
		}

		if t, ok := parseWhoisTime(value); ok {
			return t, true
		}
	}
	return time.Time{}, false
}

func splitWhoisLine(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "%") || strings.HasPrefix(line, "#") {
		return "", "", false
	}

	key, value, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(key), strings.TrimSpace(value), true
}

func isExpirationField(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if normalized == "expires" || normalized == "expiry date" || normalized == "expiration date" || normalized == "expiration time" {
		return true
	}
	return strings.Contains(normalized, "expir") && strings.Contains(normalized, "date")
}

func parseWhoisTime(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, ".")

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.999Z",
		"2006-01-02 15:04:05 MST",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006.01.02",
		"02-Jan-2006",
		"02-Jan-2006 15:04:05 UTC",
		"02-Jan-2006 15:04:05 MST",
		"Jan 02 2006",
		"January 02 2006",
		"2006/01/02",
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t, true
		}
	}

	if t, ok := parseUnixTimestamp(value); ok {
		return t, true
	}

	return time.Time{}, false
}

func parseUnixTimestamp(value string) (time.Time, bool) {
	if len(value) < 9 || len(value) > 13 {
		return time.Time{}, false
	}
	if _, err := strconv.ParseInt(value, 10, 64); err != nil {
		return time.Time{}, false
	}

	ts, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, false
	}
	if len(value) == 13 {
		return time.UnixMilli(ts).UTC(), true
	}
	return time.Unix(ts, 0).UTC(), true
}

func domainTLD(domain string) string {
	parts := strings.Split(strings.TrimSuffix(strings.ToLower(domain), "."), ".")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}
