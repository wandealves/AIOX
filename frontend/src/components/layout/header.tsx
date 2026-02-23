"use client";

import { usePathname } from "next/navigation";

const titles: Record<string, string> = {
  "/dashboard": "Dashboard",
  "/agents": "Agents",
  "/agents/new": "Create Agent",
  "/governance/quota": "Quota Usage",
  "/governance/audit": "Audit Logs",
  "/admin/users": "User Management",
  "/admin/stats": "Platform Statistics",
};

export function Header() {
  const pathname = usePathname();

  // Find the best matching title
  let title = "AIOX";
  for (const [path, t] of Object.entries(titles)) {
    if (pathname.startsWith(path)) {
      title = t;
    }
  }

  // Special case for agent detail / chat
  if (pathname.match(/^\/agents\/[^/]+\/chat$/)) {
    title = "Chat";
  } else if (pathname.match(/^\/agents\/[^/]+$/)) {
    title = "Agent Details";
  }

  return (
    <header className="flex h-16 items-center border-b bg-white px-6">
      <h1 className="text-xl font-semibold text-gray-900">{title}</h1>
    </header>
  );
}
