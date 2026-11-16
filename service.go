package main

import (
	"fmt"
	"sync"
)

type Service struct {
	tasks     map[int]*Task
	nextID    int
	processor Processor
	mu sync.RWMutex

	queue   chan int
	wg      sync.WaitGroup
	stopped bool
}

func NewService(p Processor) *Service {
		return &Service{
		tasks:     make(map[int]*Task),
		nextID:    1,
		processor: p,
		queue:     make(chan int, 10),
	}
}

func (s *Service) Create(password string) (*Task, error) {
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t := &Task{
		ID:       s.nextID,
		Password: password,
		Status:   StatusPending,
	}
	s.tasks[t.ID] = t
	s.nextID++

	return t, nil
}

func (s *Service) Get(id int) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("%w: id %d", ErrNotFound, id)
	}
	cp := *t
	return &cp, nil
}

func (s *Service) process(id int, from Status) error {
	s.mu.Lock()
	t, ok := s.tasks[id]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("%w: id %d", ErrNotFound, id)
	}
	if t.Status != from  {
		s.mu.Unlock()
		return fmt.Errorf("%w: cannot process task in status %q", ErrInvalidTransition, t.Status)
	}
	t.Status = StatusProcessing
	password := t.Password
	s.mu.Unlock()

	result, err := s.processor.Process(password)

	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		t.Status = StatusPending
		return err
	}
	t.Result = &result
	t.Status = StatusCompleted
	return nil
}

func (s *Service) Process(id int) error {
	return s.process(id, StatusPending)
}

func (s *Service) Start(workers int) {
	for i := 0; i < workers; i++ {
		s.wg.Add(1) 
		
		go func() {
			defer s.wg.Done() 
			for id := range s.queue {
				_ = s.process(id, StatusQueued)
			}
		}()
	}
}


func (s *Service) Submit(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return ErrStopped
	}

	t, ok := s.tasks[id]
	if !ok {
		return fmt.Errorf("%w: id %d", ErrNotFound, id)
	}
	
	if t.Status != StatusPending {
		return fmt.Errorf("%w: cannot submit task in status %q", ErrInvalidTransition, t.Status)
	}

	select {
	case s.queue <- id:
		t.Status = StatusQueued
		return nil
	default:
		return ErrQueueFull
	}
}

func (s *Service) Stop() {
	s.mu.Lock()
	if s.stopped {
		s.mu.Unlock()
		return
	}
	s.stopped = true
	
	close(s.queue)
	s.mu.Unlock()

	s.wg.Wait()
}