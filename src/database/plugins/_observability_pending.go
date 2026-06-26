package plugins

import (
	"context"
	"fmt"
	"time"

	"github.com/Parallels/prl-devops-service/basecontext"
	"github.com/Parallels/prl-devops-service/tracing/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

var (
	queryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_query_duration_seconds",
		Help:    "Time spent executing database queries",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation", "table", "status"})

	queryErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "db_query_errors_total",
		Help: "Total number of database query errors",
	}, []string{"operation", "table", "status"})
)

// ObservabilityPlugin is a GORM plugin that adds tracing and metrics
type ObservabilityPlugin struct{}

func NewObservabilityPlugin() *ObservabilityPlugin {
	return &ObservabilityPlugin{}
}

func (p *ObservabilityPlugin) Name() string {
	return "observability"
}

func (p *ObservabilityPlugin) Initialize(db *gorm.DB) error {
	// Register callbacks for Create
	if err := db.Callback().Create().Before("gorm:create").Register("observability:before_create", p.before("create")); err != nil {
		return err
	}
	if err := db.Callback().Create().After("gorm:create").Register("observability:after_create", p.after("create")); err != nil {
		return err
	}

	// Register callbacks for Query
	if err := db.Callback().Query().Before("gorm:query").Register("observability:before_query", p.before("query")); err != nil {
		return err
	}
	if err := db.Callback().Query().After("gorm:query").Register("observability:after_query", p.after("query")); err != nil {
		return err
	}

	// Register callbacks for Update
	if err := db.Callback().Update().Before("gorm:update").Register("observability:before_update", p.before("update")); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").Register("observability:after_update", p.after("update")); err != nil {
		return err
	}

	// Register callbacks for Delete
	if err := db.Callback().Delete().Before("gorm:delete").Register("observability:before_delete", p.before("delete")); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gorm:delete").Register("observability:after_delete", p.after("delete")); err != nil {
		return err
	}

	// Register callbacks for Row
	if err := db.Callback().Row().Before("gorm:row").Register("observability:before_row", p.before("row")); err != nil {
		return err
	}
	if err := db.Callback().Row().After("gorm:row").Register("observability:after_row", p.after("row")); err != nil {
		return err
	}

	// Register callbacks for Raw
	if err := db.Callback().Raw().Before("gorm:raw").Register("observability:before_raw", p.before("raw")); err != nil {
		return err
	}
	if err := db.Callback().Raw().After("gorm:raw").Register("observability:after_raw", p.after("raw")); err != nil {
		return err
	}

	return nil
}

func (p *ObservabilityPlugin) before(op string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		ctx := appctx.FromContext(db.Statement.Context)
		db.Statement.Context = context.WithValue(db.Statement.Context, "db_start_time", time.Now())

		// Start Tracing Span
		tracer := service.GetInstance()
		if tracer != nil {
			spanName := fmt.Sprintf("db.%s", op)
			if db.Statement.Table != "" {
				spanName = fmt.Sprintf("db.%s.%s", db.Statement.Table, op)
			}
			newCtx, span := tracer.StartSpan(ctx, "database", spanName)
			db.Statement.Context = context.WithValue(newCtx, "db_span", span)
		}
	}
}

func (p *ObservabilityPlugin) after(op string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		startTime, ok := db.Statement.Context.Value("db_start_time").(time.Time)
		if !ok {
			return
		}
		duration := time.Since(startTime).Seconds()

		tableName := db.Statement.Table
		if tableName == "" && op == "raw" {
			tableName = "raw"
		}

		status := "success"
		if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
			status = "error"
		}

		// Record Metrics
		queryDuration.WithLabelValues(op, tableName, status).Observe(duration)
		if status == "error" {
			queryErrors.WithLabelValues(op, tableName, status).Inc()
		}

		// End Tracing Span
		span, ok := db.Statement.Context.Value("db_span").(trace.Span)
		if ok && span != nil {
			if status == "error" {
				span.RecordError(db.Error)
				span.SetStatus(codes.Error, db.Error.Error())
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
			span.End()
		}
	}
}
