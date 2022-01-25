# Lighthouse

A high performance MQTT broker

[![Go](https://github.com/yunqi/lighthouse/actions/workflows/go.yml/badge.svg)](https://github.com/yunqi/lighthouse/actions/workflows/go.yml)
[![Godoc](https://img.shields.io/badge/godoc-reference-brightgreen)](https://pkg.go.dev/github.com/yunqi/lighthouse)
[![Go Report Card](https://goreportcard.com/badge/github.com/yunqi/lighthouse)](https://goreportcard.com/report/github.com/yunqi/lighthouse)
[![codecov](https://codecov.io/gh/yunqi/lighthouse/branch/master/graph/badge.svg?token=PGEOJVIkZB)](https://codecov.io/gh/yunqi/lighthouse)
[![GitHub](https://img.shields.io/github/license/yunqi/lighthouse)](https://github.com/yunqi/lighthouse/blob/master/LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/yunqi/lighthouse)](https://github.com/yunqi/lighthouse/stargazers)
[![GitHub pull requests](https://img.shields.io/github/issues-pr-raw/yunqi/lighthouse)](https://github.com/yunqi/lighthouse/pulls)

```shell
# jaeger
docker run -d   --rm   --name jaeger   -p 5775:5775/udp -p 6831:6831/udp -p 6832:6832/udp -p 5778:5778  -p 16686:16686 -p 14268:14268  -p 14269:14269   -p 9411:9411 jaegertracing/all-in-one
```