package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

// BrowserTool provides browser automation via Chrome DevTools Protocol
type BrowserTool struct {
	allocCtx context.Context
	cancel   context.CancelFunc
	timeout  time.Duration
}

type browserInput struct {
	Action   string `json:"action"`   // navigate, click, type, screenshot, text, html, evaluate, wait
	URL      string `json:"url"`      // For navigate action
	Selector string `json:"selector"` // CSS selector for element actions
	Text     string `json:"text"`     // Text to type or JS to evaluate
	Output   string `json:"output"`   // Output path for screenshot
	Timeout  int    `json:"timeout"`  // Action timeout in seconds (default: 30)
}

// BrowserConfig configures the browser tool
type BrowserConfig struct {
	Headless bool          // Run browser headlessly (default: true)
	Timeout  time.Duration // Default timeout (default: 30s)
}

func NewBrowserTool(cfg BrowserConfig) *BrowserTool {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", cfg.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	return &BrowserTool{
		allocCtx: allocCtx,
		cancel:   cancel,
		timeout:  cfg.Timeout,
	}
}

func (t *BrowserTool) Close() {
	if t.cancel != nil {
		t.cancel()
	}
}

func (t *BrowserTool) Name() string {
	return "browser"
}

func (t *BrowserTool) Description() string {
	return "Automate browser interactions via Chrome DevTools Protocol. Navigate to URLs, click elements, type text, take screenshots, extract content, and run JavaScript."
}

