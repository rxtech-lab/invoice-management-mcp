"use client";

import * as React from "react";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

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

interface ColorPickerProps {
  value: string;
  onChange: (color: string) => void;
  defaultColor?: string;
  showPresets?: boolean;
  className?: string;
}

export function ColorPicker({
  value,
  onChange,
  defaultColor = "#6B7280",
  showPresets = true,
  className,
}: ColorPickerProps) {
  const currentColor = value || defaultColor;

  const handleColorInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange(e.target.value.toUpperCase());
  };

  const handleTextInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange(e.target.value.toUpperCase());
  };

  return (
    <div className={cn("space-y-3", className)}>
      <div className="flex items-center gap-2">
        <input
          type="color"
          value={currentColor.toLowerCase()}
          onChange={handleColorInputChange}
          className="h-10 w-10 cursor-pointer rounded-md border border-input p-0.5 bg-transparent"
          aria-label="Pick a color"
        />
        <Input
          value={value}
          onChange={handleTextInputChange}
          placeholder="#FF5733"
          className="flex-1 font-mono uppercase"
        />
      </div>
      {showPresets && (
        <div className="flex flex-wrap gap-2">
          {presetColors.map((color) => (
            <button
              key={color}
              type="button"
              className={cn(
                "h-8 w-8 rounded-md border-2 transition-transform hover:scale-110",
                value === color ? "border-foreground" : "border-transparent"
              )}
              style={{ backgroundColor: color }}
              onClick={() => onChange(color)}
              aria-label={`Select color ${color}`}
            />
          ))}
        </div>
      )}
    </div>
  );
}
