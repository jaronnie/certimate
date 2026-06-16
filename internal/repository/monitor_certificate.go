package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/certimate-go/certimate/internal/app"
	"github.com/certimate-go/certimate/internal/domain"
)

type MonitorCertificateRepository struct{}

func NewMonitorCertificateRepository() *MonitorCertificateRepository {
	return &MonitorCertificateRepository{}
}

func (r *MonitorCertificateRepository) List(ctx context.Context) ([]*domain.MonitorCertificate, error) {
	records, err := app.GetApp().FindAllRecords(domain.CollectionNameMonitorCertificate)
	if err != nil {
		return nil, err
	}

	monitorCertificates := make([]*domain.MonitorCertificate, 0, len(records))
	for _, record := range records {
		monitorCertificate, err := r.castRecordToModel(record)
		if err != nil {
			return nil, err
		}
		monitorCertificates = append(monitorCertificates, monitorCertificate)
	}

	return monitorCertificates, nil
}

func (r *MonitorCertificateRepository) GetByMonitorDomain(ctx context.Context, monitorDomain string) (*domain.MonitorCertificate, error) {
	records, err := app.GetApp().FindRecordsByFilter(
		domain.CollectionNameMonitorCertificate,
		"monitorDomain={:monitorDomain}",
		"",
		1, 0,
		dbx.Params{"monitorDomain": monitorDomain},
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrRecordNotFound
		}
		return nil, err
	}
	if len(records) == 0 {
		return nil, domain.ErrRecordNotFound
	}

	return r.castRecordToModel(records[0])
}

func (r *MonitorCertificateRepository) Save(ctx context.Context, monitorCertificate *domain.MonitorCertificate) (*domain.MonitorCertificate, error) {
	collection, err := app.GetApp().FindCollectionByNameOrId(domain.CollectionNameMonitorCertificate)
	if err != nil {
		return monitorCertificate, err
	}

	var record *core.Record
	if monitorCertificate.Id == "" {
		record = core.NewRecord(collection)
	} else {
		record, err = app.GetApp().FindRecordById(collection, monitorCertificate.Id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return monitorCertificate, domain.ErrRecordNotFound
			}
			return monitorCertificate, err
		}
	}

	record.Set("monitorDomain", monitorCertificate.MonitorDomain)
	record.Set("certificateSubjectAltNames", monitorCertificate.CertificateSubjectAltNames)
	record.Set("certificateSerialNumber", monitorCertificate.CertificateSerialNumber)
	record.Set("certificateIssuerName", monitorCertificate.CertificateIssuerName)
	record.Set("certificateNotBefore", monitorCertificate.CertificateNotBefore)
	record.Set("certificateNotAfter", monitorCertificate.CertificateNotAfter)
	record.Set("workflowRef", monitorCertificate.WorkflowId)
	record.Set("workflowRunRef", monitorCertificate.WorkflowRunId)
	record.Set("workflowNodeId", monitorCertificate.WorkflowNodeId)
	if err := app.GetApp().Save(record); err != nil {
		return monitorCertificate, err
	}

	monitorCertificate.Id = record.Id
	monitorCertificate.CreatedAt = record.GetDateTime("created").Time()
	monitorCertificate.UpdatedAt = record.GetDateTime("updated").Time()
	return monitorCertificate, nil
}

func (r *MonitorCertificateRepository) castRecordToModel(record *core.Record) (*domain.MonitorCertificate, error) {
	if record == nil {
		return nil, fmt.Errorf("the record is nil")
	}

	return &domain.MonitorCertificate{
		Meta: domain.Meta{
			Id:        record.Id,
			CreatedAt: record.GetDateTime("created").Time(),
			UpdatedAt: record.GetDateTime("updated").Time(),
		},
		MonitorDomain:              record.GetString("monitorDomain"),
		CertificateSubjectAltNames: record.GetString("certificateSubjectAltNames"),
		CertificateSerialNumber:    record.GetString("certificateSerialNumber"),
		CertificateIssuerName:      record.GetString("certificateIssuerName"),
		CertificateNotBefore:       record.GetDateTime("certificateNotBefore").Time(),
		CertificateNotAfter:        record.GetDateTime("certificateNotAfter").Time(),
		WorkflowId:                 record.GetString("workflowRef"),
		WorkflowRunId:              record.GetString("workflowRunRef"),
		WorkflowNodeId:             record.GetString("workflowNodeId"),
	}, nil
}
