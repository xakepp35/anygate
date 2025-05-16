package router

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestRegister_OK(t *testing.T) {
	r := New()

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	for _, m := range methods {
		r.Register(m, "/bank", func(ctx *fasthttp.RequestCtx) {})
		r.Register(m, "/rate", func(ctx *fasthttp.RequestCtx) {})
		r.Register(m, "/ra", func(ctx *fasthttp.RequestCtx) {})
	}

	expectedPrefixCount := map[string]int{
		"/":    1,
		"bank": 1,
		"rate": 1,
		"ra":   1,
	}

	actualPrefixCount := make(map[string]int)

	calculatePrefixRecursive(&r.root, actualPrefixCount)

	assert.Equal(t, expectedPrefixCount, actualPrefixCount)

	expectedPrefixMethods := map[string][]string{
		"bank": []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		"rate": []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		"ra":   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	}

	actualPrefixMethods := make(map[string][]string)
	calculatePrefixMethodsRecursive(&r.root, actualPrefixMethods)

	for p, want := range expectedPrefixMethods {
		assert.ElementsMatch(t, want, actualPrefixMethods[p])
	}

	bankWild := getWildByPrefixRecursive(&r.root, "bank")
	assert.Nil(t, bankWild)

	rateWild := getWildByPrefixRecursive(&r.root, "rate")
	assert.Nil(t, rateWild)

	raWild := getWildByPrefixRecursive(&r.root, "ra")
	assert.Nil(t, raWild)
}

func TestRegister_AnyHandler(t *testing.T) {
	r := New()

	r.Register("GET", "/bank", func(ctx *fasthttp.RequestCtx) {})
	r.Register("POST", "/bank", func(ctx *fasthttp.RequestCtx) {})
	r.Register("PUT", "/bank", func(ctx *fasthttp.RequestCtx) {})
	r.Register("PATCH", "/bank", func(ctx *fasthttp.RequestCtx) {})
	r.Register("DELETE", "/bank", func(ctx *fasthttp.RequestCtx) {})
	r.Register("ANY", "/bank", func(ctx *fasthttp.RequestCtx) {})

	expectedPrefixCount := map[string]int{
		"/":    1,
		"bank": 1,
	}

	actualPrefixCount := make(map[string]int)
	calculatePrefixRecursive(&r.root, actualPrefixCount)

	assert.Equal(t, expectedPrefixCount, actualPrefixCount)

	expectedPrefixMethods := map[string][]string{
		"bank": []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	}

	actualPrefixMethods := make(map[string][]string)

	calculatePrefixMethodsRecursive(&r.root, actualPrefixMethods)

	for p, want := range expectedPrefixMethods {
		assert.ElementsMatchf(t, want, actualPrefixMethods[p],
			"methods mismatch for prefix %s", p)
	}

	handler := getWildByPrefixRecursive(&r.root, "bank")
	assert.NotNil(t, handler)
}

func TestRegister_NilHandler(t *testing.T) {
	r := New()

	r.Register("GET", "/bank", nil)

	expectedPrefixCount := map[string]int{
		"/":    1,
		"bank": 1,
	}

	actualPrefixCount := make(map[string]int)
	calculatePrefixRecursive(&r.root, actualPrefixCount)

	assert.Equal(t, expectedPrefixCount, actualPrefixCount)

	expectedPrefixMethods := map[string][]string{
		"bank": []string{"GET"},
	}

	actualPrefixMethods := make(map[string][]string)

	calculatePrefixMethodsRecursive(&r.root, actualPrefixMethods)

	assert.Equal(t, expectedPrefixMethods, actualPrefixMethods)
}

func TestLongestCommonPrefix_OK(t *testing.T) {
	for i := 0; i < 100; i++ {
		a := fmt.Sprintf("/path/%02d/segment", i)
		b := fmt.Sprintf("/path/%02d/segment/sub", i)
		want := computeExpectedLCP(a, b)

		t.Run(fmt.Sprintf("case-%02d", i), func(t *testing.T) {
			got := longestCommonPrefix(a, b)
			assert.Equalf(t, want, got, "longestCommonPrefix(%q,%q)", a, b)
		})
	}
}

