package wallet

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CreditTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wallet_credit_total",
		Help: "Total credit operations",
	}, []string{"status", "wallet_type"})

	DebitTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wallet_debit_total",
		Help: "Total debit operations",
	}, []string{"status", "wallet_type"})

	BalanceGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "wallet_balance_usd",
		Help: "Current wallet balance in USD",
	}, []string{"wallet_id", "wallet_type", "owner_type"})

	TransactionDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "wallet_transaction_duration_seconds",
		Help:    "Duration of wallet transactions",
		Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
	})

	BalanceDriftDetected = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wallet_balance_drift_total",
		Help: "Total balance drift incidents detected",
	})

	CacheInvalidationRetries = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wallet_cache_invalidation_retries_total",
		Help: "Total cache invalidation retry attempts",
	})

	CacheInvalidationFailures = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wallet_cache_invalidation_failures_total",
		Help: "Total cache invalidation failures after max retries",
	})

	DistributedLockAcquisitions = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wallet_distributed_lock_acquisitions_total",
		Help: "Total distributed lock acquisition attempts",
	}, []string{"status"})

	SpendCapChecks = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "wallet_spend_cap_checks_total",
		Help: "Total spend cap check results",
	}, []string{"allowed", "fail_mode"})

	LowBalanceAlerts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wallet_low_balance_alerts_total",
		Help: "Total low balance alerts sent",
	})
)