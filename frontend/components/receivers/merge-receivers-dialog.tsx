"use client";

import { useState } from "react";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Loader2, GitMerge } from "lucide-react";
import { toast } from "sonner";
import { mergeReceiversAction } from "@/lib/actions/receiver-actions";
import type { Receiver } from "@/lib/api/types";

interface MergeReceiversDialogProps {
  receivers: Receiver[];
}

export function MergeReceiversDialog({ receivers }: MergeReceiversDialogProps) {
  const [open, setOpen] = useState(false);
  const [targetId, setTargetId] = useState<string>("");
  const [sourceIds, setSourceIds] = useState<number[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const availableSourceReceivers = receivers.filter(
    (r) => r.id.toString() !== targetId
  );

  const handleSourceToggle = (receiverId: number, checked: boolean) => {
    if (checked) {
      setSourceIds([...sourceIds, receiverId]);
    } else {
      setSourceIds(sourceIds.filter((id) => id !== receiverId));
    }
  };

  const handleMerge = async () => {
    if (!targetId || sourceIds.length === 0) {
      toast.error("Please select a target and at least one source receiver");
      return;
    }

    setIsSubmitting(true);
    try {
      const result = await mergeReceiversAction(parseInt(targetId), sourceIds);
      if (result.success) {
        toast.success(
          `Merged ${result.data?.merged_count} receivers. ${result.data?.invoices_updated} invoices updated.`
        );
        setOpen(false);
        setTargetId("");
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
      setTargetId("");
      setSourceIds([]);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>
        <Button variant="outline">
          <GitMerge className="mr-2 h-4 w-4" />
          Merge Receivers
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Merge Receivers</DialogTitle>
          <DialogDescription>
            Combine multiple receivers into one. All invoices from source
            receivers will be moved to the target receiver, and source receivers
            will be deleted.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="space-y-2">
            <Label htmlFor="target">Target Receiver (keep)</Label>
            <Select value={targetId} onValueChange={setTargetId}>
              <SelectTrigger>
                <SelectValue placeholder="Select receiver to keep" />
              </SelectTrigger>
              <SelectContent>
                {receivers.map((receiver) => (
                  <SelectItem key={receiver.id} value={receiver.id.toString()}>
                    {receiver.name}
                    {receiver.is_organization && (
                      <span className="text-xs text-muted-foreground ml-1">
                        (Org)
                      </span>
                    )}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {targetId && (
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
                        className="text-sm font-normal cursor-pointer"
                      >
                        {receiver.name}
                        {receiver.is_organization && (
                          <span className="text-xs text-muted-foreground ml-1">
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
          )}
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
            disabled={!targetId || sourceIds.length === 0 || isSubmitting}
          >
            {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Merge Receivers
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
