package main

import (
	"context"
	"log"
	"strings"

	"github.com/docker/docker/api/types"
)

const traefikPrefix = "traefik."

func dockerLabelToEtcdKey(key string) string {
	return strings.ReplaceAll(key, ".", "/")
}

func main() {
	ctx := context.Background()

	etcdManager, err := newEtcdManager()
	if err != nil {
		panic(err)
	}

	dockerManager, err := newDockerManager()
	if err != nil {
		panic(err)
	}
	dockerManager.subscribeToContainerEvents(
		ctx,
		func(eventType EventType, container types.Container, oldContainer *types.Container) {
			log.Println("Received event", eventType, "for container", container.ID)
			switch eventType {
			case UpdatedLabels:
				log.Println(oldContainer.ID, "| updating labels")
				for key := range oldContainer.Labels {
					if strings.HasPrefix(key, traefikPrefix) {
						log.Println(oldContainer.ID, "| removing label", key)
						if err := etcdManager.Remove(dockerLabelToEtcdKey(key)); err != nil {
							log.Println("Error while removing key", key, ":", err.Error())
						}
					}
				}
				fallthrough
			case Created:
				if len(container.Labels) == 0 {
					return
				}
				log.Println(container.ID, "| adding labels")
				for key, value := range container.Labels {
					if strings.HasPrefix(key, traefikPrefix) {
						log.Println(container.ID, "| adding label", key)
						if err := etcdManager.Put(dockerLabelToEtcdKey(key), value); err != nil {
							log.Println("Error while putting key", key, ":", err.Error())
						}
					}
				}
			case Removed:
				if len(container.Labels) == 0 {
					return
				}
				log.Println(container.ID, "| removing labels")
				for key := range container.Labels {
					if strings.HasPrefix(key, traefikPrefix) {
						log.Println(container.ID, "| removing label", key)
						if err := etcdManager.Remove(dockerLabelToEtcdKey(key)); err != nil {
							log.Println("Error while removing key", key, ":", err.Error())
						}
					}
				}
			}
		})
}
