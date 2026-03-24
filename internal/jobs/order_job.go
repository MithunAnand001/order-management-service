package jobs

import (
	"context"

	"order-management-service/internal/service"
	"order-management-service/internal/utils"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type OrderJob interface {
	Start()
	Stop()
}

type orderJob struct {
	cron   *cron.Cron
	svc    service.OrderService
	logger *zap.Logger
	sem    chan struct{} // Semaphore to prevent overlapping
}

func NewOrderJob(svc service.OrderService, logger *zap.Logger) OrderJob {
	return &orderJob{
		cron:   cron.New(),
		svc:    svc,
		logger: logger,
		sem:    make(chan struct{}, 1), // Capacity 1 ensures only one job runs at a time
	}
}

func (j *orderJob) Start() {
	_, err := j.cron.AddFunc("*/5 * * * *", func() {
		// Try to acquire semaphore, skip if busy
		select {
		case j.sem <- struct{}{}:
			// Successfully acquired. Release when done.
			defer func() { <-j.sem }()
		default:
			j.logger.Warn("Previous background job is still running, skipping this run to prevent overlap")
			return
		}

		// Create a background context with a unique request ID for the job run
		reqID := uuid.New().String()
		ctx := context.WithValue(context.Background(), utils.RequestIDKey, reqID)

		j.logger.Info("Running background job: PENDING -> PROCESSING", zap.String("request_id", reqID))
		if err := j.svc.ProcessPendingOrders(ctx); err != nil {
			j.logger.Error("Error in background job", zap.String("request_id", reqID), zap.Error(err.Err))
		}
	})

	if err != nil {
		j.logger.Fatal("Failed to schedule background job", zap.Error(err))
	}

	j.cron.Start()
	j.logger.Info("Background cron job started")
}

func (j *orderJob) Stop() {
	j.cron.Stop()
}
