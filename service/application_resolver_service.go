package service

import (
	docker "github.com/docker/docker/client"
	"museum/persistence"
	"museum/service/impl"
	service "museum/service/interface"
	"museum/util/cache"
)

type ApplicationResolverService service.ApplicationResolverService

func NewDockerHostApplicationResolverService(exhibitService service.ExhibitService) ApplicationResolverService {
	return &impl.DockerHostApplicationResolverService{
		ExhibitService: exhibitService,
	}
}

func NewDockerExtHostApplicationResolverService(exhibitService service.ExhibitService,
	client *docker.Client,
	eventing persistence.Eventing) ApplicationResolverService {
	return &impl.DockerExtHostApplicationResolverService{
		ExhibitService: exhibitService,
		IpCache:        cache.NewLRU[string, string](1000),
		Client:         client,
		Eventing:       eventing,
	}
}
