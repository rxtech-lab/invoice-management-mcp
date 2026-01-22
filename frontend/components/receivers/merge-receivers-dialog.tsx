"use client";

import { useState, ReactNode } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Loader2, Building2, User } from "lucide-react";
import { toast } from "sonner";
import { mergeReceiversAction } from "@/lib/actions/receiver-actions";
import type { Receiver } from "@/lib/api/types";

interface MergeReceiversDialogProps {
  targetReceiver: Receiver;
  allReceivers: Receiver[];
  children: ReactNode;
}

export function MergeReceiversDialog({
  targetReceiver,
  allReceivers,
  children,
}: MergeReceiversDialogProps) {
  const [open, setOpen] = useState(false);
  const [sourceIds, setSourceIds] = useState<number[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const availableSourceReceivers = allReceivers.filter(
    (r) => r.id !== targetReceiver.id
  );

  const handleSourceToggle = (receiverId: number, checked: boolean) => {
    if (checked) {
      setSourceIds([...sourceIds, receiverId]);
    } else {
      setSourceIds(sourceIds.filter((id) => id !== receiverId));
    }
  };

  const handleMerge = async () => {
    if (sourceIds.length === 0) {
      toast.error("Please select at least one source receiver");
      return;
    }

    setIsSubmitting(true);
    try {
      const result = await mergeReceiversAction(targetReceiver.id, sourceIds);
      if (result.success) {
        toast.success(
          `Merged ${result.data?.merged_count} receivers. ${result.data?.invoices_updated} invoices updated.`
        );
        setOpen(false);
        setSourceIds([]);
      } else {
        toast.error(result.error || "Failed to merge receivers");
      }
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    setOpen(newOpen);
    if (!newOpen) {
      setSourceIds([]);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Merge into {targetReceiver.name}</DialogTitle>
          <DialogDescription>
            Select receivers to merge into{" "}
            <span className="font-medium">{targetReceiver.name}</span>. All
            invoices from selected receivers will be moved to this receiver, and
            selected receivers will be deleted.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label>Target Receiver (keep)</Label>
            <div className="flex items-center gap-2 p-3 rounded-md border bg-muted/50">
              {targetReceiver.is_organization ? (
                <Building2 className="h-4 w-4 text-muted-foreground" />
              ) : (
                <User className="h-4 w-4 text-muted-foreground" />
              )}
              <span className="font-medium">{targetReceiver.name}</span>
              {targetReceiver.is_organization && (
                <span className="text-xs text-muted-foreground">(Org)</span>
              )}
            </div>
          </div>

          <div className="space-y-2">
            <Label>Source Receivers (merge & delete)</Label>
            <div className="rounded-md border p-4 space-y-3 max-h-[200px] overflow-y-auto">
              {availableSourceReceivers.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  No other receivers available to merge
                </p>
              ) : (
                availableSourceReceivers.map((receiver) => (
                  <div
                    key={receiver.id}
                    className="flex items-center space-x-2"
                  >
                    <Checkbox
                      id={`source-${receiver.id}`}
                      checked={sourceIds.includes(receiver.id)}
                      onCheckedChange={(checked) =>
                        handleSourceToggle(receiver.id, checked as boolean)
                      }
                    />
                    <Label
                      htmlFor={`source-${receiver.id}`}
                      className="text-sm font-normal cursor-pointer flex items-center gap-1"
                    >
                      {receiver.is_organization ? (
                        <Building2 className="h-3 w-3 text-muted-foreground" />
                      ) : (
                        <User className="h-3 w-3 text-muted-foreground" />
                      )}
                      {receiver.name}
                      {receiver.is_organization && (
                        <span className="text-xs text-muted-foreground">
                          (Org)
                        </span>
                      )}
                    </Label>
                  </div>
                ))
              )}
            </div>
            {sourceIds.length > 0 && (
              <p className="text-sm text-muted-foreground">
                {sourceIds.length} receiver(s) selected to merge
              </p>
            )}
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => setOpen(false)}
            disabled={isSubmitting}
          >
            Cancel
          </Button>
          <Button
            onClick={handleMerge}
            disabled={sourceIds.length === 0 || isSubmitting}
          >
            {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Merge Receivers
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
