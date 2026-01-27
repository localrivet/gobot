package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"gobot/internal/mcp/mcpctx"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// subscriptionActions defines valid actions for subscription resource.
var subscriptionActions = map[string][]string{
	"subscription": {"get"},
}

// SubscriptionInput defines input for the unified subscription tool.
type SubscriptionInput struct {
	Resource string `json:"resource" jsonschema:"required,Resource type: subscription"`
	Action   string `json:"action" jsonschema:"required,Action to perform"`
}

// RegisterSubscriptionTool registers the unified subscription tool.
func RegisterSubscriptionTool(server *mcp.Server, toolCtx *mcpctx.ToolContext) {
	mcp.AddTool(server, &mcp.Tool{
		Name:  "subscription",
		Title: "Subscription Management",
		Description: `View subscription status.

Resources:
- subscription: Subscription information

SUBSCRIPTION RESOURCE:
- subscription.get: Get current subscription status

Examples:
  subscription(resource: subscription, action: get)`,
	}, subscriptionHandler(toolCtx))
}

func subscriptionHandler(toolCtx *mcpctx.ToolContext) func(ctx context.Context, req *mcp.CallToolRequest, input SubscriptionInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input SubscriptionInput) (*mcp.CallToolResult, any, error) {
		fmt.Printf("[MCP subscription] Handler called - Resource: %q, Action: %q\n", input.Resource, input.Action)

		// Validate resource
		validActions, ok := subscriptionActions[input.Resource]
		if !ok {
			return nil, nil, mcpctx.NewValidationError(
				fmt.Sprintf("invalid resource '%s', must be: subscription", input.Resource),
				"resource")
		}

		// Validate action
		if !slices.Contains(validActions, input.Action) {
			return nil, nil, mcpctx.NewValidationError(
				fmt.Sprintf("invalid action '%s' for resource '%s', must be: %s",
					input.Action, input.Resource, strings.Join(validActions, ", ")),
				"action")
		}

		switch input.Action {
		case "get":
			return handleSubscriptionGet(ctx, toolCtx, input)
		}
		return nil, nil, nil
	}
}

// ============================================================================
// SUBSCRIPTION HANDLERS
// ============================================================================

// SubscriptionGetOutput defines output for subscription.get.
type SubscriptionGetOutput struct {
	Status            string  `json:"status"`
	PlanID            string  `json:"plan_id"`
	CurrentPeriodEnd  *string `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd bool    `json:"cancel_at_period_end"`
}

func handleSubscriptionGet(ctx context.Context, toolCtx *mcpctx.ToolContext, input SubscriptionInput) (*mcp.CallToolResult, any, error) {
	sub, err := toolCtx.DB().GetSubscriptionByUserID(ctx, toolCtx.UserID())
	if err != nil {
		// Return free plan if no subscription
		return nil, SubscriptionGetOutput{
			Status: "none",
			PlanID: "free",
		}, nil
	}

	output := SubscriptionGetOutput{
		Status:            sub.Status,
		PlanID:            sub.PlanID,
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd == 1,
	}

	if sub.CurrentPeriodEnd.Valid {
		periodEnd := time.Unix(sub.CurrentPeriodEnd.Int64, 0).Format(time.RFC3339)
		output.CurrentPeriodEnd = &periodEnd
	}

	return nil, output, nil
}

// registerSubscriptionToolToRegistry registers subscription tool to the direct-call registry.
func registerSubscriptionToolToRegistry(registry *ToolRegistry, toolCtx *mcpctx.ToolContext) {
	registry.Register("subscription", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var input SubscriptionInput
		if err := json.Unmarshal(args, &input); err != nil {
			return nil, fmt.Errorf("invalid input: %w", err)
		}
		handler := subscriptionHandler(toolCtx)
		_, output, err := handler(ctx, nil, input)
		return output, err
	})
}
