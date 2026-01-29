// Screen capture and OCR plugin for macOS.
// Provides: capture (full screen, window, region), ocr, info (screen dimensions)
// Build with: go build -o screen
package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-plugin"
)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "GOBOT_PLUGIN",
	MagicCookieValue: "gobot-plugin-v1",
}

type ToolResult struct {
	Content string `json:"content"`
	IsError bool   `json:"is_error"`
}

type ScreenTool struct {
	tempDir string
}

func NewScreenTool() *ScreenTool {
	tmpDir := filepath.Join(os.TempDir(), "gobot-screen")
	os.MkdirAll(tmpDir, 0755)
	return &ScreenTool{tempDir: tmpDir}
}

func (t *ScreenTool) Name() string {
	return "screen"
}

func (t *ScreenTool) Description() string {
	return "Screen capture for macOS: capture full screen, specific window, or rectangular region. Supports OCR text extraction from screenshots. Returns base64-encoded image or extracted text."
}

func (t *ScreenTool) Schema() json.RawMessage {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"action": map[string]any{
				"type":        "string",
				"description": "Action: capture (screenshot), ocr (capture + extract text), info (screen dimensions)",
				"enum":        []string{"capture", "ocr", "info"},
			},
			"target": map[string]any{
				"type":        "string",
				"description": "What to capture: screen (full), window (frontmost), region, or display:<n> for specific display",
				"enum":        []string{"screen", "window", "region", "display:0", "display:1", "display:2"},
			},
			"x": map[string]any{
				"type":        "integer",
				"description": "X coordinate for region capture",
			},
			"y": map[string]any{
				"type":        "integer",
				"description": "Y coordinate for region capture",
			},
			"width": map[string]any{
				"type":        "integer",
				"description": "Width for region capture",
			},
			"height": map[string]any{
				"type":        "integer",
				"description": "Height for region capture",
			},
			"output": map[string]any{
				"type":        "string",
				"description": "Output format: base64 (default), path (save to temp file and return path)",
				"enum":        []string{"base64", "path"},
			},
			"delay": map[string]any{
				"type":        "number",
				"description": "Delay in seconds before capture (default: 0)",
			},
		},
		"required": []string{"action"},
	}
	data, _ := json.Marshal(schema)
	return data
}

type ScreenInput struct {
	Action string  `json:"action"`
	Target string  `json:"target"`
	X      int     `json:"x"`
	Y      int     `json:"y"`
	Width  int     `json:"width"`
	Height int     `json:"height"`
	Output string  `json:"output"`
	Delay  float64 `json:"delay"`
}

func (t *ScreenTool) Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
	var in ScreenInput
	if err := json.Unmarshal(input, &in); err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Failed to parse input: %v", err),
			IsError: true,
		}, nil
	}

	// Set defaults
	if in.Target == "" {
		in.Target = "screen"
	}
	if in.Output == "" {
		in.Output = "base64"
	}

	var result string
	var err error

	switch in.Action {
	case "capture":
		result, err = t.capture(in)
	case "ocr":
		result, err = t.captureWithOCR(in)
	case "info":
		result, err = t.getScreenInfo()
	default:
		return &ToolResult{
			Content: fmt.Sprintf("Unknown action: %s", in.Action),
			IsError: true,
		}, nil
	}

	if err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Action failed: %v", err),
			IsError: true,
		}, nil
	}

	return &ToolResult{
		Content: result,
		IsError: false,
	}, nil
}

