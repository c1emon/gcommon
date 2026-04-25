package service

import "sync"

func NewServiceRepo() *ServiceRepo {
	return &ServiceRepo{
		services: make([]Service, 0),
	}
}

type ServiceRepo struct {
	mu       sync.RWMutex
	services []Service
}

func (r *ServiceRepo) Register(service Service) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services = append(r.services, service)
}

func (r *ServiceRepo) Services() []ServiceRunner {
	r.mu.RLock()
	defer r.mu.RUnlock()
	runners := make([]ServiceRunner, 0, len(r.services))

	for _, svc := range r.services {
		runners = append(runners, WrapDefault(svc))
	}

	return runners
}