func TestLongestCommonPrefix_NoCommon(t *testing.T) {
	cases := []struct{ a, b string }{
		{"/a", "/b"},
		{"/α", "/β"},
		{"", "/notempty"},
		{"/only", ""},
		{"/caps", "/CAPS"},
	}
	for _, c := range cases {
		want := computeExpectedLCP(c.a, c.b)
		got := longestCommonPrefix(c.a, c.b)
		assert.Equalf(t, want, got, "lcp(%q,%q) expected %d got %d", c.a, c.b, want, got)
	}
}

func TestInsertLookup_Positive(t *testing.T) {
	var root node
	methods := []string{"GET", "POST", "PUT", "DELETE"}

	for i, p := range dbgPaths {
		i := i
		insert(&root, p, methods[i%4], func(id int) fasthttp.RequestHandler {
			return func(ctx *fasthttp.RequestCtx) { ctx.SetUserValue("id", id) }
		}(i))
	}

	for i, p := range dbgPaths {
		m := methods[i%4]
		t.Run(fmt.Sprintf("pos-%d", i), func(t *testing.T) {
			h := lookup(&root, p, m)
			assert.NotNil(t, h)
			ctx := newCtx(m, p)
			h(ctx)
			assert.Equal(t, i, ctx.UserValue("id"))
		})
	}
}

func TestInsertLookup_WrongMethod(t *testing.T) {
	var root node
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	for i, p := range dbgPaths {
		insert(&root, p, methods[i%4], func(ctx *fasthttp.RequestCtx) {})
	}
	for i, p := range dbgPaths {
		wrong := altMethod(methods[i%4])
		assert.Nilf(t, lookup(&root, p, wrong), "expected nil for %s %s", wrong, p)
	}
}

func TestInsertLookup_UnknownPath(t *testing.T) {
	var root node
	insert(&root, "/known", "GET", func(ctx *fasthttp.RequestCtx) {})
	assert.Nil(t, lookup(&root, "/unknown", "GET"))
}

func TestRouter_Positive(t *testing.T) {
	r := New()
	for i, p := range okPaths {
		status := 300 + i
		r.Register("GET", p, func(code int) fasthttp.RequestHandler {
			return func(ctx *fasthttp.RequestCtx) { ctx.SetStatusCode(code) }
		}(status))
	}
	for i, p := range okPaths {
		ctx := newCtx("GET", p)
		r.Handler(ctx)
		assert.Equalf(t, 300+i, ctx.Response.StatusCode(), p)
	}
}
func TestRouter_Tricky404(t *testing.T) {
	r := New() //
	for _, p := range trickyPaths {
		ctx := newCtx("GET", p)
		r.Handler(ctx)
		assert.Equalf(t, fasthttp.StatusNotFound, ctx.Response.StatusCode(), p)
	}
}

func TestRouterHandler_Wildcard(t *testing.T) {
	r := New()
	r.Register("ANY", "/wild", func(ctx *fasthttp.RequestCtx) { ctx.SetStatusCode(299) })
	ctx := newCtx("PATCH", "/wild")
	r.Handler(ctx)
	assert.Equal(t, 299, ctx.Response.StatusCode())
}

func TestRouterHandler_WrongMethod(t *testing.T) {
	r := New()
	r.Register("GET", "/path", func(ctx *fasthttp.RequestCtx) {})
	ctx := newCtx("POST", "/path")
	r.Handler(ctx)
	assert.Equal(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())
}

func TestRouterHandler_UnknownPath(t *testing.T) {
	r := New()
	r.Register("GET", "/exists", func(ctx *fasthttp.RequestCtx) {})
	ctx := newCtx("GET", "/nope")
	r.Handler(ctx)
	assert.Equal(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())
}

