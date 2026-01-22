import { apiClient } from "./client";
import type {
  AnalyticsPeriod,
  AnalyticsSummary,
  AnalyticsByGroup,
} from "./types";

export async function getAnalyticsSummary(
  period: AnalyticsPeriod = "1m"
): Promise<AnalyticsSummary> {
  return apiClient<AnalyticsSummary>(`/api/analytics/summary?period=${period}`);
}

export async function getAnalyticsByCategory(
  period: AnalyticsPeriod = "1m"
): Promise<AnalyticsByGroup> {
  return apiClient<AnalyticsByGroup>(
    `/api/analytics/by-category?period=${period}`
  );
}

export async function getAnalyticsByCompany(
  period: AnalyticsPeriod = "1m"
): Promise<AnalyticsByGroup> {
  return apiClient<AnalyticsByGroup>(
    `/api/analytics/by-company?period=${period}`
  );
}

export async function getAnalyticsByReceiver(
  period: AnalyticsPeriod = "1m"
): Promise<AnalyticsByGroup> {
  return apiClient<AnalyticsByGroup>(
    `/api/analytics/by-receiver?period=${period}`
  );
}

export async function getAnalyticsByTag(
  period: AnalyticsPeriod = "1m"
): Promise<AnalyticsByGroup> {
  return apiClient<AnalyticsByGroup>(`/api/analytics/by-tag?period=${period}`);
}
