"use client";

import { useState, useRef, useEffect } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Menu, Sun, Moon, Monitor, LogOut, ChevronRight } from "lucide-react";
import { useTheme } from "@/providers/theme-provider";
import { useAuthContext } from "@/providers/auth-provider";
import { useSidebar } from "./sidebar";

const titles: Record<string, string> = {
  "/dashboard": "Dashboard",
  "/agents": "Agents",
  "/agents/new": "Create Agent",
  "/governance/quota": "Quota Usage",
  "/governance/audit": "Audit Logs",
  "/admin/users": "User Management",
  "/admin/stats": "Platform Statistics",
};

function getBreadcrumbs(pathname: string) {
  const segments = pathname.split("/").filter(Boolean);
  const crumbs: { label: string; href: string }[] = [];

  let path = "";
  for (const seg of segments) {
    path += "/" + seg;
    // Skip UUID-like segments in display
    const isId = /^[0-9a-f-]{20,}$/.test(seg);
    const label = isId
      ? "Details"
      : seg.charAt(0).toUpperCase() + seg.slice(1).replace(/-/g, " ");
    crumbs.push({ label, href: path });
  }

  return crumbs;
}

export function Header() {
  const pathname = usePathname();
  const { theme, setTheme } = useTheme();
  const { userEmail, logout } = useAuthContext();
  const { openMobile } = useSidebar();
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  // Find current page title
  let title = "AIOX";
  for (const [path, t] of Object.entries(titles)) {
    if (pathname.startsWith(path)) {
      title = t;
    }
  }
  if (pathname.match(/^\/agents\/[^/]+\/chat$/)) {
    title = "Chat";
  } else if (pathname.match(/^\/agents\/[^/]+$/) && !pathname.endsWith("/new")) {
    title = "Agent Details";
  }

  const breadcrumbs = getBreadcrumbs(pathname);

  const cycleTheme = () => {
    const order: Array<"light" | "dark" | "system"> = ["light", "dark", "system"];
    const idx = order.indexOf(theme);
    setTheme(order[(idx + 1) % order.length]);
  };

  const ThemeIcon = theme === "dark" ? Moon : theme === "light" ? Sun : Monitor;

  // Close menu on click outside
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setUserMenuOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  return (
    <header className="flex h-14 items-center justify-between border-b border-[var(--card-border)] bg-[var(--card)] px-4 md:px-6"
      style={{ boxShadow: "var(--shadow-sm)" }}
    >
      <div className="flex items-center gap-3">
        {/* Mobile hamburger */}
        <button
          onClick={openMobile}
          className="rounded-lg p-2 text-[var(--foreground-muted)] hover:bg-[var(--background-tertiary)] md:hidden"
        >
          <Menu className="h-5 w-5" />
        </button>

        {/* Breadcrumbs */}
        <nav className="hidden items-center gap-1 text-sm md:flex">
          {breadcrumbs.map((crumb, i) => (
            <span key={crumb.href} className="flex items-center gap-1">
              {i > 0 && (
                <ChevronRight className="h-3.5 w-3.5 text-[var(--foreground-subtle)]" />
              )}
              {i === breadcrumbs.length - 1 ? (
                <span className="font-medium text-[var(--foreground)]">
                  {crumb.label}
                </span>
              ) : (
                <Link
                  href={crumb.href}
                  className="text-[var(--foreground-muted)] hover:text-[var(--foreground)] transition-colors"
                >
                  {crumb.label}
                </Link>
              )}
            </span>
          ))}
        </nav>

        {/* Mobile title */}
        <h1 className="text-base font-semibold text-[var(--foreground)] md:hidden">
          {title}
        </h1>
      </div>

      <div className="flex items-center gap-2">
        {/* Theme toggle */}
        <button
          onClick={cycleTheme}
          title={`Theme: ${theme}`}
          className="rounded-lg p-2 text-[var(--foreground-muted)] transition-colors hover:bg-[var(--background-tertiary)] hover:text-[var(--foreground)]"
        >
          <ThemeIcon className="h-[18px] w-[18px]" />
        </button>

        {/* User dropdown */}
        <div className="relative" ref={menuRef}>
          <button
            onClick={() => setUserMenuOpen(!userMenuOpen)}
            className="flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm text-[var(--foreground-muted)] transition-colors hover:bg-[var(--background-tertiary)]"
          >
            <div className="flex h-7 w-7 items-center justify-center rounded-full bg-primary-600 text-xs font-semibold text-white">
              {userEmail?.charAt(0).toUpperCase() || "?"}
            </div>
            <span className="hidden md:inline max-w-[150px] truncate">{userEmail}</span>
          </button>

          {userMenuOpen && (
            <div className="absolute right-0 top-full z-50 mt-1 w-56 rounded-xl border border-[var(--card-border)] bg-[var(--card)] py-1"
              style={{ boxShadow: "var(--shadow-lg)" }}
            >
              <div className="border-b border-[var(--card-border)] px-4 py-2">
                <p className="truncate text-sm font-medium text-[var(--foreground)]">
                  {userEmail}
                </p>
              </div>
              <button
                onClick={() => {
                  setUserMenuOpen(false);
                  logout();
                }}
                className="flex w-full items-center gap-2 px-4 py-2.5 text-sm text-[var(--foreground-muted)] transition-colors hover:bg-[var(--background-tertiary)] hover:text-[var(--foreground)]"
              >
                <LogOut className="h-4 w-4" />
                Logout
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}
