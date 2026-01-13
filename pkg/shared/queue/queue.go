package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"hedge-fund/pkg/shared/logger"
	"hedge-fund/pkg/shared/models"
	"hedge-fund/pkg/shared/redis"
)

type Manager struct {
	redis  *redis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

// NewManager creates a new queue manager
func NewManager(redisClient *redis.Client) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	return &Manager{
		redis:  redisClient,
		ctx:    ctx,
		cancel: cancel,
	}
}

// EnqueueJob adds a job to the appropriate queue
func (m *Manager) EnqueueJob(job *models.Job) error {
	// Generate ID if not provided
	if job.ID == "" {
		job.ID = uuid.New().String()
	}

	// Set created time
	job.CreatedAt = time.Now()

	// Determine queue based on job type
	queue := m.getQueueForJobType(job.Type)

	if err := m.redis.EnqueueJob(m.ctx, queue, job); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	logger.Info("Job enqueued successfully",
		zap.String("job_id", job.ID),
		zap.String("job_type", job.Type),
		zap.String("queue", queue))

	return nil
}

// EnqueueAIAnalysis enqueues an AI analysis job
func (m *Manager) EnqueueAIAnalysis(symbol string, agents []string, userID int) (string, error) {
	job := &models.AIAnalysisJob{
		Job: models.Job{
			ID:         uuid.New().String(),
			Type:       models.JobTypeAIAnalysis,
			Priority:   5,
			MaxRetries: 3,
			Payload: map[string]interface{}{
				"symbol":  symbol,
				"agents":  agents,
				"user_id": userID,
			},
		},
		Symbol:    symbol,
		Agents:    agents,
		UserID:    userID,
		RequestID: uuid.New().String(),
	}

	if err := m.EnqueueJob(&job.Job); err != nil {
		return "", err
	}

	return job.RequestID, nil
}

// EnqueueMarketDataUpdate enqueues a market data update job
func (m *Manager) EnqueueMarketDataUpdate(symbols []string, dataType string, immediate bool) (string, error) {
	priority := 3
	if immediate {
		priority = 8 // Higher priority for immediate updates
	}

	job := &models.MarketDataUpdateJob{
		Job: models.Job{
			ID:         uuid.New().String(),
			Type:       models.JobTypeMarketDataUpdate,
			Priority:   priority,
			MaxRetries: 5,
			Payload: map[string]interface{}{
				"symbols":   symbols,
				"data_type": dataType,
				"immediate": immediate,
			},
		},
		Symbols:   symbols,
		DataType:  dataType,
		Immediate: immediate,
	}

	if err := m.EnqueueJob(&job.Job); err != nil {
		return "", err
	}

	return job.ID, nil
}

// EnqueueRiskCalculation enqueues a risk calculation job
func (m *Manager) EnqueueRiskCalculation(userID, portfolioID int, symbols []string, riskType string) (string, error) {
	job := &models.RiskCalculationJob{
		Job: models.Job{
			ID:         uuid.New().String(),
			Type:       models.JobTypeRiskCalculation,
			Priority:   7,
			MaxRetries: 3,
			Payload: map[string]interface{}{
				"user_id":      userID,
				"portfolio_id": portfolioID,
				"symbols":      symbols,
				"risk_type":    riskType,
			},
		},
		UserID:      userID,
		PortfolioID: portfolioID,
		Symbols:     symbols,
		RiskType:    riskType,
	}

	if err := m.EnqueueJob(&job.Job); err != nil {
		return "", err
	}

	return job.ID, nil
}

// DequeueJob gets the next job from a specific queue
func (m *Manager) DequeueJob(queue string, timeout time.Duration) (*models.Job, error) {
	var job models.Job
	if err := m.redis.DequeueJob(m.ctx, queue, timeout, &job); err != nil {
		return nil, err
	}

	logger.Info("Job dequeued successfully",
		zap.String("job_id", job.ID),
		zap.String("job_type", job.Type),
		zap.String("queue", queue))

	return &job, nil
}

// SetJobStatus updates the status of a job
func (m *Manager) SetJobStatus(jobID, status string, message string, progress float64) error {
	statusKey := fmt.Sprintf("job_status:%s", jobID)

	jobStatus := models.JobStatus{
		JobID:    jobID,
		Status:   status,
		Progress: progress,
		Message:  message,
	}

	now := time.Now()
	if status == models.JobStatusRunning && progress == 0 {
		jobStatus.StartedAt = &now
	} else if status == models.JobStatusCompleted || status == models.JobStatusFailed {
		jobStatus.CompletedAt = &now
	}

	// Store status with 24-hour expiration
	if err := m.redis.SetCache(m.ctx, statusKey, jobStatus, 24*time.Hour); err != nil {
		return fmt.Errorf("failed to set job status: %w", err)
	}

	// Publish status update event
	event := models.Event{
		Type:      "job_status_updated",
		Source:    "queue_manager",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"job_id":   jobID,
			"status":   status,
			"progress": progress,
			"message":  message,
		},
	}

	if err := m.redis.PublishEvent(m.ctx, models.ChannelSystemEvents, event); err != nil {
		logger.Warn("Failed to publish job status event", zap.Error(err))
	}

	return nil
}

