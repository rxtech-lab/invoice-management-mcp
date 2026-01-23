"use server";

import { auth } from "@/auth";
import type { UploadResponse, PresignedURLResponse, FileDownloadURLResponse, ConfirmUploadRequest } from "@/lib/api/types";

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

export async function getFileDownloadURLAction(
  key: string
): Promise<{ success: boolean; data?: FileDownloadURLResponse; error?: string }> {
  try {
    const session = await auth();

    const headers: HeadersInit = {};
    if (session?.accessToken) {
      headers["Authorization"] = `Bearer ${session.accessToken}`;
    }

    const encodedKey = encodeURIComponent(key);
    const response = await fetch(
      `${API_BASE_URL}/api/files/${encodedKey}/download`,
      { headers }
    );

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || `Failed to get download URL: ${response.status}`);
    }

    const data = await response.json();
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to get download URL",
    };
  }
}

export async function confirmUploadAction(
  data: ConfirmUploadRequest
): Promise<{ success: boolean; data?: UploadResponse; error?: string }> {
  try {
    const session = await auth();

    const headers: HeadersInit = {
      "Content-Type": "application/json",
    };
    if (session?.accessToken) {
      headers["Authorization"] = `Bearer ${session.accessToken}`;
    }

    const response = await fetch(`${API_BASE_URL}/api/upload/confirm`, {
      method: "POST",
      headers,
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(error.error || `Failed to confirm upload: ${response.status}`);
    }

    const result = await response.json();
    return { success: true, data: result };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : "Failed to confirm upload",
    };
  }
}
