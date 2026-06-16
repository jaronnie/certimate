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

type MonitorDomainRepository struct{}

func NewMonitorDomainRepository() *MonitorDomainRepository {
	return &MonitorDomainRepository{}
}

func (r *MonitorDomainRepository) List(ctx context.Context) ([]*domain.MonitorDomain, error) {
	records, err := app.GetApp().FindAllRecords(domain.CollectionNameMonitorDomain)
	if err != nil {
		return nil, err
	}

	monitorDomains := make([]*domain.MonitorDomain, 0, len(records))
	for _, record := range records {
		monitorDomain, err := r.castRecordToModel(record)
		if err != nil {
			return nil, err
		}
		monitorDomains = append(monitorDomains, monitorDomain)
	}

	return monitorDomains, nil
}

func (r *MonitorDomainRepository) GetByDomainName(ctx context.Context, domainName string) (*domain.MonitorDomain, error) {
	records, err := app.GetApp().FindRecordsByFilter(
		domain.CollectionNameMonitorDomain,
		"domainName={:domainName}",
		"",
		1, 0,
		dbx.Params{"domainName": domainName},
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

func (r *MonitorDomainRepository) Save(ctx context.Context, monitorDomain *domain.MonitorDomain) (*domain.MonitorDomain, error) {
	collection, err := app.GetApp().FindCollectionByNameOrId(domain.CollectionNameMonitorDomain)
	if err != nil {
		return monitorDomain, err
	}

	var record *core.Record
	if monitorDomain.Id == "" {
		record = core.NewRecord(collection)
	} else {
		record, err = app.GetApp().FindRecordById(collection, monitorDomain.Id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return monitorDomain, domain.ErrRecordNotFound
			}
			return monitorDomain, err
		}
	}

	record.Set("monitorDomain", monitorDomain.MonitorDomain)
	record.Set("domainName", monitorDomain.DomainName)
	record.Set("domainNotAfter", monitorDomain.DomainNotAfter)
	record.Set("workflowRef", monitorDomain.WorkflowId)
	record.Set("workflowRunRef", monitorDomain.WorkflowRunId)
	record.Set("workflowNodeId", monitorDomain.WorkflowNodeId)
	if err := app.GetApp().Save(record); err != nil {
		return monitorDomain, err
	}

	monitorDomain.Id = record.Id
	monitorDomain.CreatedAt = record.GetDateTime("created").Time()
	monitorDomain.UpdatedAt = record.GetDateTime("updated").Time()
	return monitorDomain, nil
}

func (r *MonitorDomainRepository) castRecordToModel(record *core.Record) (*domain.MonitorDomain, error) {
	if record == nil {
		return nil, fmt.Errorf("the record is nil")
	}

	return &domain.MonitorDomain{
		Meta: domain.Meta{
			Id:        record.Id,
			CreatedAt: record.GetDateTime("created").Time(),
			UpdatedAt: record.GetDateTime("updated").Time(),
		},
		MonitorDomain:  record.GetString("monitorDomain"),
		DomainName:     record.GetString("domainName"),
		DomainNotAfter: record.GetDateTime("domainNotAfter").Time(),
		WorkflowId:     record.GetString("workflowRef"),
		WorkflowRunId:  record.GetString("workflowRunRef"),
		WorkflowNodeId: record.GetString("workflowNodeId"),
	}, nil
}
