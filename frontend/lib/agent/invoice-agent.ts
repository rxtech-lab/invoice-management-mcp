import { createMCPClient } from "@ai-sdk/mcp";
import { streamText, stepCountIs } from "ai";
import { invoiceAgentPrompt } from "../prompt";

export type ToolProgress = {
  status: "idle" | "calling" | "complete" | "error";
  toolName?: string;
  message: string;
  invoiceId?: number;
};

export interface RunInvoiceAgentParams {
  fileUrl: string; // Presigned download URL
  fileKey: string; // S3 key to store in invoice.original_download_link
  accessToken: string;
  onProgress: (progress: ToolProgress) => void;
}

export interface RunInvoiceAgentResult {
  invoiceId: number | null;
}

function formatToolName(toolName: string): string {
  // Convert "create_invoice" to "Creating Invoice"
  const words = toolName.replace(/_/g, " ").split(" ");
  if (words.length > 0) {
    const firstWord = words[0];
    if (firstWord.endsWith("e")) {
      words[0] = firstWord.slice(0, -1) + "ing";
    } else {
      words[0] = firstWord + "ing";
    }
  }
  return words.map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join(" ");
}

/**
 * Shared logic for running the invoice agent.
 * This function is used by both the API route and the server action.
 */
export async function runInvoiceAgent(
  params: RunInvoiceAgentParams
): Promise<RunInvoiceAgentResult> {
  const { fileUrl, fileKey, accessToken, onProgress } = params;

  let invoiceId: number | null = null;

  // Connect to MCP server
  const mcpUrl = process.env.NEXT_PUBLIC_API_URL! + "/mcp";
  const httpClient = await createMCPClient({
    transport: {
      type: "http",
      url: mcpUrl,
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    },
  });

  onProgress({
    status: "calling",
    message: "Connecting to AI agent...",
  });

  const tools = await httpClient.tools();

  const result = streamText({
    model: process.env
      .INVOICE_AGENT_MODEL! as Parameters<typeof streamText>[0]["model"],
    system: invoiceAgentPrompt,
    messages: [
      {
        role: "user",
        content: [
          {
            type: "text",
            text: `Create an invoice for the following file, make sure to use tools to add the invoice to the database. The original_download_link should be set to "${fileKey}" (this is the file key, not a URL)`,
          },
          {
            type: "file",
            data: new URL(fileUrl),
            mediaType: "application/pdf",
          },
        ],
      },
    ],
    tools: tools,
    stopWhen: stepCountIs(15),
    onChunk: ({ chunk }) => {
      if (chunk.type === "tool-call") {
        onProgress({
          status: "calling",
          toolName: chunk.toolName,
          message: `${formatToolName(chunk.toolName)}...`,
        });
      }
      // Extract invoice ID from tool result
      if (chunk.type === "tool-result" && chunk.toolName === "create_invoice") {
        try {
          // The output contains the result from the MCP tool
          const output = chunk.output as { content?: Array<{ type: string; text?: string }> };
          const textContent = output?.content?.find((c) => c.type === "text");
          if (textContent?.text) {
            const resultData = JSON.parse(textContent.text);
            if (resultData && typeof resultData.id === "number") {
              invoiceId = resultData.id;
            }
          }
        } catch {
          // Failed to parse result, invoice ID will remain null
          console.error("Failed to parse create_invoice result");
        }
      }
    },
  });

  // Wait for completion
  await result.text;

  return { invoiceId };
}
