// Invoice Status
export type InvoiceStatus = "paid" | "unpaid" | "overdue";

// Category
export interface Category {
  id: number;
  user_id: string;
  name: string;
  description: string;
  color: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCategoryRequest {
  name: string;
  description?: string;
  color?: string;
}

export interface UpdateCategoryRequest {
  name?: string;
  description?: string;
  color?: string;
}

// Company
export interface Company {
  id: number;
  user_id: string;
  name: string;
  address: string;
  email: string;
  phone: string;
  website: string;
  tax_id: string;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface CreateCompanyRequest {
  name: string;
  address?: string;
  email?: string;
  phone?: string;
  website?: string;
  tax_id?: string;
  notes?: string;
}

export interface UpdateCompanyRequest {
  name?: string;
  address?: string;
  email?: string;
  phone?: string;
  website?: string;
  tax_id?: string;
  notes?: string;
}

// Receiver
export interface Receiver {
  id: number;
  user_id: string;
  name: string;
  other_names?: string[];
  is_organization: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateReceiverRequest {
  name: string;
  is_organization?: boolean;
}

export interface UpdateReceiverRequest {
  name?: string;
  other_names?: string[];
  is_organization?: boolean;
}

export interface MergeReceiversRequest {
  target_id: number;
  source_ids: number[];
}

export interface MergeReceiversResponse {
  receiver: Receiver;
  merged_count: number;
  invoices_updated: number;
}

// Tag
export interface Tag {
  id: number;
  user_id: string;
  name: string;
  color: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTagRequest {
  name: string;
  color?: string;
}

export interface UpdateTagRequest {
  name?: string;
  color?: string;
}

export interface TagListOptions {
  keyword?: string;
  limit?: number;
  offset?: number;
}

// Invoice Item
export interface InvoiceItem {
  id: number;
  invoice_id: number;
  description: string;
  quantity: number;
  unit_price: number;
  amount: number;
  // Currency conversion fields (for USD normalization)
  target_currency: string;
  target_amount: number;
  fx_rate_used: number;
  created_at: string;
  updated_at: string;
}

export interface CreateInvoiceItemRequest {
  description: string;
  quantity?: number;
  unit_price?: number;
}

export interface UpdateInvoiceItemRequest {
  description?: string;
  quantity?: number;
  unit_price?: number;
  // Manual override for USD amount (optional, auto-calculated if not provided)
  target_amount?: number;
  // When true, forces recalculation of target_amount using latest FX rate
  auto_calculate_target_currency?: boolean;
}

// Invoice
export interface Invoice {
  id: number;
  user_id: string;
  title: string;
  description: string;
  invoice_started_at: string | null;
  invoice_ended_at: string | null;
  amount: number;
  currency: string;
  // USD-normalized total (computed as sum of item target_amounts)
  target_amount: number;
  category_id: number | null;
  category?: Category;
  company_id: number | null;
  company?: Company;
  receiver_id: number | null;
  receiver?: Receiver;
  items: InvoiceItem[];
  original_download_link: string;
  tags: { id: number; name: string; color?: string }[];
  status: InvoiceStatus;
  due_date: string | null;
  created_at: string;
  updated_at: string;
}

// Note: amount is not included - it's calculated from invoice items
export interface CreateInvoiceRequest {
  title: string;
  description?: string;
  invoice_started_at?: string;
  invoice_ended_at?: string;
  currency?: string;
  category_id?: number;
  company_id?: number;
  receiver_id?: number;
  original_download_link?: string;
  tag_ids?: number[];
  status?: InvoiceStatus;
  due_date?: string;
  items?: CreateInvoiceItemRequest[];
}

// Note: amount is not included - it's calculated from invoice items
// target_amount is computed from item target_amounts (not editable)
export interface UpdateInvoiceRequest {
  title?: string;
  description?: string;
  invoice_started_at?: string;
  invoice_ended_at?: string;
  currency?: string;
  category_id?: number;
  company_id?: number;
  receiver_id?: number;
  original_download_link?: string;
  tag_ids?: number[];
  status?: InvoiceStatus;
  due_date?: string;
}

export interface UpdateStatusRequest {
  status: InvoiceStatus;
}

// Upload
export interface UploadResponse {
  key: string;
  filename: string;
  size: number;
  content_type: string;
}

export interface PresignedURLResponse {
  upload_url: string;
  key: string;
  content_type: string;
}

export interface ConfirmUploadRequest {
  key: string;
  filename: string;
  content_type: string;
  size: number;
}

export interface FileDownloadURLResponse {
  download_url: string;
  key: string;
  filename: string;
  expires_at: string;
}

// Paginated Response
export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  limit: number;
  offset: number;
}

// List Options
export interface InvoiceListOptions {
  keyword?: string;
  category_id?: number;
  company_id?: number;
  receiver_id?: number;
  tag_ids?: number[];
  status?: InvoiceStatus;
  sort_by?: "created_at" | "updated_at" | "amount" | "due_date" | "title";
  sort_order?: "asc" | "desc";
  limit?: number;
  offset?: number;
}

export interface CategoryListOptions {
  keyword?: string;
  limit?: number;
  offset?: number;
}

export interface CompanyListOptions {
  keyword?: string;
  limit?: number;
  offset?: number;
}

export interface ReceiverListOptions {
  keyword?: string;
  limit?: number;
  offset?: number;
}

// API Error
export interface ApiError {
  error: string;
}

// Analytics Types
export type AnalyticsPeriod = "7d" | "1m" | "1y";

export interface AnalyticsSummary {
  period: string;
  start_date: string;
  end_date: string;
  total_amount: number;
  paid_amount: number;
  unpaid_amount: number;
  overdue_amount: number;
  invoice_count: number;
  paid_count: number;
  unpaid_count: number;
  overdue_count: number;
}

export interface AnalyticsGroupItem {
  id: number;
  name: string;
  color?: string;
  total_amount: number;
  paid_amount: number;
  unpaid_amount: number;
  invoice_count: number;
}

export interface AnalyticsByGroup {
  period: string;
  start_date: string;
  end_date: string;
  items: AnalyticsGroupItem[];
  uncategorized?: AnalyticsGroupItem;
}
