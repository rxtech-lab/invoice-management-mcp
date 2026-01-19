"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { apiClient } from "@/lib/api/client";
import type {
  Company,
  CreateCompanyRequest,
  UpdateCompanyRequest,
} from "@/lib/api/types";

export async function createCompanyAction(
  data: CreateCompanyRequest
): Promise<{ success: boolean; data?: Company; error?: string }> {
  try {
    const company = await apiClient<Company>("/api/companies", {
      method: "POST",
      body: JSON.stringify(data),
    });
    revalidatePath("/companies");
    return { success: true, data: company };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to create company",
    };
  }
}

export async function updateCompanyAction(
  id: number,
  data: UpdateCompanyRequest
): Promise<{ success: boolean; data?: Company; error?: string }> {
  try {
    const company = await apiClient<Company>(`/api/companies/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    });
    revalidatePath("/companies");
    revalidatePath(`/companies/${id}`);
    return { success: true, data: company };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to update company",
    };
  }
}

export async function deleteCompanyAction(
  id: number
): Promise<{ success: boolean; error?: string }> {
  try {
    await apiClient<void>(`/api/companies/${id}`, { method: "DELETE" });
    revalidatePath("/companies");
    return { success: true };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to delete company",
    };
  }
}

export async function createCompanyAndRedirect(data: CreateCompanyRequest) {
  const result = await createCompanyAction(data);
  if (result.success) {
    redirect("/companies");
  }
  return result;
}

export async function deleteCompanyAndRedirect(id: number) {
  const result = await deleteCompanyAction(id);
  if (result.success) {
    redirect("/companies");
  }
  return result;
}
