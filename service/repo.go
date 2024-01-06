package service

func NewServiceRepo() *ServiceRepo {
	return &ServiceRepo{
		services: make([]Service, 0),
	}
}

type ServiceRepo struct {
	services []Service
}

func (r *ServiceRepo) Registe(service Service) {
	r.services = append(r.services, service)
}

func (r *ServiceRepo) Services() []ServiceRunner {
	runners := make([]ServiceRunner, 0)

	for _, svc := range r.services {
		runners = append(runners, WrapDefault(svc))
	}

	return runners
}
