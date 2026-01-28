package organization

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gobot/internal/logic/organization"
	"gobot/internal/svc"
	"gobot/internal/types"
)

// Update organization
func UpdateOrganizationHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdateOrganizationRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := organization.NewUpdateOrganizationLogic(r.Context(), svcCtx)
		resp, err := l.UpdateOrganization(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
