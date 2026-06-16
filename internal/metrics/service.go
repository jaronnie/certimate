package metrics

import (
	"bytes"
	"context"
	"math"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"

	"github.com/certimate-go/certimate/internal/domain"
)

type monitorCertificateRepository interface {
	List(ctx context.Context) ([]*domain.MonitorCertificate, error)
}

type monitorDomainRepository interface {
	List(ctx context.Context) ([]*domain.MonitorDomain, error)
}

type Service struct {
	monitorCertificateRepo monitorCertificateRepository
	monitorDomainRepo      monitorDomainRepository
	now                    func() time.Time
}

func NewService(monitorCertificateRepo monitorCertificateRepository, monitorDomainRepo monitorDomainRepository) *Service {
	return &Service{
		monitorCertificateRepo: monitorCertificateRepo,
		monitorDomainRepo:      monitorDomainRepo,
		now:                    time.Now,
	}
}

func (s *Service) Render(ctx context.Context) ([]byte, string, error) {
	monitorCertificates, err := s.monitorCertificateRepo.List(ctx)
	if err != nil {
		return nil, "", err
	}

	monitorDomains, err := s.monitorDomainRepo.List(ctx)
	if err != nil {
		return nil, "", err
	}

	now := s.now()
	certificateMetrics := collectCertificateMetrics(monitorCertificates)
	domainMetrics := collectDomainMetrics(monitorDomains)

	registry := prometheus.NewRegistry()
	certificateLabels := []string{"domain", "monitor_domain", "monitor_id", "serial_number", "issuer", "source", "workflow_id"}
	daysUntilExpiryGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "certimate_certificate_days_until_expiry",
		Help: "Certificate remaining lifetime in days.",
	}, certificateLabels)
	domainDaysUntilExpiryGauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "certimate_domain_days_until_expiry",
		Help: "Domain registration remaining lifetime in days.",
	}, []string{"domain"})

	registry.MustRegister(daysUntilExpiryGauge, domainDaysUntilExpiryGauge)
	for _, metric := range certificateMetrics {
		labelValues := metric.labelValues()

		daysLeft := metric.notAfter.Sub(now).Hours() / 24
		if math.Abs(daysLeft) == 0 {
			daysLeft = 0
		}
		daysUntilExpiryGauge.WithLabelValues(labelValues...).Set(daysLeft)
	}
	for _, metric := range domainMetrics {
		labelValues := metric.labelValues()

		daysLeft := metric.notAfter.Sub(now).Hours() / 24
		if math.Abs(daysLeft) == 0 {
			daysLeft = 0
		}
		domainDaysUntilExpiryGauge.WithLabelValues(labelValues...).Set(daysLeft)
	}

	metricFamilies, err := registry.Gather()
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	encoder := expfmt.NewEncoder(&buf, expfmt.FmtText)
	for _, mf := range metricFamilies {
		if err := encoder.Encode(mf); err != nil {
			return nil, "", err
		}
	}

	return buf.Bytes(), string(expfmt.FmtText), nil
}

type certificateMetric struct {
	monitorCertificate *domain.MonitorCertificate
	notAfter           time.Time
}

func (m certificateMetric) labelValues() []string {
	return []string{
		m.monitorCertificate.CertificateSubjectAltNames,
		m.monitorCertificate.MonitorDomain,
		m.monitorCertificate.Id,
		m.monitorCertificate.CertificateSerialNumber,
		m.monitorCertificate.CertificateIssuerName,
		"monitor",
		m.monitorCertificate.WorkflowId,
	}
}

type domainMetric struct {
	monitorDomain *domain.MonitorDomain
	notAfter      time.Time
}

func (m domainMetric) labelValues() []string {
	return []string{
		m.monitorDomain.DomainName,
	}
}

func collectCertificateMetrics(monitorCertificates []*domain.MonitorCertificate) []certificateMetric {
	latestByDomain := make(map[string]*domain.MonitorCertificate)

	for _, monitorCertificate := range monitorCertificates {
		if monitorCertificate == nil || monitorCertificate.CertificateNotAfter.IsZero() || monitorCertificate.CertificateSubjectAltNames == "" {
			continue
		}

		current, ok := latestByDomain[monitorCertificate.CertificateSubjectAltNames]
		if !ok || monitorCertificateTimestamp(monitorCertificate).After(monitorCertificateTimestamp(current)) {
			latestByDomain[monitorCertificate.CertificateSubjectAltNames] = monitorCertificate
		}
	}

	metrics := make([]certificateMetric, 0, len(latestByDomain))
	for _, monitorCertificate := range latestByDomain {
		metrics = append(metrics, certificateMetric{
			monitorCertificate: monitorCertificate,
			notAfter:           monitorCertificate.CertificateNotAfter,
		})
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].monitorCertificate.CertificateSubjectAltNames < metrics[j].monitorCertificate.CertificateSubjectAltNames
	})

	return metrics
}

func collectDomainMetrics(monitorDomains []*domain.MonitorDomain) []domainMetric {
	latestByDomain := make(map[string]*domain.MonitorDomain)

	for _, monitorDomain := range monitorDomains {
		if monitorDomain == nil || monitorDomain.DomainNotAfter.IsZero() || monitorDomain.DomainName == "" {
			continue
		}

		current, ok := latestByDomain[monitorDomain.DomainName]
		if !ok || monitorDomainTimestamp(monitorDomain).After(monitorDomainTimestamp(current)) {
			latestByDomain[monitorDomain.DomainName] = monitorDomain
		}
	}

	metrics := make([]domainMetric, 0, len(latestByDomain))
	for _, monitorDomain := range latestByDomain {
		metrics = append(metrics, domainMetric{
			monitorDomain: monitorDomain,
			notAfter:      monitorDomain.DomainNotAfter,
		})
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].monitorDomain.DomainName < metrics[j].monitorDomain.DomainName
	})

	return metrics
}

func monitorCertificateTimestamp(monitorCertificate *domain.MonitorCertificate) time.Time {
	if monitorCertificate == nil {
		return time.Time{}
	}
	if !monitorCertificate.UpdatedAt.IsZero() {
		return monitorCertificate.UpdatedAt
	}
	return monitorCertificate.CreatedAt
}

func monitorDomainTimestamp(monitorDomain *domain.MonitorDomain) time.Time {
	if monitorDomain == nil {
		return time.Time{}
	}
	if !monitorDomain.UpdatedAt.IsZero() {
		return monitorDomain.UpdatedAt
	}
	return monitorDomain.CreatedAt
}
