package observe

import (
	"go.opentelemetry.io/otel/sdk/trace"
)

type DemoSampler struct{}

func (d DemoSampler) Description() string {
	return "DemoSampler"
}

func (d DemoSampler) ShouldSample(parameters trace.SamplingParameters) trace.SamplingResult {
	if parameters.Name == "/health" || parameters.Name == "/metrics" {
		return trace.SamplingResult{Decision: trace.Drop}
	}
	return trace.ParentBased(trace.TraceIDRatioBased(0.2)).ShouldSample(parameters)
}
