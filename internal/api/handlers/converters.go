package handlers

import (
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rxtech-lab/invoice-management/internal/api/generated"
	"github.com/rxtech-lab/invoice-management/internal/models"
	"github.com/rxtech-lab/invoice-management/internal/services"
)

// Category converters

func categoryModelToGenerated(cat *models.InvoiceCategory) generated.Category {
	return generated.Category{
		Id:          ptr(int(cat.ID)),
		UserId:      ptr(cat.UserID),
		Name:        ptr(cat.Name),
		Description: ptr(cat.Description),
		Color:       ptr(cat.Color),
		CreatedAt:   ptr(cat.CreatedAt),
		UpdatedAt:   ptr(cat.UpdatedAt),
	}
}

func categoryListToGenerated(categories []models.InvoiceCategory) []generated.Category {
	result := make([]generated.Category, len(categories))
	for i, cat := range categories {
		result[i] = categoryModelToGenerated(&cat)
	}
	return result
}

// Company converters

func companyModelToGenerated(comp *models.InvoiceCompany) generated.Company {
	var email *openapi_types.Email
	if comp.Email != "" {
		e := openapi_types.Email(comp.Email)
		email = &e
	}

	return generated.Company{
		Id:        ptr(int(comp.ID)),
		UserId:    ptr(comp.UserID),
		Name:      ptr(comp.Name),
		Address:   ptr(comp.Address),
		Email:     email,
		Phone:     ptr(comp.Phone),
		Website:   ptr(comp.Website),
		TaxId:     ptr(comp.TaxID),
		Notes:     ptr(comp.Notes),
		CreatedAt: ptr(comp.CreatedAt),
		UpdatedAt: ptr(comp.UpdatedAt),
	}
}

func companyListToGenerated(companies []models.InvoiceCompany) []generated.Company {
	result := make([]generated.Company, len(companies))
	for i, comp := range companies {
		result[i] = companyModelToGenerated(&comp)
	}
	return result
}

// Receiver converters

func receiverModelToGenerated(rec *models.InvoiceReceiver) generated.Receiver {
	return generated.Receiver{
		Id:             ptr(int(rec.ID)),
		UserId:         ptr(rec.UserID),
		Name:           ptr(rec.Name),
		IsOrganization: ptr(rec.IsOrganization),
		CreatedAt:      ptr(rec.CreatedAt),
		UpdatedAt:      ptr(rec.UpdatedAt),
	}
}

func receiverListToGenerated(receivers []models.InvoiceReceiver) []generated.Receiver {
	result := make([]generated.Receiver, len(receivers))
	for i, rec := range receivers {
		result[i] = receiverModelToGenerated(&rec)
	}
	return result
}

// Invoice converters

func invoiceModelToGenerated(inv *models.Invoice) generated.Invoice {
	var categoryID, companyID, receiverID *int
	var category *generated.Category
	var company *generated.Company
	var receiver *generated.Receiver

	if inv.CategoryID != nil {
		categoryID = ptr(int(*inv.CategoryID))
	}
	if inv.CompanyID != nil {
		companyID = ptr(int(*inv.CompanyID))
	}
	if inv.ReceiverID != nil {
		receiverID = ptr(int(*inv.ReceiverID))
	}

	if inv.Category != nil {
		cat := categoryModelToGenerated(inv.Category)
		category = &cat
	}
	if inv.Company != nil {
		comp := companyModelToGenerated(inv.Company)
		company = &comp
	}
	if inv.Receiver != nil {
		rec := receiverModelToGenerated(inv.Receiver)
		receiver = &rec
	}

	// Convert items
	var items *[]generated.InvoiceItem
	if len(inv.Items) > 0 {
		itemList := invoiceItemListToGenerated(inv.Items)
		items = &itemList
	}

	// Convert tags - include id and name
	var tags *[]generated.InvoiceTagReference
	if len(inv.Tags) > 0 {
		t := make([]generated.InvoiceTagReference, len(inv.Tags))
		for i, tag := range inv.Tags {
			t[i] = generated.InvoiceTagReference{
				Id:   int(tag.ID),
				Name: tag.Name,
			}
		}
		tags = &t
	}

	// Convert status
	var status *generated.InvoiceStatus
	if inv.Status != "" {
		s := generated.InvoiceStatus(inv.Status)
		status = &s
	}

	return generated.Invoice{
		Id:                   ptr(int(inv.ID)),
		UserId:               ptr(inv.UserID),
		Title:                ptr(inv.Title),
		Description:          ptr(inv.Description),
		InvoiceStartedAt:     inv.InvoiceStartedAt,
		InvoiceEndedAt:       inv.InvoiceEndedAt,
		Amount:               ptr(inv.Amount),
		Currency:             ptr(inv.Currency),
		CategoryId:           categoryID,
		Category:             category,
		CompanyId:            companyID,
		Company:              company,
		ReceiverId:           receiverID,
		Receiver:             receiver,
		Items:                items,
		OriginalDownloadLink: ptr(inv.OriginalDownloadLink),
		Tags:                 tags,
		Status:               status,
		DueDate:              inv.DueDate,
		CreatedAt:            ptr(inv.CreatedAt),
		UpdatedAt:            ptr(inv.UpdatedAt),
	}
}

