package exporter

import (
	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/h3rmt/docker-exporter/internal/osinfo"
	"github.com/prometheus/client_golang/prometheus"
)

func formatSystemInfo(ch chan<- prometheus.Metric, hostname string, version string) {
	ch <- prometheus.MustNewConstMetric(
		exporterInfoDesc,
		prometheus.CounterValue,
		1,
		hostname,
		version,
	)
}

func formatSystemHostInfo(ch chan<- prometheus.Metric, hostname string, info osinfo.OSInfo) {
	ch <- prometheus.MustNewConstMetric(
		hostOSInfoDesc,
		prometheus.CounterValue,
		1,
		hostname,
		info.Name,
		info.VersionID,
	)
}

func formatSystemDiskInfo(ch chan<- prometheus.Metric, hostname string, disk docker.DiskUsage) {
	ch <- prometheus.MustNewConstMetric(
		dockerDiskUsageContainersTotalSize,
		prometheus.GaugeValue,
		float64(disk.ContainersTotalSize),
		hostname,
	)
	ch <- prometheus.MustNewConstMetric(
		dockerDiskUsageContainersReclaimable,
		prometheus.GaugeValue,
		float64(disk.ContainersReclaimable),
		hostname,
	)
	ch <- prometheus.MustNewConstMetric(
		dockerDiskUsageImagesTotalSize,
		prometheus.GaugeValue,
		float64(disk.ImagesTotalSize),
		hostname,
	)
	ch <- prometheus.MustNewConstMetric(
		dockerDiskUsageImagesReclaimable,
		prometheus.GaugeValue,
		float64(disk.ImagesReclaimable),
		hostname,
	)
	ch <- prometheus.MustNewConstMetric(
		dockerDiskUsageBuildCacheTotalSize,
		prometheus.GaugeValue,
		float64(disk.BuildCacheTotalSize),
		hostname,
	)
	ch <- prometheus.MustNewConstMetric(
		dockerDiskUsageBuildCacheReclaimable,
		prometheus.GaugeValue,
		float64(disk.BuildCacheReclaimable),
		hostname,
	)
	ch <- prometheus.MustNewConstMetric(
		dockerDiskUsageVolumesTotalSize,
		prometheus.GaugeValue,
		float64(disk.VolumesTotalSize),
		hostname,
	)
	ch <- prometheus.MustNewConstMetric(
		dockerDiskUsageVolumesReclaimable,
		prometheus.GaugeValue,
		float64(disk.VolumesReclaimable),
		hostname,
	)
}
