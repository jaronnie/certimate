package domainexp

import (
	"testing"
	"time"
)

func TestRegisteredDomain(t *testing.T) {
	tests := map[string]string{
		"www.example.com":     "example.com",
		"*.bsc.blockrazor.io": "blockrazor.io",
		"blockrazor.xyz.":     "blockrazor.xyz",
	}

	for input, expected := range tests {
		got, err := RegisteredDomain(input)
		if err != nil {
			t.Fatalf("RegisteredDomain(%q) returned error: %v", input, err)
		}
		if got != expected {
			t.Fatalf("RegisteredDomain(%q) = %q, expected %q", input, got, expected)
		}
	}
}

func TestWhoisServerFromIANA(t *testing.T) {
	got, ok := whoisServerFromIANA(`
domain:       IO
organisation: Internet Computer Bureau Limited
whois:        whois.nic.io
status:       ACTIVE
`)
	if !ok {
		t.Fatal("expected whois server")
	}
	if got != "whois.nic.io" {
		t.Fatalf("expected whois.nic.io, got %q", got)
	}
}

func TestWhoisServerFromIANAEmpty(t *testing.T) {
	_, ok := whoisServerFromIANA(`
domain:       BUILDERS
whois:
status:       ACTIVE
`)
	if ok {
		t.Fatal("expected empty whois server to be ignored")
	}
}

func TestRDAPEndpointFromBootstrap(t *testing.T) {
	data := []byte(`{
		"services": [
			[["com"], ["https://rdap.verisign.com/com/v1/"]],
			[["builders", "business"], ["https://rdap.identitydigital.services/rdap/"]]
		]
	}`)

	got, ok := rdapEndpointFromBootstrap(data, "blockrazor.builders")
	if !ok {
		t.Fatal("expected bootstrap endpoint")
	}
	if got != "https://rdap.identitydigital.services/rdap/" {
		t.Fatalf("expected https://rdap.identitydigital.services/rdap/, got %q", got)
	}
}

func TestRDAPExpirationTime(t *testing.T) {
	got, ok := rdapExpirationTime([]rdapEvent{
		{EventAction: "registration", EventDate: "2026-01-01T00:00:00Z"},
		{EventAction: "expiration", EventDate: "2027-02-03T04:05:06Z"},
	})
	if !ok {
		t.Fatal("expected expiration event")
	}

	expected := time.Date(2027, 2, 3, 4, 5, 6, 0, time.UTC)
	if !got.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestDomainExpirationTime(t *testing.T) {
	got, ok := domainExpirationTime(`
Domain Name: BLOCKRAZOR.IO
Registrar Registration Expiration Date: 2027-02-03T04:05:06Z
Updated Date: 2026-01-01T00:00:00Z
`)
	if !ok {
		t.Fatal("expected expiration event")
	}

	expected := time.Date(2027, 2, 3, 4, 5, 6, 0, time.UTC)
	if !got.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestDomainExpirationTimeWithDateOnly(t *testing.T) {
	got, ok := domainExpirationTime(`
Domain Name: example.test
Expiry Date: 2027-02-03
`)
	if !ok {
		t.Fatal("expected expiration date")
	}

	expected := time.Date(2027, 2, 3, 0, 0, 0, 0, time.UTC)
	if !got.Equal(expected) {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}
