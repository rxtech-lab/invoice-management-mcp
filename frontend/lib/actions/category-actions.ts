"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiClient } from "@/lib/api/client";
import type {
  Category,
  CreateCategoryRequest,
  UpdateCategoryRequest,
} from "@/lib/api/types";

export async function createCategoryAction(
  data: CreateCategoryRequest
): Promise<{ success: boolean; data?: Category; error?: string }> {
  try {
    const category = await apiClient<Category>("/api/categories", {
      method: "POST",
      body: JSON.stringify(data),
    });
    revalidatePath("/categories");
    return { success: true, data: category };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to create category",
    };
  }
}

export async function updateCategoryAction(
  id: number,
  data: UpdateCategoryRequest
): Promise<{ success: boolean; data?: Category; error?: string }> {
  try {
    const category = await apiClient<Category>(`/api/categories/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
    revalidatePath("/categories");
    revalidatePath(`/categories/${id}`);
    return { success: true, data: category };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to update category",
    };
  }
}

export async function deleteCategoryAction(
  id: number
): Promise<{ success: boolean; error?: string }> {
  try {
    await apiClient<void>(`/api/categories/${id}`, { method: "DELETE" });
    revalidatePath("/categories");
    return { success: true };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to delete category",
    };
  }
}

export async function createCategoryAndRedirect(data: CreateCategoryRequest) {
  const result = await createCategoryAction(data);
  if (result.success) {
    redirect("/categories");
  }
  return result;
}

export async function deleteCategoryAndRedirect(id: number) {
  const result = await deleteCategoryAction(id);
  if (result.success) {
    redirect("/categories");
  }
  return result;
}
