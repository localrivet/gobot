// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package oauth

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gobot/internal/logic/oauth"
	"gobot/internal/svc"
	"gobot/internal/types"
)

// OAuth callback - exchange code for tokens
func OAuthCallbackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OAuthLoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := oauth.NewOAuthCallbackLogic(r.Context(), svcCtx)
		resp, err := l.OAuthCallback(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
