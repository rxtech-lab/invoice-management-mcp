"use server";

import { auth } from "@/auth";
import type { UploadResponse, PresignedURLResponse } from "@/lib/api/types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function uploadFileAction(
  formData: FormData
): Promise<{ success: boolean; data?: UploadResponse; error?: string }> {
  try {
    const session = await auth();

    const headers: HeadersInit = {};
    if (session?.accessToken) {
      headers["Authorization"] = `Bearer ${session.accessToken}`;
    }

    const response = await fetch(`${API_BASE_URL}/api/upload`, {
      method: "POST",
      headers,
      body: formData,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || `Upload failed: ${response.status}`);
    }

    const data = await response.json();
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to upload file",
    };
  }
}

export async function getPresignedURLAction(
  filename: string,
  contentType: string = "application/octet-stream"
): Promise<{ success: boolean; data?: PresignedURLResponse; error?: string }> {
  try {
    const session = await auth();

    const params = new URLSearchParams({
      filename,
      content_type: contentType,
    });

    const headers: HeadersInit = {
      "Content-Type": "application/json",
    };
    if (session?.accessToken) {
      headers["Authorization"] = `Bearer ${session.accessToken}`;
    }

    const response = await fetch(
      `${API_BASE_URL}/api/upload/presigned?${params}`,
      { headers }
    );

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || `Failed to get presigned URL: ${response.status}`);
    }

    const data = await response.json();
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to get presigned URL",
    };
  }
}
