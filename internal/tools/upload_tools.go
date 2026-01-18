package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rxtech-lab/invoice-management/internal/services"
)

// GetPresignedURLTool handles getting presigned URLs for file uploads
type GetPresignedURLTool struct {
	service services.UploadService
}

func NewGetPresignedURLTool(service services.UploadService) *GetPresignedURLTool {
	return &GetPresignedURLTool{service: service}
}

func (t *GetPresignedURLTool) GetTool() mcp.Tool {
	return mcp.NewTool("get_presigned_url",
		mcp.WithDescription("Get a presigned URL for uploading a file to S3"),
		mcp.WithString("filename", mcp.Required(), mcp.Description("Name of the file to upload")),
		mcp.WithString("content_type", mcp.Description("MIME type of the file (default: application/octet-stream)")),
	)
}

func (t *GetPresignedURLTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)
		filename, _ := args["filename"].(string)
		if filename == "" {
			return mcp.NewToolResultError("filename is required"), nil
		}

		contentType, _ := args["content_type"].(string)
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		uploadURL, key, err := t.service.GetPresignedUploadURL(ctx, userID, filename, contentType)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get presigned URL: %v", err)), nil
		}

		result, _ := json.Marshal(map[string]interface{}{
			"upload_url":   uploadURL,
			"key":          key,
			"content_type": contentType,
			"instructions": "Use PUT request with the upload_url to upload the file. The 'key' can be used as original_download_link in invoices.",
		})
		return mcp.NewToolResultText(string(result)), nil
	}
}
