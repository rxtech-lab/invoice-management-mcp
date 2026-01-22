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

// AdvancedInvoiceSearchTool handles searching invoices across multiple fields
type AdvancedInvoiceSearchTool struct {
	analyticsService services.AnalyticsService
	invoiceService   services.InvoiceService
	tagService       services.TagService
	categoryService  services.CategoryService
	companyService   services.CompanyService
	receiverService  services.ReceiverService
}

func NewAdvancedInvoiceSearchTool(
	analyticsService services.AnalyticsService,
	invoiceService services.InvoiceService,
	tagService services.TagService,
	categoryService services.CategoryService,
	companyService services.CompanyService,
	receiverService services.ReceiverService,
) *AdvancedInvoiceSearchTool {
	return &AdvancedInvoiceSearchTool{
		analyticsService: analyticsService,
		invoiceService:   invoiceService,
		tagService:       tagService,
		categoryService:  categoryService,
		companyService:   companyService,
		receiverService:  receiverService,
	}
}

func (t *AdvancedInvoiceSearchTool) GetTool() mcp.Tool {
	return mcp.NewTool("advanced_invoice_search",
		mcp.WithDescription(`Search invoices across multiple fields and get aggregated spending.

EXAMPLE QUERIES:
- "How much did I spend on Marriott hotel?" → keyword: "Marriott"
- "Total travel expenses last month" → tag_names: ["travel"], period: "last_month"
- "AWS bills in 2024" → keyword: "AWS"
- "Show all food and dining expenses" → category_name: "food"
- "What did I pay to John?" → receiver_name: "John"

Searches across: invoice title, description, category name, company name, receiver name, tag names.
Returns: matched invoices, total amount, aggregation stats (min/max/avg).`),
		mcp.WithString("keyword", mcp.Description("Search keyword for title, description")),
		mcp.WithString("category_name", mcp.Description("Filter by category name (partial match)")),
		mcp.WithString("company_name", mcp.Description("Filter by company name (partial match)")),
		mcp.WithString("receiver_name", mcp.Description("Filter by receiver name (partial match)")),
		mcp.WithArray("tag_names", mcp.Description("Filter by tag names (array of strings)")),
		mcp.WithString("period", mcp.Description("Time period: 'last_week', 'last_month', 'last_year', or custom days. Default: 'last_month'")),
		mcp.WithNumber("days", mcp.Description("Custom days lookback")),
		mcp.WithNumber("limit", mcp.Description("Maximum invoices to return (default 50)")),
		mcp.WithNumber("offset", mcp.Description("Offset for pagination")),
		mcp.WithBoolean("group_by_day", mcp.Description("Group results by day for chart data")),
	)
}

func (t *AdvancedInvoiceSearchTool) GetHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		userID := getUserIDFromContext(ctx)
		if userID == "" {
			return mcp.NewToolResultError("Authentication required"), nil
		}

		args := getArgsMap(request.Params.Arguments)

		// Extract parameters
		keyword := getStringArg(args, "keyword")
		categoryName := getStringArg(args, "category_name")
		companyName := getStringArg(args, "company_name")
		receiverName := getStringArg(args, "receiver_name")
		limit := getIntArg(args, "limit", 50)
		offset := getIntArg(args, "offset", 0)
		groupByDay := getBoolArg(args, "group_by_day", false)

		// Extract tag names
		var tagNames []string
		if tagNamesRaw, ok := args["tag_names"].([]interface{}); ok {
			for _, v := range tagNamesRaw {
				if name, ok := v.(string); ok {
					tagNames = append(tagNames, name)
				}
			}
		}

		// Look up IDs by name
		var categoryID *uint
		if categoryName != "" {
			categories, _, err := t.categoryService.ListCategories(userID, categoryName, 1, 0)
			if err == nil && len(categories) > 0 {
				categoryID = &categories[0].ID
			}
		}

		var companyID *uint
		if companyName != "" {
			companies, _, err := t.companyService.ListCompanies(userID, companyName, 1, 0)
			if err == nil && len(companies) > 0 {
				companyID = &companies[0].ID
			}
		}

		var receiverID *uint
		if receiverName != "" {
			receivers, _, err := t.receiverService.ListReceivers(userID, receiverName, 1, 0)
			if err == nil && len(receivers) > 0 {
				receiverID = &receivers[0].ID
			}
		}

		// Look up tag IDs
		var tagIDs []uint
		for _, tagName := range tagNames {
			tags, _, err := t.tagService.ListTags(userID, tagName, 1, 0)
			if err == nil && len(tags) > 0 {
				tagIDs = append(tagIDs, tags[0].ID)
			}
		}

		// Build invoice list options
		invoiceOpts := services.InvoiceListOptions{
			Keyword:    keyword,
			CategoryID: categoryID,
			CompanyID:  companyID,
			ReceiverID: receiverID,
			TagIDs:     tagIDs,
			Limit:      limit,
			Offset:     offset,
			SortBy:     "due_date",
			SortOrder:  "desc",
		}

		// Get matching invoices
		invoices, total, err := t.invoiceService.ListInvoices(userID, invoiceOpts)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to search invoices: %v", err)), nil
		}

		// Calculate aggregations
		var totalAmount, minAmount, maxAmount float64
		var maxInvoiceID uint
		var maxInvoiceTitle string
		for i, inv := range invoices {
			totalAmount += inv.Amount
			if i == 0 || inv.Amount < minAmount {
				minAmount = inv.Amount
			}
			if inv.Amount > maxAmount {
				maxAmount = inv.Amount
				maxInvoiceID = inv.ID
				maxInvoiceTitle = inv.Title
			}
		}

		avgAmount := 0.0
		if len(invoices) > 0 {
			avgAmount = totalAmount / float64(len(invoices))
		}

		// Build response
		response := map[string]interface{}{
			"invoices":    invoices,
			"total_count": total,
			"limit":       limit,
			"offset":      offset,
			"aggregations": map[string]interface{}{
				"total_amount":      totalAmount,
				"min_amount":        minAmount,
				"max_amount":        maxAmount,
				"avg_amount":        avgAmount,
				"max_invoice_id":    maxInvoiceID,
				"max_invoice_title": maxInvoiceTitle,
			},
		}

		// If group by day is requested, get statistics
		if groupByDay {
			statsOpts := services.StatisticsOptions{
				Period:     services.PeriodLastMonth, // Default
				CategoryID: categoryID,
				CompanyID:  companyID,
				ReceiverID: receiverID,
				TagIDs:     tagIDs,
				Keyword:    keyword,
				GroupBy:    services.GroupByDay,
			}

			// Handle period parameter
			periodStr := getStringArg(args, "period")
			if periodStr != "" {
				switch periodStr {
				case "last_week":
					statsOpts.Period = services.PeriodLastWeek
				case "last_month":
					statsOpts.Period = services.PeriodLastMonth
				case "last_year":
					statsOpts.Period = services.PeriodLastYear
				}
			} else if days := getIntArg(args, "days", 0); days > 0 {
				statsOpts.Period = services.PeriodCustom
				statsOpts.Days = days
			}

			stats, err := t.analyticsService.GetStatistics(userID, statsOpts)
			if err == nil && stats != nil {
				response["daily_breakdown"] = stats.Breakdown
			}
		}

		result, _ := json.Marshal(response)
		return mcp.NewToolResultText(string(result)), nil
	}
}
