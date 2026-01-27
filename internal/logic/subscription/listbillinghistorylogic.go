package subscription

import (
	"context"
	"fmt"

	"gobot/internal/auth"
	"gobot/internal/svc"
	"gobot/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListBillingHistoryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListBillingHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListBillingHistoryLogic {
	return &ListBillingHistoryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListBillingHistoryLogic) ListBillingHistory(req *types.ListBillingHistoryRequest) (resp *types.ListBillingHistoryResponse, err error) {
	if l.svcCtx.Levee == nil {
		return nil, fmt.Errorf("levee service not configured")
	}

	// Get email from JWT context
	email, err := auth.GetEmailFromContext(l.ctx)
	if err != nil {
		l.Errorf("Failed to get email from context: %v", err)
		return nil, err
	}

	// Set defaults
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	// Get invoices from Levee SDK
	invoicesResp, err := l.svcCtx.Levee.Customers.ListCustomerInvoices(l.ctx, email, pageSize)
	if err != nil {
		l.Errorf("Failed to get invoices for %s: %v", email, err)
		return nil, err
	}

	// Convert to response format
	records := make([]types.BillingRecord, 0, len(invoicesResp.Invoices))
	for _, inv := range invoicesResp.Invoices {
		records = append(records, types.BillingRecord{
			Id:               inv.ID,
			Amount:           int(inv.AmountDue),
			Currency:         inv.Currency,
			Status:           inv.Status,
			Description:      inv.Description,
			InvoiceDate:      inv.CreatedAt,
			PaidAt:           inv.PaidAt,
			InvoicePdfUrl:    inv.InvoicePdfUrl,
			HostedInvoiceUrl: inv.HostedUrl,
		})
	}

	return &types.ListBillingHistoryResponse{
		Records:    records,
		TotalCount: len(records),
		Page:       page,
		PageSize:   pageSize,
	}, nil
}
