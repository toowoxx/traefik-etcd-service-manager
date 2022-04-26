package main

import (
	"context"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
	"github.com/pkg/errors"
)

type EventType int

const (
	Created EventType = iota
	Removed
	UpdatedLabels
)

func (t EventType) String() string {
	switch t {
	case Created:
		return "created"
	case Removed:
		return "removed"
	case UpdatedLabels:
		return "updated labels"
	}
	return "unknown"
}

type DockerManager struct {
	client *docker.Client

	containers []types.Container
}

func newDockerManager() (*DockerManager, error) {
	client, err := docker.NewClientWithOpts(docker.FromEnv)
	if err != nil {
		return nil, errors.Wrap(err, "could not create docker manager")
	}
	return &DockerManager{client: client}, nil
}

func (d *DockerManager) watchContainers(ctx context.Context) chan []types.Container {
	ch := make(chan []types.Container)
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			_ = <-ticker.C
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			containers, err := d.client.ContainerList(ctx, types.ContainerListOptions{})
			cancel()

			if err != nil {
				log.Println("error while getting containers", err)
			} else {
				ch <- containers
			}
		}
	}()
	return ch
}

func (d *DockerManager) subscribeToContainerEvents(
	ctx context.Context, onEvent func(eventType EventType, container types.Container, oldContainer *types.Container),
) {
	log.Println("Subscribing to container events")
	containersChannel := d.watchContainers(ctx)
	for {
		containers := <-containersChannel
		if len(containers) != len(d.containers) {
			log.Println("Container count changed. Old:", len(d.containers), "new: ", len(containers))
		}

		var addedContainers []types.Container
		var removedContainers []types.Container
		var labelsChangedContainers map[*types.Container]*types.Container

		// This code looks a bit convoluted but there's not much I
		// can do about it since it's idiomatic Go to do slice comparisons
		// this way.
		// This code compares the old container list with the new container list
		// and works out added containers, removed containers and containers whose
		// labels have been changed, added or removed, either fully or in part.
		for _, oldContainer := range d.containers {
			oldExists := false
			var newContainer types.Container
			for _, newContainer = range containers {
				if oldContainer.ID == newContainer.ID {
					oldExists = true
					break
				}
			}
			if !oldExists {
				removedContainers = append(removedContainers, oldContainer)
			} else {
				for key, oldValue := range oldContainer.Labels {
					foundOldKey := false
					for newKey, newValue := range newContainer.Labels {
						if key == newKey {
							if oldValue != newValue {
								log.Println("Label", key, "was changed for container", newContainer.ID)
								labelsChangedContainers[&oldContainer] = &newContainer
								goto endLabelCheck
							}
							foundOldKey = true
							break
						}
					}
					if !foundOldKey {
						log.Println("Label", key, "was removed from container", oldContainer.ID)
						labelsChangedContainers[&oldContainer] = &newContainer
						goto endLabelCheck
					}
				}
				for key := range oldContainer.Labels {
					foundNewKey := false
					for oldKey := range oldContainer.Labels {
						if key == oldKey {
							foundNewKey = true
							break
						}
					}
					if !foundNewKey {
						log.Println("Label", key, "was added to container", newContainer.ID)
						labelsChangedContainers[&oldContainer] = &newContainer
						goto endLabelCheck
					}
				}
			endLabelCheck:
			}
		}

		for _, newContainer := range containers {
			newExists := false
			for _, oldContainer := range d.containers {
				if newContainer.ID == oldContainer.ID {
					newExists = true
					break
				}
			}

			if !newExists {
				addedContainers = append(addedContainers, newContainer)
			}
		}

		if len(addedContainers) > 0 {
			log.Println("Added containers:", len(addedContainers))
		}
		if len(removedContainers) > 0 {
			log.Println("Removed containers:", len(removedContainers))
		}
		if len(labelsChangedContainers) > 0 {
			log.Println("Label changed containers:", len(labelsChangedContainers))
		}
		for _, container := range removedContainers {
			onEvent(Removed, container, nil)
		}
		for _, container := range addedContainers {
			onEvent(Created, container, nil)
		}
		for oldContainer, newContainer := range labelsChangedContainers {
			onEvent(UpdatedLabels, *oldContainer, newContainer)
		}

		d.containers = containers
	}
}
