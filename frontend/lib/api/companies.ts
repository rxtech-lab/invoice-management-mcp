import { apiClient } from "./client";
import type {
  Company,
  PaginatedResponse,
  CompanyListOptions,
  CreateCompanyRequest,
  UpdateCompanyRequest,
} from "./types";

export async function getCompanies(
  options: CompanyListOptions = {}
): Promise<PaginatedResponse<Company>> {
  const params = new URLSearchParams();
  Object.entries(options).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      params.set(key, String(value));
    }
  });
  const query = params.toString();
  return apiClient<PaginatedResponse<Company>>(
    `/api/companies${query ? `?${query}` : ""}`
  );
}

export async function getCompany(id: number): Promise<Company> {
  return apiClient<Company>(`/api/companies/${id}`);
}

export async function createCompany(
  data: CreateCompanyRequest
): Promise<Company> {
  return apiClient<Company>("/api/companies", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateCompany(
  id: number,
  data: UpdateCompanyRequest
): Promise<Company> {
  return apiClient<Company>(`/api/companies/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteCompany(id: number): Promise<void> {
  return apiClient<void>(`/api/companies/${id}`, { method: "DELETE" });
}
