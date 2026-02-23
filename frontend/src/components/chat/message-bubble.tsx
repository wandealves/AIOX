"use client";

import { Bot, User, AlertCircle, Loader2 } from "lucide-react";

interface MessageBubbleProps {
  role: "user" | "assistant";
  content: string;
  timestamp: string;
  status?: "sending" | "sent" | "error";
}

export function MessageBubble({
  role,
  content,
  timestamp,
  status,
}: MessageBubbleProps) {
  const isUser = role === "user";

  return (
    <div className={`flex gap-3 ${isUser ? "flex-row-reverse" : ""}`}>
      <div
        className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full ${
          isUser ? "bg-blue-600" : "bg-gray-200"
        }`}
      >
        {isUser ? (
          <User className="h-4 w-4 text-white" />
        ) : (
          <Bot className="h-4 w-4 text-gray-600" />
        )}
      </div>
      <div
        className={`max-w-[70%] rounded-2xl px-4 py-2 ${
          isUser
            ? "bg-blue-600 text-white"
            : "bg-gray-100 text-gray-900"
        }`}
      >
        <p className="whitespace-pre-wrap text-sm">{content}</p>
        <div
          className={`mt-1 flex items-center gap-1 text-xs ${
            isUser ? "text-blue-200" : "text-gray-400"
          }`}
        >
          <span>
            {new Date(timestamp).toLocaleTimeString([], {
              hour: "2-digit",
              minute: "2-digit",
            })}
          </span>
          {status === "sending" && <Loader2 className="h-3 w-3 animate-spin" />}
          {status === "error" && <AlertCircle className="h-3 w-3 text-red-300" />}
        </div>
      </div>
    </div>
  );
}
