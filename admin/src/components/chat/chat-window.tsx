"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { Bot, Wifi, WifiOff, ChevronDown, MessageCircle } from "lucide-react";
import { useChat } from "@/hooks/use-chat";
import { MessageBubble } from "./message-bubble";
import { MessageInput } from "./message-input";
import { TypingIndicator } from "./typing-indicator";

interface ChatWindowProps {
  agentId: string;
  agentName: string;
}

export function ChatWindow({ agentId, agentName }: ChatWindowProps) {
  const { messages, sendMessage, isConnected } = useChat(agentId);
  const bottomRef = useRef<HTMLDivElement>(null);
  const scrollAreaRef = useRef<HTMLDivElement>(null);
  const [showScrollBtn, setShowScrollBtn] = useState(false);

  const isWaitingForResponse =
    messages.length > 0 &&
    messages[messages.length - 1].role === "user" &&
    messages[messages.length - 1].status === "sending";

  const scrollToBottom = useCallback(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages, scrollToBottom]);

  const handleScroll = () => {
    if (!scrollAreaRef.current) return;
    const { scrollTop, scrollHeight, clientHeight } = scrollAreaRef.current;
    setShowScrollBtn(scrollHeight - scrollTop - clientHeight > 100);
  };

  return (
    <div
      className="flex h-[calc(100vh-8rem)] flex-col rounded-xl border border-[var(--card-border)] bg-[var(--card)] overflow-hidden"
      style={{ boxShadow: "var(--shadow-sm)" }}
    >
      {/* Header */}
      <div className="flex items-center justify-between border-b border-[var(--card-border)] px-4 py-3">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-primary-400 to-primary-600">
            <Bot className="h-4.5 w-4.5 text-white" />
          </div>
          <div>
            <h3 className="font-semibold text-[var(--foreground)]">
              {agentName}
            </h3>
            <div className="flex items-center gap-1.5 text-xs text-[var(--foreground-muted)]">
              {isConnected ? (
                <>
                  <span className="relative flex h-2 w-2">
                    <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75" />
                    <span className="relative inline-flex h-2 w-2 rounded-full bg-green-500" />
                  </span>
                  <span>Online</span>
                </>
              ) : (
                <>
                  <WifiOff className="h-3 w-3 text-red-400" />
                  <span>Disconnected</span>
                </>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Messages */}
      <div
        ref={scrollAreaRef}
        onScroll={handleScroll}
        className="relative flex-1 space-y-4 overflow-y-auto p-4 scrollbar-thin"
      >
        {messages.length === 0 && (
          <div className="flex h-full flex-col items-center justify-center text-center">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary-500/10">
              <MessageCircle className="h-7 w-7 text-primary-500" />
            </div>
            <h4 className="mt-3 font-medium text-[var(--foreground)]">
              Start a conversation
            </h4>
            <p className="mt-1 text-sm text-[var(--foreground-muted)]">
              Send a message to begin chatting with {agentName}
            </p>
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
        {isWaitingForResponse && <TypingIndicator />}
        <div ref={bottomRef} />

        {/* Scroll-to-bottom FAB */}
        {showScrollBtn && (
          <button
            onClick={scrollToBottom}
            className="sticky bottom-2 left-1/2 z-10 flex h-8 w-8 -translate-x-1/2 items-center justify-center rounded-full border border-[var(--card-border)] bg-[var(--card)] text-[var(--foreground-muted)] transition-colors hover:bg-[var(--background-tertiary)]"
            style={{ boxShadow: "var(--shadow-md)" }}
          >
            <ChevronDown className="h-4 w-4" />
          </button>
        )}
      </div>

      {/* Input */}
      <MessageInput onSend={sendMessage} />
    </div>
  );
}
