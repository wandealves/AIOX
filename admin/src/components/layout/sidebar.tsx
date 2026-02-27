"use client";

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
} from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Bot,
  Shield,
  FileText,
  Users,
  BarChart3,
  Wrench,
  Puzzle,
  LogOut,
  ChevronLeft,
  ChevronRight,
  X,
} from "lucide-react";
import { useAuthContext } from "@/providers/auth-provider";

const navItems = [
  { href: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { href: "/agents", label: "Agents", icon: Bot },
  { href: "/governance/quota", label: "Quotas", icon: Shield },
  { href: "/governance/audit", label: "Audit Logs", icon: FileText },
];

const adminItems = [
  { href: "/admin/users", label: "Users", icon: Users },
  { href: "/admin/stats", label: "Statistics", icon: BarChart3 },
  { href: "/admin/system-tools", label: "System Tools", icon: Wrench },
  { href: "/admin/plugins", label: "Plugins", icon: Puzzle },
];

interface SidebarContextValue {
  isCollapsed: boolean;
  isMobileOpen: boolean;
  toggleCollapse: () => void;
  openMobile: () => void;
  closeMobile: () => void;
}

const SidebarContext = createContext<SidebarContextValue>({
  isCollapsed: false,
  isMobileOpen: false,
  toggleCollapse: () => {},
  openMobile: () => {},
  closeMobile: () => {},
});

export function useSidebar() {
  return useContext(SidebarContext);
}

export function SidebarProvider({ children }: { children: React.ReactNode }) {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [isMobileOpen, setIsMobileOpen] = useState(false);

  useEffect(() => {
    const stored = localStorage.getItem("aiox-sidebar-collapsed");
    if (stored === "true") setIsCollapsed(true);
  }, []);

  const toggleCollapse = useCallback(() => {
    setIsCollapsed((prev) => {
      const next = !prev;
      localStorage.setItem("aiox-sidebar-collapsed", String(next));
      return next;
    });
  }, []);

  const openMobile = useCallback(() => setIsMobileOpen(true), []);
  const closeMobile = useCallback(() => setIsMobileOpen(false), []);

  return (
    <SidebarContext.Provider
      value={{ isCollapsed, isMobileOpen, toggleCollapse, openMobile, closeMobile }}
    >
      {children}
    </SidebarContext.Provider>
  );
}

export function Sidebar() {
  const pathname = usePathname();
  const { userEmail, logout } = useAuthContext();
  const { isCollapsed, isMobileOpen, toggleCollapse, closeMobile } = useSidebar();

  const userInitial = userEmail ? userEmail.charAt(0).toUpperCase() : "?";

  const renderNavItem = (item: (typeof navItems)[0]) => {
    const isActive = pathname.startsWith(item.href);
    return (
      <Link
        key={item.href}
        href={item.href}
        onClick={closeMobile}
        title={isCollapsed ? item.label : undefined}
        className={`group relative flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all duration-200 ${
          isActive
            ? "bg-sidebar-active-bg text-sidebar-text-active"
            : "text-sidebar-text hover:bg-sidebar-hover-bg hover:text-[var(--foreground)]"
        }`}
      >
        {isActive && (
          <span className="absolute left-0 top-1/2 h-6 w-[3px] -translate-y-1/2 rounded-r-full bg-primary-500" />
        )}
        <item.icon className="h-[18px] w-[18px] shrink-0" />
        {!isCollapsed && <span>{item.label}</span>}
        {isCollapsed && (
          <span className="pointer-events-none absolute left-full ml-2 hidden rounded-md bg-[var(--foreground)] px-2 py-1 text-xs text-[var(--background)] shadow-lg group-hover:block">
            {item.label}
          </span>
        )}
      </Link>
    );
  };

  const sidebarContent = (
    <>
      {/* Logo */}
      <div className="flex h-16 items-center justify-between border-b border-sidebar-border px-4">
        <Link href="/dashboard" className="flex items-center gap-2.5" onClick={closeMobile}>
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary-600">
            <Bot className="h-4.5 w-4.5 text-white" />
          </div>
          {!isCollapsed && (
            <span className="text-lg font-bold text-[var(--foreground)]">AIOX</span>
          )}
        </Link>
        {/* Desktop collapse toggle */}
        <button
          onClick={toggleCollapse}
          className="hidden rounded-lg p-1.5 text-sidebar-text hover:bg-sidebar-hover-bg md:flex"
        >
          {isCollapsed ? (
            <ChevronRight className="h-4 w-4" />
          ) : (
            <ChevronLeft className="h-4 w-4" />
          )}
        </button>
        {/* Mobile close */}
        <button
          onClick={closeMobile}
          className="rounded-lg p-1.5 text-sidebar-text hover:bg-sidebar-hover-bg md:hidden"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      {/* Navigation */}
      <nav className="flex-1 space-y-1 overflow-y-auto p-3">
        {!isCollapsed && (
          <div className="mb-2 px-3 text-[10px] font-semibold uppercase tracking-wider text-[var(--foreground-subtle)]">
            Main
          </div>
        )}
        {navItems.map(renderNavItem)}

        <div className="my-4 border-t border-sidebar-border" />

        {!isCollapsed && (
          <div className="mb-2 px-3 text-[10px] font-semibold uppercase tracking-wider text-[var(--foreground-subtle)]">
            Admin
          </div>
        )}
        {adminItems.map(renderNavItem)}
      </nav>

      {/* User section */}
      <div className="border-t border-sidebar-border p-3">
        <div className="flex items-center gap-3 rounded-lg px-3 py-2">
          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary-600 text-xs font-semibold text-white">
            {userInitial}
          </div>
          {!isCollapsed && (
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium text-[var(--foreground)]">
                {userEmail}
              </p>
            </div>
          )}
        </div>
        <button
          onClick={logout}
          title={isCollapsed ? "Logout" : undefined}
          className="mt-1 flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-sidebar-text transition-colors hover:bg-sidebar-hover-bg hover:text-[var(--foreground)]"
        >
          <LogOut className="h-[18px] w-[18px] shrink-0" />
          {!isCollapsed && <span>Logout</span>}
        </button>
      </div>
    </>
  );

  return (
    <>
      {/* Desktop sidebar */}
      <aside
        className={`hidden md:flex h-screen flex-col border-r border-sidebar-border bg-sidebar-bg transition-[width] duration-300 ease-in-out ${
          isCollapsed ? "w-16" : "w-64"
        }`}
      >
        {sidebarContent}
      </aside>

      {/* Mobile overlay */}
      {isMobileOpen && (
        <div className="fixed inset-0 z-40 md:hidden">
          <div
            className="absolute inset-0 bg-black/40 backdrop-blur-sm"
            onClick={closeMobile}
          />
          <aside className="absolute left-0 top-0 z-50 flex h-full w-64 flex-col bg-sidebar-bg shadow-xl animate-in slide-in-from-left duration-300">
            {sidebarContent}
          </aside>
        </div>
      )}
    </>
  );
}
