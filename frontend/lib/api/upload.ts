import { getSession } from "next-auth/react";
import type { UploadResponse, PresignedURLResponse, FileDownloadURLResponse } from "./types";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

export async function uploadFile(file: File): Promise<UploadResponse> {
  const session = await getSession();

  const formData = new FormData();
  formData.append("file", file);

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

  return response.json();
}

export async function getPresignedURL(
  filename: string,
  contentType: string = "application/octet-stream"
): Promise<PresignedURLResponse> {
  const session = await getSession();

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

  return response.json();
}

export async function uploadToPresignedURL(
  presignedUrl: string,
  file: File,
  contentType: string
): Promise<void> {
  const response = await fetch(presignedUrl, {
    method: "PUT",
    headers: {
      "Content-Type": contentType,
    },
    body: file,
  });

  if (!response.ok) {
    throw new Error(`Upload to presigned URL failed: ${response.status}`);
  }
}

/**
 * Upload a file from an external URL to S3.
 * For use in API routes where a token is passed explicitly.
 */
export async function uploadFromURL(
  fileUrl: string,
  token: string
): Promise<{ success: boolean; data?: UploadResponse; error?: string }> {
  try {
    // Fetch file from external URL
    const fileResponse = await fetch(fileUrl);
    if (!fileResponse.ok) {
      throw new Error(`Failed to fetch file from URL: ${fileResponse.status}`);
    }

    const blob = await fileResponse.blob();

    // Extract filename from URL or use default
    const urlPath = new URL(fileUrl).pathname;
    const filename = urlPath.split("/").pop() || "invoice.pdf";

    // Create FormData with the file
    const formData = new FormData();
    formData.append("file", blob, filename);

    // Upload to backend
    const response = await fetch(`${API_BASE_URL}/api/upload`, {
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
      },
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
      error:
        error instanceof Error ? error.message : "Failed to upload file from URL",
    };
  }
}

/**
 * Get file download URL using a token.
 * For use in API routes where a token is passed explicitly.
 */
export async function getFileDownloadURLWithToken(
  key: string,
  token: string
): Promise<{ success: boolean; data?: FileDownloadURLResponse; error?: string }> {
  try {
    const encodedKey = encodeURIComponent(key);
    const response = await fetch(
      `${API_BASE_URL}/api/files/${encodedKey}/download`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    );

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(
        error.error || `Failed to get download URL: ${response.status}`
      );
    }

    const data = await response.json();
    return { success: true, data };
  } catch (error) {
    return {
      success: false,
      error:
        error instanceof Error ? error.message : "Failed to get download URL",
    };
  }
}
