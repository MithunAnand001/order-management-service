package jobs

import (
	"context"
	"log"

	"order-management-service/internal/service"

	"github.com/robfig/cron/v3"
)

type OrderJob interface {
	Start()
	Stop()
}

type orderJob struct {
	cron *cron.Cron
	svc  service.OrderService
}

func NewOrderJob(svc service.OrderService) OrderJob {
	return &orderJob{
		cron: cron.New(),
		svc:  svc,
	}
}

func (j *orderJob) Start() {
	_, err := j.cron.AddFunc("*/5 * * * *", func() {
		log.Println("Running background job: PENDING -> PROCESSING")
		if err := j.svc.ProcessPendingOrders(context.Background()); err != nil {
			log.Printf("Error in background job: %v", err)
		}
	})

	if err != nil {
		log.Fatalf("Failed to schedule background job: %v", err)
	}

	j.cron.Start()
	log.Println("Background cron job started")
}

func (j *orderJob) Stop() {
	j.cron.Stop()
}
