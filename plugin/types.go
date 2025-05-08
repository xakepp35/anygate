package plugin

import "github.com/valyala/fasthttp"

type (
	Handler     = fasthttp.RequestHandler
	Func        = func(Handler) Handler
	SpecArgs    = map[string]any
	Constructor = func(args SpecArgs) (Func, error)
)

type Spec struct {
	Kind string   `yaml:"kind"`
	Args SpecArgs `yaml:"args"`
}
