"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiClient } from "@/lib/api/client";
import type {
  Receiver,
  CreateReceiverRequest,
  UpdateReceiverRequest,
} from "@/lib/api/types";

export async function createReceiverAction(
  data: CreateReceiverRequest
): Promise<{ success: boolean; data?: Receiver; error?: string }> {
  try {
    const receiver = await apiClient<Receiver>("/api/receivers", {
      method: "POST",
      body: JSON.stringify(data),
    });
    revalidatePath("/receivers");
    return { success: true, data: receiver };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to create receiver",
    };
  }
}

export async function updateReceiverAction(
  id: number,
  data: UpdateReceiverRequest
): Promise<{ success: boolean; data?: Receiver; error?: string }> {
  try {
    const receiver = await apiClient<Receiver>(`/api/receivers/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
    revalidatePath("/receivers");
    revalidatePath(`/receivers/${id}`);
    return { success: true, data: receiver };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to update receiver",
    };
  }
}

export async function deleteReceiverAction(
  id: number
): Promise<{ success: boolean; error?: string }> {
  try {
    await apiClient<void>(`/api/receivers/${id}`, { method: "DELETE" });
    revalidatePath("/receivers");
    return { success: true };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to delete receiver",
    };
  }
}

export async function createReceiverAndRedirect(data: CreateReceiverRequest) {
  const result = await createReceiverAction(data);
  if (result.success) {
    redirect("/receivers");
  }
  return result;
}

export async function deleteReceiverAndRedirect(id: number) {
  const result = await deleteReceiverAction(id);
  if (result.success) {
    redirect("/receivers");
  }
  return result;
}
