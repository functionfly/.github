package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
)

type RevenueRecognitionService struct {
	repo *storage.RevenueRecognitionRepository
}

func NewRevenueRecognitionService(repo *storage.RevenueRecognitionRepository) *RevenueRecognitionService {
	return &RevenueRecognitionService{repo: repo}
}

type AllocationRequest struct {
	TenantID           uuid.UUID
	InvoiceID          uuid.UUID
	InvoiceAmountCents int
	Currency           string
	LineItems          []LineItem
}

type LineItem struct {
	Description          string
	AmountCents          int
	RevenueType          string
	SSPCents             int
	RecognitionMethod    string
	RecognitionStartDate time.Time
	RecognitionEndDate   *time.Time
	DeliveryPattern      string
}

func (s *RevenueRecognitionService) AllocateAndSchedule(ctx context.Context, req *AllocationRequest) error {
	if req.InvoiceAmountCents <= 0 {
		return fmt.Errorf("invoice amount must be positive")
	}

	var totalSSP int
	for _, item := range req.LineItems {
		totalSSP += item.SSPCents
	}

	if totalSSP == 0 {
		return fmt.Errorf("total SSP cannot be zero")
	}

	var obligations []*storage.PerformanceObligation

	for _, item := range req.LineItems {
		allocatedPrice := (req.InvoiceAmountCents * item.SSPCents) / totalSSP

		po := &storage.PerformanceObligation{
			ID:                    uuid.New(),
			TenantID:              req.TenantID,
			InvoiceID:             req.InvoiceID,
			Name:                  item.Description,
			Type:                  s.mapRevenueTypeToObligationType(item.RevenueType),
			TransactionPriceCents: req.InvoiceAmountCents,
			AllocatedPriceCents:   allocatedPrice,
			SSPCents:              item.SSPCents,
			SSPCurrency:           req.Currency,
			SSPBasis:              "total",
			RecognitionMethod:     item.RecognitionMethod,
			RecognitionStartDate:  item.RecognitionStartDate,
			RecognitionEndDate:    item.RecognitionEndDate,
			DeliveryPattern:       item.DeliveryPattern,
		}

		if err := s.repo.CreatePerformanceObligation(ctx, po); err != nil {
			return fmt.Errorf("failed to create performance obligation: %w", err)
		}
		obligations = append(obligations, po)
	}

	if err := s.createSchedules(ctx, req.InvoiceID, req.TenantID, req.Currency, obligations); err != nil {
		return fmt.Errorf("failed to create recognition schedules: %w", err)
	}

	if err := s.createContractAsset(ctx, req.InvoiceID, req.TenantID, req.Currency, obligations); err != nil {
		return fmt.Errorf("failed to create contract asset: %w", err)
	}

	return nil
}

func (s *RevenueRecognitionService) mapRevenueTypeToObligationType(revType string) string {
	switch revType {
	case "subscription":
		return "access"
	case "usage":
		return "usage"
	case "one_time":
		return "license"
	default:
		return "access"
	}
}

