# Changelog

## [1.4.2](https://github.com/H3rmt/docker-exporter/compare/v1.4.1...v1.4.2) (2026-02-21)


### Bug Fixes

* rename docker disk usage metrics to have _bytes at end ([8c87636](https://github.com/H3rmt/docker-exporter/commit/8c876365bec73bc201c0507b32ed379741dc53be))

## [1.4.1](https://github.com/H3rmt/docker-exporter/compare/v1.4.0...v1.4.1) (2026-02-21)


### Bug Fixes

* use internal cache to speed up and improve cpu usage calculation ([c1097da](https://github.com/H3rmt/docker-exporter/commit/c1097da19351eccd39389c488f0d968fecacc31e))

## [1.4.0](https://github.com/H3rmt/docker-exporter/compare/v1.3.1...v1.4.0) (2026-02-21)


### Features

* added disk usage metrics ([908f77a](https://github.com/H3rmt/docker-exporter/commit/908f77ae2a68b94b67d28f2e2a2e92bdef135da1))
* better status reporting: ([3e256f3](https://github.com/H3rmt/docker-exporter/commit/3e256f357bde743aea02cb999f871ba3a4e7e5b9))
* better status reporting: ([a518e9c](https://github.com/H3rmt/docker-exporter/commit/a518e9c6b3cc7c86cd3e3ae64d3cc212fa643538))

## [1.3.1](https://github.com/H3rmt/docker-exporter/compare/v1.3.0...v1.3.1) (2026-02-10)


### Bug Fixes

* improve homepage version info ([c31028f](https://github.com/H3rmt/docker-exporter/commit/c31028f46c3190ead3b69405076321cb0bb522f4))

## [1.3.0](https://github.com/H3rmt/docker-exporter/compare/v1.2.0...v1.3.0) (2026-02-05)


### Features

* add CPU and memory limit statistics to homepage ([bd5eb14](https://github.com/H3rmt/docker-exporter/commit/bd5eb14cf2bb75ffa42617127d21ad4cb59b3cfd))
* added `docker_container_cpu_percent` and `docker_container_cpu_percent_host` ([da3540b](https://github.com/H3rmt/docker-exporter/commit/da3540bcdac3f60cc1d472910bbbc07b00ae23bf))
* added docker_container_cpu_percent and docker_container_cpu_percent_host ([21875a6](https://github.com/H3rmt/docker-exporter/commit/21875a6a43e11d274a43ee2aac5c9d5cb13253ec))
* optimize container data collection with parallel processing ([c93ea9f](https://github.com/H3rmt/docker-exporter/commit/c93ea9f18ad3925841a37915ccea564a7cf019b6))

## [1.2.0](https://github.com/H3rmt/docker-exporter/compare/v1.1.4...v1.2.0) (2026-01-30)


### Features

* enhance web UI, update status API, and refine documentation ([cdf446c](https://github.com/H3rmt/docker-exporter/commit/cdf446cfcc56be783d3dc056c0efc88f55a481e4))


### Bug Fixes

* expose IP addr in dashboard ([630ac3f](https://github.com/H3rmt/docker-exporter/commit/630ac3ffcf2625ebfb93d2552b74e0df1e537e17))
* remove custom meminfo handling ([630ac3f](https://github.com/H3rmt/docker-exporter/commit/630ac3ffcf2625ebfb93d2552b74e0df1e537e17))
* switch to cobra ([75fa2c6](https://github.com/H3rmt/docker-exporter/commit/75fa2c639df4a1cd11b591a4ebc9de639ea851a5))

## [1.1.4](https://github.com/H3rmt/docker-exporter/compare/v1.1.3...v1.1.4) (2026-01-09)


### Bug Fixes

* add hostname to all labels ([#42](https://github.com/H3rmt/docker-exporter/issues/42)) ([9ff5391](https://github.com/H3rmt/docker-exporter/commit/9ff53913eb67d6cd655626d1b6020f5158db9930))
* add hostname to all labels ([#42](https://github.com/H3rmt/docker-exporter/issues/42)) ([9ff5391](https://github.com/H3rmt/docker-exporter/commit/9ff53913eb67d6cd655626d1b6020f5158db9930))

## [1.1.3](https://github.com/H3rmt/docker-exporter/compare/v1.1.2...v1.1.3) (2026-01-09)


### Bug Fixes

* special handling for meminfo in socket ([9291ed3](https://github.com/H3rmt/docker-exporter/commit/9291ed300afbae0135fc6af16e00d9f173cbbfa8))

## [1.1.2](https://github.com/H3rmt/docker-exporter/compare/v1.1.1...v1.1.2) (2026-01-09)


### Bug Fixes

* add verbose mode ([70ed1c5](https://github.com/H3rmt/docker-exporter/commit/70ed1c59da98047aa25c41c798104cba90a59293))

## [1.1.1](https://github.com/H3rmt/docker-exporter/compare/v1.1.0...v1.1.1) (2026-01-08)


### Bug Fixes

* lager timeout for bg data collection for charts ([b39f267](https://github.com/H3rmt/docker-exporter/commit/b39f267db6af1e1a4a73308d88fb33fbcfb0d174))
* show total memory, show mounted hostname ([101172e](https://github.com/H3rmt/docker-exporter/commit/101172ede60097f28018bfc799ec32ebb372b74d))

## [1.0.2](https://github.com/H3rmt/docker-exporter/compare/v1.0.1...v1.0.2) (2025-12-10)


### Bug Fixes

* **deps:** update module github.com/docker/docker to v28.3.3+incompatible [security] ([d54c1c0](https://github.com/H3rmt/docker-exporter/commit/d54c1c045adaf59512733999f7ae2f903322218c))
* **deps:** update module github.com/docker/docker to v28.5.2+incompatible ([8c82db6](https://github.com/H3rmt/docker-exporter/commit/8c82db6566e54378754a7ccc62fdc57532d64dbf))
* **deps:** update module github.com/prometheus/client_golang to v1.23.2 ([00d8b86](https://github.com/H3rmt/docker-exporter/commit/00d8b8611e05c2fba1287024dfd2f0ddd14773e9))

## [1.0.1](https://github.com/H3rmt/docker-exporter/compare/v1.0.0...v1.0.1) (2025-07-03)


### Bug Fixes

* fix mem computation ([740b4ea](https://github.com/H3rmt/docker-exporter/commit/740b4ea9b913f5b572fe46e348b89bcf18305acc))

## [1.0.0](https://github.com/H3rmt/docker-exporter/compare/v0.1.1...v1.0.0) (2025-07-03)


### Features

* added many new metrics ([137f96b](https://github.com/H3rmt/docker-exporter/commit/137f96b80609e779a9bfb0388a970441d5843bc4))

## [0.1.1](https://github.com/H3rmt/docker-exporter/compare/v0.1.0...v0.1.1) (2025-07-01)


### Bug Fixes

* fix logger file refs ([57eb50c](https://github.com/H3rmt/docker-exporter/commit/57eb50c5621250d2b9e880c0ced41a4fe7a3976d))

## 0.1.0 (2025-07-01)


### Features

* added cli args ([ba9c948](https://github.com/H3rmt/docker-exporter/commit/ba9c948472c3aa550b1210d556385f121f848ce0))
* added docker_container_info, docker_container_name, docker_container_state, docker_container_created metrics ([7137cc0](https://github.com/H3rmt/docker-exporter/commit/7137cc0c4f99fd10584ad8b53e95e7f42d63eff1))
* added docker_container_ports ([4f89915](https://github.com/H3rmt/docker-exporter/commit/4f89915cb8f52ebf3bafbb20131119ed45120035))
* added docker-host cli arg ([ce654f5](https://github.com/H3rmt/docker-exporter/commit/ce654f59a86a4587190f56d4685308b4187f9e84))
* added quiet flag ([014f3f8](https://github.com/H3rmt/docker-exporter/commit/014f3f86387fb0828d4f17636b8fe859bae3bb23))
* initial ([66251c3](https://github.com/H3rmt/docker-exporter/commit/66251c3c0080b7f93968d8872cac41b4897710fd))

## 1.0.0 (2025-07-01)


### Features

* added cli args ([ba9c948](https://github.com/H3rmt/docker-exporter/commit/ba9c948472c3aa550b1210d556385f121f848ce0))
* added docker_container_info, docker_container_name, docker_container_state, docker_container_created metrics ([7137cc0](https://github.com/H3rmt/docker-exporter/commit/7137cc0c4f99fd10584ad8b53e95e7f42d63eff1))
* added docker-host cli arg ([ce654f5](https://github.com/H3rmt/docker-exporter/commit/ce654f59a86a4587190f56d4685308b4187f9e84))
* added quiet flag ([014f3f8](https://github.com/H3rmt/docker-exporter/commit/014f3f86387fb0828d4f17636b8fe859bae3bb23))
* initial ([66251c3](https://github.com/H3rmt/docker-exporter/commit/66251c3c0080b7f93968d8872cac41b4897710fd))
