"use client";

import { useState, useRef, useCallback, useEffect } from "react";
import { Upload, FileText, Download, X, Loader2, AlertCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { getPresignedURLAction, getFileDownloadURLAction, confirmUploadAction } from "@/lib/actions/upload-actions";

interface FileUploadProps {
  value?: string | null;
  onChange: (key: string | null) => void;
  accept?: string;
  maxSizeMB?: number;
  disabled?: boolean;
  className?: string;
}

type UploadState = "idle" | "uploading" | "uploaded" | "error";

export function FileUpload({
  value,
  onChange,
  accept = "application/pdf,.pdf",
  maxSizeMB = 10,
  disabled = false,
  className,
}: FileUploadProps) {
  const [state, setState] = useState<UploadState>(() => (value ? "uploaded" : "idle"));
  const [filename, setFilename] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isDragOver, setIsDragOver] = useState(false);
  const [isDownloading, setIsDownloading] = useState(false);

  const fileInputRef = useRef<HTMLInputElement>(null);

  // Fetch filename when value (key) changes and we don't have a filename
  useEffect(() => {
    if (value && !filename) {
      setState("uploaded");
      // Extract filename from key as fallback
      const keyFilename = value.split("/").pop() || "Uploaded file";
      setFilename(keyFilename);
    } else if (!value) {
      setState("idle");
      setFilename(null);
    }
  }, [value, filename]);

  const handleFileSelect = useCallback(
    async (file: File) => {
      // Validate file type
      const acceptedTypes = accept.split(",").map((t) => t.trim());
      const isValidType = acceptedTypes.some((type) => {
        if (type.startsWith(".")) {
          return file.name.toLowerCase().endsWith(type.toLowerCase());
        }
        if (type.includes("*")) {
          const [mainType] = type.split("/");
          return file.type.startsWith(mainType + "/");
        }
        return file.type === type;
      });

      if (!isValidType) {
        setErrorMessage("Invalid file type");
        setState("error");
        return;
      }

      // Validate file size
      const maxBytes = maxSizeMB * 1024 * 1024;
      if (file.size > maxBytes) {
        setErrorMessage(`File too large. Maximum size is ${maxSizeMB}MB`);
        setState("error");
        return;
      }

      // Proceed with upload using presigned URL
      setState("uploading");
      setFilename(file.name);
      setErrorMessage(null);

      // Get presigned URL from backend
      const contentType = file.type || "application/octet-stream";
      const presignedResult = await getPresignedURLAction(file.name, contentType);

      if (!presignedResult.success || !presignedResult.data) {
        setState("error");
        setErrorMessage(presignedResult.error || "Failed to get upload URL");
        return;
      }

      // Upload directly to S3 using presigned URL
      const { upload_url, key } = presignedResult.data;
      try {
        const uploadResponse = await fetch(upload_url, {
          method: "PUT",
          headers: { "Content-Type": contentType },
          body: file,
        });

        if (!uploadResponse.ok) {
          setState("error");
          setErrorMessage(`Upload failed: ${uploadResponse.status}`);
          return;
        }

        // Confirm upload to register file in database
        const confirmResult = await confirmUploadAction({
          key,
          filename: file.name,
          content_type: contentType,
          size: file.size,
        });

        if (!confirmResult.success) {
          setState("error");
          setErrorMessage(confirmResult.error || "Failed to confirm upload");
          return;
        }

        setState("uploaded");
        setFilename(file.name);
        onChange(key);
      } catch {
        setState("error");
        setErrorMessage("Upload failed: Network error");
      }
    },
    [accept, maxSizeMB, onChange]
  );

  const handleInputChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0];
      if (file) {
        handleFileSelect(file);
      }
      // Reset input for re-selection
      e.target.value = "";
    },
    [handleFileSelect]
  );

  const handleDragOver = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      if (!disabled) setIsDragOver(true);
    },
    [disabled]
  );

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);
  }, []);

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      setIsDragOver(false);

      if (disabled) return;

      const file = e.dataTransfer.files[0];
      if (file) handleFileSelect(file);
    },
    [disabled, handleFileSelect]
  );

  const handleRemove = useCallback(() => {
    setState("idle");
    setFilename(null);
    setErrorMessage(null);
    onChange(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = "";
    }
  }, [onChange]);

  const handleDownload = useCallback(async () => {
    if (!value) return;

    setIsDownloading(true);
    try {
      const result = await getFileDownloadURLAction(value);
      if (result.success && result.data) {
        window.open(result.data.download_url, "_blank");
      } else {
        setErrorMessage(result.error || "Failed to get download URL");
      }
    } catch {
      setErrorMessage("Failed to download file");
    } finally {
      setIsDownloading(false);
    }
  }, [value]);

  const handleRetry = useCallback(() => {
    setState("idle");
    setErrorMessage(null);
    fileInputRef.current?.click();
  }, []);

  const handleClick = useCallback(() => {
    if (!disabled && state === "idle") {
      fileInputRef.current?.click();
    }
  }, [disabled, state]);

  // Render based on state
  if (state === "uploading") {
    return (
      <div
        className={cn(
          "flex items-center gap-3 rounded-lg border border-dashed p-4",
          className
        )}
      >
        <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        <span className="text-sm text-muted-foreground">
          Uploading {filename}...
        </span>
      </div>
    );
  }

  if (state === "uploaded" && value) {
    return (
      <div
        className={cn(
          "flex items-center justify-between gap-3 rounded-lg border p-4",
          className
        )}
      >
        <div className="flex items-center gap-3 min-w-0">
          <FileText className="h-5 w-5 flex-shrink-0 text-muted-foreground" />
          <span className="text-sm truncate">{filename}</span>
        </div>
        <div className="flex items-center gap-2 flex-shrink-0">
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={handleDownload}
            disabled={isDownloading}
          >
            {isDownloading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Download className="h-4 w-4" />
            )}
          </Button>
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={handleRemove}
            disabled={disabled}
          >
            <X className="h-4 w-4" />
          </Button>
        </div>
      </div>
    );
  }

  if (state === "error") {
    return (
      <div
        className={cn(
          "flex flex-col items-center gap-2 rounded-lg border border-destructive p-6 text-center",
          className
        )}
      >
        <AlertCircle className="h-8 w-8 text-destructive" />
        <span className="text-sm text-destructive">{errorMessage}</span>
        <Button type="button" variant="outline" size="sm" onClick={handleRetry}>
          Try Again
        </Button>
      </div>
    );
  }

  // Idle state - dropzone
  return (
    <>
      <input
        type="file"
        ref={fileInputRef}
        className="hidden"
        accept={accept}
        onChange={handleInputChange}
        disabled={disabled}
      />
      <div
        className={cn(
          "flex cursor-pointer flex-col items-center gap-2 rounded-lg border border-dashed p-6 text-center transition-colors",
          isDragOver && "border-primary bg-primary/5",
          disabled && "cursor-not-allowed opacity-50",
          className
        )}
        onClick={handleClick}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        <Upload className="h-8 w-8 text-muted-foreground" />
        <div className="text-sm text-muted-foreground">
          <span className="font-medium text-foreground">Click to upload</span>{" "}
          or drag and drop
        </div>
        <p className="text-xs text-muted-foreground">
          PDF up to {maxSizeMB}MB
        </p>
      </div>
    </>
  );
}
