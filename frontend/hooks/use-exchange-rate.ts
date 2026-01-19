"use client";

import { useState, useEffect, useCallback } from "react";
import { getExchangeRateAction } from "@/lib/actions/exchange-rate-actions";
import type { ExchangeRate, CurrencyCode } from "@/lib/currency";

interface UseExchangeRateOptions {
  fromCurrency: string;
  toCurrency: CurrencyCode | null;
}

interface UseExchangeRateReturn {
  rate: ExchangeRate | null;
  isLoading: boolean;
  error: string | null;
  convert: (amount: number) => number;
}

export function useExchangeRate({
  fromCurrency,
  toCurrency,
}: UseExchangeRateOptions): UseExchangeRateReturn {
  const [rate, setRate] = useState<ExchangeRate | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // No conversion needed
    if (!toCurrency || toCurrency === fromCurrency) {
      setRate(null);
      setError(null);
      return;
    }

    const fetchRate = async () => {
      setIsLoading(true);
      setError(null);

      const result = await getExchangeRateAction(fromCurrency, toCurrency);

      if (result.success && result.data) {
        setRate(result.data);
      } else {
        setError(result.error || "Failed to fetch rate");
        setRate(null);
      }

      setIsLoading(false);
    };

    fetchRate();
  }, [fromCurrency, toCurrency]);

  const convert = useCallback(
    (amount: number): number => {
      if (!rate || rate.rate === 1) return amount;
      return amount * rate.rate;
    },
    [rate]
  );

  return { rate, isLoading, error, convert };
}
