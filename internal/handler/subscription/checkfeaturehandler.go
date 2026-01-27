// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package subscription

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gobot/internal/logic/subscription"
	"gobot/internal/svc"
	"gobot/internal/types"
)

// Check if user has access to a feature
func CheckFeatureHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CheckFeatureRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := subscription.NewCheckFeatureLogic(r.Context(), svcCtx)
		resp, err := l.CheckFeature(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
