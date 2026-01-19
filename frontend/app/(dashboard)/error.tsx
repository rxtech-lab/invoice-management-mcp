"use client";

import { useEffect } from "react";
import { signOut } from "next-auth/react";
import { AlertCircle, LogOut, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function ErrorPage({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Dashboard error:", error);
  }, [error]);

  return (
    <div className="flex min-h-[50vh] flex-col items-center justify-center gap-6 p-4">
      <div className="flex flex-col items-center gap-2 text-center">
        <AlertCircle className="h-12 w-12 text-destructive" />
        <h1 className="text-2xl font-semibold">Something went wrong</h1>
        <p className="text-muted-foreground max-w-md">
          An error occurred while processing your request. You can try refreshing
          the page or sign out and back in.
        </p>
        {error.digest && (
          <p className="text-xs text-muted-foreground">Error ID: {error.digest}</p>
        )}
      </div>
      <div className="flex gap-3">
        <Button variant="outline" onClick={() => reset()}>
          <RefreshCw className="mr-2 h-4 w-4" />
          Try again
        </Button>
        <Button
          variant="destructive"
          onClick={() => signOut({ callbackUrl: "/login" })}
        >
          <LogOut className="mr-2 h-4 w-4" />
          Sign out
        </Button>
      </div>
    </div>
  );
}