func TestLookup_WildcardAndParentFallback(t *testing.T) {
	var root node

	// wildcard на /api
	insert(&root, "/api", "", func(ctx *fasthttp.RequestCtx) { ctx.SetUserValue("wild", true) })
	// конкретный GET на /api/users
	insert(&root, "/api/users", "GET", func(ctx *fasthttp.RequestCtx) { ctx.SetUserValue("user", true) })

	// 3a. POST /api → должен вернуть wildcard
	ctx1 := newCtx("POST", "/api")
	h1 := lookup(&root, "/api", "POST")
	assert.NotNil(t, h1)
	h1(ctx1)
	assert.True(t, ctx1.UserValue("wild").(bool))

	// 3b. GET /api/unknown → нет ребёнка, но у /api есть wildcard → вернётся wildcard
	ctx2 := newCtx("GET", "/api/unknown")
	h2 := lookup(&root, "/api/unknown", "GET")
	assert.NotNil(t, h2)
	h2(ctx2)
	assert.True(t, ctx2.UserValue("wild").(bool))

	// 3c. GET /api/users → конкретный
	ctx3 := newCtx("GET", "/api/users")
	h3 := lookup(&root, "/api/users", "GET")
	assert.NotNil(t, h3)
	h3(ctx3)
	assert.True(t, ctx3.UserValue("user").(bool))
}

func TestInsert_SplitPrefix(t *testing.T) {
	var root node
	idBank := 1
	idBase := 2
	insert(&root, "/bank", "GET", func(ctx *fasthttp.RequestCtx) { ctx.SetUserValue("id", idBank) })
	insert(&root, "/base", "GET", func(ctx *fasthttp.RequestCtx) { ctx.SetUserValue("id", idBase) }) // заставит split "/ba"

	ctx1 := newCtx("GET", "/bank")
	lookup(&root, "/bank", "GET")(ctx1)
	assert.Equal(t, idBank, ctx1.UserValue("id"))

	ctx2 := newCtx("GET", "/base")
	lookup(&root, "/base", "GET")(ctx2)
	assert.Equal(t, idBase, ctx2.UserValue("id"))
}

func TestInsert_HandlersMapCreated(t *testing.T) {
	var root node
	insert(&root, "/root", "GET", func(ctx *fasthttp.RequestCtx) {})

	// handler карта хранится в узле, чей prefix="/root"
	var leaf *node
	for _, child := range root.children {
		if child.prefix == "/root" {
			leaf = child
			break
		}
	}
	assert.NotNil(t, leaf)
	assert.NotNil(t, leaf.handlers)
	assert.Contains(t, leaf.handlers, "GET")
}

func TestInsert_AddWildcardToExistingNode(t *testing.T) {
	var root node
	// первый раз добавляем конкретный метод
	insert(&root, "/same", "GET", func(ctx *fasthttp.RequestCtx) {})
	// второй раз wildcard на тот же путь
	insert(&root, "/same", "", func(ctx *fasthttp.RequestCtx) { ctx.SetUserValue("wild", true) })

	ctx := newCtx("POST", "/same")
	h := lookup(&root, "/same", "POST")
	assert.NotNil(t, h) // должен вернуться только что добавленный wildcard
	h(ctx)
	assert.True(t, ctx.UserValue("wild").(bool))
}

func calculatePrefixRecursive(root *node, actual map[string]int) {
	if root == nil {
		return
	}

	actual[root.prefix] += 1
	for _, child := range root.children {
		calculatePrefixRecursive(child, actual)
	}
}

func calculatePrefixMethodsRecursive(root *node, actual map[string][]string) {
	if root == nil {
		return
	}

	for method, _ := range root.handlers {
		actual[root.prefix] = append(actual[root.prefix], method)
	}

	for _, child := range root.children {
		calculatePrefixMethodsRecursive(child, actual)
	}
}

func getWildByPrefixRecursive(root *node, prefix string) fasthttp.RequestHandler {
	if root == nil {
		return nil
	}

	if root.prefix == prefix && len(root.children) == 0 {
		return root.wild
	}

	for _, child := range root.children {
		handler := getWildByPrefixRecursive(child, prefix)
		if handler != nil {
			return handler
		}
	}

	return nil
}

