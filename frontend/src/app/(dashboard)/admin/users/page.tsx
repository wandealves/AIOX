"use client";

import { useState } from "react";
import toast from "react-hot-toast";
import {
  useAdminUsers,
  useAdminUpdateRole,
  useAdminDisableUser,
  useAdminEnableUser,
} from "@/hooks/use-admin";
import { Badge } from "@/components/ui/badge";
import { Pagination } from "@/components/ui/pagination";
import { Skeleton } from "@/components/ui/skeleton";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";

interface PendingAction {
  type: "disable" | "enable";
  userId: string;
  email: string;
}

export default function AdminUsersPage() {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useAdminUsers(page, 20);
  const updateRole = useAdminUpdateRole();
  const disableUser = useAdminDisableUser();
  const enableUser = useAdminEnableUser();
  const [pendingAction, setPendingAction] = useState<PendingAction | null>(null);

  const users = data?.data || [];
  const totalCount = data?.total_count || 0;
  const totalPages = Math.ceil(totalCount / 20);

  const handleRoleChange = async (userId: string, currentRole: string) => {
    const newRole = currentRole === "admin" ? "user" : "admin";
    try {
      await updateRole.mutateAsync({ userId, role: newRole });
      toast.success(`User role changed to ${newRole}`);
    } catch {
      toast.error("Failed to update role");
    }
  };

  const handleConfirmAction = async () => {
    if (!pendingAction) return;
    try {
      if (pendingAction.type === "disable") {
        await disableUser.mutateAsync(pendingAction.userId);
        toast.success("User disabled");
      } else {
        await enableUser.mutateAsync(pendingAction.userId);
        toast.success("User enabled");
      }
    } catch {
      toast.error(`Failed to ${pendingAction.type} user`);
    }
    setPendingAction(null);
  };

  return (
    <div>
      <h2 className="mb-6 text-lg font-semibold text-[var(--foreground)]">
        User Management ({totalCount})
      </h2>

      <div
        className="overflow-x-auto rounded-xl border border-[var(--card-border)] bg-[var(--card)]"
        style={{ boxShadow: "var(--shadow-sm)" }}
      >
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[var(--card-border)] bg-[var(--background-tertiary)]">
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Email
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Role
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Status
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Created
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold uppercase tracking-wider text-[var(--foreground-muted)]">
                Actions
              </th>
            </tr>
          </thead>
          <tbody>
            {isLoading &&
              Array.from({ length: 5 }).map((_, i) => (
                <tr key={i} className="border-b border-[var(--card-border)]">
                  {Array.from({ length: 5 }).map((_, j) => (
                    <td key={j} className="px-4 py-3">
                      <Skeleton className="h-5 w-full" variant="text" />
                    </td>
                  ))}
                </tr>
              ))}
            {!isLoading &&
              users.map((user, i) => (
                <tr
                  key={user.id}
                  className={`border-b border-[var(--card-border)] last:border-0 transition-colors hover:bg-[var(--card-hover)] ${
                    i % 2 === 1 ? "bg-[var(--background-secondary)]" : ""
                  }`}
                >
                  <td className="px-4 py-3 font-medium text-[var(--foreground)]">
                    {user.email}
                  </td>
                  <td className="px-4 py-3">
                    <Badge variant={user.role === "admin" ? "purple" : "default"}>
                      {user.role || "user"}
                    </Badge>
                  </td>
                  <td className="px-4 py-3">
                    <Badge variant={user.disabled_at ? "error" : "success"}>
                      {user.disabled_at ? "Disabled" : "Active"}
                    </Badge>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-[var(--foreground-subtle)]">
                    {new Date(user.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <button
                        onClick={() =>
                          handleRoleChange(user.id, user.role || "user")
                        }
                        className="rounded-lg px-2.5 py-1 text-xs font-medium text-primary-600 transition-colors hover:bg-primary-500/10"
                      >
                        {user.role === "admin" ? "Demote" : "Promote"}
                      </button>
                      {user.disabled_at ? (
                        <button
                          onClick={() =>
                            setPendingAction({
                              type: "enable",
                              userId: user.id,
                              email: user.email,
                            })
                          }
                          className="rounded-lg px-2.5 py-1 text-xs font-medium text-green-600 transition-colors hover:bg-green-500/10"
                        >
                          Enable
                        </button>
                      ) : (
                        <button
                          onClick={() =>
                            setPendingAction({
                              type: "disable",
                              userId: user.id,
                              email: user.email,
                            })
                          }
                          className="rounded-lg px-2.5 py-1 text-xs font-medium text-red-600 transition-colors hover:bg-red-500/10"
                        >
                          Disable
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
          </tbody>
        </table>
      </div>

      <Pagination page={page} totalPages={totalPages} onPageChange={setPage} />

      <ConfirmDialog
        open={!!pendingAction}
        title={
          pendingAction?.type === "disable" ? "Disable User" : "Enable User"
        }
        message={
          pendingAction?.type === "disable"
            ? `Are you sure you want to disable ${pendingAction.email}? They will not be able to log in.`
            : `Are you sure you want to enable ${pendingAction?.email}?`
        }
        variant={pendingAction?.type === "disable" ? "danger" : "warning"}
        confirmLabel={pendingAction?.type === "disable" ? "Disable" : "Enable"}
        onConfirm={handleConfirmAction}
        onCancel={() => setPendingAction(null)}
      />
    </div>
  );
}