func (t *ScreenTool) capture(in ScreenInput) (string, error) {
	// Apply delay if specified
	if in.Delay > 0 {
		time.Sleep(time.Duration(in.Delay * float64(time.Second)))
	}

	// Create temp file for screenshot
	tmpFile := filepath.Join(t.tempDir, fmt.Sprintf("screenshot_%d.png", time.Now().UnixNano()))

	// Build screencapture command
	args := []string{"-x"} // -x = no sound

	switch {
	case in.Target == "window":
		args = append(args, "-l")
		// Get frontmost window ID
		windowID, err := t.getFrontmostWindowID()
		if err != nil {
			// Fallback to interactive window selection
			args = []string{"-x", "-w"}
		} else {
			args = append(args, windowID)
		}
	case in.Target == "region":
		if in.Width <= 0 || in.Height <= 0 {
			return "", fmt.Errorf("width and height required for region capture")
		}
		args = append(args, "-R", fmt.Sprintf("%d,%d,%d,%d", in.X, in.Y, in.Width, in.Height))
	case strings.HasPrefix(in.Target, "display:"):
		displayNum := strings.TrimPrefix(in.Target, "display:")
		args = append(args, "-D", displayNum)
	}
	// "screen" is default (full screen), no special args needed

	args = append(args, tmpFile)

	cmd := exec.Command("screencapture", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("screencapture failed: %v, output: %s", err, string(out))
	}

	// Check if file was created
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		return "", fmt.Errorf("screenshot file was not created")
	}
	defer func() {
		if in.Output == "base64" {
			os.Remove(tmpFile)
		}
	}()

	if in.Output == "path" {
		return fmt.Sprintf("Screenshot saved to: %s", tmpFile), nil
	}

	// Read and encode as base64
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		return "", fmt.Errorf("failed to read screenshot: %v", err)
	}

	// Get image dimensions
	img, err := png.Decode(bytes.NewReader(data))
	var dimensions string
	if err == nil {
		bounds := img.Bounds()
		dimensions = fmt.Sprintf("%dx%d", bounds.Dx(), bounds.Dy())
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return fmt.Sprintf("Screenshot captured (%s)\nBase64 length: %d\nData: data:image/png;base64,%s",
		dimensions, len(encoded), encoded), nil
}

func (t *ScreenTool) captureWithOCR(in ScreenInput) (string, error) {
	// First capture the screenshot
	in.Output = "path" // We need the path for OCR
	pathResult, err := t.capture(in)
	if err != nil {
		return "", err
	}

	// Extract path from result
	tmpFile := strings.TrimPrefix(pathResult, "Screenshot saved to: ")
	defer os.Remove(tmpFile)

	// Try macOS Vision framework via shortcuts/swift
	text, err := t.extractTextVision(tmpFile)
	if err != nil {
		// Try tesseract as fallback
		text, err = t.extractTextTesseract(tmpFile)
		if err != nil {
			return "", fmt.Errorf("OCR failed (Vision: %v, Tesseract not available or failed)", err)
		}
	}

	if text == "" {
		return "No text detected in screenshot", nil
	}

	return fmt.Sprintf("Extracted text from screenshot:\n\n%s", text), nil
}

func (t *ScreenTool) extractTextVision(imagePath string) (string, error) {
	// Use macOS Shortcuts or swift to access Vision framework
	// This uses a small inline Swift script
	swiftCode := fmt.Sprintf(`
import Vision
import AppKit

let url = URL(fileURLWithPath: "%s")
guard let image = NSImage(contentsOf: url),
      let cgImage = image.cgImage(forProposedRect: nil, context: nil, hints: nil) else {
    print("Error: Could not load image")
    exit(1)
}

let request = VNRecognizeTextRequest { request, error in
    guard let observations = request.results as? [VNRecognizedTextObservation] else { return }
    for observation in observations {
        if let topCandidate = observation.topCandidates(1).first {
            print(topCandidate.string)
        }
    }
}
request.recognitionLevel = .accurate
request.usesLanguageCorrection = true

let handler = VNImageRequestHandler(cgImage: cgImage, options: [:])
try? handler.perform([request])
`, imagePath)

	// Write Swift code to temp file
	swiftFile := filepath.Join(t.tempDir, "ocr.swift")
	if err := os.WriteFile(swiftFile, []byte(swiftCode), 0644); err != nil {
		return "", err
	}
	defer os.Remove(swiftFile)

	// Execute Swift
	out, err := exec.Command("swift", swiftFile).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("swift OCR failed: %v, output: %s", err, string(out))
	}

	return strings.TrimSpace(string(out)), nil
}

