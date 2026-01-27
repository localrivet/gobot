// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package subscription

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gobot/internal/logic/subscription"
	"gobot/internal/svc"
)

// List all available subscription plans
func ListPlansHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := subscription.NewListPlansLogic(r.Context(), svcCtx)
		resp, err := l.ListPlans()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
