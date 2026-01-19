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

// Invoice Item
export interface InvoiceItem {
  id: number;
  invoice_id: number;
  description: string;
  quantity: number;
  unit_price: number;
  amount: number;
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
  category_id: number | null;
  category?: Category;
  company_id: number | null;
  company?: Company;
  items: InvoiceItem[];
  original_download_link: string;
  tags: string[];
  status: InvoiceStatus;
  due_date: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateInvoiceRequest {
  title: string;
  description?: string;
  invoice_started_at?: string;
  invoice_ended_at?: string;
  amount?: number;
  currency?: string;
  category_id?: number;
  company_id?: number;
  original_download_link?: string;
  tags?: string[];
  status?: InvoiceStatus;
  due_date?: string;
  items?: CreateInvoiceItemRequest[];
}

export interface UpdateInvoiceRequest {
  title?: string;
  description?: string;
  invoice_started_at?: string;
  invoice_ended_at?: string;
  amount?: number;
  currency?: string;
  category_id?: number;
  company_id?: number;
  original_download_link?: string;
  tags?: string[];
  status?: InvoiceStatus;
  due_date?: string;
}

export interface UpdateStatusRequest {
  status: InvoiceStatus;
}

// Upload
export interface UploadResponse {
  key: string;
  download_url: string;
  filename: string;
  size: number;
  content_type: string;
}

export interface PresignedURLResponse {
  upload_url: string;
  key: string;
  content_type: string;
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

// API Error
export interface ApiError {
  error: string;
}
