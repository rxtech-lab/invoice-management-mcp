"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useDebounce } from "@uidotdev/usehooks";
import { useState, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Search, X } from "lucide-react";
import type { Category, Company, Receiver } from "@/lib/api/types";

interface InvoiceFiltersProps {
  categories: Category[];
  companies: Company[];
  receivers: Receiver[];
}

export function InvoiceFilters({
  categories,
  companies,
  receivers,
}: InvoiceFiltersProps) {
  const router = useRouter();
  const searchParams = useSearchParams();

  // Current filter values from URL (use "all" as default for selects)
  const keyword = searchParams.get("keyword") || "";
  const categoryId = searchParams.get("category_id") || "all";
  const companyId = searchParams.get("company_id") || "all";
  const receiverId = searchParams.get("receiver_id") || "all";
  const status = searchParams.get("status") || "all";

  // Local state for search input with debounce
  const [searchTerm, setSearchTerm] = useState(keyword);
  const debouncedSearchTerm = useDebounce(searchTerm, 300);

  // Update URL when debounced search term changes
  useEffect(() => {
    if (debouncedSearchTerm !== keyword) {
      updateParams({ keyword: debouncedSearchTerm || null });
    }
  }, [debouncedSearchTerm]);

  // Sync search input when URL changes externally
  useEffect(() => {
    setSearchTerm(keyword);
  }, [keyword]);

  // Update URL with new params (treat "all" as clearing the filter)
  const updateParams = (updates: Record<string, string | null>) => {
    const params = new URLSearchParams(searchParams.toString());
    Object.entries(updates).forEach(([key, value]) => {
      if (value && value !== "all") {
        params.set(key, value);
      } else {
        params.delete(key);
      }
    });
    router.push(`/invoices?${params.toString()}`);
  };

  // Clear all filters
  const clearFilters = () => {
    setSearchTerm("");
    router.push("/invoices");
  };

  const hasFilters =
    keyword ||
    categoryId !== "all" ||
    companyId !== "all" ||
    receiverId !== "all" ||
    status !== "all";

  return (
    <div className="flex flex-wrap items-center gap-4">
      {/* Search Input */}
      <div className="relative flex-1 min-w-[200px] max-w-sm">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search invoices..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          className="pl-9"
        />
      </div>

      {/* Category Filter */}
      <Select
        value={categoryId}
        onValueChange={(v) => updateParams({ category_id: v })}
      >
        <SelectTrigger className="w-[180px]">
          <SelectValue placeholder="All Categories" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Categories</SelectItem>
          {categories.map((cat) => (
            <SelectItem key={cat.id} value={String(cat.id)}>
              {cat.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* Company Filter */}
      <Select
        value={companyId}
        onValueChange={(v) => updateParams({ company_id: v })}
      >
        <SelectTrigger className="w-[180px]">
          <SelectValue placeholder="All Companies" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Companies</SelectItem>
          {companies.map((comp) => (
            <SelectItem key={comp.id} value={String(comp.id)}>
              {comp.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* Receiver Filter */}
      <Select
        value={receiverId}
        onValueChange={(v) => updateParams({ receiver_id: v })}
      >
        <SelectTrigger className="w-[180px]">
          <SelectValue placeholder="All Receivers" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Receivers</SelectItem>
          {receivers.map((rec) => (
            <SelectItem key={rec.id} value={String(rec.id)}>
              {rec.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      {/* Status Filter */}
      <Select
        value={status}
        onValueChange={(v) => updateParams({ status: v })}
      >
        <SelectTrigger className="w-[140px]">
          <SelectValue placeholder="All Status" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="all">All Status</SelectItem>
          <SelectItem value="paid">Paid</SelectItem>
          <SelectItem value="unpaid">Unpaid</SelectItem>
          <SelectItem value="overdue">Overdue</SelectItem>
        </SelectContent>
      </Select>

      {/* Clear Filters */}
      {hasFilters && (
        <Button variant="ghost" size="sm" onClick={clearFilters}>
          <X className="h-4 w-4 mr-1" />
          Clear
        </Button>
      )}
    </div>
  );
}
