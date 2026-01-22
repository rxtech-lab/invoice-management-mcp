"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import {
  Category,
  Company,
  Receiver,
  Invoice,
  InvoiceStatus,
  Tag,
} from "@/lib/api/types";
import {
  createInvoiceAction,
  updateInvoiceAction,
} from "@/lib/actions/invoice-actions";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { useMemo, useState, useEffect } from "react";
import { Loader2 } from "lucide-react";
import { useDebounce } from "@uidotdev/usehooks";
import { searchTagsAction } from "@/lib/actions/tag-actions";
import { FileUpload } from "@/components/ui/file-upload";
import {
  Combobox,
  ComboboxChips,
  ComboboxChip,
  ComboboxChipsInput,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxItem,
  ComboboxList,
  useComboboxAnchor,
} from "@/components/ui/combobox";

// Note: amount and target_amount are not in the schema - they're calculated from invoice items
const invoiceSchema = z.object({
  title: z.string().min(1, "Title is required"),
  description: z.string().optional(),
  currency: z.string().min(1),
  category_id: z.coerce.number().optional().nullable(),
  company_id: z.coerce.number().optional().nullable(),
  receiver_id: z.coerce.number().optional().nullable(),
  tag_ids: z.array(z.number()).optional(),
  status: z.enum(["paid", "unpaid", "overdue"]),
  due_date: z.string().optional().nullable(),
  invoice_started_at: z.string().optional().nullable(),
  invoice_ended_at: z.string().optional().nullable(),
  original_download_link: z.string().optional().nullable(),
});

type InvoiceFormData = z.output<typeof invoiceSchema>;

interface InvoiceFormProps {
  invoice?: Invoice;
  categories: Category[];
  companies: Company[];
  receivers: Receiver[];
  tags: Tag[];
}

