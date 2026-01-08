package exporter

import (
	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/prometheus/client_golang/prometheus"
)

func formatContainerStarted(ch chan<- prometheus.Metric, containerID string, inspect docker.ContainerInspect) {
	ch <- prometheus.MustNewConstMetric(
		containerStartedDesc,
		prometheus.GaugeValue,
		float64(inspect.StartedAt),
		containerID,
	)
}

func formatContainerExitCode(ch chan<- prometheus.Metric, containerID string, inspect docker.ContainerInspect) {
	ch <- prometheus.MustNewConstMetric(
		containerExitCodeDesc,
		prometheus.GaugeValue,
		float64(inspect.ExitCode),
		containerID,
	)
}

func formatContainerRestartCount(ch chan<- prometheus.Metric, containerID string, inspect docker.ContainerInspect) {
	ch <- prometheus.MustNewConstMetric(
		containerRestartCountDesc,
		prometheus.CounterValue,
		float64(inspect.RestartCount),
		containerID,
	)
}

func formatContainerFinished(ch chan<- prometheus.Metric, containerID string, inspect docker.ContainerInspect) {
	ch <- prometheus.MustNewConstMetric(
		containerFinishedAtDesc,
		prometheus.GaugeValue,
		float64(inspect.FinishedAt),
		containerID,
	)
}

func formatContainerSizeRootFs(ch chan<- prometheus.Metric, containerID string, inspect docker.ContainerInspect) {
	ch <- prometheus.MustNewConstMetric(
		containerSizeRootFsDesc,
		prometheus.GaugeValue,
		float64(inspect.SizeRootFs),
		containerID,
	)
}

func formatContainerSizeRw(ch chan<- prometheus.Metric, containerID string, inspect docker.ContainerInspect) {
	ch <- prometheus.MustNewConstMetric(
		containerSizeRwDesc,
		prometheus.GaugeValue,
		float64(inspect.SizeRw),
		containerID,
	)
}
