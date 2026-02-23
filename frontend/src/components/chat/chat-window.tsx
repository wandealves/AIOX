"use client";

import { useEffect, useRef } from "react";
import { Wifi, WifiOff } from "lucide-react";
import { useChat } from "@/hooks/use-chat";
import { MessageBubble } from "./message-bubble";
import { MessageInput } from "./message-input";

interface ChatWindowProps {
  agentId: string;
  agentName: string;
}

export function ChatWindow({ agentId, agentName }: ChatWindowProps) {
  const { messages, sendMessage, isConnected } = useChat(agentId);
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  return (
    <div className="flex h-[calc(100vh-8rem)] flex-col rounded-xl border bg-white">
      {/* Header */}
      <div className="flex items-center justify-between border-b px-4 py-3">
        <div>
          <h3 className="font-medium text-gray-900">{agentName}</h3>
          <div className="flex items-center gap-1.5 text-xs text-gray-500">
            {isConnected ? (
              <>
                <Wifi className="h-3 w-3 text-green-500" />
                Connected
              </>
            ) : (
              <>
                <WifiOff className="h-3 w-3 text-red-500" />
                Disconnected
              </>
            )}
          </div>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 space-y-4 overflow-y-auto p-4">
        {messages.length === 0 && (
          <div className="flex h-full items-center justify-center text-sm text-gray-400">
            Send a message to start the conversation
          </div>
        )}
        {messages.map((msg) => (
          <MessageBubble
            key={msg.id}
            role={msg.role}
            content={msg.content}
            timestamp={msg.timestamp}
            status={msg.status}
          />
        ))}
        <div ref={bottomRef} />
      </div>

      {/* Input */}
      <MessageInput onSend={sendMessage} />
    </div>
  );
}