// GetJobStatus retrieves the status of a job
func (m *Manager) GetJobStatus(jobID string) (*models.JobStatus, error) {
	statusKey := fmt.Sprintf("job_status:%s", jobID)

	var status models.JobStatus
	if err := m.redis.GetCache(m.ctx, statusKey, &status); err != nil {
		return nil, fmt.Errorf("job status not found: %s", jobID)
	}

	return &status, nil
}

// GetQueueLength returns the number of jobs in a queue
func (m *Manager) GetQueueLength(queue string) (int64, error) {
	return m.redis.QueueLength(m.ctx, queue)
}

// GetAllQueueLengths returns the length of all queues
func (m *Manager) GetAllQueueLengths() (map[string]int64, error) {
	queues := []string{
		models.QueueAIAnalysis,
		models.QueueRiskCalc,
		models.QueueNotifications,
		models.QueueMarketData,
		models.QueueReports,
		models.QueueCleanup,
		models.QueueMaintenance,
	}

	lengths := make(map[string]int64)
	for _, queue := range queues {
		length, err := m.GetQueueLength(queue)
		if err != nil {
			logger.Warn("Failed to get queue length",
				zap.String("queue", queue),
				zap.Error(err))
			continue
		}
		lengths[queue] = length
	}

	return lengths, nil
}

// Worker represents a job worker
type Worker struct {
	manager   *Manager
	queue     string
	handler   JobHandler
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
}

// JobHandler defines the interface for handling jobs
type JobHandler interface {
	Handle(ctx context.Context, job *models.Job) error
	CanHandle(jobType string) bool
}

// NewWorker creates a new job worker
func (m *Manager) NewWorker(queue string, handler JobHandler) *Worker {
	ctx, cancel := context.WithCancel(m.ctx)
	return &Worker{
		manager: m,
		queue:   queue,
		handler: handler,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start starts the worker
func (w *Worker) Start() error {
	if w.isRunning {
		return fmt.Errorf("worker is already running")
	}

	w.isRunning = true
	logger.Info("Starting job worker", zap.String("queue", w.queue))

	go w.run()
	return nil
}

// Stop stops the worker
func (w *Worker) Stop() {
	if !w.isRunning {
		return
	}

	logger.Info("Stopping job worker", zap.String("queue", w.queue))
	w.cancel()
	w.isRunning = false
}

// run is the main worker loop
func (w *Worker) run() {
	defer func() {
		w.isRunning = false
		logger.Info("Job worker stopped", zap.String("queue", w.queue))
	}()

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			// Try to get a job with a timeout
			job, err := w.manager.DequeueJob(w.queue, 5*time.Second)
			if err != nil {
				// Timeout is expected, continue
				continue
			}

			// Check if handler can process this job type
			if !w.handler.CanHandle(job.Type) {
				logger.Warn("Handler cannot process job type",
					zap.String("job_type", job.Type),
					zap.String("job_id", job.ID))
				continue
			}

			// Process the job
			w.processJob(job)
		}
	}
}

// processJob processes a single job
func (w *Worker) processJob(job *models.Job) {
	logger.Info("Processing job",
		zap.String("job_id", job.ID),
		zap.String("job_type", job.Type))

	// Update status to running
	w.manager.SetJobStatus(job.ID, models.JobStatusRunning, "Processing job", 0)

	// Create job context with timeout
	ctx, cancel := context.WithTimeout(w.ctx, 10*time.Minute)
	defer cancel()

	// Handle the job
	err := w.handler.Handle(ctx, job)
	if err != nil {
		logger.Error("Job processing failed",
			zap.String("job_id", job.ID),
			zap.Error(err))

		// Check if we should retry
		if job.Retries < job.MaxRetries {
			job.Retries++
			w.manager.SetJobStatus(job.ID, models.JobStatusRetrying,
				fmt.Sprintf("Retrying job (attempt %d/%d)", job.Retries, job.MaxRetries), 0)

			// Re-enqueue with exponential backoff
			go func() {
				backoff := time.Duration(job.Retries) * time.Minute
				time.Sleep(backoff)
				w.manager.EnqueueJob(job)
			}()
		} else {
			w.manager.SetJobStatus(job.ID, models.JobStatusFailed,
				fmt.Sprintf("Job failed after %d retries: %v", job.MaxRetries, err), 100)
		}
		return
	}

	// Mark as completed
	w.manager.SetJobStatus(job.ID, models.JobStatusCompleted, "Job completed successfully", 100)
	logger.Info("Job completed successfully", zap.String("job_id", job.ID))
}

// getQueueForJobType returns the appropriate queue for a job type
func (m *Manager) getQueueForJobType(jobType string) string {
	switch jobType {
	case models.JobTypeAIAnalysis:
		return models.QueueAIAnalysis
	case models.JobTypeRiskCalculation:
		return models.QueueRiskCalc
	case models.JobTypeNotification:
		return models.QueueNotifications
	case models.JobTypeMarketDataUpdate:
		return models.QueueMarketData
	case models.JobTypeReportGeneration:
		return models.QueueReports
	case models.JobTypeCleanup:
		return models.QueueCleanup
	default:
		return models.QueueMaintenance
	}
}

// Close shuts down the queue manager
func (m *Manager) Close() {
	logger.Info("Shutting down queue manager")
	m.cancel()
}