func newCtx(method, path string) *fasthttp.RequestCtx {
	ctx := &fasthttp.RequestCtx{}
	ctx.Request.Header.SetMethod(method)
	ctx.Request.Header.SetRequestURI(path)
	return ctx
}

func computeExpectedLCP(a, b string) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	i := 0
	for i < minLen && a[i] == b[i] {
		i++
	}
	return i
}

func pathForIndex(i int) string {
	switch i % 5 {
	case 0:
		return fmt.Sprintf("/simple/%d", i)
	case 1:
		return fmt.Sprintf("/level/%d/deeper", i)
	case 2:
		return fmt.Sprintf("/unicode/π/%d", i)
	case 3:
		return fmt.Sprintf("/caps/ABC/%d", i)
	default:
		return fmt.Sprintf("/trailing/%d/", i)
	}
}

func altMethod(m string) string {
	switch m {
	case "GET":
		return "POST"
	case "POST":
		return "PUT"
	case "PUT":
		return "DELETE"
	default:
		return "GET"
	}
}

var okPaths = []string{
	"/a", "/a/b", "/a/b/c", "/caps/ABC", "/unicode/π",
	"/bank", "/rate", "/x/y", "/alpha", "/beta",
	"/a/b/c/d", "/x/y/z", "/deep/1/2/3", "/bank/loan", "/rate/usd",
	"/ra/deep", "/long/segment1/segment2", "/numeric/123456",
	"/emoji/😀", "/unicode/λ/π",
	"/ab", "/aa", "/aB", "/a123", "/123a",
	"/path.with.dot", "/path_with_underscore", "/path-with-dash",
	"/cases/UPPER", "/cases/lower",
	"/alpha/beta/gamma", "/numeric/987654321",
	"/bank/account", "/bank/branch", "/bank/loan/approved",
	"/rate/eur", "/rate/gbp", "/rate/jpy",
	"/ra/deepest/level", "/x/y/z/0", "/x/long", "/x/y_long",
	"/edge/nochange", "/reserved;semicolon", "/colon/with:colon",
	"/backslash/\\", "/tilde/~", "/asterisk/*",
	"/plus/+", "/minus/-",
	"/unicode/π/extra", "/unicode/λ", "/unicode/µ",
	"/alpha/β", "/beta/γ", "/gamma/δ",
	"/ra", "/ra/deeper", "/ra/deeper/still", "/ok/final",
}

var trickyPaths = []string{
	"/trailing//", "/slashes//double", "/triple//slash///", "/x//y//z",
	"/double//middle//slash", "/lead//slash", "//double/lead", "//",
	"/edge/./dot", "/edge/../parent", "/./root", "/level/one/../two",
	"/a/b/../../c", "/dot/././keep", "/trail/../",
	"/percent/%25", "/percent/%25extra", "/specialchars/%40", "/specialchars/%23",
	"/decode/%2Fslash", "/decode/%2E", "/decode/%2e%2e/",
	"/path/%2e%2e/home", "/mixed/%25%32%30", "/percent/%25encoded",
	"/hash/fragment#", "/hash/frag#ment", "/fragment/#end",
	"/empty//", "/path//", "/root//", "/multi///slash//",
	"/a//b//c", "/dots/..", "/dots/.", "/dots/.../file",
	"/cleanup/./../clean", "/compress//./dir/../file",
	"/../../escape", "/up/../..", "/../upper", "/depth////../end",
	"/uni/%E2%9C%93", "/uni/%e2%9c%93", "/emoji/%F0%9F%98%80",
	"/mix/%2e/./", "/mix/%2E/../", "/mix/%2f", "/mix/%2F",
	"/edge/%2E%2E", "/edge/%2E./", "/edge/.%2E/",
	"/collapse/a/./b/../c//", "/collapse//",
	"/dots/filename.", "/dots/dir./file", "/trailing/space/ ",
	"/space/%20here", "/space/there%20", "/space/%20",
	"/plusencoded/%2B", "/plusencoded/text%2Bend",
	"/reserved/%3Bsemicolon", "/reserved/%3Acolon",
	"/reserved/%40at", "/reserved/%23hash",
	"/rep/%2e%2e/%2e%2e/%2e%2e", "/rep/%2E%2E/dir",
	"/combo/%2E/..//", "/combo/%2e//./", "/combo/%2f/",
	"/frag/%23hash#", "/frag/%2523hash", "/frag/%2Fslash#frag",
	"/dotdot/%2e%2e", "/cleanup/%2E%2E/",
}

