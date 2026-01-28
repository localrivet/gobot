package notification

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"gobot/internal/logic/notification"
	"gobot/internal/svc"
)

// Mark all notifications as read
func MarkAllNotificationsReadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := notification.NewMarkAllNotificationsReadLogic(r.Context(), svcCtx)
		resp, err := l.MarkAllNotificationsRead()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
