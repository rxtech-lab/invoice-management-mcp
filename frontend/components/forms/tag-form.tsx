"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Tag } from "@/lib/api/types";
import { createTagAction, updateTagAction } from "@/lib/actions/tag-actions";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Loader2 } from "lucide-react";

const tagSchema = z.object({
  name: z.string().min(1, "Name is required").max(100, "Name must be 100 characters or less"),
  color: z
    .string()
    .regex(/^#[0-9A-Fa-f]{6}$/, "Must be a valid hex color (e.g., #FF5733)")
    .optional()
    .or(z.literal("")),
});

type TagFormData = z.infer<typeof tagSchema>;

const presetColors = [
  "#EF4444", // Red
  "#F97316", // Orange
  "#EAB308", // Yellow
  "#22C55E", // Green
  "#14B8A6", // Teal
  "#3B82F6", // Blue
  "#8B5CF6", // Violet
  "#EC4899", // Pink
  "#6B7280", // Gray
];

interface TagFormProps {
  tag?: Tag;
}

export function TagForm({ tag }: TagFormProps) {
  const router = useRouter();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const isEditing = !!tag;

  const {
    register,
    handleSubmit,
    watch,
    setValue,
    formState: { errors },
  } = useForm<TagFormData>({
    resolver: zodResolver(tagSchema),
    defaultValues: {
      name: tag?.name || "",
      color: tag?.color || "#6B7280",
    },
  });

  const selectedColor = watch("color");

  const onSubmit = async (data: TagFormData) => {
    setIsSubmitting(true);
    try {
      const payload = {
        ...data,
        color: data.color || undefined,
      };

      const result = isEditing
        ? await updateTagAction(tag.id, payload)
        : await createTagAction(payload);

      if (result.success) {
        toast.success(isEditing ? "Tag updated" : "Tag created");
        router.push("/tags");
      } else {
        toast.error(result.error || "Failed to save tag");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>{isEditing ? "Edit Tag" : "Create Tag"}</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          <div className="space-y-2">
            <Label htmlFor="name">Tag Name *</Label>
            <Input id="name" {...register("name")} placeholder="Tag name" />
            {errors.name && (
              <p className="text-sm text-destructive">{errors.name.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="color">Color</Label>
            <div className="space-y-3">
              <div className="flex items-center gap-2">
                <div
                  className="h-10 w-10 rounded-md border"
                  style={{ backgroundColor: selectedColor || "#6B7280" }}
                />
                <Input
                  id="color"
                  {...register("color")}
                  placeholder="#FF5733"
                  className="flex-1"
                />
              </div>
              <div className="flex flex-wrap gap-2">
                {presetColors.map((color) => (
                  <button
                    key={color}
                    type="button"
                    className={`h-8 w-8 rounded-md border-2 transition-transform hover:scale-110 ${
                      selectedColor === color
                        ? "border-foreground"
                        : "border-transparent"
                    }`}
                    style={{ backgroundColor: color }}
                    onClick={() => setValue("color", color)}
                  />
                ))}
              </div>
              {errors.color && (
                <p className="text-sm text-destructive">{errors.color.message}</p>
              )}
            </div>
          </div>

          <div className="flex gap-4">
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {isEditing ? "Update Tag" : "Create Tag"}
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
