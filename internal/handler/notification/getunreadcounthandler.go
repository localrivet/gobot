package notification

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gobot/internal/logic/notification"
	"gobot/internal/svc"
)

// Get unread notification count
func GetUnreadCountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := notification.NewGetUnreadCountLogic(r.Context(), svcCtx)
		resp, err := l.GetUnreadCount()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
