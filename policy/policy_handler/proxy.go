package policy_handler

import (
	"context"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data/proxy"
)

func WithProxyClient(url string) Opt {
	return func(h *EventHandler) {
		provider := getProxyClientProvider(url)
		h.proxyClientProvider = &provider
	}
}

func getProxyClientProvider(url string) proxyClientProvider {
	return func(ctx context.Context, req skill.RequestContext) proxy.ProxyClient {
		return proxy.NewProxyClientFromSkillRequest(ctx, url, req)
	}
}
