"use server";

import type { ExchangeRate, ExchangeRateResponse } from "@/lib/currency";

const FRANKFURTER_API = "https://api.frankfurter.dev/v1";

// In-memory cache for server action (persists across requests within same server instance)
const rateCache = new Map<string, { rate: ExchangeRate; timestamp: number }>();
const CACHE_TTL = 60 * 60 * 1000; // 1 hour in milliseconds

export async function getExchangeRateAction(
  fromCurrency: string,
  toCurrency: string
): Promise<{ success: boolean; data?: ExchangeRate; error?: string }> {
  // Same currency, no conversion needed
  if (fromCurrency === toCurrency) {
    return {
      success: true,
      data: {
        from: fromCurrency,
        to: toCurrency,
        rate: 1,
        date: new Date().toISOString().split("T")[0],
      },
    };
  }

  const cacheKey = `${fromCurrency}-${toCurrency}`;
  const cached = rateCache.get(cacheKey);

  // Return cached if valid
  if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
    return { success: true, data: cached.rate };
  }

  try {
    const response = await fetch(
      `${FRANKFURTER_API}/latest?base=${fromCurrency}&symbols=${toCurrency}`,
      { next: { revalidate: 3600 } } // Next.js fetch cache: 1 hour
    );

    if (!response.ok) {
      throw new Error(`Exchange rate API error: ${response.status}`);
    }

    const data: ExchangeRateResponse = await response.json();

    const rate: ExchangeRate = {
      from: fromCurrency,
      to: toCurrency,
      rate: data.rates[toCurrency],
      date: data.date,
    };

    // Update cache
    rateCache.set(cacheKey, { rate, timestamp: Date.now() });

    return { success: true, data: rate };
  } catch (error) {
    return {
      success: false,
      error:
        error instanceof Error ? error.message : "Failed to fetch exchange rate",
    };
  }
}
