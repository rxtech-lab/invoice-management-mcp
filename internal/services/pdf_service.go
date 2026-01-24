package services

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/microcosm-cc/bluemonday"
)

// PDFService handles HTML to PDF conversion
type PDFService interface {
	ConvertHTMLToPDF(ctx context.Context, html string, options PDFOptions) ([]byte, error)
}

// PDFOptions contains options for PDF generation
type PDFOptions struct {
	PaperWidth   float64 // in inches
	PaperHeight  float64 // in inches
	MarginTop    float64 // in inches
	MarginBottom float64 // in inches
	MarginLeft   float64 // in inches
	MarginRight  float64 // in inches
	Landscape    bool
}

// DefaultPDFOptions returns sensible defaults (US Letter)
func DefaultPDFOptions() PDFOptions {
	return PDFOptions{
		PaperWidth:   8.5,
		PaperHeight:  11,
		MarginTop:    0.4,
		MarginBottom: 0.4,
		MarginLeft:   0.4,
		MarginRight:  0.4,
		Landscape:    false,
	}
}

// PDFServiceConfig holds configuration for the PDF service
type PDFServiceConfig struct {
	ChromeURL string        // WebSocket URL to headless Chrome (e.g., ws://localhost:9222)
	Timeout   time.Duration // Request timeout (default: 30s)
}

type pdfService struct {
	config    PDFServiceConfig
	sanitizer *bluemonday.Policy
}

// NewPDFService creates a new PDF service
func NewPDFService(config PDFServiceConfig) PDFService {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Create a sanitizer policy that allows safe HTML for rendering
	// UGCPolicy allows common HTML elements but strips scripts, iframes, etc.
	sanitizer := bluemonday.UGCPolicy()
	// Allow additional elements commonly used in invoices/documents
	sanitizer.AllowElements("style", "table", "thead", "tbody", "tfoot", "tr", "th", "td", "colgroup", "col")
	sanitizer.AllowAttrs("style").Globally()
	sanitizer.AllowAttrs("class", "id").Globally()
	sanitizer.AllowAttrs("colspan", "rowspan", "width", "height", "align", "valign").OnElements("td", "th", "col")

	return &pdfService{
		config:    config,
		sanitizer: sanitizer,
	}
}

// ConvertHTMLToPDF converts HTML to PDF using headless Chrome
func (s *pdfService) ConvertHTMLToPDF(ctx context.Context, html string, options PDFOptions) ([]byte, error) {
	// Sanitize HTML to prevent XSS and other attacks
	sanitizedHTML := s.sanitizer.Sanitize(html)

	// Wrap in basic HTML structure if not already a complete document
	if !containsHTMLTag(sanitizedHTML) {
		sanitizedHTML = fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<style>
@import url('https://fonts.googleapis.com/css2?family=Noto+Sans+SC:wght@400;500;700&display=swap');
body { font-family: 'Noto Sans SC', Arial, sans-serif; }
</style>
</head>
<body>
%s
</body>
</html>`, sanitizedHTML)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.config.Timeout)
	defer cancel()

	// Connect to remote Chrome instance
	allocatorCtx, allocatorCancel := chromedp.NewRemoteAllocator(ctx, s.config.ChromeURL)
	defer allocatorCancel()

	// Create a new browser context
	taskCtx, taskCancel := chromedp.NewContext(allocatorCtx)
	defer taskCancel()

	var pdfContent []byte

	// Run the PDF generation
	err := chromedp.Run(taskCtx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, sanitizedHTML).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfContent, _, err = page.PrintToPDF().
				WithPaperWidth(options.PaperWidth).
				WithPaperHeight(options.PaperHeight).
				WithMarginTop(options.MarginTop).
				WithMarginBottom(options.MarginBottom).
				WithMarginLeft(options.MarginLeft).
				WithMarginRight(options.MarginRight).
				WithLandscape(options.Landscape).
				WithPrintBackground(true).
				Do(ctx)
			return err
		}),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return pdfContent, nil
}

// containsHTMLTag checks if the string contains an html tag
func containsHTMLTag(s string) bool {
	for i := 0; i < len(s)-5; i++ {
		if s[i] == '<' && (s[i+1] == 'h' || s[i+1] == 'H') &&
			(s[i+2] == 't' || s[i+2] == 'T') &&
			(s[i+3] == 'm' || s[i+3] == 'M') &&
			(s[i+4] == 'l' || s[i+4] == 'L') {
			return true
		}
	}
	return false
}

// MockPDFService is a mock implementation for testing
type MockPDFService struct{}

// NewMockPDFService creates a new mock PDF service
func NewMockPDFService() PDFService {
	return &MockPDFService{}
}

// ConvertHTMLToPDF returns a minimal valid PDF for testing
func (m *MockPDFService) ConvertHTMLToPDF(ctx context.Context, html string, options PDFOptions) ([]byte, error) {
	// Return a minimal valid PDF for testing
	return []byte("%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\n3 0 obj<</Type/Page/MediaBox[0 0 612 792]/Parent 2 0 R>>endobj\nxref\n0 4\n0000000000 65535 f \n0000000009 00000 n \n0000000052 00000 n \n0000000101 00000 n \ntrailer<</Size 4/Root 1 0 R>>\nstartxref\n170\n%%EOF"), nil
}
