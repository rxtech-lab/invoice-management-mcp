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
import { toast } from "sonner";
import { Loader2, Plus, Pencil, Trash, Check, X } from "lucide-react";
import { CurrencyPicker } from "@/components/currency/currency-picker";
import { ConvertedAmount } from "@/components/currency/converted-amount";
import { useExchangeRate } from "@/hooks/use-exchange-rate";
import type { CurrencyCode } from "@/lib/currency";

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
  });
  const [editItem, setEditItem] = useState<EditingItem>({
    id: null,
    description: "",
    quantity: 1,
    unit_price: 0,
  });
  const [displayCurrency, setDisplayCurrency] = useState<CurrencyCode | null>(
    null
  );

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
      setNewItem({ id: null, description: "", quantity: 1, unit_price: 0 });
      setIsAdding(false);
    } else {
      toast.error(result.error || "Failed to add item");
    }
    setLoading(null);
  };

  const handleStartEdit = (item: InvoiceItem) => {
    setEditingId(item.id);
    setEditItem({
      id: item.id,
      description: item.description,
      quantity: item.quantity,
      unit_price: item.unit_price,
    });
  };

  const handleSaveEdit = async () => {
    if (!editItem.id || !editItem.description.trim()) {
      toast.error("Description is required");
      return;
    }

    setLoading(editItem.id);
    const result = await updateInvoiceItemAction(invoiceId, editItem.id, {
      description: editItem.description,
      quantity: editItem.quantity,
      unit_price: editItem.unit_price,
    });

    if (result.success) {
      toast.success("Item updated");
      setEditingId(null);
    } else {
      toast.error(result.error || "Failed to update item");
    }
    setLoading(null);
  };

  const handleCancelEdit = () => {
    setEditingId(null);
    setEditItem({ id: null, description: "", quantity: 1, unit_price: 0 });
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
              <TableHead className="w-[40%]">Description</TableHead>
              <TableHead className="w-[15%]">Quantity</TableHead>
              <TableHead className="w-[20%]">Unit Price</TableHead>
              <TableHead className="w-[15%]">Amount</TableHead>
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
                    <TableCell>
                      <div className="flex gap-1">
                        <Button
                          size="icon"
                          variant="ghost"
                          onClick={handleSaveEdit}
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
                <TableCell colSpan={5} className="h-24 text-center">
                  No items yet. Click "Add Item" to add one.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
        {items.length > 0 && (
          <div className="mt-4 flex justify-end border-t pt-4">
            <div className="text-lg font-semibold">
              Total:{" "}
              <ConvertedAmount
                originalAmount={totalAmount}
                originalCurrency={currency}
                convertedAmount={rate ? convert(totalAmount) : null}
                displayCurrency={displayCurrency}
              />
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
