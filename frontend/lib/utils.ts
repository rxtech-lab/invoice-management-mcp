import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"
import { format, formatDistanceToNow } from "date-fns"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatCurrency(amount: number, currency: string = "USD"): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: currency,
  }).format(amount)
}

export function formatDate(date: string | Date | null): string {
  if (!date) return "-"
  const d = typeof date === "string" ? new Date(date) : date
  return format(d, "MMM d, yyyy")
}

export function formatDateTime(date: string | Date | null): string {
  if (!date) return "-"
  const d = typeof date === "string" ? new Date(date) : date
  return format(d, "MMM d, yyyy h:mm a")
}

export function formatRelativeTime(date: string | Date | null): string {
  if (!date) return "-"
  const d = typeof date === "string" ? new Date(date) : date
  return formatDistanceToNow(d, { addSuffix: true })
}
