import { apiClient } from "./client";
import type {
  Receiver,
  PaginatedResponse,
  ReceiverListOptions,
  CreateReceiverRequest,
  UpdateReceiverRequest,
} from "./types";

export async function getReceivers(
  options: ReceiverListOptions = {}
): Promise<PaginatedResponse<Receiver>> {
  const params = new URLSearchParams();
  Object.entries(options).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== "") {
      params.set(key, String(value));
    }
  });
  const query = params.toString();
  return apiClient<PaginatedResponse<Receiver>>(
    `/api/receivers${query ? `?${query}` : ""}`
  );
}

export async function getReceiver(id: number): Promise<Receiver> {
  return apiClient<Receiver>(`/api/receivers/${id}`);
}

export async function createReceiver(
  data: CreateReceiverRequest
): Promise<Receiver> {
  return apiClient<Receiver>("/api/receivers", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateReceiver(
  id: number,
  data: UpdateReceiverRequest
): Promise<Receiver> {
  return apiClient<Receiver>(`/api/receivers/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteReceiver(id: number): Promise<void> {
  return apiClient<void>(`/api/receivers/${id}`, { method: "DELETE" });
}
