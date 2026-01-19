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
import { Loader2 } from "lucide-react";

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

  const onSubmit = async (data: ReceiverFormData) => {
    setIsSubmitting(true);
    try {
      const payload = {
        name: data.name,
        is_organization: data.is_organization || false,
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
