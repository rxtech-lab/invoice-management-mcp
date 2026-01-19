"use client";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { DISPLAY_CURRENCIES, type CurrencyCode } from "@/lib/currency";
import { RefreshCw } from "lucide-react";

interface CurrencyPickerProps {
  value: CurrencyCode | null;
  onChange: (currency: CurrencyCode | null) => void;
  originalCurrency: string;
  isLoading?: boolean;
}

export function CurrencyPicker({
  value,
  onChange,
  originalCurrency,
  isLoading,
}: CurrencyPickerProps) {
  return (
    <div className="flex items-center gap-2">
      <span className="text-sm text-muted-foreground">Display as:</span>
      <Select
        value={value || "original"}
        onValueChange={(val) =>
          onChange(val === "original" ? null : (val as CurrencyCode))
        }
      >
        <SelectTrigger className="w-[160px] h-8">
          {isLoading ? (
            <RefreshCw className="h-4 w-4 animate-spin" />
          ) : (
            <SelectValue placeholder="Original" />
          )}
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="original">
            Original ({originalCurrency})
          </SelectItem>
          {DISPLAY_CURRENCIES.filter((c) => c.code !== originalCurrency).map(
            (currency) => (
              <SelectItem key={currency.code} value={currency.code}>
                {currency.code} - {currency.name}
              </SelectItem>
            )
          )}
        </SelectContent>
      </Select>
    </div>
  );
}
