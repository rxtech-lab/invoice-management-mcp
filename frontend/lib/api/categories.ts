import { apiClient } from "./client";
import type {
  Category,
  PaginatedResponse,
  CategoryListOptions,
  CreateCategoryRequest,
  UpdateCategoryRequest,
} from "./types";

export async function getCategories(
  options: CategoryListOptions = {}
): Promise<PaginatedResponse<Category>> {
  const params = new URLSearchParams();
  Object.entries(options).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      params.set(key, String(value));
    }
  });
  const query = params.toString();
  return apiClient<PaginatedResponse<Category>>(
    `/api/categories${query ? `?${query}` : ""}`
  );
}

export async function getCategory(id: number): Promise<Category> {
  return apiClient<Category>(`/api/categories/${id}`);
}

export async function createCategory(
  data: CreateCategoryRequest
): Promise<Category> {
  return apiClient<Category>("/api/categories", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateCategory(
  id: number,
  data: UpdateCategoryRequest
): Promise<Category> {
  return apiClient<Category>(`/api/categories/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteCategory(id: number): Promise<void> {
  return apiClient<void>(`/api/categories/${id}`, { method: "DELETE" });
}
