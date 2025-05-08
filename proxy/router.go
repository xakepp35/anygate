package proxy

import (
	"github.com/fasthttp/router"
	"github.com/xakepp35/anygate/plugin"
)

func BuildGroup(r *router.Router, group RoutesGroup, inheritedPlugins []plugin.Spec, inheritedConfig ProxyConfig) {
	// Наследуем конфиг, если не задан явно
	cfg := inheritedConfig
	if group.Config != (ProxyConfig{}) {
		cfg = group.Config
	}
	// Склеиваем middleware цепочку текущий группы
	fullChain := append([]plugin.Spec{}, inheritedPlugins...)
	fullChain = append(fullChain, group.Plugins...)
	// Регистрируем маршруты текущей группы
	for prefix, target := range group.Routes {
		base := Handler(prefix, target, cfg)
		final := plugin.BuildChain(base, fullChain...)
		r.ANY(prefix, final)
	}
	// Рекурсивно обрабатываем подгруппы
	for _, child := range group.Groups {
		BuildGroup(r, child, fullChain, cfg)
	}
}
