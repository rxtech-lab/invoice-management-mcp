import {
  streamText,
  convertToModelMessages,
  UIMessage,
  tool,
  stepCountIs,
} from "ai";
import { createMCPClient } from "@ai-sdk/mcp";
import { z } from "zod";
import { auth } from "@/auth";
import { searchAgentPrompt } from "@/lib/search/prompt";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";
export const maxDuration = 60;

// Schema for display_invoices tool
const displayInvoicesSchema = z.object({
  invoices: z.array(
    z.object({
      id: z.number().describe("Invoice ID"),
      title: z.string().describe("Invoice title"),
      description: z.string().optional().describe("Invoice description"),
      amount: z.number().describe("Invoice amount"),
      currency: z.string().describe("Currency code"),
      status: z
        .enum(["paid", "unpaid", "overdue"])
        .describe("Invoice status"),
      company: z
        .object({
          id: z.number(),
          name: z.string(),
        })
        .optional()
        .describe("Company info"),
      category: z
        .object({
          id: z.number(),
          name: z.string(),
          color: z.string().optional(),
        })
        .optional()
        .describe("Category info"),
      receiver: z
        .object({
          id: z.number(),
          name: z.string(),
        })
        .optional()
        .describe("Receiver info"),
      due_date: z.string().nullable().describe("Due date"),
      created_at: z.string().describe("Creation date"),
    })
  ),
  total: z.number().describe("Total number of matching invoices"),
  query: z
    .string()
    .optional()
    .describe("The search query used"),
});

type DisplayInvoicesInput = z.infer<typeof displayInvoicesSchema>;

// Schema for display_statistics tool
const displayStatisticsSchema = z.object({
  period: z.string().describe("Period description (e.g., 'last_week', 'last_month')"),
  start_date: z.string().describe("Start date of the period (ISO format)"),
  end_date: z.string().describe("End date of the period (ISO format)"),
  total_amount: z.number().describe("Total amount across all invoices"),
  total_count: z.number().describe("Total number of invoices"),
  currency: z.string().optional().describe("Currency code (default: USD)"),
  breakdown: z.array(
    z.object({
      date: z.string().optional().describe("Date for time-based breakdown"),
      name: z.string().optional().describe("Name for entity-based breakdown"),
      id: z.number().optional().describe("ID for entity-based breakdown"),
      amount: z.number().describe("Amount for this breakdown item"),
      count: z.number().describe("Number of invoices for this breakdown item"),
    })
  ).optional().describe("Breakdown data for charts"),
  status_breakdown: z.object({
    paid: z.object({ amount: z.number(), count: z.number() }),
    unpaid: z.object({ amount: z.number(), count: z.number() }),
    overdue: z.object({ amount: z.number(), count: z.number() }),
  }).optional().describe("Breakdown by status"),
  aggregations: z.object({
    max_amount: z.number(),
    min_amount: z.number(),
    avg_amount: z.number(),
    max_invoice: z.object({
      id: z.number(),
      title: z.string(),
    }).optional(),
  }).optional().describe("Aggregation statistics"),
});

type DisplayStatisticsInput = z.infer<typeof displayStatisticsSchema>;

// Client-side tool for displaying invoices in the UI
const displayInvoicesTool = {
  display_invoices: tool<DisplayInvoicesInput, DisplayInvoicesInput>({
    description:
      "Display the relevant invoices to the user based on their query. ALWAYS use this tool to show invoice search results after using search_invoices or list_invoices. Only include invoices that are most relevant to the user's specific request.",
    inputSchema: displayInvoicesSchema,
    execute: async (input) => {
      return { invoices: input.invoices, total: input.total, query: input.query };
    },
  }),
};

// Client-side tool for displaying statistics charts
const displayStatisticsTool = {
  display_statistics: tool<DisplayStatisticsInput, DisplayStatisticsInput>({
    description:
      "Display statistics with visual charts to the user. ALWAYS use this tool after calling invoice_statistics to show the results as charts. Pass the data from invoice_statistics directly to this tool.",
    inputSchema: displayStatisticsSchema,
    execute: async (input) => {
      return input;
    },
  }),
};

export async function POST(req: Request) {
  const session = await auth();
  if (!session?.accessToken) {
    return Response.json({ error: "Unauthorized" }, { status: 401 });
  }

  let mcpClient: Awaited<ReturnType<typeof createMCPClient>> | null = null;

  try {
    const { messages }: { messages: UIMessage[] } = await req.json();

    const url = process.env.NEXT_PUBLIC_API_URL + "/mcp";

    mcpClient = await createMCPClient({
      transport: {
        type: "http",
        url: url,
        headers: {
          Authorization: `Bearer ${session.accessToken}`,
        },
      },
    });

    const mcpTools = await mcpClient.tools();

    // Merge MCP tools with display tools
    const tools = { ...mcpTools, ...displayInvoicesTool, ...displayStatisticsTool };

    // Use model from env
    const modelId =
      process.env.SEARCH_AGENT_MODEL || "openai/gpt-4o";

    const result = streamText({
      model: modelId as Parameters<typeof streamText>[0]["model"],
      system: searchAgentPrompt,
      messages: await convertToModelMessages(messages),
      tools,
      stopWhen: stepCountIs(10),
      onFinish: async () => {
        if (mcpClient) {
          await mcpClient.close();
        }
      },
    });

    return result.toUIMessageStreamResponse();
  } catch (error) {
    console.error("[search-agent] Error:", error);
    if (mcpClient) {
      await mcpClient.close();
    }
    return Response.json(
      { error: error instanceof Error ? error.message : "Internal server error" },
      { status: 500 }
    );
  }
}
