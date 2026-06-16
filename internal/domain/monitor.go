package domain

import "time"

const (
	CollectionNameMonitorCertificate = "monitor_certificate"
	CollectionNameMonitorDomain      = "monitor_domain"
)

type MonitorCertificate struct {
	Meta
	MonitorDomain              string    `db:"monitorDomain"              json:"monitorDomain"`
	CertificateSubjectAltNames string    `db:"certificateSubjectAltNames" json:"certificateSubjectAltNames"`
	CertificateSerialNumber    string    `db:"certificateSerialNumber"    json:"certificateSerialNumber"`
	CertificateIssuerName      string    `db:"certificateIssuerName"      json:"certificateIssuerName"`
	CertificateNotBefore       time.Time `db:"certificateNotBefore"       json:"certificateNotBefore"`
	CertificateNotAfter        time.Time `db:"certificateNotAfter"        json:"certificateNotAfter"`
	WorkflowId                 string    `db:"workflowRef"                json:"workflowId"`
	WorkflowRunId              string    `db:"workflowRunRef"             json:"workflowRunId"`
	WorkflowNodeId             string    `db:"workflowNodeId"             json:"workflowNodeId"`
}

type MonitorDomain struct {
	Meta
	MonitorDomain  string    `db:"monitorDomain"  json:"monitorDomain"`
	DomainName     string    `db:"domainName"     json:"domainName"`
	DomainNotAfter time.Time `db:"domainNotAfter" json:"domainNotAfter"`
	WorkflowId     string    `db:"workflowRef"    json:"workflowId"`
	WorkflowRunId  string    `db:"workflowRunRef" json:"workflowRunId"`
	WorkflowNodeId string    `db:"workflowNodeId" json:"workflowNodeId"`
}
