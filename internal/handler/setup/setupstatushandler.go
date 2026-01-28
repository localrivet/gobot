package setup

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gobot/internal/logic/setup"
	"gobot/internal/svc"
)

// Check if setup is required (no admin exists)
func SetupStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := setup.NewSetupStatusLogic(r.Context(), svcCtx)
		resp, err := l.SetupStatus()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
