"use client";

import { useState, useEffect, useCallback, useRef } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { WSClient } from "@/lib/ws";
import type {
  ApiResponse,
  PaginatedResponse,
  Conversation,
  SendMessageResponse,
  WSServerMessage,
} from "@/lib/types";

interface ChatMessage {
  id: string;
  role: "user" | "assistant";
  content: string;
  timestamp: string;
  status?: "sending" | "sent" | "error";
  request_id?: string;
}

export function useConversations(agentId: string, page = 1, pageSize = 50) {
  return useQuery({
    queryKey: ["conversations", agentId, page, pageSize],
    queryFn: () =>
      api.get<PaginatedResponse<Conversation>>(
        `/api/v1/agents/${agentId}/conversations?page=${page}&page_size=${pageSize}`
      ),
    enabled: !!agentId,
  });
}

export function useChat(agentId: string) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const wsRef = useRef<WSClient | null>(null);

  // Connect WebSocket
  useEffect(() => {
    const token = localStorage.getItem("access_token");
    if (!token || !agentId) return;

    const ws = new WSClient(token);
    wsRef.current = ws;

    const unsub = ws.subscribe((msg: WSServerMessage) => {
      if (msg.type === "message" && msg.agent_id === agentId) {
        setMessages((prev) => {
          // Find pending message by request_id and update, or add new
          const idx = prev.findIndex(
            (m) => m.request_id === msg.request_id && m.role === "user"
          );
          const assistantMsg: ChatMessage = {
            id: msg.request_id || Date.now().toString(),
            role: "assistant",
            content: msg.body || "",
            timestamp: new Date().toISOString(),
            request_id: msg.request_id,
          };

          if (idx >= 0) {
            // Update user message status and add assistant reply
            const updated = [...prev];
            updated[idx] = { ...updated[idx], status: "sent" };
            return [...updated, assistantMsg];
          }
          return [...prev, assistantMsg];
        });
      }
    });

    ws.connect();
    setIsConnected(true);

    return () => {
      unsub();
      ws.disconnect();
      setIsConnected(false);
    };
  }, [agentId]);

  // Load history
  const { data: history } = useConversations(agentId, 1, 100);

  useEffect(() => {
    if (history?.data && messages.length === 0) {
      const historyMessages: ChatMessage[] = [];
      // Conversations are ordered DESC, reverse for chronological
      const sorted = [...history.data].reverse();
      for (const conv of sorted) {
        if (conv.input) {
          historyMessages.push({
            id: conv.id + "-user",
            role: "user",
            content: conv.input,
            timestamp: conv.created_at,
            status: "sent",
          });
        }
        if (conv.output) {
          historyMessages.push({
            id: conv.id + "-agent",
            role: "assistant",
            content: conv.output,
            timestamp: conv.created_at,
          });
        }
      }
      if (historyMessages.length > 0) {
        setMessages(historyMessages);
      }
    }
  }, [history, messages.length]);

  const sendMessage = useCallback(
    async (body: string) => {
      if (!body.trim()) return;

      const tempId = Date.now().toString();
      const userMsg: ChatMessage = {
        id: tempId,
        role: "user",
        content: body,
        timestamp: new Date().toISOString(),
        status: "sending",
      };

      setMessages((prev) => [...prev, userMsg]);

      try {
        const res = await api.post<ApiResponse<SendMessageResponse>>(
          `/api/v1/agents/${agentId}/messages`,
          { body }
        );

        if (res.data) {
          setMessages((prev) =>
            prev.map((m) =>
              m.id === tempId
                ? { ...m, status: "sent", request_id: res.data!.request_id }
                : m
            )
          );
        }
      } catch {
        setMessages((prev) =>
          prev.map((m) =>
            m.id === tempId ? { ...m, status: "error" } : m
          )
        );
      }
    },
    [agentId]
  );

  return { messages, sendMessage, isConnected };
}
