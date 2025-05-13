// echo.go — чистый 200 OK, никаких лишних движений
package main

import (
	"log"

	"github.com/valyala/fasthttp"
	"github.com/xakepp35/anygate/handler"
)

func main() {
	addr := ":9000"
	log.Println("✅ fasthttp echo up at", addr)
	err := fasthttp.ListenAndServe(addr, handler.Ok)
	if err != nil {
		log.Fatalln("🔥 fasthttp echo crash:", err)
	}
}
