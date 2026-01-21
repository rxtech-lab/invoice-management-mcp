package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
)

// InvoiceStatisticsTool handles invoice statistics queries
type InvoiceStatisticsTool struct {
	service services.AnalyticsService
}

func NewInvoiceStatisticsTool(service services.AnalyticsService) *InvoiceStatisticsTool {
	return &InvoiceStatisticsTool{service: service}
}

func (t *InvoiceStatisticsTool) GetTool() mcp.Tool {
	return mcp.NewTool("invoice_statistics",
		mcp.WithDescription(`Get invoice statistics with time period filtering, grouping, and aggregations.

EXAMPLE QUERIES:
- "How much did I spend last week?" → invoice_statistics(period: "last_week")
- "Show daily spending for 7 days (bar chart data)" → invoice_statistics(period: "last_week", group_by: "day")
- "What is the max spend last week?" → invoice_statistics(period: "last_week", include_aggregations: true)
- "Show electricity invoices last month" → invoice_statistics(period: "last_month", keyword: "electricity")
- "Compare spending by category" → invoice_statistics(period: "last_month", group_by: "category")
- "Which company did I pay most to?" → invoice_statistics(period: "last_year", group_by: "company")
- "Daily electricity costs last month" → invoice_statistics(period: "last_month", keyword: "electricity", group_by: "day")
- "Highest electricity bill last year" → invoice_statistics(period: "last_year", keyword: "electricity", include_aggregations: true)

PERIODS: last_day, last_week, last_month, last_year, or custom days
GROUPING: day (for charts), week, month, category, company, receiver
FILTERS: category_id, company_id, receiver_id, status (paid/unpaid/overdue), keyword`),
		mcp.WithString("period", mcp.Description("Natural time period: 'last_day', 'last_week', 'last_month', 'last_year'. Default: 'last_month'")),
		mcp.WithNumber("days", mcp.Description("Custom days lookback (e.g., 90 for last 90 days)")),
		mcp.WithNumber("category_id", mcp.Description("Filter by category ID")),
		mcp.WithNumber("company_id", mcp.Description("Filter by company ID")),
		mcp.WithNumber("receiver_id", mcp.Description("Filter by receiver ID")),
		mcp.WithString("status", mcp.Description("Filter by status: 'paid', 'unpaid', 'overdue'")),
		mcp.WithString("keyword", mcp.Description("Search keyword for title/description (e.g., 'electricity', 'consulting')")),
		mcp.WithString("group_by", mcp.Description("Group results by: 'day' (for charts), 'week', 'month', 'category', 'company', 'receiver'")),
		mcp.WithBoolean("include_aggregations", mcp.Description("Include min/max/avg amounts and references to max invoice (default: false)")),
	)
}

func (t *InvoiceStatisticsTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)

		// Build options
		opts := services.StatisticsOptions{
			Period:              services.PeriodLastMonth, // Default
			CategoryID:          getUintPtrArg(args, "category_id"),
			CompanyID:           getUintPtrArg(args, "company_id"),
			ReceiverID:          getUintPtrArg(args, "receiver_id"),
			Keyword:             getStringArg(args, "keyword"),
			IncludeAggregations: getBoolArg(args, "include_aggregations", false),
		}

		// Handle status parameter
		if statusStr := getStringArg(args, "status"); statusStr != "" {
			switch statusStr {
			case "paid":
				status := models.InvoiceStatusPaid
				opts.Status = &status
			case "unpaid":
				status := models.InvoiceStatusUnpaid
				opts.Status = &status
			case "overdue":
				status := models.InvoiceStatusOverdue
				opts.Status = &status
			default:
				return mcp.NewToolResultError(fmt.Sprintf("Invalid status '%s'. Valid values: paid, unpaid, overdue", statusStr)), nil
			}
		}

		// Handle period parameter
		periodStr := getStringArg(args, "period")
		if periodStr != "" {
			switch periodStr {
			case "last_day":
				opts.Period = services.PeriodLastDay
			case "last_week":
				opts.Period = services.PeriodLastWeek
			case "last_month":
				opts.Period = services.PeriodLastMonth
			case "last_year":
				opts.Period = services.PeriodLastYear
			default:
				return mcp.NewToolResultError(fmt.Sprintf("Invalid period '%s'. Valid values: last_day, last_week, last_month, last_year", periodStr)), nil
			}
		} else if days := getIntArg(args, "days", 0); days > 0 {
			// Use custom days if period not specified
			opts.Period = services.PeriodCustom
			opts.Days = days
		}

		// Handle group_by parameter
		groupByStr := getStringArg(args, "group_by")
		if groupByStr != "" {
			switch groupByStr {
			case "day":
				opts.GroupBy = services.GroupByDay
			case "week":
				opts.GroupBy = services.GroupByWeek
			case "month":
				opts.GroupBy = services.GroupByMonth
			case "category":
				opts.GroupBy = services.GroupByCategory
			case "company":
				opts.GroupBy = services.GroupByCompany
			case "receiver":
				opts.GroupBy = services.GroupByReceiver
			default:
				return mcp.NewToolResultError(fmt.Sprintf("Invalid group_by '%s'. Valid values: day, week, month, category, company, receiver", groupByStr)), nil
			}
		}

		stats, err := t.service.GetStatistics(userID, opts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get statistics: %v", err)), nil
		}

		result, _ := json.Marshal(stats)
		return mcp.NewToolResultText(string(result)), nil
	}
}

// getBoolArg extracts a boolean argument with a default value
func getBoolArg(args map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return defaultVal
}
