"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { Receiver } from "@/lib/api/types";
import { createReceiverAction, updateReceiverAction } from "@/lib/actions/receiver-actions";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Loader2, X, Edit2, Check } from "lucide-react";

const receiverSchema = z.object({
  name: z.string().min(1, "Name is required"),
  is_organization: z.boolean().optional(),
});

type ReceiverFormData = z.infer<typeof receiverSchema>;

interface ReceiverFormProps {
  receiver?: Receiver;
}

export function ReceiverForm({ receiver }: ReceiverFormProps) {
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const isEditing = !!receiver;

  // State for managing other_names (aliases)
  const [otherNames, setOtherNames] = useState<string[]>(receiver?.other_names || []);
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editValue, setEditValue] = useState("");

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<ReceiverFormData>({
    resolver: zodResolver(receiverSchema),
    defaultValues: {
      name: receiver?.name || "",
      is_organization: receiver?.is_organization || false,
    },
  });

  const isOrganization = watch("is_organization");

  const handleStartEdit = (index: number, currentValue: string) => {
    setEditingIndex(index);
    setEditValue(currentValue);
  };

  const handleSaveEdit = (index: number) => {
    if (editValue.trim()) {
      const newNames = [...otherNames];
      newNames[index] = editValue.trim();
      setOtherNames(newNames);
    }
    setEditingIndex(null);
    setEditValue("");
  };

  const handleCancelEdit = () => {
    setEditingIndex(null);
    setEditValue("");
  };

  const handleRemoveName = (index: number) => {
    setOtherNames(otherNames.filter((_, i) => i !== index));
  };

  const onSubmit = async (data: ReceiverFormData) => {
    setIsSubmitting(true);
    try {
      const payload = {
        name: data.name,
        is_organization: data.is_organization || false,
        ...(isEditing && { other_names: otherNames }),
      };

      const result = isEditing
        ? await updateReceiverAction(receiver.id, payload)
        : await createReceiverAction(payload);

      if (result.success) {
        toast.success(isEditing ? "Receiver updated" : "Receiver created");
        router.push("/receivers");
      } else {
        toast.error(result.error || "Failed to save receiver");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{isEditing ? "Edit Receiver" : "Create Receiver"}</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="name">Receiver Name *</Label>
            <Input id="name" {...register("name")} placeholder="Receiver name" />
            {errors.name && (
              <p className="text-sm text-destructive">{errors.name.message}</p>
            )}
          </div>

          <div className="flex items-center space-x-3 rounded-lg border p-4">
            <Checkbox
              id="is_organization"
              checked={isOrganization}
              onCheckedChange={(checked) => setValue("is_organization", checked === true)}
            />
            <div className="space-y-0.5">
              <Label htmlFor="is_organization" className="cursor-pointer">Organization</Label>
              <p className="text-sm text-muted-foreground">
                Is this receiver an organization (company, business, etc.)?
              </p>
            </div>
          </div>

          {/* Other Names Section - Only show when editing and has aliases */}
          {isEditing && otherNames.length > 0 && (
            <div className="space-y-3">
              <div>
                <Label>Alternative Names (Aliases)</Label>
                <p className="text-sm text-muted-foreground">
                  These are alternative names collected from merged receivers. You can edit or remove them.
                </p>
              </div>
              <div className="space-y-2">
                {otherNames.map((name, index) => (
                  <div key={index} className="flex items-center gap-2">
                    {editingIndex === index ? (
                      <>
                        <Input
                          value={editValue}
                          onChange={(e) => setEditValue(e.target.value)}
                          className="flex-1"
                          autoFocus
                          onKeyDown={(e) => {
                            if (e.key === "Enter") {
                              e.preventDefault();
                              handleSaveEdit(index);
                            } else if (e.key === "Escape") {
                              handleCancelEdit();
                            }
                          }}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => handleSaveEdit(index)}
                        >
                          <Check className="h-4 w-4" />
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={handleCancelEdit}
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      </>
                    ) : (
                      <>
                        <div className="flex-1 rounded-md border bg-muted/50 px-3 py-2 text-sm">
                          {name}
                        </div>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => handleStartEdit(index, name)}
                          title="Edit alias"
                        >
                          <Edit2 className="h-4 w-4" />
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => handleRemoveName(index)}
                          title="Remove alias"
                        >
                          <X className="h-4 w-4" />
                        </Button>
                      </>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}

          <div className="flex gap-4">
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {isEditing ? "Update Receiver" : "Create Receiver"}
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
