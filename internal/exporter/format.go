package exporter

import (
	"docker-exporter/internal/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/prometheus/client_golang/prometheus"
)

func formatContainerInfo(ch chan<- prometheus.Metric, containerInfo []docker.ContainerInfo) {
	for _, c := range containerInfo {
		ch <- prometheus.MustNewConstMetric(
			containerInfoDesc,
			prometheus.GaugeValue,
			1,
			c.ID,
			c.Names[0],
			c.ImageID,
			c.Command,
		)
	}
}

func formatContainerNames(ch chan<- prometheus.Metric, containerInfo []docker.ContainerInfo) {
	for _, c := range containerInfo {
		for _, name := range c.Names {
			ch <- prometheus.MustNewConstMetric(
				containerNameDesc,
				prometheus.GaugeValue,
				1,
				c.ID,
				name,
			)
		}
	}
}

func formatContainerState(ch chan<- prometheus.Metric, containerInfo []docker.ContainerInfo) {
	for _, c := range containerInfo {
		var stateVal float64
		switch c.State {
		case container.StateCreated:
			stateVal = 0
		case container.StateRunning:
			stateVal = 1
		case container.StatePaused:
			stateVal = 2
		case container.StateRestarting:
			stateVal = 3
		case container.StateRemoving:
			stateVal = 4
		case container.StateExited:
			stateVal = 5
		case container.StateDead:
			stateVal = 6
		}
		ch <- prometheus.MustNewConstMetric(
			containerStateDesc,
			prometheus.GaugeValue,
			stateVal,
			c.ID,
		)
	}
}

func formatContainerCreated(ch chan<- prometheus.Metric, containerInfo []docker.ContainerInfo) {
	for _, c := range containerInfo {
		ch <- prometheus.MustNewConstMetric(
			containerCreatedDesc,
			prometheus.CounterValue,
			float64(c.Created),
			c.ID,
		)
	}
}
