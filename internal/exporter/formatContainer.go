package exporter

import (
	"strconv"

	"github.com/h3rmt/docker-exporter/internal/docker"

	"github.com/moby/moby/api/types/container"
	"github.com/prometheus/client_golang/prometheus"
)

func formatContainerInfo(ch chan<- prometheus.Metric, hostname string, containerInfo []docker.ContainerInfo) {
	for _, c := range containerInfo {
		ch <- prometheus.MustNewConstMetric(
			containerInfoDesc,
			prometheus.CounterValue,
			1,
			hostname,
			c.ID,
			c.Names[0],
			c.ImageID,
			c.Command,
			c.NetworkMode,
		)
	}
}

func formatContainerNames(ch chan<- prometheus.Metric, hostname string, containerInfo []docker.ContainerInfo) {
	for _, c := range containerInfo {
		for _, name := range c.Names {
			ch <- prometheus.MustNewConstMetric(
				containerNameDesc,
				prometheus.CounterValue,
				1,
				hostname,
				c.ID,
				name,
			)
		}
	}
}

func formatContainerState(ch chan<- prometheus.Metric, hostname string, containerInfo []docker.ContainerInfo) {
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
			hostname,
			c.ID,
		)
	}
}

func formatContainerCreated(ch chan<- prometheus.Metric, hostname string, containerInfo []docker.ContainerInfo) {
	for _, c := range containerInfo {
		ch <- prometheus.MustNewConstMetric(
			containerCreatedDesc,
			prometheus.CounterValue,
			float64(c.Created),
			hostname,
			c.ID,
		)
	}
}

func formatContainerPorts(ch chan<- prometheus.Metric, hostname string, containerInfo []docker.ContainerInfo) {
	for _, c := range containerInfo {
		for _, port := range c.Ports {
			ch <- prometheus.MustNewConstMetric(
				containerPortsDesc,
				prometheus.GaugeValue,
				1,
				hostname,
				c.ID,
				strconv.Itoa(int(port.PublicPort)),
				strconv.Itoa(int(port.PrivatePort)),
				port.IP.String(),
				port.Type,
			)
		}
	}
}
