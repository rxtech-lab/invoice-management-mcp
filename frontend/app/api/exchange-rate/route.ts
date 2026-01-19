import { NextResponse } from "next/server";
import type { ExchangeRate, ExchangeRateResponse } from "@/lib/currency";

const FRANKFURTER_API = "https://api.frankfurter.dev/v1";

// In-memory cache for exchange rates (persists across requests within same server instance)
const rateCache = new Map<string, { rate: ExchangeRate; timestamp: number }>();
const CACHE_TTL = 60 * 60 * 1000; // 1 hour in milliseconds

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const from = searchParams.get("from");
  const to = searchParams.get("to");

  // Validate required parameters
  if (!from || !to) {
    return NextResponse.json(
      { success: false, error: "Missing required parameters: from and to" },
      { status: 400 }
    );
  }

  // Same currency - no conversion needed
  if (from === to) {
    return NextResponse.json({
      success: true,
      data: {
        from,
        to,
        rate: 1,
        date: new Date().toISOString().split("T")[0],
      },
    });
  }

  // Check cache for existing rate
  const cacheKey = `${from}-${to}`;
  const cached = rateCache.get(cacheKey);

  if (cached && Date.now() - cached.timestamp < CACHE_TTL) {
    return NextResponse.json({ success: true, data: cached.rate });
  }

  // Fetch from Frankfurter API
  try {
    const response = await fetch(
      `${FRANKFURTER_API}/latest?base=${from}&symbols=${to}`,
      { next: { revalidate: 3600 } } // Next.js fetch cache: 1 hour
    );

    if (!response.ok) {
      throw new Error(`Exchange rate API error: ${response.status}`);
    }

    const data: ExchangeRateResponse = await response.json();

    const rate: ExchangeRate = {
      from,
      to,
      rate: data.rates[to],
      date: data.date,
    };

    // Update cache
    rateCache.set(cacheKey, { rate, timestamp: Date.now() });

    return NextResponse.json({ success: true, data: rate });
  } catch (error) {
    return NextResponse.json(
      {
        success: false,
        error:
          error instanceof Error ? error.message : "Failed to fetch exchange rate",
      },
      { status: 500 }
    );
  }
}
