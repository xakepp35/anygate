package utils

import (
	"strings"
	"sync"
)

type PathBuilder struct {
	rootPath string
	fromLen  int
	pool     *sync.Pool
}

func NewPathBuilder(rootPath string, fromLen, capHint int) *PathBuilder {
	return &PathBuilder{
		rootPath: rootPath,
		fromLen:  fromLen,
		pool: &sync.Pool{
			New: func() any {
				b := strings.Builder{}
				b.Grow(capHint)
				return &b
			},
		},
	}
}

func (s *PathBuilder) Build(ctxPath []byte) string {
	builder := s.pool.Get().(*strings.Builder)
	defer s.pool.Put(builder)
	builder.Reset()
	builder.WriteString(s.rootPath)
	if len(ctxPath) >= s.fromLen {
		builder.Write(ctxPath[s.fromLen:])
	}
	return builder.String()
}
