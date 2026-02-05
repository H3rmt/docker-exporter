package exporter

import (
	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/prometheus/client_golang/prometheus"
)

func formatContainerPids(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerPidsDesc,
		prometheus.GaugeValue,
		float64(stat.PIds),
		hostname,
		containerID,
	)
}

func formatContainerCpuUserMicroSeconds(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerCpuUserNSDesc,
		prometheus.CounterValue,
		float64(stat.Cpu.UsageUserNS),
		hostname,
		containerID,
	)
}

func formatContainerCpuKernelMicroSeconds(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerCpuKernelNSDesc,
		prometheus.CounterValue,
		float64(stat.Cpu.UsageKernelNS),
		hostname,
		containerID,
	)
}

func formatContainerCpuMicroSeconds(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerCpuNSDesc,
		prometheus.CounterValue,
		float64(stat.Cpu.UsageNS),
		hostname,
		containerID,
	)
}

func formatContainerMemLimitKiB(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerMemLimitKiBDesc,
		prometheus.GaugeValue,
		float64(stat.MemoryLimitKiB),
		hostname,
		containerID,
	)
}

func formatContainerMemUsageKiB(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerMemUsageKiBDesc,
		prometheus.GaugeValue,
		float64(stat.MemoryUsageKiB),
		hostname,
		containerID,
	)
}

func formatContainerNetSendBytes(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerNetSendBytesDesc,
		prometheus.CounterValue,
		float64(stat.Net.SendBytes),
		hostname,
		containerID,
	)
}

func formatContainerNetSendDropped(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerNetSendDroppedDesc,
		prometheus.CounterValue,
		float64(stat.Net.SendDropped),
		hostname,
		containerID,
	)
}

func formatContainerNetSendErrors(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerNetSendErrorsDesc,
		prometheus.CounterValue,
		float64(stat.Net.SendErrors),
		hostname,
		containerID,
	)
}

func formatContainerNetRecvBytes(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerNetRecvBytesDesc,
		prometheus.CounterValue,
		float64(stat.Net.RecvBytes),
		hostname,
		containerID,
	)
}

func formatContainerNetRecvDropped(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerNetRecvDroppedDesc,
		prometheus.CounterValue,
		float64(stat.Net.RecvDropped),
		hostname,
		containerID,
	)
}

func formatContainerNetRecvErrors(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerNetRecvErrorsDesc,
		prometheus.CounterValue,
		float64(stat.Net.RecvErrors),
		hostname,
		containerID,
	)
}

func formatBlockInputBytes(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerBlockInputBytesDesc,
		prometheus.CounterValue,
		float64(stat.BlockInputBytes),
		hostname,
		containerID,
	)
}

func formatBlockOutputBytes(ch chan<- prometheus.Metric, hostname string, containerID string, stat docker.ContainerStats) {
	ch <- prometheus.MustNewConstMetric(
		containerBlockOutputBytesDesc,
		prometheus.CounterValue,
		float64(stat.BlockOutputBytes),
		hostname,
		containerID,
	)
}
