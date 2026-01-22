import { apiClient } from "./client";
import type {
  Tag,
  PaginatedResponse,
  TagListOptions,
  CreateTagRequest,
  UpdateTagRequest,
} from "./types";

export async function getTags(
  options: TagListOptions = {}
): Promise<PaginatedResponse<Tag>> {
  const params = new URLSearchParams();
  Object.entries(options).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      params.set(key, String(value));
    }
  });
  const query = params.toString();
  return apiClient<PaginatedResponse<Tag>>(`/api/tags${query ? `?${query}` : ""}`);
}

export async function getTag(id: number): Promise<Tag> {
  return apiClient<Tag>(`/api/tags/${id}`);
}

export async function createTag(data: CreateTagRequest): Promise<Tag> {
  return apiClient<Tag>("/api/tags", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateTag(
  id: number,
  data: UpdateTagRequest
): Promise<Tag> {
  return apiClient<Tag>(`/api/tags/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteTag(id: number): Promise<void> {
  return apiClient<void>(`/api/tags/${id}`, { method: "DELETE" });
}
