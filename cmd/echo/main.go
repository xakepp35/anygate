// echo.go — чистый 200 OK, никаких лишних движений
package main

import (
	"log"

	"github.com/valyala/fasthttp"
)

func main() {
	addr := ":9000"
	log.Println("✅ fasthttp echo up at", addr)
	err := fasthttp.ListenAndServe(addr, handlerOK)
	if err != nil {
		log.Fatalln("🔥 fasthttp echo crash:", err)
	}
}

func handlerOK(ctx *fasthttp.RequestCtx) {
	// log.Println("request")
	ctx.SetStatusCode(fasthttp.StatusOK)
}
