package relay

import "go.opentelemetry.io/otel"

var (
	tracer = otel.Tracer("github.com/datachainlab/ethereum-ibc-relay-prover/relay")
)
