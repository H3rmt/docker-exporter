package exporter

import (
	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/prometheus/client_golang/prometheus"
)

func formatContainerStarted(ch chan<- prometheus.Metric, hostname string, containerID string, inspect docker.ContainerInspect) {
	ch <- prometheus.MustNewConstMetric(
		containerStartedDesc,
		prometheus.GaugeValue,
		float64(inspect.StartedAt),
		hostname,
		containerID,
	)
}

func formatContainerFinished(ch chan<- prometheus.Metric, hostname string, containerID string, inspect docker.ContainerInspect) {
	ch <- prometheus.MustNewConstMetric(
		containerFinishedAtDesc,
		prometheus.GaugeValue,
		float64(inspect.FinishedAt),
		hostname,
		containerID,
	)
}

func formatContainerSizeRootFs(ch chan<- prometheus.Metric, hostname string, containerID string, inspect docker.ContainerInspect) {
	ch <- prometheus.MustNewConstMetric(
		containerSizeRootFsDesc,
		prometheus.GaugeValue,
		float64(inspect.SizeRootFs),
		hostname,
		containerID,
	)
}

func formatContainerSizeRw(ch chan<- prometheus.Metric, hostname string, containerID string, inspect docker.ContainerInspect) {
	ch <- prometheus.MustNewConstMetric(
		containerSizeRwDesc,
		prometheus.GaugeValue,
		float64(inspect.SizeRw),
		hostname,
		containerID,
	)
}

func formatContainerRestartCount(ch chan<- prometheus.Metric, hostname string, containerID string, inspect docker.ContainerInspect) {
	ch <- prometheus.MustNewConstMetric(
		containerRestartCountDesc,
		prometheus.CounterValue,
		float64(inspect.RestartCount),
		hostname,
		containerID,
	)
}

func formatContainerExitCode(ch chan<- prometheus.Metric, hostname string, containerID string, inspect docker.ContainerInspect) {
	ch <- prometheus.MustNewConstMetric(
		containerExitCodeDesc,
		prometheus.GaugeValue,
		float64(inspect.ExitCode),
		hostname,
		containerID,
	)
}
