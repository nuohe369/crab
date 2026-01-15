package cron

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// RedisClient defines the Redis client interface.
type RedisClient interface {
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error)
	Del(ctx context.Context, keys ...string) error
}

// Job defines a cron job.
type Job struct {
	Name    string        // job name
	Spec    string        // cron expression
	Func    func()        // job function
	Timeout time.Duration // job timeout (for lock TTL), default 5 minutes
}

// Scheduler manages cron jobs.
type Scheduler struct {
	cron  *cron.Cron
	redis RedisClient
	jobs  map[string]cron.EntryID
	mu    sync.RWMutex
}

var defaultScheduler *Scheduler

// Init initializes the default scheduler.
func Init(redis RedisClient) {
	defaultScheduler = New(redis)
}

// Get returns the default scheduler.
func Get() *Scheduler {
	return defaultScheduler
}

// New creates a new scheduler.
func New(redis RedisClient) *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds(), cron.WithChain(
			cron.Recover(cron.DefaultLogger),
		)),
		redis: redis,
		jobs:  make(map[string]cron.EntryID),
	}
}

// Register registers a job with the scheduler.
func (s *Scheduler) Register(job Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove existing job with the same name
	if id, ok := s.jobs[job.Name]; ok {
		s.cron.Remove(id)
	}

	// Wrap job function
	fn := s.wrapJob(job)

	id, err := s.cron.AddFunc(job.Spec, fn)
	if err != nil {
		return err
	}

	s.jobs[job.Name] = id
	log.Printf("cron: registered job [%s] spec=%s", job.Name, job.Spec)
	return nil
}

// wrapJob wraps a job to handle distributed locking.
func (s *Scheduler) wrapJob(job Job) func() {
	return func() {
		ttl := job.Timeout
		if ttl == 0 {
			ttl = 5 * time.Minute
		}

		lockKey := "cron:lock:" + job.Name
		ctx := context.Background()

		// Try to acquire lock
		ok, err := s.redis.SetNX(ctx, lockKey, "1", ttl)
		if err != nil || !ok {
			log.Printf("cron: job [%s] skipped (another instance is executing)", job.Name)
			return
		}
		defer s.redis.Del(ctx, lockKey)

		s.executeJob(job)
	}
}

// executeJob executes a job.
func (s *Scheduler) executeJob(job Job) {
	start := time.Now()
	log.Printf("cron: job [%s] started", job.Name)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("cron: job [%s] panic: %v", job.Name, r)
		}
	}()

	job.Func()
	log.Printf("cron: job [%s] completed in %v", job.Name, time.Since(start))
}

// Remove removes a job from the scheduler.
func (s *Scheduler) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if id, ok := s.jobs[name]; ok {
		s.cron.Remove(id)
		delete(s.jobs, name)
		log.Printf("cron: removed job [%s]", name)
	}
}

// Start starts the scheduler.
func (s *Scheduler) Start() {
	s.cron.Start()
	log.Println("cron: scheduler started (distributed mode)")
}

// Stop stops the scheduler.
func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	<-ctx.Done()
	log.Println("cron: scheduler stopped")
}

// ============ Convenience methods ============

// Register registers a job with the default scheduler.
func Register(job Job) error {
	return defaultScheduler.Register(job)
}

// Remove removes a job from the default scheduler.
func Remove(name string) {
	defaultScheduler.Remove(name)
}

// Start starts the default scheduler.
func Start() {
	defaultScheduler.Start()
}

// Stop stops the default scheduler.
func Stop() {
	defaultScheduler.Stop()
}
