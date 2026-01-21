"use client";

import React from "react";
import { usePathname } from "next/navigation";
import { SidebarTrigger } from "@/components/ui/sidebar";
import { Separator } from "@/components/ui/separator";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { GlobalSearch } from "@/components/search/global-search";

const routeTitles: Record<string, string> = {
  "/dashboard": "Dashboard",
  "/invoices": "Invoices",
  "/invoices/new": "New Invoice",
  "/companies": "Companies",
  "/companies/new": "New Company",
  "/categories": "Categories",
  "/categories/new": "New Category",
};

export function SiteHeader() {
  const pathname = usePathname();

  const getBreadcrumbs = () => {
    const segments = pathname.split("/").filter(Boolean);
    const breadcrumbs: { label: string; href: string; isLast: boolean }[] = [];

    // Skip if we're on dashboard
    if (pathname === "/dashboard") {
      return [];
    }

    // Always start with Dashboard for non-dashboard pages
    breadcrumbs.push({ label: "Dashboard", href: "/dashboard", isLast: false });

    let currentPath = "";
    segments.forEach((segment, index) => {
      currentPath += `/${segment}`;
      const isLast = index === segments.length - 1;

      // Check if this is an ID segment (numeric or could be "new")
      const isIdSegment = /^\d+$/.test(segment);

      let label = routeTitles[currentPath];
      if (!label) {
        if (isIdSegment) {
          label = "Details";
        } else {
          label = segment.charAt(0).toUpperCase() + segment.slice(1);
        }
      }

      breadcrumbs.push({ label, href: currentPath, isLast });
    });

    return breadcrumbs;
  };

  const breadcrumbs = getBreadcrumbs();
  const pageTitle = breadcrumbs.length > 0
    ? breadcrumbs[breadcrumbs.length - 1].label
    : "Dashboard";

  return (
    <header className="flex h-14 shrink-0 items-center gap-2 border-b px-4">
      <div className="flex items-center gap-2">
        <SidebarTrigger className="-ml-1" />
        <Breadcrumb>
          <BreadcrumbList>
            {breadcrumbs.length === 0 ? (
              <BreadcrumbItem>
                <BreadcrumbPage>Dashboard</BreadcrumbPage>
              </BreadcrumbItem>
            ) : (
              breadcrumbs.map((crumb, index) => (
                <React.Fragment key={crumb.href}>
                  <BreadcrumbItem>
                    {crumb.isLast ? (
                      <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
                    ) : (
                      <BreadcrumbLink href={crumb.href}>{crumb.label}</BreadcrumbLink>
                    )}
                  </BreadcrumbItem>
                  {!crumb.isLast && <BreadcrumbSeparator />}
                </React.Fragment>
              ))
            )}
          </BreadcrumbList>
        </Breadcrumb>
        <Separator orientation="vertical" className="mx-2 h-6 my-auto" />
        <GlobalSearch />
      </div>
    </header>
  );
}
