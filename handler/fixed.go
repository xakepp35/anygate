package handler

import (
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

func NewFixed(code int, body []byte) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(code)
		if body != nil {
			ctx.SetBody(body)
		}
	}
}

// Ok Edge case, 200 bodyless (for benchmarks)
func Ok(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
}

// Echo Edge case, 200 echoed (for benchmarks)
func Echo(ctx *fasthttp.RequestCtx) {
	ctx.Response.SetBodyRaw(ctx.Request.Body()) // избегает копирования
}

// isFixedResponse проверяет, начинается ли строка с числового HTTP-кода
func isFixedResponse(to string) bool {
	if to == "" {
		return false
	}
	// Найти первую позицию пробела
	spaceIdx := strings.IndexByte(to, ' ')
	if spaceIdx == -1 {
		// Нет пробела — пытаемся распарсить всю строку как код
		_, err := strconv.Atoi(to)
		return err == nil
	}
	// Есть пробел — пытаемся распарсить часть до пробела как код
	_, err := strconv.Atoi(to[:spaceIdx])
	return err == nil
}

// parseFixedResponse парсит строку "200" или "200 текст"
// Возвращает код и тело (если есть)
func parseFixedResponse(to string) (int, []byte) {
	if to == "" {
		return 200, nil
	}
	spaceIdx := strings.IndexByte(to, ' ')
	if spaceIdx == -1 {
		// Только код, без тела
		code, _ := strconv.Atoi(to)
		return code, nil
	}
	code, _ := strconv.Atoi(to[:spaceIdx])
	body := to[spaceIdx+1:]
	return code, []byte(body)
}