func (t *ScreenTool) extractTextTesseract(imagePath string) (string, error) {
	// Check if tesseract is installed
	if _, err := exec.LookPath("tesseract"); err != nil {
		return "", fmt.Errorf("tesseract not installed")
	}

	out, err := exec.Command("tesseract", imagePath, "stdout", "-l", "eng").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tesseract failed: %v", err)
	}

	return strings.TrimSpace(string(out)), nil
}

func (t *ScreenTool) getFrontmostWindowID() (string, error) {
	script := `
		tell application "System Events"
			set frontApp to name of first application process whose frontmost is true
			tell process frontApp
				set winID to id of window 1
				return winID as string
			end tell
		end tell
	`

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func (t *ScreenTool) getScreenInfo() (string, error) {
	// Get screen info using system_profiler or AppleScript
	script := `
		tell application "Finder"
			set screenBounds to bounds of window of desktop
			set screenWidth to item 3 of screenBounds
			set screenHeight to item 4 of screenBounds
		end tell

		-- Also try to get all displays
		set displayInfo to do shell script "system_profiler SPDisplaysDataType 2>/dev/null | grep -E 'Resolution|Display Type|Retina' | head -20"

		return "Primary: " & screenWidth & "x" & screenHeight & "
" & displayInfo
	`

	out, err := exec.Command("osascript", "-e", script).Output()
	if err != nil {
		// Fallback to system_profiler only
		out2, err2 := exec.Command("system_profiler", "SPDisplaysDataType").Output()
		if err2 != nil {
			return "", fmt.Errorf("failed to get screen info: %v", err)
		}
		return t.parseDisplayInfo(string(out2)), nil
	}

	return strings.TrimSpace(string(out)), nil
}

func (t *ScreenTool) parseDisplayInfo(info string) string {
	lines := strings.Split(info, "\n")
	var result []string
	var currentDisplay string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, ":") && !strings.Contains(line, "Graphics") {
			currentDisplay = strings.TrimSuffix(line, ":")
		}
		if strings.Contains(line, "Resolution:") {
			result = append(result, fmt.Sprintf("%s: %s", currentDisplay, line))
		}
		if strings.Contains(line, "Retina:") || strings.Contains(line, "Display Type:") {
			result = append(result, fmt.Sprintf("  %s", line))
		}
	}

	if len(result) == 0 {
		return "Could not parse display information"
	}

	return strings.Join(result, "\n")
}

func (t *ScreenTool) RequiresApproval() bool {
	return false // Screenshot is relatively safe for tool use
}

// Unused but needed to avoid import error
var _ = strconv.Itoa
var _ image.Image = nil

// RPC Server implementation
type ToolRPCServer struct {
	Impl *ScreenTool
}

func (s *ToolRPCServer) Name(_ struct{}, resp *string) error {
	*resp = s.Impl.Name()
	return nil
}

func (s *ToolRPCServer) Description(_ struct{}, resp *string) error {
	*resp = s.Impl.Description()
	return nil
}

func (s *ToolRPCServer) Schema(_ struct{}, resp *json.RawMessage) error {
	*resp = s.Impl.Schema()
	return nil
}

type ExecuteArgs struct {
	Input json.RawMessage
}

type ExecuteReply struct {
	Result *ToolResult
	Error  string
}

func (s *ToolRPCServer) Execute(args ExecuteArgs, reply *ExecuteReply) error {
	result, err := s.Impl.Execute(context.Background(), args.Input)
	reply.Result = result
	if err != nil {
		reply.Error = err.Error()
	}
	return nil
}

func (s *ToolRPCServer) RequiresApproval(_ struct{}, resp *bool) error {
	*resp = s.Impl.RequiresApproval()
	return nil
}

type ToolPlugin struct {
	Impl *ScreenTool
}

func (p *ToolPlugin) Server(*plugin.MuxBroker) (any, error) {
	return &ToolRPCServer{Impl: p.Impl}, nil
}

func (p *ToolPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (any, error) {
	return nil, fmt.Errorf("client not implemented")
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"tool": &ToolPlugin{Impl: NewScreenTool()},
		},
	})
}
