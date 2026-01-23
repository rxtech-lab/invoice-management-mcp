"use client";

import { useState, useRef } from "react";
import Link from "next/link";
import { Plus, ChevronDown, Bot, FileText, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { readStreamableValue } from "@ai-sdk/rsc";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { createInvoiceWithAgentAction, type ToolProgress } from "@/lib/actions/invoice-actions";
import { getPresignedURLAction, confirmUploadAction } from "@/lib/actions/upload-actions";
import { useRouter } from "next/navigation";

type LoadingState = "idle" | "uploading" | "processing";

export function NewInvoiceButton() {
  const [isLoading, setIsLoading] = useState(false);
  const [loadingState, setLoadingState] = useState<LoadingState>("idle");
  const fileInputRef = useRef<HTMLInputElement>(null);
  const router = useRouter();

  const handleCreateWithAgentClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileSelected = async (
    event: React.ChangeEvent<HTMLInputElement>
  ) => {
    const file = event.target.files?.[0];
    if (!file) return;

    // Reset input for re-selection
    event.target.value = "";

    // Validate PDF
    if (file.type !== "application/pdf") {
      toast.error("Please select a PDF file");
      return;
    }

    setIsLoading(true);
    setLoadingState("uploading");
    const toastId = toast.loading("Uploading invoice...");

    try {
      // Get presigned URL for upload
      const contentType = file.type || "application/pdf";
      const presignedResult = await getPresignedURLAction(file.name, contentType);

      if (!presignedResult.success || !presignedResult.data) {
        toast.error(presignedResult.error || "Failed to get upload URL", {
          id: toastId,
        });
        return;
      }

      // Upload directly to S3 using presigned URL
      const { upload_url, key } = presignedResult.data;
      const uploadResponse = await fetch(upload_url, {
        method: "PUT",
        headers: { "Content-Type": contentType },
        body: file,
      });

      if (!uploadResponse.ok) {
        toast.error(`Upload failed: ${uploadResponse.status}`, {
          id: toastId,
        });
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
        toast.error(confirmResult.error || "Failed to confirm upload", {
          id: toastId,
        });
        return;
      }

      // Process with agent (streaming)
      setLoadingState("processing");
      toast.loading("Processing invoice with AI agent...", { id: toastId });

      const { progress } = await createInvoiceWithAgentAction(key);

      let lastStatus: ToolProgress["status"] = "idle";

      for await (const update of readStreamableValue(progress)) {
        if (update) {
          lastStatus = update.status;
          if (update.status === "error") {
            toast.error(update.message, { id: toastId });
          } else if (update.status === "complete") {
            toast.success(update.message, { id: toastId });
            router.refresh();
          } else {
            toast.loading(update.message, { id: toastId });
          }
        }
      }

      // Handle case where stream ends without explicit complete/error
      if (lastStatus !== "complete" && lastStatus !== "error") {
        toast.success("Invoice processed!", { id: toastId });
        router.refresh();
      }
    } catch (error) {
      toast.error("An unexpected error occurred: " + error, { id: toastId });
    } finally {
      setIsLoading(false);
      setLoadingState("idle");
    }
  };

  return (
    <>
      <input
        type="file"
        ref={fileInputRef}
        className="hidden"
        accept="application/pdf,.pdf"
        onChange={handleFileSelected}
      />
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button disabled={isLoading}>
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                {loadingState === "uploading" ? "Uploading..." : "Processing..."}
              </>
            ) : (
              <>
                <Plus className="mr-2 h-4 w-4" />
                New Invoice
                <ChevronDown className="ml-2 h-4 w-4" />
              </>
            )}
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="min-w-48">
          <DropdownMenuItem asChild>
            <Link href="/invoices/new">
              <FileText className="mr-2 h-4 w-4" />
              New Invoice
            </Link>
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={handleCreateWithAgentClick}
            disabled={isLoading}
          >
            <Bot className="mr-2 h-4 w-4" />
            Create with Agent
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </>
  );
}