func (s *RevenueRecognitionService) createSchedules(ctx context.Context, invoiceID, tenantID uuid.UUID, currency string, obligations []*storage.PerformanceObligation) error {
	now := time.Now()

	for _, po := range obligations {
		var schedules []*storage.RevenueRecognitionSchedule

		switch po.RecognitionMethod {
		case storage.RecognitionMethodPointInTime:
			periodEnd := time.Now()
			if po.RecognitionEndDate != nil {
				periodEnd = *po.RecognitionEndDate
			}
			schedule := &storage.RevenueRecognitionSchedule{
				ID:                      uuid.New(),
				TenantID:                tenantID,
				InvoiceID:               invoiceID,
				PerformanceObligationID: po.ID,
				RecognitionMonth:        now.Format("2006-01"),
				PeriodStartDate:         po.RecognitionStartDate,
				PeriodEndDate:           periodEnd,
				AllocatedAmountCents:    po.AllocatedPriceCents,
				RecognizedAmountCents:   0,
				DeferredAmountCents:     po.AllocatedPriceCents,
				RevenueType:             s.mapObligationTypeToRevenueType(po.Type),
				OriginalTotalCents:      po.AllocatedPriceCents,
			}
			schedules = append(schedules, schedule)

		case storage.RecognitionMethodOverTime:
			schedules = s.createOverTimeSchedules(po, tenantID, invoiceID)
		}

		for _, schedule := range schedules {
			if err := s.repo.CreateRecognitionSchedule(ctx, schedule); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *RevenueRecognitionService) createOverTimeSchedules(po *storage.PerformanceObligation, tenantID, invoiceID uuid.UUID) []*storage.RevenueRecognitionSchedule {
	var schedules []*storage.RevenueRecognitionSchedule

	if po.RecognitionEndDate == nil {
		periodEnd := po.RecognitionStartDate.AddDate(1, 0, 0)
		po.RecognitionEndDate = &periodEnd
	}

	startDate := po.RecognitionStartDate
	totalMonths := s.monthsBetween(startDate, *po.RecognitionEndDate)

	if totalMonths <= 0 {
		totalMonths = 1
	}

	monthlyAmount := po.AllocatedPriceCents / totalMonths
	remainder := po.AllocatedPriceCents % totalMonths

	currentDate := startDate
	for i := 0; i < totalMonths; i++ {
		monthStr := currentDate.Format("2006-01")
		periodEndDate := currentDate.AddDate(0, 1, 0).AddDate(0, 0, -1)

		isLastMonth := (i == totalMonths-1)
		adjustedMonthlyAmount := monthlyAmount
		if isLastMonth {
			adjustedMonthlyAmount = po.AllocatedPriceCents - (monthlyAmount * (totalMonths - 1))
		}
		if isLastMonth && remainder > 0 {
			adjustedMonthlyAmount += remainder
		}

		schedule := &storage.RevenueRecognitionSchedule{
			ID:                      uuid.New(),
			TenantID:                tenantID,
			InvoiceID:               invoiceID,
			PerformanceObligationID: po.ID,
			RecognitionMonth:        monthStr,
			PeriodStartDate:         currentDate,
			PeriodEndDate:           periodEndDate,
			AllocatedAmountCents:    adjustedMonthlyAmount,
			RecognizedAmountCents:   0,
			DeferredAmountCents:     adjustedMonthlyAmount,
			RevenueType:             s.mapObligationTypeToRevenueType(po.Type),
			OriginalTotalCents:      po.AllocatedPriceCents,
		}
		schedules = append(schedules, schedule)

		currentDate = currentDate.AddDate(0, 1, 0)
	}

	return schedules
}

func (s *RevenueRecognitionService) monthsBetween(start, end time.Time) int {
	yearDiff := end.Year() - start.Year()
	monthDiff := int(end.Month()) - int(start.Month())
	return (yearDiff * 12) + monthDiff
}

func (s *RevenueRecognitionService) mapObligationTypeToRevenueType(obType string) string {
	switch obType {
	case "access":
		return "subscription"
	case "usage":
		return "usage"
	case "license":
		return "one_time"
	default:
		return "subscription"
	}
}

func (s *RevenueRecognitionService) createContractAsset(ctx context.Context, invoiceID, tenantID uuid.UUID, currency string, obligations []*storage.PerformanceObligation) error {
	totalDeferred := 0
	for _, po := range obligations {
		if po.RecognitionMethod == storage.RecognitionMethodPointInTime {
			totalDeferred += po.AllocatedPriceCents
		} else {
			totalDeferred += po.AllocatedPriceCents
		}
	}

	if totalDeferred > 0 {
		asset := &storage.ContractAsset{
			ID:              uuid.New(),
			TenantID:        tenantID,
			InvoiceID:       &invoiceID,
			CustomerID:      "",
			AssetType:       storage.ContractAssetType,
			AmountCents:     totalDeferred,
			Currency:        currency,
			Description:     "Deferred revenue from invoice",
			ReportingPeriod: time.Now().Format("2006-01"),
			Status:          "active",
		}
		return s.repo.CreateContractAsset(ctx, asset)
	}

	return nil
}

func (s *RevenueRecognitionService) ProcessRecognition(ctx context.Context, tenantID uuid.UUID, period string) (*storage.RevenueRecognitionReport, error) {
	schedules, err := s.repo.GetRecognitionSchedulesByPeriod(ctx, tenantID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get schedules: %w", err)
	}

	report := &storage.RevenueRecognitionReport{
		ReportID:        uuid.New(),
		GeneratedAt:     time.Now(),
		ReportingPeriod: period,
	}

	periodStart, _ := time.Parse("2006-01", period)
	report.PeriodStart = periodStart
	report.PeriodEnd = periodStart.AddDate(0, 1, 0).AddDate(0, 0, -1)

	previousPeriod := periodStart.AddDate(0, -1, 0).Format("2006-01")
	openingDef, err := s.getOpeningDeferred(ctx, tenantID, previousPeriod)
	if err == nil {
		report.OpeningDeferredCents = openingDef
	}

	for _, schedule := range schedules {
		if !schedule.IsRecognized && schedule.RecognitionMonth == period {
			report.NewDeferredCents += schedule.AllocatedAmountCents
			report.ClosingDeferredCents += schedule.DeferredAmountCents

			if schedule.RevenueType == "subscription" {
				report.TotalDeferredCents += schedule.DeferredAmountCents
			}
		}

		if schedule.IsRecognized {
			report.RecognizedFromDeferredCents += schedule.RecognizedAmountCents
			report.TotalRecognizedCents += schedule.RecognizedAmountCents

			report.OverTimeRevenueCents += schedule.RecognizedAmountCents

			if schedule.RevenueType == "subscription" {
				report.TotalRevenueCents += schedule.RecognizedAmountCents
			}
		}
	}

	if report.OpeningDeferredCents == 0 && report.NewDeferredCents > 0 {
		report.OpeningDeferredCents = report.NewDeferredCents
	}

	report.ClosingDeferredCents = report.OpeningDeferredCents + report.NewDeferredCents - report.RecognizedFromDeferredCents

	return report, nil
}

func (s *RevenueRecognitionService) getOpeningDeferred(ctx context.Context, tenantID uuid.UUID, period string) (int, error) {
	schedules, err := s.repo.GetRecognitionSchedulesByPeriod(ctx, tenantID, period)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, schedule := range schedules {
		if !schedule.IsRecognized {
			total += schedule.DeferredAmountCents
		}
	}

	return total, nil
}

func (s *RevenueRecognitionService) RecognizeRevenue(ctx context.Context, scheduleID uuid.UUID) error {
	schedule, err := s.getScheduleByID(ctx, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to get schedule: %w", err)
	}

	if schedule.IsRecognized {
		return fmt.Errorf("schedule already recognized")
	}

	if err := s.repo.MarkScheduleRecognized(ctx, scheduleID); err != nil {
		if strings.Contains(err.Error(), "already recognized") {
			return nil
		}
		return fmt.Errorf("failed to mark schedule recognized: %w", err)
	}

	event := &storage.RevenueRecognitionEvent{
		ID:                    uuid.New(),
		TenantID:              schedule.TenantID,
		InvoiceID:             schedule.InvoiceID,
		EventType:             "milestone_reached",
		RevenueType:           schedule.RevenueType,
		GrossAmountCents:      schedule.AllocatedAmountCents,
		DeferredAmountCents:   0,
		RecognizedAmountCents: schedule.AllocatedAmountCents,
		EventDate:             time.Now(),
		ReportingPeriod:       time.Now().Format("2006-01"),
		ScheduleID:            &scheduleID,
		Description:           fmt.Sprintf("Revenue recognized for period %s", schedule.RecognitionMonth),
	}

	return s.repo.CreateRecognitionEvent(ctx, event)
}

func (s *RevenueRecognitionService) getScheduleByID(ctx context.Context, id uuid.UUID) (*storage.RevenueRecognitionSchedule, error) {
	return s.repo.GetScheduleByID(ctx, id)
}

func (s *RevenueRecognitionService) GetDeferredRevenueSummary(ctx context.Context, tenantID uuid.UUID, period string) (*storage.DeferredRevenueSummary, error) {
	return s.repo.GetDeferredRevenueSummary(ctx, tenantID, period)
}

func (s *RevenueRecognitionService) GetRecognizedRevenueSummary(ctx context.Context, tenantID uuid.UUID, period string) (*storage.RecognizedRevenueSummary, error) {
	return s.repo.GetRecognizedRevenueSummary(ctx, tenantID, period)
}

func (s *RevenueRecognitionService) GetUnbilledRevenue(ctx context.Context, tenantID uuid.UUID) (int, error) {
	return s.repo.GetUnbilledRevenue(ctx, tenantID)
}

func (s *RevenueRecognitionService) GetPerformanceObligationIDsByInvoice(ctx context.Context, invoiceID uuid.UUID) ([]uuid.UUID, error) {
	obligations, err := s.repo.GetPerformanceObligationsByInvoice(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(obligations))
	for _, po := range obligations {
		ids = append(ids, po.ID)
	}
	return ids, nil
}

func (s *RevenueRecognitionService) RecordInvoicePaid(ctx context.Context, invoiceID, tenantID uuid.UUID, amountCents int, currency string) error {
	event := &storage.RevenueRecognitionEvent{
		ID:                    uuid.New(),
		TenantID:              tenantID,
		InvoiceID:             invoiceID,
		EventType:             "invoice_paid",
		RevenueType:           "subscription",
		GrossAmountCents:      amountCents,
		DeferredAmountCents:   amountCents,
		RecognizedAmountCents: 0,
		EventDate:             time.Now(),
		ReportingPeriod:       time.Now().Format("2006-01"),
		Description:           fmt.Sprintf("Invoice paid: %d %s", amountCents, currency),
	}

	return s.repo.CreateRecognitionEvent(ctx, event)
}

func (s *RevenueRecognitionService) RecordDeliveryCompleted(ctx context.Context, poID uuid.UUID, tenantID uuid.UUID) error {
	if err := s.repo.UpdatePerformanceObligationDeliveryStatus(ctx, poID, true); err != nil {
		return fmt.Errorf("failed to update delivery status: %w", err)
	}

	event := &storage.RevenueRecognitionEvent{
		ID:                      uuid.New(),
		TenantID:                tenantID,
		InvoiceID:               uuid.Nil,
		EventType:               "delivery_completed",
		RevenueType:             "subscription",
		GrossAmountCents:        0,
		DeferredAmountCents:     0,
		RecognizedAmountCents:   0,
		EventDate:               time.Now(),
		ReportingPeriod:         time.Now().Format("2006-01"),
		PerformanceObligationID: &poID,
		Description:             "Performance obligation delivery completed",
	}

	return s.repo.CreateRecognitionEvent(ctx, event)
}

func (s *RevenueRecognitionService) ProcessAllPendingRecognition(ctx context.Context, period string) error {
	schedules, err := s.repo.GetAllUnrecognizedSchedules(ctx, period)
	if err != nil {
		return fmt.Errorf("failed to get schedules: %w", err)
	}

	for _, schedule := range schedules {
		if err := s.RecognizeRevenue(ctx, schedule.ID); err != nil {
			continue
		}
	}

	return nil
}