func invoiceListToGenerated(invoices []models.Invoice) []generated.Invoice {
	result := make([]generated.Invoice, len(invoices))
	for i, inv := range invoices {
		result[i] = invoiceModelToGenerated(&inv)
	}
	return result
}

// InvoiceItem converters

func invoiceItemModelToGenerated(item *models.InvoiceItem) generated.InvoiceItem {
	return generated.InvoiceItem{
		Id:          ptr(int(item.ID)),
		InvoiceId:   ptr(int(item.InvoiceID)),
		Description: ptr(item.Description),
		Quantity:    ptr(item.Quantity),
		UnitPrice:   ptr(item.UnitPrice),
		Amount:      ptr(item.Amount),
		CreatedAt:   ptr(item.CreatedAt),
		UpdatedAt:   ptr(item.UpdatedAt),
	}
}

func invoiceItemListToGenerated(items []models.InvoiceItem) []generated.InvoiceItem {
	result := make([]generated.InvoiceItem, len(items))
	for i, item := range items {
		result[i] = invoiceItemModelToGenerated(&item)
	}
	return result
}

// Analytics converters

func analyticsSummaryToGenerated(summary *services.AnalyticsSummary) generated.AnalyticsSummary {
	return generated.AnalyticsSummary{
		Period:        ptr(summary.Period),
		StartDate:     ptr(summary.StartDate),
		EndDate:       ptr(summary.EndDate),
		TotalAmount:   ptr(summary.TotalAmount),
		PaidAmount:    ptr(summary.PaidAmount),
		UnpaidAmount:  ptr(summary.UnpaidAmount),
		OverdueAmount: ptr(summary.OverdueAmount),
		InvoiceCount:  ptr(int(summary.InvoiceCount)),
		PaidCount:     ptr(int(summary.PaidCount)),
		UnpaidCount:   ptr(int(summary.UnpaidCount)),
		OverdueCount:  ptr(int(summary.OverdueCount)),
	}
}

func analyticsGroupItemToGenerated(item *services.AnalyticsGroupItem) generated.AnalyticsGroupItem {
	return generated.AnalyticsGroupItem{
		Id:           ptr(int(item.ID)),
		Name:         ptr(item.Name),
		Color:        ptr(item.Color),
		TotalAmount:  ptr(item.TotalAmount),
		PaidAmount:   ptr(item.PaidAmount),
		UnpaidAmount: ptr(item.UnpaidAmount),
		InvoiceCount: ptr(int(item.InvoiceCount)),
	}
}

func analyticsByGroupToGenerated(group *services.AnalyticsByGroup) generated.AnalyticsByGroup {
	items := make([]generated.AnalyticsGroupItem, len(group.Items))
	for i, item := range group.Items {
		items[i] = analyticsGroupItemToGenerated(&item)
	}

	var uncategorized *generated.AnalyticsGroupItem
	if group.Uncategorized != nil {
		u := analyticsGroupItemToGenerated(group.Uncategorized)
		uncategorized = &u
	}

	return generated.AnalyticsByGroup{
		Period:        ptr(group.Period),
		StartDate:     ptr(group.StartDate),
		EndDate:       ptr(group.EndDate),
		Items:         &items,
		Uncategorized: uncategorized,
	}
}

// Period converter
func periodParamToService(period string) services.AnalyticsPeriod {
	switch period {
	case "7d":
		return services.Period7Days
	case "1y":
		return services.Period1Year
	default:
		return services.Period1Month
	}
}
