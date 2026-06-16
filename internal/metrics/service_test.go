package metrics

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
)

type fakeMonitorCertificateRepository struct {
	monitorCertificates []*domain.MonitorCertificate
}

func (r fakeMonitorCertificateRepository) List(ctx context.Context) ([]*domain.MonitorCertificate, error) {
	return r.monitorCertificates, nil
}

type fakeMonitorDomainRepository struct {
	monitorDomains []*domain.MonitorDomain
}

func (r fakeMonitorDomainRepository) List(ctx context.Context) ([]*domain.MonitorDomain, error) {
	return r.monitorDomains, nil
}

func TestServiceRender(t *testing.T) {
	service := NewService(
		fakeMonitorCertificateRepository{
			monitorCertificates: []*domain.MonitorCertificate{
				{
					Meta: domain.Meta{
						Id:        "old-cert",
						UpdatedAt: time.Unix(10, 0),
					},
					MonitorDomain:              "www.example.com",
					CertificateSubjectAltNames: "example.com;www.example.com",
					CertificateNotAfter:        time.Unix(86400, 0),
				},
				{
					Meta: domain.Meta{
						Id:        "new-cert",
						UpdatedAt: time.Unix(20, 0),
					},
					MonitorDomain:              "api.example.com",
					CertificateSubjectAltNames: "example.com;www.example.com",
					CertificateNotAfter:        time.Unix(172800, 0),
				},
			},
		},
		fakeMonitorDomainRepository{
			monitorDomains: []*domain.MonitorDomain{
				{
					Meta: domain.Meta{
						Id:        "old-domain",
						UpdatedAt: time.Unix(10, 0),
					},
					MonitorDomain:  "www.example.com",
					DomainName:     "example.com",
					DomainNotAfter: time.Unix(86400, 0),
				},
				{
					Meta: domain.Meta{
						Id:        "new-domain",
						UpdatedAt: time.Unix(20, 0),
					},
					MonitorDomain:  "api.example.com",
					DomainName:     "example.com",
					DomainNotAfter: time.Unix(172800, 0),
				},
			},
		},
	)
	service.now = func() time.Time {
		return time.Unix(0, 0)
	}

	res, _, err := service.Render(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	body := string(res)
	if !strings.Contains(body, "certimate_certificate_days_until_expiry") {
		t.Fatalf("expected certificate days metric, got:\n%s", body)
	}
	if !strings.Contains(body, "certimate_domain_days_until_expiry") {
		t.Fatalf("expected domain days metric, got:\n%s", body)
	}
	if strings.Contains(body, "certimate_certificate_not_after_timestamp_seconds") {
		t.Fatalf("unexpected not after timestamp metric, got:\n%s", body)
	}
	if strings.Count(body, `domain="example.com;www.example.com"`) != 1 {
		t.Fatalf("expected certificate domain to be de-duplicated, got:\n%s", body)
	}
	if strings.Count(body, `certimate_domain_days_until_expiry{domain="example.com"}`) != 1 {
		t.Fatalf("expected registered domain to be de-duplicated, got:\n%s", body)
	}
	if !strings.Contains(body, `monitor_id="new-cert"`) {
		t.Fatalf("expected latest monitor certificate record to win, got:\n%s", body)
	}
	if strings.Contains(body, `monitor_domain="api.example.com"`) && strings.Contains(body, `certimate_domain_days_until_expiry{domain="example.com",`) {
		t.Fatalf("expected domain metric to use only the domain label, got:\n%s", body)
	}
}

func TestCollectCertificateMetrics(t *testing.T) {
	oldMonitor := &domain.MonitorCertificate{
		Meta: domain.Meta{
			Id:        "old",
			UpdatedAt: time.Unix(100, 0),
		},
		CertificateSubjectAltNames: "example.com;www.example.com",
		CertificateNotAfter:        time.Unix(1000, 0),
	}
	newMonitor := &domain.MonitorCertificate{
		Meta: domain.Meta{
			Id:        "new",
			UpdatedAt: time.Unix(200, 0),
		},
		CertificateSubjectAltNames: "example.com;www.example.com",
		CertificateNotAfter:        time.Unix(2000, 0),
	}

	metrics := collectCertificateMetrics([]*domain.MonitorCertificate{oldMonitor, newMonitor})

	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].monitorCertificate.Id != "new" {
		t.Fatalf("expected latest monitor metric, got id=%q", metrics[0].monitorCertificate.Id)
	}
}
