import { runInvoiceAgent } from "@/lib/agent/invoice-agent";
import {
  uploadFromURL,
  getFileDownloadURLWithToken,
} from "@/lib/api/upload";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

interface AgentProgress {
  status: "idle" | "calling" | "complete" | "error";
  toolName?: string;
  message: string;
  invoice_id?: number;
}

export async function POST(request: Request) {
  // Extract Bearer token from Authorization header
  const authHeader = request.headers.get("Authorization");
  const token = authHeader?.replace("Bearer ", "");

  if (!token) {
    return new Response(JSON.stringify({ error: "Unauthorized" }), {
      status: 401,
      headers: { "Content-Type": "application/json" },
    });
  }

  // Parse request body - accept file_url OR file_key
  let fileUrl: string | undefined;
  let fileKey: string | undefined;
  try {
    const body = await request.json();
    fileUrl = body.file_url;
    fileKey = body.file_key;
    if (!fileUrl && !fileKey) {
      return new Response(
        JSON.stringify({ error: "Missing file_url or file_key in request body" }),
        { status: 400, headers: { "Content-Type": "application/json" } }
      );
    }
  } catch {
    return new Response(JSON.stringify({ error: "Invalid JSON body" }), {
      status: 400,
      headers: { "Content-Type": "application/json" },
    });
  }

  // Create SSE response stream
  const encoder = new TextEncoder();
  const stream = new ReadableStream({
    async start(controller) {
      const sendEvent = (event: string, data: AgentProgress) => {
        controller.enqueue(
          encoder.encode(`event: ${event}\ndata: ${JSON.stringify(data)}\n\n`)
        );
      };

      try {
        // Step 1: Upload file from external URL to S3 (skip if file_key provided)
        if (!fileKey && fileUrl) {
          sendEvent("progress", {
            status: "calling",
            message: "Uploading file to storage...",
          });

          const uploadResult = await uploadFromURL(fileUrl, token);
          if (!uploadResult.success || !uploadResult.data) {
            sendEvent("error", {
              status: "error",
              message: uploadResult.error || "Failed to upload file",
            });
            controller.close();
            return;
          }

          fileKey = uploadResult.data.key;
        }

        // Ensure fileKey is defined (should always be true at this point)
        if (!fileKey) {
          sendEvent("error", {
            status: "error",
            message: "No file key available",
          });
          controller.close();
          return;
        }

        // Step 2: Get presigned download URL for the file
        sendEvent("progress", {
          status: "calling",
          message: "Preparing file for processing...",
        });

        const downloadResult = await getFileDownloadURLWithToken(fileKey, token);
        if (!downloadResult.success || !downloadResult.data) {
          sendEvent("error", {
            status: "error",
            message: downloadResult.error || "Failed to get file download URL",
          });
          controller.close();
          return;
        }

        const downloadUrl = downloadResult.data.download_url;

        // Step 3: Run the invoice agent with the S3 file
        const result = await runInvoiceAgent({
          fileUrl: downloadUrl,
          fileKey: fileKey,
          accessToken: token,
          onProgress: (progress) => {
            sendEvent("progress", progress);
          },
        });

        // Send completion event with invoice ID
        sendEvent("complete", {
          status: "complete",
          message: "Invoice created successfully!",
          invoice_id: result.invoiceId ?? undefined,
        });
      } catch (error) {
        console.error("Agent error:", error);
        sendEvent("error", {
          status: "error",
          message:
            error instanceof Error ? error.message : "Failed to create invoice",
        });
      } finally {
        controller.close();
      }
    },
  });

  return new Response(stream, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      Connection: "keep-alive",
    },
  });
}
