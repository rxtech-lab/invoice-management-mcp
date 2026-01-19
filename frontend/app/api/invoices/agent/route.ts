import { createMCPClient } from "@ai-sdk/mcp";
import { streamText, stepCountIs } from "ai";
import { invoiceAgentPrompt } from "@/lib/prompt";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

interface AgentProgress {
  status: "idle" | "calling" | "complete" | "error";
  toolName?: string;
  message: string;
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

  // Parse request body
  let fileUrl: string;
  try {
    const body = await request.json();
    fileUrl = body.file_url;
    if (!fileUrl) {
      return new Response(
        JSON.stringify({ error: "Missing file_url in request body" }),
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
        // Connect to MCP server with the OAuth token from header
        const mcpUrl = process.env.NEXT_PUBLIC_API_URL! + "/mcp";
        const httpClient = await createMCPClient({
          transport: {
            type: "http",
            url: mcpUrl,
            headers: {
              Authorization: `Bearer ${token}`, // Pass OAuth token to backend
            },
          },
        });

        sendEvent("progress", {
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
                  text: `Create an invoice for the following file, make sure to use tools to add the invoice to the database. The original download link is ${fileUrl}, make sure add this link to the invoice`,
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
              sendEvent("progress", {
                status: "calling",
                toolName: chunk.toolName,
                message: `${formatToolName(chunk.toolName)}...`,
              });
            }
          },
        });

        // Wait for completion
        await result.text;

        sendEvent("complete", {
          status: "complete",
          message: "Invoice created successfully!",
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
