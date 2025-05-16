package router

import (
	"fmt"
	"github.com/xakepp35/anygate/router"
	"testing"

	"github.com/valyala/fasthttp"
)

const (
	regSmall  = 1_000
	regMedium = 10_000
	regLarge  = 100_000
)

func genPath(i int) string { return fmt.Sprintf("/r%d", i) }

func BenchmarkRouterRegister(b *testing.B) {
	sizes := []int{regSmall, regMedium, regLarge}

	for _, n := range sizes {
		b.Run(fmt.Sprintf("%dk", n/1_000), func(b *testing.B) {
			paths := make([]string, n)
			for i := range paths {
				paths[i] = genPath(i)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				r := router.New()
				for _, p := range paths {
					r.Register("GET", p, func(*fasthttp.RequestCtx) {})
				}
			}
		})
	}
}

// фиксированный датасет для lookup'ов
var benchRouter *router.Router
var benchCtx fasthttp.RequestCtx

func init() {
	benchRouter = router.New()
	for i := 0; i < regMedium; i++ {
		benchRouter.Register("GET", genPath(i), func(*fasthttp.RequestCtx) {})
	}
}

// параллельный бенчмарку lookup; каждый worker перебирает пути по кругу
func BenchmarkRouterLookup(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			path := genPath(i % regMedium)
			benchCtx.Request.Header.SetMethod("GET")
			benchCtx.Request.SetRequestURI(path)
			benchRouter.Handler(&benchCtx)
			i++
		}
	})
}