func (t *BrowserTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"action": {
				"type": "string",
				"enum": ["navigate", "click", "type", "screenshot", "text", "html", "evaluate", "wait"],
				"description": "Browser action: navigate (go to URL), click (click element), type (enter text), screenshot (capture page), text (get text content), html (get HTML), evaluate (run JS), wait (wait for element)"
			},
			"url": {
				"type": "string",
				"description": "URL to navigate to (required for 'navigate' action)"
			},
			"selector": {
				"type": "string",
				"description": "CSS selector for element (required for click, type, text, html, wait actions)"
			},
			"text": {
				"type": "string",
				"description": "Text to type (for 'type' action) or JavaScript code (for 'evaluate' action)"
			},
			"output": {
				"type": "string",
				"description": "File path to save screenshot (for 'screenshot' action). If empty, returns base64."
			},
			"timeout": {
				"type": "integer",
				"description": "Action timeout in seconds. Default: 30",
				"default": 30
			}
		},
		"required": ["action"]
	}`)
}

func (t *BrowserTool) RequiresApproval() bool {
	return true // Browser automation can be dangerous
}

func (t *BrowserTool) Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
	var params browserInput
	if err := json.Unmarshal(input, &params); err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Failed to parse input: %v", err),
			IsError: true,
		}, nil
	}

	timeout := t.timeout
	if params.Timeout > 0 {
		timeout = time.Duration(params.Timeout) * time.Second
	}

	// Create browser context
	browserCtx, cancel := chromedp.NewContext(t.allocCtx)
	defer cancel()

	// Add timeout
	browserCtx, cancel = context.WithTimeout(browserCtx, timeout)
	defer cancel()

	var result string
	var err error

	switch params.Action {
	case "navigate":
		result, err = t.navigate(browserCtx, params.URL)
	case "click":
		result, err = t.click(browserCtx, params.Selector)
	case "type":
		result, err = t.typeText(browserCtx, params.Selector, params.Text)
	case "screenshot":
		result, err = t.screenshot(browserCtx, params.Output)
	case "text":
		result, err = t.getText(browserCtx, params.Selector)
	case "html":
		result, err = t.getHTML(browserCtx, params.Selector)
	case "evaluate":
		result, err = t.evaluate(browserCtx, params.Text)
	case "wait":
		result, err = t.waitFor(browserCtx, params.Selector)
	default:
		return &ToolResult{
			Content: fmt.Sprintf("Unknown action: %s", params.Action),
			IsError: true,
		}, nil
	}

	if err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Browser action failed: %v", err),
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: result,
		IsError: false,
	}, nil
}

func (t *BrowserTool) navigate(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("URL is required for navigate action")
	}

	var title string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Title(&title),
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Navigated to: %s\nPage title: %s", url, title), nil
}

func (t *BrowserTool) click(ctx context.Context, selector string) (string, error) {
	if selector == "" {
		return "", fmt.Errorf("selector is required for click action")
	}

	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
		chromedp.Click(selector),
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Clicked element: %s", selector), nil
}

func (t *BrowserTool) typeText(ctx context.Context, selector, text string) (string, error) {
	if selector == "" {
		return "", fmt.Errorf("selector is required for type action")
	}
	if text == "" {
		return "", fmt.Errorf("text is required for type action")
	}

	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
		chromedp.Clear(selector),
		chromedp.SendKeys(selector, text),
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Typed text into element: %s", selector), nil
}

func (t *BrowserTool) screenshot(ctx context.Context, outputPath string) (string, error) {
	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.FullScreenshot(&buf, 90), // 90% quality
	)
	if err != nil {
		return "", err
	}

	if outputPath == "" {
		// Return as base64
		b64 := base64.StdEncoding.EncodeToString(buf)
		return fmt.Sprintf("Screenshot captured (%d bytes)\ndata:image/png;base64,%s", len(buf), b64), nil
	}

	// Expand ~ to home directory
	if strings.HasPrefix(outputPath, "~/") {
		homeDir, _ := os.UserHomeDir()
		outputPath = filepath.Join(homeDir, outputPath[2:])
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(outputPath, buf, 0644); err != nil {
		return "", fmt.Errorf("failed to save screenshot: %w", err)
	}

	return fmt.Sprintf("Screenshot saved to: %s (%d bytes)", outputPath, len(buf)), nil
}

func (t *BrowserTool) getText(ctx context.Context, selector string) (string, error) {
	if selector == "" {
		// Get all visible text from body
		selector = "body"
	}

	var text string
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
		chromedp.Text(selector, &text),
	)
	if err != nil {
		return "", err
	}

	// Truncate if too long
	if len(text) > 10000 {
		text = text[:10000] + "\n... (truncated)"
	}

	return text, nil
}

func (t *BrowserTool) getHTML(ctx context.Context, selector string) (string, error) {
	if selector == "" {
		selector = "html"
	}

	var html string
	err := chromedp.Run(ctx,
		chromedp.WaitReady(selector),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}

			if selector == "html" {
				html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
				return err
			}

			// Find element by selector
			var nodes []*cdp.Node
			if err := chromedp.Nodes(selector, &nodes).Do(ctx); err != nil {
				return err
			}
			if len(nodes) == 0 {
				return fmt.Errorf("no element found for selector: %s", selector)
			}
			html, err = dom.GetOuterHTML().WithNodeID(nodes[0].NodeID).Do(ctx)
			return err
		}),
	)
	if err != nil {
		return "", err
	}

	// Truncate if too long
	if len(html) > 50000 {
		html = html[:50000] + "\n... (truncated)"
	}

	return html, nil
}

func (t *BrowserTool) evaluate(ctx context.Context, js string) (string, error) {
	if js == "" {
		return "", fmt.Errorf("JavaScript code is required for evaluate action")
	}

	var result any
	err := chromedp.Run(ctx,
		chromedp.Evaluate(js, &result),
	)
	if err != nil {
		return "", err
	}

	// Convert result to string
	switch v := result.(type) {
	case string:
		return v, nil
	case nil:
		return "undefined", nil
	default:
		jsonResult, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Sprintf("%v", result), nil
		}
		return string(jsonResult), nil
	}
}

func (t *BrowserTool) waitFor(ctx context.Context, selector string) (string, error) {
	if selector == "" {
		return "", fmt.Errorf("selector is required for wait action")
	}

	start := time.Now()
	err := chromedp.Run(ctx,
		chromedp.WaitVisible(selector),
	)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Element '%s' appeared after %v", selector, time.Since(start).Round(time.Millisecond)), nil
}
