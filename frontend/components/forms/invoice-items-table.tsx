"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { InvoiceItem } from "@/lib/api/types";
import {
  addInvoiceItemAction,
  updateInvoiceItemAction,
  deleteInvoiceItemAction,
} from "@/lib/actions/invoice-item-actions";
import { FXRecalculateDialog } from "@/components/forms/fx-recalculate-dialog";
import { toast } from "sonner";
import { Loader2, Plus, Pencil, Trash, Check, X } from "lucide-react";
import { CurrencyPicker } from "@/components/currency/currency-picker";
import { ConvertedAmount } from "@/components/currency/converted-amount";
import { useExchangeRate } from "@/hooks/use-exchange-rate";
import type { CurrencyCode } from "@/lib/currency";
import { formatCurrency } from "@/lib/utils";

interface InvoiceItemsTableProps {
  invoiceId: number;
  items: InvoiceItem[];
  currency: string;
}

interface EditingItem {
  id: number | null;
  description: string;
  quantity: number;
  unit_price: number;
  target_amount: number | null; // null = auto-calculate, number = manual override
  target_amount_modified: boolean; // track if user explicitly changed target_amount
}

export function InvoiceItemsTable({
  invoiceId,
  items,
  currency,
}: InvoiceItemsTableProps) {
  const [isAdding, setIsAdding] = useState(false);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [loading, setLoading] = useState<number | null>(null);
  const [newItem, setNewItem] = useState<EditingItem>({
    id: null,
    description: "",
    quantity: 1,
    unit_price: 0,
    target_amount: null,
    target_amount_modified: false,
  });
  const [editItem, setEditItem] = useState<EditingItem>({
    id: null,
    description: "",
    quantity: 1,
    unit_price: 0,
    target_amount: null,
    target_amount_modified: false,
  });
  const [displayCurrency, setDisplayCurrency] = useState<CurrencyCode | null>(
    null
  );

  // FX recalculation dialog state
  const [showFXDialog, setShowFXDialog] = useState(false);
  const [originalUnitPrice, setOriginalUnitPrice] = useState<number | null>(null);

  const { rate, isLoading: rateLoading, convert } = useExchangeRate({
    fromCurrency: currency,
    toCurrency: displayCurrency,
  });

  const totalAmount = items.reduce((sum, item) => sum + item.amount, 0);

  const handleAddItem = async () => {
    if (!newItem.description.trim()) {
      toast.error("Description is required");
      return;
    }

    setLoading(-1);
    const result = await addInvoiceItemAction(invoiceId, {
      description: newItem.description,
      quantity: newItem.quantity,
      unit_price: newItem.unit_price,
    });

    if (result.success) {
      toast.success("Item added");
      setNewItem({ id: null, description: "", quantity: 1, unit_price: 0, target_amount: null, target_amount_modified: false });
      setIsAdding(false);
    } else {
      toast.error(result.error || "Failed to add item");
    }
    setLoading(null);
  };

  const handleStartEdit = (item: InvoiceItem) => {
    setEditingId(item.id);
    setOriginalUnitPrice(item.unit_price); // Track original price for FX dialog
    setEditItem({
      id: item.id,
      description: item.description,
      quantity: item.quantity,
      unit_price: item.unit_price,
      target_amount: item.target_amount, // Initialize with current value
      target_amount_modified: false, // Track if user explicitly changes this
    });
  };

  const handleSaveEdit = async (autoRecalculate: boolean = false) => {
    if (!editItem.id || !editItem.description.trim()) {
      toast.error("Description is required");
      return;
    }

    const itemId = editItem.id;

    setLoading(itemId);
    const result = await updateInvoiceItemAction(invoiceId, itemId, {
      description: editItem.description,
      quantity: editItem.quantity,
      unit_price: editItem.unit_price,
      // Only pass target_amount if user explicitly modified it (for manual USD override)
      target_amount: editItem.target_amount_modified ? (editItem.target_amount ?? undefined) : undefined,
      // Include auto_calculate flag if user confirmed recalculation
      auto_calculate_target_currency: autoRecalculate || undefined,
    });

    if (result.success) {
      toast.success(autoRecalculate ? "Item updated with recalculated USD amount" : "Item updated");
      setEditingId(null);
      setOriginalUnitPrice(null);
    } else {
      toast.error(result.error || "Failed to update item");
    }
    setLoading(null);
  };

  // Check if we should show FX dialog before saving
  const handleTrySaveEdit = () => {
    if (!editItem.id || !editItem.description.trim()) {
      toast.error("Description is required");
      return;
    }

    const priceChanged = originalUnitPrice !== editItem.unit_price;
    const isNonUSD = currency !== "USD";
    const targetNotManuallyModified = !editItem.target_amount_modified;

    // Show FX dialog if price changed on a non-USD invoice and target wasn't manually modified
    if (priceChanged && isNonUSD && targetNotManuallyModified) {
      setShowFXDialog(true);
    } else {
      // No dialog needed, save directly
      handleSaveEdit(false);
    }
  };

  const handleCancelEdit = () => {
    setEditingId(null);
    setEditItem({ id: null, description: "", quantity: 1, unit_price: 0, target_amount: null, target_amount_modified: false });
  };

  const handleDelete = async (itemId: number) => {
    if (!confirm("Are you sure you want to delete this item?")) return;

    setLoading(itemId);
    const result = await deleteInvoiceItemAction(invoiceId, itemId);

    if (result.success) {
      toast.success("Item deleted");
    } else {
      toast.error(result.error || "Failed to delete item");
    }
    setLoading(null);
  };

  const handleFXRecalculateConfirm = () => {
    setShowFXDialog(false);
    handleSaveEdit(true); // Save with auto_calculate_target_currency=true
  };

  const handleFXRecalculateCancel = () => {
    setShowFXDialog(false);
    handleSaveEdit(false); // Save without recalculating
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>Invoice Items</CardTitle>
        <div className="flex items-center gap-4">
          <CurrencyPicker
            value={displayCurrency}
            onChange={setDisplayCurrency}
            originalCurrency={currency}
            isLoading={rateLoading}
          />
          <Button
            size="sm"
            onClick={() => setIsAdding(true)}
            disabled={isAdding}
          >
            <Plus className="mr-2 h-4 w-4" />
            Add Item
          </Button>
        </div>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[35%]">Description</TableHead>
              <TableHead className="w-[10%]">Quantity</TableHead>
              <TableHead className="w-[15%]">Unit Price</TableHead>
              <TableHead className="w-[15%]">Amount</TableHead>
              {currency !== "USD" && (
                <TableHead className="w-[15%]">USD Amount</TableHead>
              )}
              <TableHead className="w-[10%]">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => (
              <TableRow key={item.id}>
                {editingId === item.id ? (
                  <>
                    <TableCell>
                      <Input
                        value={editItem.description}
                        onChange={(e) =>
                          setEditItem({ ...editItem, description: e.target.value })
                        }
                        placeholder="Item description"
                      />
                    </TableCell>
                    <TableCell>
                      <Input
                        type="number"
                        min="0"
                        step="0.01"
                        value={editItem.quantity}
                        onChange={(e) =>
                          setEditItem({
                            ...editItem,
                            quantity: parseFloat(e.target.value) || 0,
                          })
                        }
                      />
                    </TableCell>
                    <TableCell>
                      <Input
                        type="number"
                        min="0"
                        step="0.01"
                        value={editItem.unit_price}
                        onChange={(e) =>
                          setEditItem({
                            ...editItem,
                            unit_price: parseFloat(e.target.value) || 0,
                          })
                        }
                      />
                    </TableCell>
                    <TableCell className="font-medium">
                      <ConvertedAmount
                        originalAmount={editItem.quantity * editItem.unit_price}
                        originalCurrency={currency}
                        convertedAmount={
                          rate
                            ? convert(editItem.quantity * editItem.unit_price)
                            : null
                        }
                        displayCurrency={displayCurrency}
                      />
                    </TableCell>
                    {currency !== "USD" && (
                      <TableCell>
                        <Input
                          type="number"
                          min="0"
                          step="0.01"
                          value={editItem.target_amount ?? ""}
                          onChange={(e) =>
                            setEditItem({
                              ...editItem,
                              target_amount: e.target.value ? parseFloat(e.target.value) : null,
                              target_amount_modified: true, // Mark as modified when user changes value
                            })
                          }
                          placeholder="Auto"
                          className="w-24"
                        />
                      </TableCell>
                    )}
                    <TableCell>
                      <div className="flex gap-1">
                        <Button
                          size="icon"
                          variant="ghost"
                          onClick={handleTrySaveEdit}
                          disabled={loading === item.id}
                        >
                          {loading === item.id ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <Check className="h-4 w-4 text-green-600" />
                          )}
                        </Button>
                        <Button
                          size="icon"
                          variant="ghost"
                          onClick={handleCancelEdit}
                        >
                          <X className="h-4 w-4 text-red-600" />
                        </Button>
                      </div>
                    </TableCell>
                  </>
                ) : (
                  <>
                    <TableCell>{item.description}</TableCell>
                    <TableCell>{item.quantity}</TableCell>
                    <TableCell>
                      <ConvertedAmount
                        originalAmount={item.unit_price}
                        originalCurrency={currency}
                        convertedAmount={rate ? convert(item.unit_price) : null}
                        displayCurrency={displayCurrency}
                        showOriginal={false}
                      />
                    </TableCell>
                    <TableCell className="font-medium">
                      <ConvertedAmount
                        originalAmount={item.amount}
                        originalCurrency={currency}
                        convertedAmount={rate ? convert(item.amount) : null}
                        displayCurrency={displayCurrency}
                      />
                    </TableCell>
                    {currency !== "USD" && (
                      <TableCell className="text-muted-foreground">
                        {formatCurrency(item.target_amount || item.amount, "USD")}
                      </TableCell>
                    )}
                    <TableCell>
                      <div className="flex gap-1">
                        <Button
                          size="icon"
                          variant="ghost"
                          onClick={() => handleStartEdit(item)}
                          disabled={loading === item.id}
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                          size="icon"
                          variant="ghost"
                          onClick={() => handleDelete(item.id)}
                          disabled={loading === item.id}
                        >
                          {loading === item.id ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <Trash className="h-4 w-4 text-destructive" />
                          )}
                        </Button>
                      </div>
                    </TableCell>
                  </>
                )}
              </TableRow>
            ))}
            {isAdding && (
              <TableRow>
                <TableCell>
                  <Input
                    value={newItem.description}
                    onChange={(e) =>
                      setNewItem({ ...newItem, description: e.target.value })
                    }
                    placeholder="Item description"
                    autoFocus
                  />
                </TableCell>
                <TableCell>
                  <Input
                    type="number"
                    min="0"
                    step="0.01"
                    value={newItem.quantity}
                    onChange={(e) =>
                      setNewItem({
                        ...newItem,
                        quantity: parseFloat(e.target.value) || 0,
                      })
                    }
                  />
                </TableCell>
                <TableCell>
                  <Input
                    type="number"
                    min="0"
                    step="0.01"
                    value={newItem.unit_price}
                    onChange={(e) =>
                      setNewItem({
                        ...newItem,
                        unit_price: parseFloat(e.target.value) || 0,
                      })
                    }
                  />
                </TableCell>
                <TableCell className="font-medium">
                  <ConvertedAmount
                    originalAmount={newItem.quantity * newItem.unit_price}
                    originalCurrency={currency}
                    convertedAmount={
                      rate
                        ? convert(newItem.quantity * newItem.unit_price)
                        : null
                    }
                    displayCurrency={displayCurrency}
                  />
                </TableCell>
                {currency !== "USD" && (
                  <TableCell className="text-muted-foreground text-sm">
                    (calculated on save)
                  </TableCell>
                )}
                <TableCell>
                  <div className="flex gap-1">
                    <Button
                      size="icon"
                      variant="ghost"
                      onClick={handleAddItem}
                      disabled={loading === -1}
                    >
                      {loading === -1 ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <Check className="h-4 w-4 text-green-600" />
                      )}
                    </Button>
                    <Button
                      size="icon"
                      variant="ghost"
                      onClick={() => {
                        setIsAdding(false);
                        setNewItem({
                          id: null,
                          description: "",
                          quantity: 1,
                          unit_price: 0,
                          target_amount: null,
                          target_amount_modified: false,
                        });
                      }}
                    >
                      <X className="h-4 w-4 text-red-600" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            )}
            {items.length === 0 && !isAdding && (
              <TableRow>
                <TableCell colSpan={currency !== "USD" ? 6 : 5} className="h-24 text-center">
                  No items yet. Click &quot;Add Item&quot; to add one.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
        {items.length > 0 && (
          <div className="mt-4 flex justify-end gap-6 border-t pt-4">
            <div className="text-lg font-semibold">
              Total:{" "}
              <ConvertedAmount
                originalAmount={totalAmount}
                originalCurrency={currency}
                convertedAmount={rate ? convert(totalAmount) : null}
                displayCurrency={displayCurrency}
              />
            </div>
            {currency !== "USD" && (
              <div className="text-lg font-semibold text-muted-foreground">
                USD Total: {formatCurrency(
                  items.reduce((sum, item) => sum + (item.target_amount || item.amount), 0),
                  "USD"
                )}
              </div>
            )}
          </div>
        )}
      </CardContent>

      {/* FX Recalculation Dialog */}
      <FXRecalculateDialog
        open={showFXDialog}
        onOpenChange={setShowFXDialog}
        onConfirm={handleFXRecalculateConfirm}
        onCancel={handleFXRecalculateCancel}
        currency={currency}
      />
    </Card>
  );
}
