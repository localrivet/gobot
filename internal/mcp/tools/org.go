package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"gobot/internal/db"
	"gobot/internal/mcp/mcpctx"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// orgActions defines valid actions for each org resource.
var orgActions = map[string][]string{
	"org": {"list", "select", "get", "update", "create"},
}

// OrgInput defines input for the unified org tool.
type OrgInput struct {
	Resource string `json:"resource" jsonschema:"required,Resource type: org"`
	Action   string `json:"action" jsonschema:"required,Action to perform"`

	// Common
	ID string `json:"id,omitempty" jsonschema:"Resource ID (for get, update)"`

	// Select-specific
	Slug string `json:"slug,omitempty" jsonschema:"Organization slug to select (alternative to id). For org.select."`

	// Create/Update-specific
	Name    string `json:"name,omitempty" jsonschema:"Organization name. For org.create, org.update."`
	LogoURL string `json:"logo_url,omitempty" jsonschema:"Organization logo URL. For org.update."`
}

// RegisterOrgTool registers the unified org tool.
func RegisterOrgTool(server *mcp.Server, toolCtx *mcpctx.ToolContext) {
	mcp.AddTool(server, &mcp.Tool{
		Name:  "org",
		Title: "Organization Management",
		Description: `Manage organizations and select which organization to work with.

IMPORTANT: You must use org.select before using other org-scoped tools.

Resources:
- org: Organization management

ORG RESOURCE:
- org.list: List all organizations you have access to
- org.select: Select an organization to work with (requires: id or slug)
- org.get: Get current organization details
- org.update: Update organization settings (optional: name, logo_url)
- org.create: Create a new organization (requires: name)

Examples:
  org(resource: org, action: list)
  org(resource: org, action: select, slug: "my-company")
  org(resource: org, action: select, id: "uuid")
  org(resource: org, action: get)
  org(resource: org, action: update, name: "Updated Name")
  org(resource: org, action: create, name: "New Company")`,
	}, orgHandler(toolCtx))
}

func orgHandler(toolCtx *mcpctx.ToolContext) func(ctx context.Context, req *mcp.CallToolRequest, input OrgInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest, input OrgInput) (*mcp.CallToolResult, any, error) {
		fmt.Printf("[MCP org] Handler called - Resource: %q, Action: %q, ID: %q, Slug: %q, Name: %q\n",
			input.Resource, input.Action, input.ID, input.Slug, input.Name)

		// Validate resource
		validActions, ok := orgActions[input.Resource]
		if !ok {
			fmt.Printf("[MCP org] ERROR: invalid resource %q\n", input.Resource)
			return nil, nil, mcpctx.NewValidationError(
				fmt.Sprintf("invalid resource '%s', must be: org", input.Resource),
				"resource")
		}

		// Validate action
		if !slices.Contains(validActions, input.Action) {
			fmt.Printf("[MCP org] ERROR: invalid action %q for resource %q\n", input.Action, input.Resource)
			return nil, nil, mcpctx.NewValidationError(
				fmt.Sprintf("invalid action '%s' for resource '%s', must be: %s",
					input.Action, input.Resource, strings.Join(validActions, ", ")),
				"action")
		}

		switch input.Resource {
		case "org":
			return handleOrg(ctx, toolCtx, input)
		}
		return nil, nil, nil // unreachable
	}
}

// ============================================================================
// ORG HANDLERS
// ============================================================================

// OrgListItem represents an organization in the list.
type OrgListItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	LogoURL  string `json:"logo_url,omitempty"`
	Selected bool   `json:"selected"`
}

// OrgListOutput defines output for org.list.
type OrgListOutput struct {
	Organizations []OrgListItem `json:"organizations"`
	Total         int           `json:"total"`
}

// OrgSelectOutput defines output for org.select.
type OrgSelectOutput struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Selected bool   `json:"selected"`
	Message  string `json:"message,omitempty"`
}

