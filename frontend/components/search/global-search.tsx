"use client";

import { useState, useCallback, useEffect } from "react";
import { useRouter } from "next/navigation";
import {
  SearchTrigger,
  SearchCommand,
  type ToolResultRendererProps,
  type ToolAction,
} from "@rx-lab/dashboard-searching-ui";
import { keywordSearch } from "@/lib/search/keyword-search";
import { InvoiceResult } from "./invoice-result";
import { StatisticsRenderer } from "./statistics-renderer";
import { InvoiceListRenderer } from "./invoice-list-renderer";
import type { InvoiceSearchResult, InvoiceStatistics, DisplayInvoicesResult } from "@/lib/search/types";
import { Search, Loader2, FileText } from "lucide-react";

// Helper to extract data from MCP tool result format
// MCP tools return { content: [{ type: "text", text: "JSON string" }] }
function extractMCPToolData<T>(output: unknown): T | null {
  if (!output) return null;

  // If output is already the data object (not MCP format), return it
  if (typeof output === "object" && output !== null) {
    const obj = output as Record<string, unknown>;

    // Check if it's MCP format with content array
    if (Array.isArray(obj.content)) {
      const textContent = obj.content.find(
        (c: { type?: string; text?: string }) => c.type === "text"
      );
      if (textContent?.text) {
        try {
          return JSON.parse(textContent.text) as T;
        } catch {
          console.error("Failed to parse MCP tool result:", textContent.text);
          return null;
        }
      }
    }

    // If it has expected fields, it's already parsed data
    if ("total_amount" in obj || "invoices" in obj || "breakdown" in obj) {
      return output as T;
    }
  }

  return null;
}

// Wrapper for statistics renderer to match ToolResultRendererProps
function StatisticsRendererWrapper({ output, onAction }: ToolResultRendererProps) {
  const handleAction = useCallback(
    (action: { type: string; path?: string; id?: number }) => {
      if (onAction) {
        onAction({ type: action.type, payload: action });
      }
    },
    [onAction]
  );

  const data = extractMCPToolData<InvoiceStatistics>(output);

  if (!data) {
    return (
      <div className="p-4 text-sm text-muted-foreground">
        Unable to display statistics data
      </div>
    );
  }

  return (
    <StatisticsRenderer
      output={data}
      onAction={handleAction}
    />
  );
}

// Wrapper for invoice list renderer to match ToolResultRendererProps
function InvoiceListRendererWrapper({ output, onAction }: ToolResultRendererProps) {
  const handleAction = useCallback(
    (action: { type: string; path?: string; id?: number }) => {
      if (onAction) {
        onAction({ type: action.type, payload: action });
      }
    },
    [onAction]
  );

  const data = extractMCPToolData<DisplayInvoicesResult>(output);

  if (!data) {
    return (
      <div className="p-4 text-sm text-muted-foreground">
        Unable to display invoice data
      </div>
    );
  }

  return (
    <InvoiceListRenderer
      output={data}
      onAction={handleAction}
    />
  );
}

export function GlobalSearch() {
  const [open, setOpen] = useState(false);
  const router = useRouter();

  // Handle keyboard shortcut
  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((open) => !open);
      }
    };
    document.addEventListener("keydown", down);
    return () => document.removeEventListener("keydown", down);
  }, []);

  // Keyword search handler
  const handleSearch = useCallback(
    async (params: {
      query: string;
      limit?: number;
    }): Promise<InvoiceSearchResult[]> => {
      return keywordSearch({ query: params.query, limit: params.limit });
    },
    []
  );

  // Handle result selection
  const handleResultSelect = useCallback(
    (result: InvoiceSearchResult) => {
      setOpen(false);
      router.push(`/invoices/${result.id}`);
    },
    [router]
  );

  // Handle tool actions from AI agent
  const handleToolAction = useCallback(
    (action: ToolAction) => {
      const payload = action.payload as {
        type?: string;
        path?: string;
        id?: number;
      };

      setOpen(false);

      switch (payload.type || action.type) {
        case "navigate":
          if (payload.path) router.push(payload.path);
          break;
        case "view_invoice":
          if (payload.id) router.push(`/invoices/${payload.id}`);
          break;
        case "filter":
          if (payload.path) router.push(payload.path);
          break;
      }
    },
    [router]
  );

  // Custom result renderer
  const renderResult = useCallback(
    (result: InvoiceSearchResult, onSelect: () => void) => (
      <InvoiceResult key={result.id} invoice={result} onSelect={onSelect} />
    ),
    []
  );

  // Empty state renderer
  const renderEmpty = useCallback((query: string, hasResults: boolean) => {
    if (hasResults) return null;
    return (
      <div className="py-6 text-center text-sm text-muted-foreground">
        <FileText className="h-8 w-8 mx-auto mb-2 opacity-50" />
        <p>No invoices found for &quot;{query}&quot;</p>
        <p className="text-xs mt-1">Try AI mode for more advanced queries</p>
      </div>
    );
  }, []);

  // Loading state renderer
  const renderLoading = useCallback(
    () => (
      <div className="py-6 text-center text-sm text-muted-foreground">
        <Loader2 className="h-6 w-6 mx-auto mb-2 animate-spin" />
        <p>Searching invoices...</p>
      </div>
    ),
    []
  );

  return (
    <>
      <SearchTrigger
        onClick={() => setOpen(true)}
        placeholder="Search invoices..."
        shortcut={{ key: "K", modifier: "⌘" }}
        icon={Search}
        className="h-9 w-64"
        variant="outline"
      />

      <SearchCommand
        open={open}
        onOpenChange={setOpen}
        onSearch={handleSearch}
        onResultSelect={handleResultSelect}
        renderResult={renderResult}
        renderEmpty={renderEmpty}
        renderLoading={renderLoading}
        debounceMs={2000}
        limit={10}
        placeholder="Search invoices or ask AI..."
        className="max-w-[80vw] md:min-w-2xl lg:min-w-4xl"
        enableAgentMode
        agentConfig={{
          apiEndpoint: "/api/search-agent",
          toolResultRenderers: {
            display_statistics: StatisticsRendererWrapper,
            display_invoices: InvoiceListRendererWrapper,
          },
          onToolAction: handleToolAction,
          header: {
            title: "Invoice AI Assistant",
            showBackButton: true,
            showClearButton: true,
          },
          input: {
            placeholder: "Ask about your invoices...",
            placeholderProcessing: "Analyzing...",
            streamingText: "Thinking...",
          },
        }}
        chatHistoryStorageKey="invoice-search-history"
      />
    </>
  );
}
