package config

type Trace struct {
	Name     string  `yaml:"name"`
	Endpoint string  `yaml:"endpoint"`
	Sampler  float64 `yaml:"sampler"`
	Batcher  string  `yaml:"batcher"  validate:"eq=jaeger|eq=zipkin"` //jaeger|zipkin
}
