import type { InvoiceStatus } from "@/lib/api/types";

// Search result type for keyword search
export interface InvoiceSearchResult {
  id: number;
  title: string;
  description: string;
  amount: number;
  currency: string;
  status: InvoiceStatus;
  company?: { id: number; name: string };
  category?: { id: number; name: string; color?: string };
  receiver?: { id: number; name: string };
  due_date: string | null;
  created_at: string;
}

// Statistics types matching backend response
export interface StatusStats {
  count: number;
  amount: number;
}

export interface StatusBreakdown {
  paid: StatusStats;
  unpaid: StatusStats;
  overdue: StatusStats;
}

export interface BreakdownItem {
  date?: string;
  id?: number;
  name?: string;
  amount: number;
  count: number;
}

export interface InvoiceReference {
  id: number;
  title: string;
}

export interface DayReference {
  date: string;
  amount: number;
}

export interface EntityReference {
  id: number;
  name: string;
  amount: number;
}

export interface AggregationStats {
  max_amount: number;
  min_amount: number;
  avg_amount: number;
  max_invoice?: InvoiceReference;
  max_day?: DayReference;
  max_category?: EntityReference;
  max_company?: EntityReference;
}

export interface StatisticsFilters {
  category_id?: number;
  company_id?: number;
  receiver_id?: number;
  status?: InvoiceStatus;
  keyword?: string;
}

export interface InvoiceStatistics {
  period: string;
  start_date: string;
  end_date: string;
  total_amount: number;
  // Backend may use either name
  invoice_count?: number;
  total_count?: number;
  // Backend may use either name
  by_status?: StatusBreakdown;
  status_breakdown?: StatusBreakdown;
  breakdown?: BreakdownItem[];
  aggregations?: AggregationStats;
  filters?: StatisticsFilters;
  currency?: string;
}

// Display invoices tool result type
export interface DisplayInvoicesResult {
  invoices: Array<{
    id: number;
    title: string;
    description?: string;
    amount: number;
    currency: string;
    status: InvoiceStatus;
    company?: { id: number; name: string };
    category?: { id: number; name: string; color?: string };
    receiver?: { id: number; name: string };
    due_date: string | null;
    created_at: string;
  }>;
  total: number;
  query?: string;
}

// Tool action types
export type SearchToolAction =
  | { type: "navigate"; path: string }
  | { type: "filter"; path: string }
  | { type: "view_invoice"; id: number };
