package exporter

import (
	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/prometheus/client_golang/prometheus"
)

func formatImageInfoCreated(ch chan<- prometheus.Metric, hostname string, imageInfo docker.ImageInfo) {
	ch <- prometheus.MustNewConstMetric(
		imageCreated,
		prometheus.GaugeValue,
		float64(imageInfo.Created),
		hostname,
		imageInfo.Name,
		imageInfo.ID,
	)
}

func formatImageInfoContainers(ch chan<- prometheus.Metric, hostname string, imageInfo docker.ImageInfo) {
	ch <- prometheus.MustNewConstMetric(
		imageContainers,
		prometheus.GaugeValue,
		float64(imageInfo.Containers),
		hostname,
		imageInfo.Name,
		imageInfo.ID,
	)
}

func formatImageInfoSize(ch chan<- prometheus.Metric, hostname string, imageInfo docker.ImageInfo) {
	ch <- prometheus.MustNewConstMetric(
		imageSize,
		prometheus.GaugeValue,
		float64(imageInfo.Size),
		hostname,
		imageInfo.Name,
		imageInfo.ID,
	)
}
