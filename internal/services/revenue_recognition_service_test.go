package services

import (
	"context"
	"testing"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOverTimeSchedules_RemainderDistribution(t *testing.T) {
	svc := &RevenueRecognitionService{}

	tests := []struct {
		name              string
		allocatedCents    int
		totalMonths       int
		expectedRemainder int
	}{
		{
			name:              "evenly divisible - no remainder",
			allocatedCents:    1200,
			totalMonths:       12,
			expectedRemainder: 0,
		},
		{
			name:              "remainder of 1 cent",
			allocatedCents:    1001,
			totalMonths:       3,
			expectedRemainder: 1,
		},
		{
			name:              "remainder distributed to last month",
			allocatedCents:    100,
			totalMonths:       3,
			expectedRemainder: 1,
		},
		{
			name:              "large remainder",
			allocatedCents:    10000,
			totalMonths:       7,
			expectedRemainder: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := &storage.PerformanceObligation{
				ID:                   uuid.New(),
				AllocatedPriceCents:  tt.allocatedCents,
				RecognitionStartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				RecognitionEndDate:   timePtr(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)),
				Type:                 "access",
			}

			schedules := svc.createOverTimeSchedules(po, uuid.New(), uuid.New())

			assert.Equal(t, tt.totalMonths, len(schedules), "should create correct number of schedules")

			remainder := tt.allocatedCents % tt.totalMonths
			assert.Equal(t, tt.expectedRemainder, remainder, "remainder calculation should match")

			totalFromSchedules := 0
			for i, schedule := range schedules {
				totalFromSchedules += schedule.AllocatedAmountCents

				if i == tt.totalMonths-1 {
					expectedLastMonth := (tt.allocatedCents / tt.totalMonths)
					if remainder > 0 {
						expectedLastMonth += remainder
					}
					assert.Equal(t, expectedLastMonth, schedule.AllocatedAmountCents,
						"last month should receive remainder")
				}
			}

			assert.Equal(t, tt.allocatedCents, totalFromSchedules,
				"sum of all schedule amounts should equal allocated amount")
		})
	}
}

func TestCreateOverTimeSchedules_MonthsCalculation(t *testing.T) {
	svc := &RevenueRecognitionService{}

	tests := []struct {
		name            string
		startDate       time.Time
		endDate        time.Time
		expectedMonths int
	}{
		{
			name:            "one month",
			startDate:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
			expectedMonths: 1,
		},
		{
			name:            "three months",
			startDate:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC),
			expectedMonths: 3,
		},
		{
			name:            "one year",
			startDate:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			expectedMonths: 12,
		},
		{
			name:            "cross year boundary",
			startDate:       time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
			endDate:        time.Date(2027, 5, 31, 0, 0, 0, 0, time.UTC),
			expectedMonths: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			po := &storage.PerformanceObligation{
				ID:                   uuid.New(),
				AllocatedPriceCents:  1000,
				RecognitionStartDate: tt.startDate,
				RecognitionEndDate:   timePtr(tt.endDate),
				Type:                 "access",
			}

			schedules := svc.createOverTimeSchedules(po, uuid.New(), uuid.New())

			assert.Equal(t, tt.expectedMonths, len(schedules), "should create correct number of monthly schedules")
		})
	}
}

func TestAllocateAndSchedule_Validation(t *testing.T) {
	svc := &RevenueRecognitionService{
		repo:      nil,
		txManager: nil,
	}

	tests := []struct {
		name        string
		req         *AllocationRequest
		expectError bool
		errorMsg   string
	}{
		{
			name: "zero invoice amount",
			req: &AllocationRequest{
				InvoiceAmountCents: 0,
				LineItems: []LineItem{
					{SSPCents: 100},
				},
			},
			expectError: true,
			errorMsg:   "invoice amount must be positive",
		},
		{
			name: "negative invoice amount",
			req: &AllocationRequest{
				InvoiceAmountCents: -100,
				LineItems: []LineItem{
					{SSPCents: 100},
				},
			},
			expectError: true,
			errorMsg:   "invoice amount must be positive",
		},
		{
			name: "zero total SSP",
			req: &AllocationRequest{
				InvoiceAmountCents: 1000,
				LineItems: []LineItem{
					{SSPCents: 0},
				},
			},
			expectError: true,
			errorMsg:   "total SSP cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.AllocateAndSchedule(context.Background(), tt.req)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMonthsBetween(t *testing.T) {
	svc := &RevenueRecognitionService{}

	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{
			name:     "same month",
			start:    time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 1, 28, 0, 0, 0, 0, time.UTC),
			expected: 0,
		},
		{
			name:     "one month apart",
			start:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "three months",
			start:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			expected: 3,
		},
		{
			name:     "cross year boundary",
			start:    time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC),
			expected: 3,
		},
		{
			name:     "one year",
			start:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: 12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.monthsBetween(tt.start, tt.end)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapRevenueTypeToObligationType(t *testing.T) {
	svc := &RevenueRecognitionService{}

	tests := []struct {
		revType   string
		expected  string
	}{
		{"subscription", "access"},
		{"usage", "usage"},
		{"one_time", "license"},
		{"unknown", "access"},
	}

	for _, tt := range tests {
		t.Run(tt.revType, func(t *testing.T) {
			result := svc.mapRevenueTypeToObligationType(tt.revType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMapObligationTypeToRevenueType(t *testing.T) {
	svc := &RevenueRecognitionService{}

	tests := []struct {
		obType   string
		expected string
	}{
		{"access", "subscription"},
		{"usage", "usage"},
		{"license", "one_time"},
		{"unknown", "subscription"},
	}

	for _, tt := range tests {
		t.Run(tt.obType, func(t *testing.T) {
			result := svc.mapObligationTypeToRevenueType(tt.obType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
