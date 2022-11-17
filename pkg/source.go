package pkg

import (
	"context"
	"io"
)

type Provider interface {
	Name() string
	ListSources() []Source
}

type Source interface {
	Name() string
	Tail(ctx context.Context, follow bool, tail string) (io.ReadCloser, io.ReadCloser)
}
