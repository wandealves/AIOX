"use client";

import { useState } from "react";
import { Bot, User, AlertCircle, Loader2, Copy, Check } from "lucide-react";
import { motion } from "framer-motion";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { oneDark, oneLight } from "react-syntax-highlighter/dist/esm/styles/prism";
import { useTheme } from "@/providers/theme-provider";
import { formatDistanceToNow } from "date-fns";

interface MessageBubbleProps {
  role: "user" | "assistant";
  content: string;
  timestamp: string;
  status?: "sending" | "sent" | "error";
}

function CodeBlock({ language, value }: { language: string; value: string }) {
  const { resolvedTheme } = useTheme();
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="group relative my-2 overflow-hidden rounded-lg border border-[var(--card-border)]">
      <div className="flex items-center justify-between bg-[var(--background-tertiary)] px-3 py-1.5">
        <span className="text-[10px] font-medium uppercase tracking-wider text-[var(--foreground-subtle)]">
          {language || "code"}
        </span>
        <button
          onClick={handleCopy}
          className="flex items-center gap-1 rounded px-1.5 py-0.5 text-[10px] text-[var(--foreground-subtle)] transition-colors hover:bg-[var(--card-hover)] hover:text-[var(--foreground)]"
        >
          {copied ? (
            <>
              <Check className="h-3 w-3" /> Copied
            </>
          ) : (
            <>
              <Copy className="h-3 w-3" /> Copy
            </>
          )}
        </button>
      </div>
      <SyntaxHighlighter
        language={language || "text"}
        style={resolvedTheme === "dark" ? oneDark : oneLight}
        customStyle={{
          margin: 0,
          borderRadius: 0,
          fontSize: "0.8125rem",
          background: resolvedTheme === "dark" ? "#1a1a2e" : "#fafafa",
        }}
      >
        {value}
      </SyntaxHighlighter>
    </div>
  );
}

export function MessageBubble({
  role,
  content,
  timestamp,
  status,
}: MessageBubbleProps) {
  const isUser = role === "user";

  const timeAgo = (() => {
    try {
      return formatDistanceToNow(new Date(timestamp), { addSuffix: true });
    } catch {
      return "";
    }
  })();

  const absoluteTime = (() => {
    try {
      return new Date(timestamp).toLocaleString();
    } catch {
      return "";
    }
  })();

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.2 }}
      className={`flex gap-3 ${isUser ? "flex-row-reverse" : ""}`}
    >
      <div
        className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full ${
          isUser
            ? "bg-primary-600"
            : "bg-gradient-to-br from-primary-400 to-primary-600"
        }`}
      >
        {isUser ? (
          <User className="h-4 w-4 text-white" />
        ) : (
          <Bot className="h-4 w-4 text-white" />
        )}
      </div>
      <div
        className={`max-w-[75%] px-4 py-2.5 ${
          isUser
            ? "rounded-2xl rounded-br-md bg-primary-600 text-white"
            : "rounded-2xl rounded-bl-md bg-[var(--background-tertiary)] text-[var(--foreground)]"
        }`}
      >
        {isUser ? (
          <p className="whitespace-pre-wrap text-sm">{content}</p>
        ) : (
          <div className="prose-chat text-sm">
            <ReactMarkdown
              remarkPlugins={[remarkGfm]}
              components={{
                code(props) {
                  const { children, className, ...rest } = props;
                  const match = /language-(\w+)/.exec(className || "");
                  const codeString = String(children).replace(/\n$/, "");

                  if (match) {
                    return (
                      <CodeBlock
                        language={match[1]}
                        value={codeString}
                      />
                    );
                  }

                  return (
                    <code className={className} {...rest}>
                      {children}
                    </code>
                  );
                },
              }}
            >
              {content}
            </ReactMarkdown>
          </div>
        )}
        <div
          className={`mt-1 flex items-center gap-1.5 text-[10px] ${
            isUser ? "text-primary-200" : "text-[var(--foreground-subtle)]"
          }`}
        >
          <span title={absoluteTime}>{timeAgo}</span>
          {status === "sending" && (
            <Loader2 className="h-3 w-3 animate-spin" />
          )}
          {status === "error" && (
            <AlertCircle className="h-3 w-3 text-red-400" />
          )}
        </div>
      </div>
    </motion.div>
  );
}
