"use client";

import { useState, useRef } from "react";
import Link from "next/link";
import { Plus, ChevronDown, Bot, FileText, Loader2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { createInvoiceWithAgentAction } from "@/lib/actions/invoice-actions";
import { uploadFileAction } from "@/lib/actions/upload-actions";
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
      // Upload file
      const formData = new FormData();
      formData.append("file", file);
      const uploadResult = await uploadFileAction(formData);

      if (!uploadResult.success || !uploadResult.data) {
        toast.error(uploadResult.error || "Failed to upload file", {
          id: toastId,
        });
        return;
      }

      // Process with agent
      setLoadingState("processing");
      toast.loading("Processing invoice with AI agent...", { id: toastId });

      const agentResult = await createInvoiceWithAgentAction(
        uploadResult.data.download_url
      );

      if (agentResult.success) {
        toast.success("Invoice created successfully!", { id: toastId });
        router.refresh();
      } else {
        toast.error(agentResult.error || "Failed to process invoice", {
          id: toastId,
        });
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