export function InvoiceForm({
  invoice,
  categories,
  companies,
  receivers,
  tags,
}: InvoiceFormProps) {
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [tagSearch, setTagSearch] = useState("");
  const debouncedTagSearch = useDebounce(tagSearch, 300);
  const [searchedTags, setSearchedTags] = useState<Tag[]>(tags);
  const [isLoadingTags, setIsLoadingTags] = useState(false);
  const tagAnchorRef = useComboboxAnchor();
  const isEditing = !!invoice;

  // Fetch tags from server when debounced search changes
  useEffect(() => {
    const fetchTags = async () => {
      setIsLoadingTags(true);
      try {
        const response = await searchTagsAction(
          debouncedTagSearch || undefined,
          20,
        );
        setSearchedTags(response.data || []);
      } catch (error) {
        console.error("Failed to fetch tags:", error);
        setSearchedTags([]);
      } finally {
        setIsLoadingTags(false);
      }
    };
    fetchTags();
  }, [debouncedTagSearch]);

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors, isValid },
  } = useForm<InvoiceFormData>({
    mode: "onChange",
    resolver: zodResolver(invoiceSchema),
    defaultValues: {
      title: invoice?.title || "",
      description: invoice?.description || "",
      currency: invoice?.currency || "USD",
      category_id: invoice?.category_id || null,
      company_id: invoice?.company_id || null,
      receiver_id: invoice?.receiver_id || null,
      tag_ids:
        invoice?.tags
          ?.map((t) => t.id)
          .filter((id): id is number => typeof id === "number") || [],
      status: invoice?.status || "unpaid",
      due_date: invoice?.due_date?.split("T")[0] || "",
      invoice_started_at: invoice?.invoice_started_at?.split("T")[0] || "",
      invoice_ended_at: invoice?.invoice_ended_at?.split("T")[0] || "",
      original_download_link: invoice?.original_download_link || "",
    },
  });

  // Combine searched tags with selected tags to ensure selected tags are always visible
  const selectedTagIds = watch("tag_ids") || [];
  const displayTags = useMemo(() => {
    const selectedTags = tags.filter((t) => selectedTagIds.includes(t.id));
    const allTags = [...searchedTags];

    // Add selected tags that aren't in search results
    selectedTags.forEach((selectedTag) => {
      if (!allTags.find((t) => t.id === selectedTag.id)) {
        allTags.push(selectedTag);
      }
    });

    return allTags;
  }, [searchedTags, tags, selectedTagIds]);

  const onSubmit = async (data: InvoiceFormData) => {
    setIsSubmitting(true);
    try {
      // Note: amount and target_amount are not included - they're calculated from invoice items
      const payload = {
        title: data.title,
        description: data.description || undefined,
        currency: data.currency,
        status: data.status,
        category_id: data.category_id || undefined,
        company_id: data.company_id || undefined,
        receiver_id: data.receiver_id || undefined,
        tag_ids: data.tag_ids?.length ? data.tag_ids : undefined,
        due_date: data.due_date
          ? new Date(data.due_date).toISOString()
          : undefined,
        invoice_started_at: data.invoice_started_at
          ? new Date(data.invoice_started_at).toISOString()
          : undefined,
        invoice_ended_at: data.invoice_ended_at
          ? new Date(data.invoice_ended_at).toISOString()
          : undefined,
        original_download_link: data.original_download_link || undefined,
      };

      const result = isEditing
        ? await updateInvoiceAction(invoice.id, payload)
        : await createInvoiceAction(payload);

      if (result.success) {
        toast.success(isEditing ? "Invoice updated" : "Invoice created");
        if (result.data) {
          router.push(`/invoices/${result.data.id}`);
        }
      } else {
        toast.error(result.error || "Failed to save invoice");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{isEditing ? "Edit Invoice" : "Create Invoice"}</CardTitle>
      </CardHeader>
      <CardContent>
        <form
          onSubmit={handleSubmit(onSubmit, (errors) => {
            const firstError = Object.values(errors)[0];
            const message = Array.isArray(firstError)
              ? "Invalid form data"
              : firstError?.message || "Please fix the form errors";
            toast.error(message);
          })}
          className="space-y-6"
        >
          <div className="space-y-2">
            <Label htmlFor="title">Title *</Label>
            <Input
              id="title"
              {...register("title")}
              placeholder="Invoice title"
            />
            {errors.title && (
              <p className="text-sm text-destructive">{errors.title.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              {...register("description")}
              placeholder="Invoice description"
              rows={3}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="currency">Currency</Label>
            <Input
              id="currency"
              {...register("currency")}
              placeholder="USD"
              className="max-w-xs"
            />
            <p className="text-sm text-muted-foreground">
              Amount and USD equivalent are calculated from invoice items
            </p>
          </div>

          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label>Category</Label>
              <Select
                value={watch("category_id")?.toString() || ""}
                onValueChange={(v) =>
                  setValue("category_id", v ? Number(v) : null)
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select category" />
                </SelectTrigger>
                <SelectContent>
                  {categories.map((cat) => (
                    <SelectItem key={cat.id} value={cat.id.toString()}>
                      <div className="flex items-center gap-2">
                        {cat.color && (
                          <div
                            className="h-3 w-3 rounded-full"
                            style={{ backgroundColor: cat.color }}
                          />
                        )}
                        {cat.name}
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label>Company</Label>
              <Select
                value={watch("company_id")?.toString() || ""}
                onValueChange={(v) =>
                  setValue("company_id", v ? Number(v) : null)
                }
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select company" />
                </SelectTrigger>
                <SelectContent>
                  {companies.map((comp) => (
                    <SelectItem key={comp.id} value={comp.id.toString()}>
                      {comp.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          <div className="space-y-2">
            <Label>Receiver</Label>
            <Select
              value={watch("receiver_id")?.toString() || ""}
              onValueChange={(v) =>
                setValue("receiver_id", v ? Number(v) : null)
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="Select receiver" />
              </SelectTrigger>
              <SelectContent>
                {receivers.map((rec) => (
                  <SelectItem key={rec.id} value={rec.id.toString()}>
                    <div className="flex items-center gap-2">
                      {rec.name}
                      {rec.is_organization && (
                        <span className="text-xs text-muted-foreground">
                          (Org)
                        </span>
                      )}
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>Tags</Label>
            <Combobox
              multiple
              value={watch("tag_ids") || []}
              onValueChange={(newValue) =>
                setValue("tag_ids", newValue as number[])
              }
            >
              <ComboboxChips ref={tagAnchorRef}>
                {(watch("tag_ids") || []).map((tagId) => {
                  const tag = tags.find((t) => t.id === tagId);
                  return tag ? (
                    <ComboboxChip key={tagId} value={tagId}>
                      <div
                        className="h-2 w-2 rounded-full"
                        style={{ backgroundColor: tag.color }}
                      />
                      {tag.name}
                    </ComboboxChip>
                  ) : null;
                })}
                <ComboboxChipsInput
                  placeholder={watch("tag_ids")?.length ? "" : "Search tags..."}
                  value={tagSearch}
                  onChange={(e) => setTagSearch(e.target.value)}
                />
              </ComboboxChips>
              <ComboboxContent anchor={tagAnchorRef}>
                <ComboboxList>
                  {!isLoadingTags && displayTags.length === 0 && (
                    <ComboboxEmpty>No tags found</ComboboxEmpty>
                  )}
                  {isLoadingTags && <ComboboxEmpty>Loading...</ComboboxEmpty>}
                  {displayTags.map((tag) => (
                    <ComboboxItem key={tag.id} value={tag.id}>
                      <div
                        className="h-2 w-2 rounded-full"
                        style={{ backgroundColor: tag.color }}
                      />
                      {tag.name}
                    </ComboboxItem>
                  ))}
                </ComboboxList>
              </ComboboxContent>
            </Combobox>
          </div>

          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label>Status</Label>
              <Select
                value={watch("status")}
                onValueChange={(v) => setValue("status", v as InvoiceStatus)}
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="unpaid">Unpaid</SelectItem>
                  <SelectItem value="paid">Paid</SelectItem>
                  <SelectItem value="overdue">Overdue</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="due_date">Due Date</Label>
              <Input id="due_date" type="date" {...register("due_date")} />
            </div>
          </div>

          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="invoice_started_at">Billing Start Date</Label>
              <Input
                id="invoice_started_at"
                type="date"
                {...register("invoice_started_at")}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="invoice_ended_at">Billing End Date</Label>
              <Input
                id="invoice_ended_at"
                type="date"
                {...register("invoice_ended_at")}
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label>Invoice File</Label>
            <FileUpload
              value={watch("original_download_link")}
              onChange={(key) => setValue("original_download_link", key || "")}
              accept="application/pdf,.pdf,image/*"
              maxSizeMB={10}
            />
          </div>

          <div className="flex gap-4">
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              {isEditing ? "Update Invoice" : "Create Invoice"}
            </Button>
            <Button
              type="button"
              variant="outline"
              onClick={() => router.back()}
            >
              Cancel
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
