package policy_handler

import (
	"context"

	"github.com/atomist-skills/go-skill"
	"github.com/atomist-skills/go-skill/policy/data/proxy"
)

func WithProxyClient() Opt {
	return func(h *EventHandler) {
		h.proxyClientProvider = func(ctx context.Context, req skill.RequestContext) proxy.ProxyClient {
			return proxy.NewProxyClientFromSkillRequest(ctx, req)
		}
	}
}
