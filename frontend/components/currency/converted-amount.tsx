"use client";

import { formatCurrency } from "@/lib/utils";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface ConvertedAmountProps {
  originalAmount: number;
  originalCurrency: string;
  convertedAmount: number | null;
  displayCurrency: string | null;
  showOriginal?: boolean;
  className?: string;
}

export function ConvertedAmount({
  originalAmount,
  originalCurrency,
  convertedAmount,
  displayCurrency,
  showOriginal = true,
  className,
}: ConvertedAmountProps) {
  // No conversion - show original
  if (!displayCurrency || convertedAmount === null) {
    return (
      <span className={className}>
        {formatCurrency(originalAmount, originalCurrency)}
      </span>
    );
  }

  // Show converted with original in tooltip
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <span className={className}>
          {formatCurrency(convertedAmount, displayCurrency)}
          {showOriginal && (
            <span className="ml-1 text-xs text-muted-foreground">
              ({formatCurrency(originalAmount, originalCurrency)})
            </span>
          )}
        </span>
      </TooltipTrigger>
      <TooltipContent>
        <p>Original: {formatCurrency(originalAmount, originalCurrency)}</p>
        <p className="text-xs text-muted-foreground">
          Exchange rate may vary from actual transaction rate
        </p>
      </TooltipContent>
    </Tooltip>
  );
}