// OrgGetOutput defines output for org.get.
type OrgGetOutput struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	LogoURL string `json:"logo_url,omitempty"`
	OwnerID string `json:"owner_id"`
}

// OrgUpdateOutput defines output for org.update.
type OrgUpdateOutput struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	LogoURL string `json:"logo_url,omitempty"`
	Updated bool   `json:"updated"`
}

// OrgCreateOutput defines output for org.create.
type OrgCreateOutput struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Created bool   `json:"created"`
}

func handleOrg(ctx context.Context, toolCtx *mcpctx.ToolContext, input OrgInput) (*mcp.CallToolResult, any, error) {
	switch input.Action {
	case "list":
		return handleOrgList(ctx, toolCtx, input)
	case "select":
		return handleOrgSelect(ctx, toolCtx, input)
	case "get":
		return handleOrgGet(ctx, toolCtx, input)
	case "update":
		return handleOrgUpdate(ctx, toolCtx, input)
	case "create":
		return handleOrgCreate(ctx, toolCtx, input)
	}
	return nil, nil, nil
}

func handleOrgList(ctx context.Context, toolCtx *mcpctx.ToolContext, input OrgInput) (*mcp.CallToolResult, any, error) {
	items := make([]OrgListItem, 0)

	// List all organizations the user has access to
	orgs, err := toolCtx.DB().ListUserOrganizations(ctx, toolCtx.UserID())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	currentOrgID := toolCtx.OrgID()
	for _, org := range orgs {
		items = append(items, OrgListItem{
			ID:       org.ID,
			Name:     org.Name,
			Slug:     org.Slug,
			LogoURL:  org.LogoUrl.String,
			Selected: org.ID == currentOrgID,
		})
	}

	return nil, OrgListOutput{
		Organizations: items,
		Total:         len(items),
	}, nil
}

func handleOrgSelect(ctx context.Context, toolCtx *mcpctx.ToolContext, input OrgInput) (*mcp.CallToolResult, any, error) {
	orgID := input.ID

	fmt.Printf("[MCP org.select] Starting - SessionID: %s, ID: %q, Slug: %q\n",
		toolCtx.SessionID(), input.ID, input.Slug)

	if orgID == "" && input.Slug == "" {
		fmt.Println("[MCP org.select] ERROR: neither id nor slug provided")
		return nil, nil, mcpctx.NewValidationError("either id or slug is required for org.select", "id")
	}

	var fullOrg db.Organization
	var err error

	if orgID != "" {
		fmt.Printf("[MCP org.select] Looking up by ID: %s\n", orgID)
		fullOrg, err = toolCtx.DB().GetOrganizationByID(ctx, orgID)
	} else {
		fmt.Printf("[MCP org.select] Looking up by slug: %s\n", input.Slug)
		fullOrg, err = toolCtx.DB().GetOrganizationBySlug(ctx, input.Slug)
	}
	if err != nil {
		fmt.Printf("[MCP org.select] ERROR: org lookup failed: %v\n", err)
		if orgID != "" {
			return nil, nil, mcpctx.NewNotFoundError(fmt.Sprintf("organization with ID '%s' not found", orgID))
		}
		return nil, nil, mcpctx.NewNotFoundError(fmt.Sprintf("organization with slug '%s' not found", input.Slug))
	}

	// Verify user has access to this org
	_, err = toolCtx.DB().GetOrganizationMember(ctx, db.GetOrganizationMemberParams{
		OrganizationID: fullOrg.ID,
		UserID:         toolCtx.UserID(),
	})
	if err != nil {
		return nil, nil, mcpctx.NewUnauthorizedError("you don't have access to this organization")
	}

	fmt.Printf("[MCP org.select] Found org: %s (%s), calling SelectOrg\n", fullOrg.Name, fullOrg.ID)
	toolCtx.SelectOrg(fullOrg)
	fmt.Printf("[MCP org.select] SelectOrg completed, HasOrg: %v\n", toolCtx.HasOrg())

	return nil, OrgSelectOutput{
		ID:       fullOrg.ID,
		Name:     fullOrg.Name,
		Slug:     fullOrg.Slug,
		Selected: true,
		Message:  "Organization selected. All subsequent operations will use this organization.",
	}, nil
}

