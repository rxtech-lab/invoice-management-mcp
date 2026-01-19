import { getSession } from "next-auth/react";
import type { UploadResponse, PresignedURLResponse } from "./types";

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
