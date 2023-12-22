package tracing

import "context"

// TODO(KB): Attach real service instead of the stub.

type Span struct{}

func (*Span) Close() {}

func NewSpan(ctx context.Context, name string) (*Span, context.Context) {
	return &Span{}, ctx
}