func handleOrgGet(ctx context.Context, toolCtx *mcpctx.ToolContext, input OrgInput) (*mcp.CallToolResult, any, error) {
	if err := toolCtx.RequireOrg(); err != nil {
		return nil, nil, err
	}
	org := toolCtx.Org()

	return nil, OrgGetOutput{
		ID:      org.ID,
		Name:    org.Name,
		Slug:    org.Slug,
		LogoURL: org.LogoUrl.String,
		OwnerID: org.OwnerID,
	}, nil
}

func handleOrgUpdate(ctx context.Context, toolCtx *mcpctx.ToolContext, input OrgInput) (*mcp.CallToolResult, any, error) {
	if err := toolCtx.RequireOrg(); err != nil {
		return nil, nil, err
	}
	org := toolCtx.Org()

	// Build update params
	params := db.UpdateOrganizationParams{
		ID: toolCtx.OrgID(),
	}

	if input.Name != "" {
		params.Name = sql.NullString{String: input.Name, Valid: true}
	}
	if input.LogoURL != "" {
		params.LogoUrl = sql.NullString{String: input.LogoURL, Valid: true}
	}

	err := toolCtx.DB().UpdateOrganization(ctx, params)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to update organization: %w", err)
	}

	updatedOrg, err := toolCtx.DB().GetOrganizationByID(ctx, toolCtx.OrgID())
	if err != nil {
		return nil, OrgUpdateOutput{
			ID:      org.ID,
			Name:    org.Name,
			Slug:    org.Slug,
			Updated: true,
		}, nil
	}

	return nil, OrgUpdateOutput{
		ID:      updatedOrg.ID,
		Name:    updatedOrg.Name,
		Slug:    updatedOrg.Slug,
		LogoURL: updatedOrg.LogoUrl.String,
		Updated: true,
	}, nil
}

func handleOrgCreate(ctx context.Context, toolCtx *mcpctx.ToolContext, input OrgInput) (*mcp.CallToolResult, any, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, nil, mcpctx.NewValidationError("name is required", "name")
	}

	// Generate slug from name
	slug := strings.ToLower(input.Name)
	slug = strings.ReplaceAll(slug, " ", "-")
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	slug = result.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	slug = strings.Trim(slug, "-")

	orgID := uuid.New().String()

	org, err := toolCtx.DB().CreateOrganization(ctx, db.CreateOrganizationParams{
		ID:      orgID,
		Name:    input.Name,
		Slug:    slug,
		OwnerID: toolCtx.UserID(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Add the creator as owner
	_, err = toolCtx.DB().AddOrganizationMember(ctx, db.AddOrganizationMemberParams{
		ID:             uuid.New().String(),
		OrganizationID: org.ID,
		UserID:         toolCtx.UserID(),
		Role:           "owner",
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add member to organization: %w", err)
	}

	// Auto-select the newly created org so subsequent operations work immediately
	toolCtx.SelectOrg(org)

	return nil, OrgCreateOutput{
		ID:      org.ID,
		Name:    org.Name,
		Slug:    org.Slug,
		Created: true,
	}, nil
}

// registerOrgToolToRegistry registers org tool to the direct-call registry.
func registerOrgToolToRegistry(registry *ToolRegistry, toolCtx *mcpctx.ToolContext) {
	registry.Register("org", func(ctx context.Context, args json.RawMessage) (interface{}, error) {
		var input OrgInput
		if err := json.Unmarshal(args, &input); err != nil {
			return nil, fmt.Errorf("invalid input: %w", err)
		}
		handler := orgHandler(toolCtx)
		_, output, err := handler(ctx, nil, input)
		return output, err
	})
}