var dbgPaths = []string{
	"/a", "/a/b", "/a/b/c", "/unicode/π", "/Caps", "/trailing/", "/x/y", "/bank", "/rate", "/ra", "/", "/a/bd",
	"/a/b/c/d", "/aB", "/ab", "/aa", "/unicode/λ", "/unicode/π/extra", "/caps/ABC", "/caps/Abc", "/caps/abc",
	"/trailing//slash", "/trailing//", "/x/y/z", "/x", "/bank/branch", "/bank/account", "/rate/usd", "/rate/eur",
	"/ra/deep", "/a/b/c/d/e", "/a/d", "/aa/bb", "/ab/cd", "/alpha", "/beta/gamma", "/numbers/0", "/numbers/123/456",
	"/hex/0xFF", "/symbols/exclaim!", "/symbols/plus+", "/symbols/tilde~", "/unicode/λ/π", "/unicode/测试",
	"/unicode/🤖", "/emoji/😀", "/long/segment1/segment2/segment3/segment4", "/triple//slash///", "/endslash",
	"/endslash/", "/multi/byte/тест", "/cases/UPPER", "/cases/lower", "/cases/MixedCase", "/path.with.dot",
	"/path_with_underscore", "/path-with-dash", "/path,comma", "/sql/SELECT", "/json/{}", "/percent/%25",
	"/paren/(test)", "/brackets/[1]", "/semicolon;test", "/colon:test", "/pipe/pipe|test", "/space/here",
	"/slashes//double", "/a123", "/123a", "/0/1/2/3", "/deep/1/2/3/4/5", "/bank/loan", "/bank/loan/approved",
	"/rate/gbp", "/rate/jpy", "/ra/deepest/level", "/x/y/z/0", "/x/long", "/x/y_long", "/deep/0/1/2/3/4/5",
	"/deep2/level1/level2", "/alpha/beta/gamma", "/numeric/1234567890", "/specialchars/%40", "/specialchars/%23",
	"/mix/слово", "/emoji/😂", "/edge/./dot", "/edge/../parent", "/reserved;semicolon", "/colon/with:colon",
	"/pipe/|/bar", "/tilde/~", "/asterisk/*", "/plus/+", "/minus/-", "/percent/%25encoded", "/hash/fragment#",
}

func BenchmarkRouterHandler(b *testing.B) {
	r := New()
	for i := 0; i < benchRoutes; i++ {
		path := benchPath(i)
		r.Register("GET", path, func(*fasthttp.RequestCtx) {})
	}

	ctx := &fasthttp.RequestCtx{}
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		id := 0
		for pb.Next() {
			path := benchPath(id % benchRoutes)
			ctx.Request.Header.SetMethod("GET")
			ctx.Request.SetRequestURI(path)
			r.Handler(ctx)
			id++
		}
	})
}

func BenchmarkLookup(b *testing.B) {
	var root node
	for i := 0; i < benchRoutes; i++ {
		insert(&root, benchPath(i), "GET", func(*fasthttp.RequestCtx) {})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		id := 0
		for pb.Next() {
			path := benchPath(id % benchRoutes)
			lookup(&root, path, "GET")
			id++
		}
	})
}

const benchRoutes = 10_000

func benchPath(i int) string { return fmt.Sprintf("/p%d", i) }

func BenchmarkInsert(b *testing.B) {
	for _, n := range []int{1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("%d-routes", n), func(b *testing.B) {
			paths := make([]string, n)
			for i := range paths {
				paths[i] = benchPath(i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var root node
				for j, p := range paths {
					insert(&root, p, "GET",
						func(id int) fasthttp.RequestHandler { return func(*fasthttp.RequestCtx) {} }(j))
				}
			}
		})
	}
}
