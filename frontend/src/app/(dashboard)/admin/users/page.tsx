"use client";

import { useState } from "react";
import {
  useAdminUsers,
  useAdminUpdateRole,
  useAdminDisableUser,
  useAdminEnableUser,
} from "@/hooks/use-admin";

export default function AdminUsersPage() {
  const [page, setPage] = useState(1);
  const { data, isLoading } = useAdminUsers(page, 20);
  const updateRole = useAdminUpdateRole();
  const disableUser = useAdminDisableUser();
  const enableUser = useAdminEnableUser();

  const users = data?.data || [];
  const totalCount = data?.total_count || 0;
  const totalPages = Math.ceil(totalCount / 20);

  return (
    <div>
      <h2 className="mb-6 text-lg font-semibold text-gray-900">
        User Management ({totalCount})
      </h2>

      <div className="overflow-x-auto rounded-xl border bg-white">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-gray-50">
              <th className="px-4 py-3 text-left font-medium text-gray-600">Email</th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">Role</th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">Status</th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">Created</th>
              <th className="px-4 py-3 text-left font-medium text-gray-600">Actions</th>
            </tr>
          </thead>
          <tbody>
            {isLoading && (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-gray-400">Loading...</td>
              </tr>
            )}
            {users.map((user) => (
              <tr key={user.id} className="border-b last:border-0">
                <td className="px-4 py-3 font-medium text-gray-900">{user.email}</td>
                <td className="px-4 py-3">
                  <span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
                    user.role === "admin" ? "bg-purple-50 text-purple-700" : "bg-gray-50 text-gray-700"
                  }`}>
                    {user.role || "user"}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${
                    user.disabled_at ? "bg-red-50 text-red-700" : "bg-green-50 text-green-700"
                  }`}>
                    {user.disabled_at ? "Disabled" : "Active"}
                  </span>
                </td>
                <td className="whitespace-nowrap px-4 py-3 text-gray-500">
                  {new Date(user.created_at).toLocaleDateString()}
                </td>
                <td className="px-4 py-3">
                  <div className="flex gap-2">
                    <button
                      onClick={() => updateRole.mutate({ userId: user.id, role: user.role === "admin" ? "user" : "admin" })}
                      className="rounded px-2 py-1 text-xs text-blue-600 hover:bg-blue-50"
                    >
                      {user.role === "admin" ? "Demote" : "Promote"}
                    </button>
                    {user.disabled_at ? (
                      <button
                        onClick={() => enableUser.mutate(user.id)}
                        className="rounded px-2 py-1 text-xs text-green-600 hover:bg-green-50"
                      >
                        Enable
                      </button>
                    ) : (
                      <button
                        onClick={() => disableUser.mutate(user.id)}
                        className="rounded px-2 py-1 text-xs text-red-600 hover:bg-red-50"
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

      {totalPages > 1 && (
        <div className="mt-4 flex justify-center gap-2">
          <button onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page <= 1} className="rounded-lg border px-3 py-1.5 text-sm hover:bg-gray-50 disabled:opacity-50">Previous</button>
          <span className="px-3 py-1.5 text-sm text-gray-500">Page {page} of {totalPages}</span>
          <button onClick={() => setPage((p) => Math.min(totalPages, p + 1))} disabled={page >= totalPages} className="rounded-lg border px-3 py-1.5 text-sm hover:bg-gray-50 disabled:opacity-50">Next</button>
        </div>
      )}
    </div>
  );
}
