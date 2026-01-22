"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiClient } from "@/lib/api/client";
import type { Tag, CreateTagRequest, UpdateTagRequest } from "@/lib/api/types";

export async function createTagAction(
  data: CreateTagRequest
): Promise<{ success: boolean; data?: Tag; error?: string }> {
  try {
    const tag = await apiClient<Tag>("/api/tags", {
      method: "POST",
      body: JSON.stringify(data),
    });
    revalidatePath("/tags");
    return { success: true, data: tag };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to create tag",
    };
  }
}

export async function updateTagAction(
  id: number,
  data: UpdateTagRequest
): Promise<{ success: boolean; data?: Tag; error?: string }> {
  try {
    const tag = await apiClient<Tag>(`/api/tags/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
    revalidatePath("/tags");
    revalidatePath(`/tags/${id}`);
    return { success: true, data: tag };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to update tag",
    };
  }
}

export async function deleteTagAction(
  id: number
): Promise<{ success: boolean; error?: string }> {
  try {
    await apiClient<void>(`/api/tags/${id}`, { method: "DELETE" });
    revalidatePath("/tags");
    return { success: true };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to delete tag",
    };
  }
}

export async function createTagAndRedirect(data: CreateTagRequest) {
  const result = await createTagAction(data);
  if (result.success) {
    redirect("/tags");
  }
  return result;
}

export async function deleteTagAndRedirect(id: number) {
  const result = await deleteTagAction(id);
  if (result.success) {
    redirect("/tags");
  }
  return result;
}
