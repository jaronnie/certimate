package engine

import (
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/certimate-go/certimate/internal/domain"
	"github.com/certimate-go/certimate/internal/domainexp"
	"github.com/certimate-go/certimate/internal/repository"
)

type bizMonitorDomainNodeExecutor struct {
	nodeExecutor

	monitorDomainRepo monitorDomainRepository
}

func (ne *bizMonitorDomainNodeExecutor) Execute(execCtx *NodeExecutionContext) (*NodeExecutionResult, error) {
	execRes := newNodeExecutionResult(execCtx.Node)

	nodeCfg := execCtx.Node.Data.Config.AsBizMonitorDomain()
	ne.logger.Info("ready to monitor domain ...", slog.Any("config", nodeCfg))

	monitorDomain := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(nodeCfg.Domain)), ".")
	if monitorDomain == "" {
		return execRes, fmt.Errorf("domain is required")
	}

	domainExpiry, err := domainexp.Lookup(execCtx.Context(), monitorDomain)
	if err != nil {
		return execRes, fmt.Errorf("failed to retrieve domain expiration: %w", err)
	}

	record := &domain.MonitorDomain{
		MonitorDomain:  monitorDomain,
		DomainName:     domainExpiry.Domain,
		DomainNotAfter: domainExpiry.NotAfter,
		WorkflowId:     execCtx.WorkflowId,
		WorkflowRunId:  execCtx.RunId,
		WorkflowNodeId: execCtx.Node.Id,
	}

	lastRecord, err := ne.monitorDomainRepo.GetByDomainName(execCtx.Context(), domainExpiry.Domain)
	if err != nil && !domain.IsRecordNotFoundError(err) {
		return execRes, err
	}
	if lastRecord != nil {
		record.Id = lastRecord.Id
	}

	if record, err = ne.monitorDomainRepo.Save(execCtx.Context(), record); err != nil {
		return execRes, err
	}

	ne.setVariablesOfResult(execCtx, execRes, record)

	ne.logger.Info(
		"monitor domain saved",
		slog.String("recordId", record.Id),
		slog.String("monitorDomain", record.MonitorDomain),
		slog.String("domain", record.DomainName),
	)
	ne.logger.Info("monitoring completed")
	return execRes, nil
}

func (ne *bizMonitorDomainNodeExecutor) setVariablesOfResult(execCtx *NodeExecutionContext, execRes *NodeExecutionResult, monitorDomain *domain.MonitorDomain) {
	var vDomainName string
	var vNotAfter time.Time
	var vHoursLeft int32
	var vDaysLeft int32
	var vValidity bool

	if monitorDomain != nil {
		vDomainName = monitorDomain.DomainName
		vNotAfter = monitorDomain.DomainNotAfter
		vHoursLeft = int32(math.Floor(time.Until(monitorDomain.DomainNotAfter).Hours()))
		vDaysLeft = int32(math.Floor(time.Until(monitorDomain.DomainNotAfter).Hours() / 24))
		vValidity = monitorDomain.DomainNotAfter.After(time.Now())
	}

	execRes.AddVariable(stateVarKeyDomainName, vDomainName, stateValTypeString)
	execRes.AddVariable(stateVarKeyDomainNotAfter, vNotAfter, stateValTypeDateTime)
	execRes.AddVariable(stateVarKeyDomainHoursLeft, vHoursLeft, stateValTypeNumber)
	execRes.AddVariable(stateVarKeyDomainDaysLeft, vDaysLeft, stateValTypeNumber)
	execRes.AddVariable(stateVarKeyDomainValidity, vValidity, stateValTypeBoolean)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyDomainName, vDomainName, stateValTypeString)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyDomainNotAfter, vNotAfter, stateValTypeDateTime)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyDomainHoursLeft, vHoursLeft, stateValTypeNumber)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyDomainDaysLeft, vDaysLeft, stateValTypeNumber)
	execRes.AddVariableWithScope(execCtx.Node.Id, stateVarKeyDomainValidity, vValidity, stateValTypeBoolean)
}

func newBizMonitorDomainNodeExecutor() NodeExecutor {
	return &bizMonitorDomainNodeExecutor{
		nodeExecutor:      nodeExecutor{logger: slog.Default()},
		monitorDomainRepo: repository.NewMonitorDomainRepository(),
	}
}
