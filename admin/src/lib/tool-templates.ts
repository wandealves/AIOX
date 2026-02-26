import type { LucideIcon } from "lucide-react";
import {
  Search,
  Calendar,
  Mail,
  Hash,
  Code,
  Sparkles,
  Database,
  Wrench,
  MessageCircle,
  Globe,
  Send,
  FileText,
  BarChart3,
  Cloud,
  Image,
  Languages,
  AudioLines,
  Newspaper,
  TrendingUp,
  Calculator,
  Terminal,
  Braces,
  Webhook,
  Table,
} from "lucide-react";
import type { CreateToolRequest } from "./types";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type ToolCategory =
  | "search"
  | "productivity"
  | "communication"
  | "development"
  | "ai-media"
  | "data"
  | "utilities";

export type ConfigFieldType = "text" | "password" | "url" | "select" | "number" | "json";

export interface ConfigField {
  key: string;
  label: string;
  type: ConfigFieldType;
  placeholder?: string;
  required: boolean;
  helpText?: string;
  options?: { label: string; value: string }[];
  defaultValue?: string;
}

export interface ToolTemplate {
  id: string;
  name: string;
  description: string;
  category: ToolCategory;
  iconName: string;
  authType: "none" | "bearer" | "api_key";
  defaults: {
    endpoint_url: string;
    http_method: string;
    parameters: Record<string, unknown>;
    headers?: Record<string, string>;
    timeout_sec: number;
  };
  configFields: ConfigField[];
  buildRequest: (values: Record<string, string>) => CreateToolRequest;
}

export interface CategoryInfo {
  label: string;
  iconName: string;
  color: string;
}

// ---------------------------------------------------------------------------
// Icon Map (tree-shake friendly)
// ---------------------------------------------------------------------------

export const TOOL_ICONS: Record<string, LucideIcon> = {
  Search,
  Calendar,
  Mail,
  Hash,
  Code,
  Sparkles,
  Database,
  Wrench,
  MessageCircle,
  Globe,
  Send,
  FileText,
  BarChart3,
  Cloud,
  Image,
  Languages,
  AudioLines,
  Newspaper,
  TrendingUp,
  Calculator,
  Terminal,
  Braces,
  Webhook,
  Table,
};

// ---------------------------------------------------------------------------
// Categories
// ---------------------------------------------------------------------------

export const TOOL_CATEGORIES: Record<ToolCategory, CategoryInfo> = {
  search:        { label: "Search & Web",   iconName: "Search",        color: "blue"   },
  productivity:  { label: "Productivity",   iconName: "Calendar",      color: "green"  },
  communication: { label: "Communication",  iconName: "MessageCircle", color: "purple" },
  development:   { label: "Development",    iconName: "Code",          color: "amber"  },
  "ai-media":    { label: "AI & Media",     iconName: "Sparkles",      color: "pink"   },
  data:          { label: "Data & Info",    iconName: "Database",      color: "cyan"   },
  utilities:     { label: "Utilities",      iconName: "Wrench",        color: "gray"   },
};

// ---------------------------------------------------------------------------
// Category color helpers
// ---------------------------------------------------------------------------

export const CATEGORY_COLORS: Record<string, { bg: string; text: string; accent: string }> = {
  blue:   { bg: "bg-blue-500/10",   text: "text-blue-600",   accent: "bg-blue-500"   },
  green:  { bg: "bg-green-500/10",  text: "text-green-600",  accent: "bg-green-500"  },
  purple: { bg: "bg-purple-500/10", text: "text-purple-600", accent: "bg-purple-500" },
  amber:  { bg: "bg-amber-500/10",  text: "text-amber-600",  accent: "bg-amber-500"  },
  pink:   { bg: "bg-pink-500/10",   text: "text-pink-600",   accent: "bg-pink-500"   },
  cyan:   { bg: "bg-cyan-500/10",   text: "text-cyan-600",   accent: "bg-cyan-500"   },
  gray:   { bg: "bg-gray-500/10",   text: "text-gray-600",   accent: "bg-gray-500"   },
};

// ---------------------------------------------------------------------------
// Templates
// ---------------------------------------------------------------------------

export const TOOL_TEMPLATES: ToolTemplate[] = [
  // ── Search & Web ──────────────────────────────────────────────────────
  {
    id: "google-search",
    name: "Google Search",
    description: "Search the web using Google Custom Search API",
    category: "search",
    iconName: "Search",
    authType: "api_key",
    defaults: {
      endpoint_url: "https://www.googleapis.com/customsearch/v1",
      http_method: "GET",
      parameters: { type: "object", properties: { q: { type: "string", description: "Search query" }, cx: { type: "string", description: "Custom search engine ID" } }, required: ["q", "cx"] },
      timeout_sec: 15,
    },
    configFields: [
      { key: "api_key", label: "Google API Key", type: "password", required: true, placeholder: "AIza...", helpText: "From Google Cloud Console" },
      { key: "cx", label: "Search Engine ID", type: "text", required: true, placeholder: "017576662512...", helpText: "Custom Search Engine CX" },
    ],
    buildRequest: (v) => ({
      name: "google-search",
      description: "Search the web using Google Custom Search API",
      endpoint_url: "https://www.googleapis.com/customsearch/v1",
      http_method: "GET",
      parameters: { type: "object", properties: { q: { type: "string", description: "Search query" }, cx: { type: "string", description: "Custom search engine ID", default: v.cx } }, required: ["q"] },
      auth_type: "api_key",
      auth_config: { key: "key", value: v.api_key, in: "query" },
      timeout_sec: 15,
    }),
  },
  {
    id: "brave-search",
    name: "Brave Search",
    description: "Privacy-focused web search via Brave Search API",
    category: "search",
    iconName: "Search",
    authType: "api_key",
    defaults: {
      endpoint_url: "https://api.search.brave.com/res/v1/web/search",
      http_method: "GET",
      parameters: { type: "object", properties: { q: { type: "string", description: "Search query" } }, required: ["q"] },
      timeout_sec: 15,
    },
    configFields: [
      { key: "api_key", label: "API Key", type: "password", required: true, placeholder: "BSA...", helpText: "From brave.com/search/api" },
    ],
    buildRequest: (v) => ({
      name: "brave-search",
      description: "Privacy-focused web search via Brave Search API",
      endpoint_url: "https://api.search.brave.com/res/v1/web/search",
      http_method: "GET",
      parameters: { type: "object", properties: { q: { type: "string", description: "Search query" } }, required: ["q"] },
      auth_type: "api_key",
      auth_config: { key: "X-Subscription-Token", value: v.api_key, in: "header" },
      timeout_sec: 15,
    }),
  },
  {
    id: "tavily-search",
    name: "Tavily Search",
    description: "AI-optimized search API for LLM applications",
    category: "search",
    iconName: "Search",
    authType: "api_key",
    defaults: {
      endpoint_url: "https://api.tavily.com/search",
      http_method: "POST",
      parameters: { type: "object", properties: { query: { type: "string", description: "Search query" } }, required: ["query"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 20,
    },
    configFields: [
      { key: "api_key", label: "API Key", type: "password", required: true, placeholder: "tvly-...", helpText: "From tavily.com" },
    ],
    buildRequest: (v) => ({
      name: "tavily-search",
      description: "AI-optimized search API for LLM applications",
      endpoint_url: "https://api.tavily.com/search",
      http_method: "POST",
      parameters: { type: "object", properties: { query: { type: "string", description: "Search query" }, api_key: { type: "string", description: "API key", default: v.api_key } }, required: ["query"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 20,
    }),
  },
  {
    id: "wikipedia",
    name: "Wikipedia",
    description: "Search and retrieve Wikipedia articles",
    category: "search",
    iconName: "Globe",
    authType: "none",
    defaults: {
      endpoint_url: "https://en.wikipedia.org/w/api.php",
      http_method: "GET",
      parameters: { type: "object", properties: { action: { type: "string", default: "query" }, list: { type: "string", default: "search" }, srsearch: { type: "string", description: "Search term" }, format: { type: "string", default: "json" } }, required: ["srsearch"] },
      timeout_sec: 10,
    },
    configFields: [
      { key: "language", label: "Language", type: "select", required: false, defaultValue: "en", options: [{ label: "English", value: "en" }, { label: "Spanish", value: "es" }, { label: "Portuguese", value: "pt" }, { label: "French", value: "fr" }, { label: "German", value: "de" }] },
    ],
    buildRequest: (v) => ({
      name: "wikipedia",
      description: "Search and retrieve Wikipedia articles",
      endpoint_url: `https://${v.language || "en"}.wikipedia.org/w/api.php`,
      http_method: "GET",
      parameters: { type: "object", properties: { action: { type: "string", default: "query" }, list: { type: "string", default: "search" }, srsearch: { type: "string", description: "Search term" }, format: { type: "string", default: "json" } }, required: ["srsearch"] },
      timeout_sec: 10,
    }),
  },
  {
    id: "web-scraper",
    name: "Web Scraper",
    description: "Extract content from web pages via scraping API",
    category: "search",
    iconName: "Globe",
    authType: "api_key",
    defaults: {
      endpoint_url: "https://api.scrapingbee.com/api/v1",
      http_method: "GET",
      parameters: { type: "object", properties: { url: { type: "string", description: "URL to scrape" }, render_js: { type: "string", default: "false" } }, required: ["url"] },
      timeout_sec: 30,
    },
    configFields: [
      { key: "api_key", label: "API Key", type: "password", required: true, helpText: "ScrapingBee or similar scraping API key" },
      { key: "endpoint_url", label: "Endpoint URL", type: "url", required: false, defaultValue: "https://api.scrapingbee.com/api/v1", helpText: "Custom scraper endpoint (optional)" },
    ],
    buildRequest: (v) => ({
      name: "web-scraper",
      description: "Extract content from web pages via scraping API",
      endpoint_url: v.endpoint_url || "https://api.scrapingbee.com/api/v1",
      http_method: "GET",
      parameters: { type: "object", properties: { url: { type: "string", description: "URL to scrape" }, render_js: { type: "string", default: "false" } }, required: ["url"] },
      auth_type: "api_key",
      auth_config: { key: "api_key", value: v.api_key, in: "query" },
      timeout_sec: 30,
    }),
  },

  // ── Productivity ──────────────────────────────────────────────────────
  {
    id: "google-calendar",
    name: "Google Calendar",
    description: "Manage events and schedules via Google Calendar API",
    category: "productivity",
    iconName: "Calendar",
    authType: "bearer",
    defaults: {
      endpoint_url: "https://www.googleapis.com/calendar/v3/calendars/primary/events",
      http_method: "GET",
      parameters: { type: "object", properties: { timeMin: { type: "string", description: "Start time (ISO 8601)" }, maxResults: { type: "number", default: 10 } } },
      timeout_sec: 15,
    },
    configFields: [
      { key: "token", label: "OAuth Access Token", type: "password", required: true, helpText: "Google OAuth 2.0 Bearer token" },
    ],
    buildRequest: (v) => ({
      name: "google-calendar",
      description: "Manage events and schedules via Google Calendar API",
      endpoint_url: "https://www.googleapis.com/calendar/v3/calendars/primary/events",
      http_method: "GET",
      parameters: { type: "object", properties: { timeMin: { type: "string", description: "Start time (ISO 8601)" }, maxResults: { type: "number", default: 10 } } },
      auth_type: "bearer",
      auth_config: { token: v.token },
      timeout_sec: 15,
    }),
  },
  {
    id: "google-sheets",
    name: "Google Sheets",
    description: "Read and write data in Google Sheets",
    category: "productivity",
    iconName: "Table",
    authType: "bearer",
    defaults: {
      endpoint_url: "https://sheets.googleapis.com/v4/spreadsheets",
      http_method: "GET",
      parameters: { type: "object", properties: { spreadsheetId: { type: "string", description: "Spreadsheet ID" }, range: { type: "string", description: "Cell range (e.g. Sheet1!A1:D10)" } }, required: ["spreadsheetId"] },
      timeout_sec: 15,
    },
    configFields: [
      { key: "token", label: "OAuth Access Token", type: "password", required: true, helpText: "Google OAuth 2.0 Bearer token" },
    ],
    buildRequest: (v) => ({
      name: "google-sheets",
      description: "Read and write data in Google Sheets",
      endpoint_url: "https://sheets.googleapis.com/v4/spreadsheets",
      http_method: "GET",
      parameters: { type: "object", properties: { spreadsheetId: { type: "string", description: "Spreadsheet ID" }, range: { type: "string", description: "Cell range (e.g. Sheet1!A1:D10)" } }, required: ["spreadsheetId"] },
      auth_type: "bearer",
      auth_config: { token: v.token },
      timeout_sec: 15,
    }),
  },
  {
    id: "gmail",
    name: "Gmail",
    description: "Send and read emails via Gmail API",
    category: "productivity",
    iconName: "Mail",
    authType: "bearer",
    defaults: {
      endpoint_url: "https://gmail.googleapis.com/gmail/v1/users/me/messages",
      http_method: "GET",
      parameters: { type: "object", properties: { q: { type: "string", description: "Search query" }, maxResults: { type: "number", default: 10 } } },
      timeout_sec: 15,
    },
    configFields: [
      { key: "token", label: "OAuth Access Token", type: "password", required: true, helpText: "Google OAuth 2.0 Bearer token with Gmail scope" },
    ],
    buildRequest: (v) => ({
      name: "gmail",
      description: "Send and read emails via Gmail API",
      endpoint_url: "https://gmail.googleapis.com/gmail/v1/users/me/messages",
      http_method: "GET",
      parameters: { type: "object", properties: { q: { type: "string", description: "Search query" }, maxResults: { type: "number", default: 10 } } },
      auth_type: "bearer",
      auth_config: { token: v.token },
      timeout_sec: 15,
    }),
  },
  {
    id: "notion",
    name: "Notion",
    description: "Search and manage Notion pages and databases",
    category: "productivity",
    iconName: "FileText",
    authType: "bearer",
    defaults: {
      endpoint_url: "https://api.notion.com/v1/search",
      http_method: "POST",
      parameters: { type: "object", properties: { query: { type: "string", description: "Search query" } } },
      headers: { "Content-Type": "application/json", "Notion-Version": "2022-06-28" },
      timeout_sec: 15,
    },
    configFields: [
      { key: "token", label: "Integration Token", type: "password", required: true, placeholder: "ntn_...", helpText: "From notion.so/my-integrations" },
    ],
    buildRequest: (v) => ({
      name: "notion",
      description: "Search and manage Notion pages and databases",
      endpoint_url: "https://api.notion.com/v1/search",
      http_method: "POST",
      parameters: { type: "object", properties: { query: { type: "string", description: "Search query" } } },
      headers: { "Content-Type": "application/json", "Notion-Version": "2022-06-28" },
      auth_type: "bearer",
      auth_config: { token: v.token },
      timeout_sec: 15,
    }),
  },
  {
    id: "trello",
    name: "Trello",
    description: "Manage Trello boards, lists, and cards",
    category: "productivity",
    iconName: "BarChart3",
    authType: "api_key",
    defaults: {
      endpoint_url: "https://api.trello.com/1/members/me/boards",
      http_method: "GET",
      parameters: { type: "object", properties: { fields: { type: "string", default: "name,url" } } },
      timeout_sec: 15,
    },
    configFields: [
      { key: "api_key", label: "API Key", type: "password", required: true, helpText: "From trello.com/power-ups/admin" },
      { key: "token", label: "Token", type: "password", required: true, helpText: "Generated OAuth token" },
    ],
    buildRequest: (v) => ({
      name: "trello",
      description: "Manage Trello boards, lists, and cards",
      endpoint_url: "https://api.trello.com/1/members/me/boards",
      http_method: "GET",
      parameters: { type: "object", properties: { fields: { type: "string", default: "name,url" }, key: { type: "string", default: v.api_key }, token: { type: "string", default: v.token } } },
      timeout_sec: 15,
    }),
  },

  // ── Communication ─────────────────────────────────────────────────────
  {
    id: "slack",
    name: "Slack",
    description: "Send messages and interact with Slack workspaces",
    category: "communication",
    iconName: "Hash",
    authType: "bearer",
    defaults: {
      endpoint_url: "https://slack.com/api/chat.postMessage",
      http_method: "POST",
      parameters: { type: "object", properties: { channel: { type: "string", description: "Channel ID" }, text: { type: "string", description: "Message text" } }, required: ["channel", "text"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 10,
    },
    configFields: [
      { key: "token", label: "Bot Token", type: "password", required: true, placeholder: "xoxb-...", helpText: "Slack Bot User OAuth Token" },
    ],
    buildRequest: (v) => ({
      name: "slack",
      description: "Send messages and interact with Slack workspaces",
      endpoint_url: "https://slack.com/api/chat.postMessage",
      http_method: "POST",
      parameters: { type: "object", properties: { channel: { type: "string", description: "Channel ID" }, text: { type: "string", description: "Message text" } }, required: ["channel", "text"] },
      headers: { "Content-Type": "application/json" },
      auth_type: "bearer",
      auth_config: { token: v.token },
      timeout_sec: 10,
    }),
  },
  {
    id: "discord",
    name: "Discord Webhook",
    description: "Send messages to Discord channels via webhooks",
    category: "communication",
    iconName: "MessageCircle",
    authType: "none",
    defaults: {
      endpoint_url: "",
      http_method: "POST",
      parameters: { type: "object", properties: { content: { type: "string", description: "Message content" } }, required: ["content"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 10,
    },
    configFields: [
      { key: "webhook_url", label: "Webhook URL", type: "url", required: true, placeholder: "https://discord.com/api/webhooks/...", helpText: "Discord channel webhook URL" },
    ],
    buildRequest: (v) => ({
      name: "discord",
      description: "Send messages to Discord channels via webhooks",
      endpoint_url: v.webhook_url,
      http_method: "POST",
      parameters: { type: "object", properties: { content: { type: "string", description: "Message content" } }, required: ["content"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 10,
    }),
  },
  {
    id: "telegram",
    name: "Telegram",
    description: "Send messages via Telegram Bot API",
    category: "communication",
    iconName: "Send",
    authType: "api_key",
    defaults: {
      endpoint_url: "https://api.telegram.org",
      http_method: "POST",
      parameters: { type: "object", properties: { chat_id: { type: "string", description: "Chat ID" }, text: { type: "string", description: "Message text" } }, required: ["chat_id", "text"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 10,
    },
    configFields: [
      { key: "bot_token", label: "Bot Token", type: "password", required: true, placeholder: "123456:ABC-DEF...", helpText: "From @BotFather on Telegram" },
    ],
    buildRequest: (v) => ({
      name: "telegram",
      description: "Send messages via Telegram Bot API",
      endpoint_url: `https://api.telegram.org/bot${v.bot_token}/sendMessage`,
      http_method: "POST",
      parameters: { type: "object", properties: { chat_id: { type: "string", description: "Chat ID" }, text: { type: "string", description: "Message text" } }, required: ["chat_id", "text"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 10,
    }),
  },
  {
    id: "email-smtp",
    name: "Email (SMTP)",
    description: "Send emails via SMTP relay API",
    category: "communication",
    iconName: "Mail",
    authType: "api_key",
    defaults: {
      endpoint_url: "https://api.sendgrid.com/v3/mail/send",
      http_method: "POST",
      parameters: { type: "object", properties: { to: { type: "string", description: "Recipient email" }, subject: { type: "string", description: "Email subject" }, body: { type: "string", description: "Email body (HTML or text)" } }, required: ["to", "subject", "body"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 15,
    },
    configFields: [
      { key: "api_key", label: "API Key", type: "password", required: true, helpText: "SendGrid or similar SMTP relay API key" },
      { key: "endpoint_url", label: "Endpoint URL", type: "url", required: false, defaultValue: "https://api.sendgrid.com/v3/mail/send", helpText: "Custom SMTP relay endpoint (optional)" },
    ],
    buildRequest: (v) => ({
      name: "email-smtp",
      description: "Send emails via SMTP relay API",
      endpoint_url: v.endpoint_url || "https://api.sendgrid.com/v3/mail/send",
      http_method: "POST",
      parameters: { type: "object", properties: { to: { type: "string", description: "Recipient email" }, subject: { type: "string", description: "Email subject" }, body: { type: "string", description: "Email body" } }, required: ["to", "subject", "body"] },
      headers: { "Content-Type": "application/json" },
      auth_type: "bearer",
      auth_config: { token: v.api_key },
      timeout_sec: 15,
    }),
  },

  // ── Development ───────────────────────────────────────────────────────
  {
    id: "github",
    name: "GitHub",
    description: "Interact with GitHub repositories, issues, and PRs",
    category: "development",
    iconName: "Code",
    authType: "bearer",
    defaults: {
      endpoint_url: "https://api.github.com",
      http_method: "GET",
      parameters: { type: "object", properties: { endpoint: { type: "string", description: "API path (e.g. /repos/owner/repo/issues)" } }, required: ["endpoint"] },
      headers: { Accept: "application/vnd.github+json" },
      timeout_sec: 15,
    },
    configFields: [
      { key: "token", label: "Personal Access Token", type: "password", required: true, placeholder: "ghp_...", helpText: "From github.com/settings/tokens" },
    ],
    buildRequest: (v) => ({
      name: "github",
      description: "Interact with GitHub repositories, issues, and PRs",
      endpoint_url: "https://api.github.com",
      http_method: "GET",
      parameters: { type: "object", properties: { endpoint: { type: "string", description: "API path (e.g. /repos/owner/repo/issues)" } }, required: ["endpoint"] },
      headers: { Accept: "application/vnd.github+json" },
      auth_type: "bearer",
      auth_config: { token: v.token },
      timeout_sec: 15,
    }),
  },
  {
    id: "jira",
    name: "Jira",
    description: "Manage Jira issues and projects",
    category: "development",
    iconName: "BarChart3",
    authType: "bearer",
    defaults: {
      endpoint_url: "https://your-domain.atlassian.net/rest/api/3/search",
      http_method: "GET",
      parameters: { type: "object", properties: { jql: { type: "string", description: "JQL query" }, maxResults: { type: "number", default: 10 } } },
      headers: { Accept: "application/json" },
      timeout_sec: 15,
    },
    configFields: [
      { key: "domain", label: "Jira Domain", type: "text", required: true, placeholder: "your-company.atlassian.net", helpText: "Your Atlassian domain" },
      { key: "email", label: "Email", type: "text", required: true, helpText: "Atlassian account email" },
      { key: "api_token", label: "API Token", type: "password", required: true, helpText: "From id.atlassian.com/manage-profile/security/api-tokens" },
    ],
    buildRequest: (v) => ({
      name: "jira",
      description: "Manage Jira issues and projects",
      endpoint_url: `https://${v.domain}/rest/api/3/search`,
      http_method: "GET",
      parameters: { type: "object", properties: { jql: { type: "string", description: "JQL query" }, maxResults: { type: "number", default: 10 } } },
      headers: { Accept: "application/json" },
      auth_type: "bearer",
      auth_config: { token: btoa(`${v.email}:${v.api_token}`) },
      timeout_sec: 15,
    }),
  },
  {
    id: "rest-api",
    name: "REST API",
    description: "Generic REST API connector for any HTTP endpoint",
    category: "development",
    iconName: "Webhook",
    authType: "none",
    defaults: {
      endpoint_url: "",
      http_method: "GET",
      parameters: {},
      timeout_sec: 30,
    },
    configFields: [
      { key: "name", label: "Tool Name", type: "text", required: true, placeholder: "my-api", helpText: "Unique identifier for this tool" },
      { key: "description", label: "Description", type: "text", required: false, placeholder: "What does this API do?" },
      { key: "endpoint_url", label: "Endpoint URL", type: "url", required: true, placeholder: "https://api.example.com/v1" },
      { key: "http_method", label: "HTTP Method", type: "select", required: true, defaultValue: "GET", options: [{ label: "GET", value: "GET" }, { label: "POST", value: "POST" }, { label: "PUT", value: "PUT" }, { label: "DELETE", value: "DELETE" }, { label: "PATCH", value: "PATCH" }] },
      { key: "headers", label: "Headers (JSON)", type: "json", required: false, placeholder: '{"Content-Type": "application/json"}' },
      { key: "auth_type", label: "Auth Type", type: "select", required: false, options: [{ label: "None", value: "" }, { label: "Bearer Token", value: "bearer" }, { label: "API Key", value: "api_key" }] },
      { key: "auth_token", label: "Auth Token / API Key", type: "password", required: false, helpText: "Token or API key value" },
    ],
    buildRequest: (v) => {
      const req: CreateToolRequest = {
        name: v.name || "rest-api",
        description: v.description || "Generic REST API connector",
        endpoint_url: v.endpoint_url,
        http_method: v.http_method || "GET",
        timeout_sec: 30,
      };
      if (v.headers) {
        try { req.headers = JSON.parse(v.headers); } catch { /* skip */ }
      }
      if (v.auth_type === "bearer") {
        req.auth_type = "bearer";
        req.auth_config = { token: v.auth_token };
      } else if (v.auth_type === "api_key") {
        req.auth_type = "api_key";
        req.auth_config = { key: "X-API-Key", value: v.auth_token, in: "header" };
      }
      return req;
    },
  },
  {
    id: "graphql",
    name: "GraphQL",
    description: "Execute GraphQL queries against any endpoint",
    category: "development",
    iconName: "Braces",
    authType: "none",
    defaults: {
      endpoint_url: "",
      http_method: "POST",
      parameters: { type: "object", properties: { query: { type: "string", description: "GraphQL query" }, variables: { type: "object", description: "Query variables" } }, required: ["query"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 30,
    },
    configFields: [
      { key: "endpoint_url", label: "GraphQL Endpoint", type: "url", required: true, placeholder: "https://api.example.com/graphql" },
      { key: "auth_type", label: "Auth Type", type: "select", required: false, options: [{ label: "None", value: "" }, { label: "Bearer Token", value: "bearer" }, { label: "API Key", value: "api_key" }] },
      { key: "auth_token", label: "Auth Token / API Key", type: "password", required: false },
    ],
    buildRequest: (v) => {
      const req: CreateToolRequest = {
        name: "graphql",
        description: "Execute GraphQL queries against any endpoint",
        endpoint_url: v.endpoint_url,
        http_method: "POST",
        parameters: { type: "object", properties: { query: { type: "string", description: "GraphQL query" }, variables: { type: "object", description: "Query variables" } }, required: ["query"] },
        headers: { "Content-Type": "application/json" },
        timeout_sec: 30,
      };
      if (v.auth_type === "bearer") {
        req.auth_type = "bearer";
        req.auth_config = { token: v.auth_token };
      } else if (v.auth_type === "api_key") {
        req.auth_type = "api_key";
        req.auth_config = { key: "X-API-Key", value: v.auth_token, in: "header" };
      }
      return req;
    },
  },

  // ── AI & Media ────────────────────────────────────────────────────────
  {
    id: "dall-e",
    name: "DALL-E",
    description: "Generate images using OpenAI DALL-E API",
    category: "ai-media",
    iconName: "Image",
    authType: "api_key",
    defaults: {
      endpoint_url: "https://api.openai.com/v1/images/generations",
      http_method: "POST",
      parameters: { type: "object", properties: { prompt: { type: "string", description: "Image description" }, size: { type: "string", default: "1024x1024" }, n: { type: "number", default: 1 } }, required: ["prompt"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 60,
    },
    configFields: [
      { key: "api_key", label: "OpenAI API Key", type: "password", required: true, placeholder: "sk-...", helpText: "From platform.openai.com" },
    ],
    buildRequest: (v) => ({
      name: "dall-e",
      description: "Generate images using OpenAI DALL-E API",
      endpoint_url: "https://api.openai.com/v1/images/generations",
      http_method: "POST",
      parameters: { type: "object", properties: { prompt: { type: "string", description: "Image description" }, size: { type: "string", default: "1024x1024" }, n: { type: "number", default: 1 } }, required: ["prompt"] },
      headers: { "Content-Type": "application/json" },
      auth_type: "bearer",
      auth_config: { token: v.api_key },
      timeout_sec: 60,
    }),
  },
  {
    id: "elevenlabs-tts",
    name: "ElevenLabs TTS",
    description: "Convert text to speech with realistic AI voices",
    category: "ai-media",
    iconName: "AudioLines",
    authType: "api_key",
    defaults: {
      endpoint_url: "https://api.elevenlabs.io/v1/text-to-speech",
      http_method: "POST",
      parameters: { type: "object", properties: { text: { type: "string", description: "Text to convert to speech" }, voice_id: { type: "string", description: "Voice ID" } }, required: ["text"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 30,
    },
    configFields: [
      { key: "api_key", label: "API Key", type: "password", required: true, helpText: "From elevenlabs.io" },
      { key: "voice_id", label: "Default Voice ID", type: "text", required: false, placeholder: "21m00Tcm4TlvDq8ikWAM", helpText: "Default voice to use (optional)" },
    ],
    buildRequest: (v) => ({
      name: "elevenlabs-tts",
      description: "Convert text to speech with realistic AI voices",
      endpoint_url: `https://api.elevenlabs.io/v1/text-to-speech/${v.voice_id || "21m00Tcm4TlvDq8ikWAM"}`,
      http_method: "POST",
      parameters: { type: "object", properties: { text: { type: "string", description: "Text to convert to speech" } }, required: ["text"] },
      headers: { "Content-Type": "application/json" },
      auth_type: "api_key",
      auth_config: { key: "xi-api-key", value: v.api_key, in: "header" },
      timeout_sec: 30,
    }),
  },
  {
    id: "deepl",
    name: "DeepL Translation",
    description: "Translate text between languages using DeepL API",
    category: "ai-media",
    iconName: "Languages",
    authType: "api_key",
    defaults: {
      endpoint_url: "https://api-free.deepl.com/v2/translate",
      http_method: "POST",
      parameters: { type: "object", properties: { text: { type: "string", description: "Text to translate" }, target_lang: { type: "string", description: "Target language code (e.g. EN, DE, FR)" } }, required: ["text", "target_lang"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 15,
    },
    configFields: [
      { key: "api_key", label: "Auth Key", type: "password", required: true, helpText: "From deepl.com/account" },
      { key: "plan", label: "Plan", type: "select", required: false, defaultValue: "free", options: [{ label: "Free", value: "free" }, { label: "Pro", value: "pro" }], helpText: "Determines API endpoint" },
    ],
    buildRequest: (v) => ({
      name: "deepl",
      description: "Translate text between languages using DeepL API",
      endpoint_url: v.plan === "pro" ? "https://api.deepl.com/v2/translate" : "https://api-free.deepl.com/v2/translate",
      http_method: "POST",
      parameters: { type: "object", properties: { text: { type: "string", description: "Text to translate" }, target_lang: { type: "string", description: "Target language code (e.g. EN, DE, FR)" } }, required: ["text", "target_lang"] },
      headers: { "Content-Type": "application/json" },
      auth_type: "api_key",
      auth_config: { key: "Authorization", value: `DeepL-Auth-Key ${v.api_key}`, in: "header" },
      timeout_sec: 15,
    }),
  },

  // ── Data & Info ───────────────────────────────────────────────────────
  {
    id: "openweathermap",
    name: "OpenWeatherMap",
    description: "Get current weather data for any location",
    category: "data",
    iconName: "Cloud",
    authType: "api_key",
    defaults: {
      endpoint_url: "https://api.openweathermap.org/data/2.5/weather",
      http_method: "GET",
      parameters: { type: "object", properties: { q: { type: "string", description: "City name" }, units: { type: "string", default: "metric" } }, required: ["q"] },
      timeout_sec: 10,
    },
    configFields: [
      { key: "api_key", label: "API Key", type: "password", required: true, helpText: "From openweathermap.org/api" },
    ],
    buildRequest: (v) => ({
      name: "openweathermap",
      description: "Get current weather data for any location",
      endpoint_url: "https://api.openweathermap.org/data/2.5/weather",
      http_method: "GET",
      parameters: { type: "object", properties: { q: { type: "string", description: "City name" }, units: { type: "string", default: "metric" } }, required: ["q"] },
      auth_type: "api_key",
      auth_config: { key: "appid", value: v.api_key, in: "query" },
      timeout_sec: 10,
    }),
  },
  {
    id: "news-api",
    name: "News API",
    description: "Search and retrieve news articles from around the world",
    category: "data",
    iconName: "Newspaper",
    authType: "api_key",
    defaults: {
      endpoint_url: "https://newsapi.org/v2/everything",
      http_method: "GET",
      parameters: { type: "object", properties: { q: { type: "string", description: "Search keywords" }, language: { type: "string", default: "en" }, sortBy: { type: "string", default: "publishedAt" } }, required: ["q"] },
      timeout_sec: 15,
    },
    configFields: [
      { key: "api_key", label: "API Key", type: "password", required: true, helpText: "From newsapi.org" },
    ],
    buildRequest: (v) => ({
      name: "news-api",
      description: "Search and retrieve news articles from around the world",
      endpoint_url: "https://newsapi.org/v2/everything",
      http_method: "GET",
      parameters: { type: "object", properties: { q: { type: "string", description: "Search keywords" }, language: { type: "string", default: "en" }, sortBy: { type: "string", default: "publishedAt" } }, required: ["q"] },
      auth_type: "api_key",
      auth_config: { key: "X-Api-Key", value: v.api_key, in: "header" },
      timeout_sec: 15,
    }),
  },
  {
    id: "alpha-vantage",
    name: "Alpha Vantage",
    description: "Real-time and historical stock market data",
    category: "data",
    iconName: "TrendingUp",
    authType: "api_key",
    defaults: {
      endpoint_url: "https://www.alphavantage.co/query",
      http_method: "GET",
      parameters: { type: "object", properties: { function: { type: "string", description: "API function (e.g. TIME_SERIES_DAILY)" }, symbol: { type: "string", description: "Stock symbol (e.g. AAPL)" } }, required: ["function", "symbol"] },
      timeout_sec: 15,
    },
    configFields: [
      { key: "api_key", label: "API Key", type: "password", required: true, helpText: "From alphavantage.co" },
    ],
    buildRequest: (v) => ({
      name: "alpha-vantage",
      description: "Real-time and historical stock market data",
      endpoint_url: "https://www.alphavantage.co/query",
      http_method: "GET",
      parameters: { type: "object", properties: { function: { type: "string", description: "API function (e.g. TIME_SERIES_DAILY)" }, symbol: { type: "string", description: "Stock symbol (e.g. AAPL)" } }, required: ["function", "symbol"] },
      auth_type: "api_key",
      auth_config: { key: "apikey", value: v.api_key, in: "query" },
      timeout_sec: 15,
    }),
  },

  // ── Utilities ─────────────────────────────────────────────────────────
  {
    id: "calculator",
    name: "Calculator",
    description: "Evaluate mathematical expressions via API",
    category: "utilities",
    iconName: "Calculator",
    authType: "none",
    defaults: {
      endpoint_url: "https://api.mathjs.org/v4/",
      http_method: "POST",
      parameters: { type: "object", properties: { expr: { type: "string", description: "Mathematical expression to evaluate" } }, required: ["expr"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 10,
    },
    configFields: [],
    buildRequest: () => ({
      name: "calculator",
      description: "Evaluate mathematical expressions via API",
      endpoint_url: "https://api.mathjs.org/v4/",
      http_method: "POST",
      parameters: { type: "object", properties: { expr: { type: "string", description: "Mathematical expression to evaluate" } }, required: ["expr"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 10,
    }),
  },
  {
    id: "code-execution",
    name: "Code Execution",
    description: "Execute code snippets via a sandboxed execution API",
    category: "utilities",
    iconName: "Terminal",
    authType: "api_key",
    defaults: {
      endpoint_url: "",
      http_method: "POST",
      parameters: { type: "object", properties: { language: { type: "string", description: "Programming language" }, code: { type: "string", description: "Code to execute" } }, required: ["language", "code"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 30,
    },
    configFields: [
      { key: "endpoint_url", label: "Execution API URL", type: "url", required: true, placeholder: "https://api.judge0.com/submissions", helpText: "Sandboxed code execution endpoint" },
      { key: "api_key", label: "API Key", type: "password", required: false, helpText: "API key if required" },
    ],
    buildRequest: (v) => {
      const req: CreateToolRequest = {
        name: "code-execution",
        description: "Execute code snippets via a sandboxed execution API",
        endpoint_url: v.endpoint_url,
        http_method: "POST",
        parameters: { type: "object", properties: { language: { type: "string", description: "Programming language" }, code: { type: "string", description: "Code to execute" } }, required: ["language", "code"] },
        headers: { "Content-Type": "application/json" },
        timeout_sec: 30,
      };
      if (v.api_key) {
        req.auth_type = "api_key";
        req.auth_config = { key: "X-API-Key", value: v.api_key, in: "header" };
      }
      return req;
    },
  },
  {
    id: "json-transform",
    name: "JSON Transform",
    description: "Transform and query JSON data using JSONPath or JQ expressions",
    category: "utilities",
    iconName: "Braces",
    authType: "none",
    defaults: {
      endpoint_url: "https://api.jsonata.org/evaluate",
      http_method: "POST",
      parameters: { type: "object", properties: { expression: { type: "string", description: "JSONata expression" }, data: { type: "object", description: "Input JSON data" } }, required: ["expression", "data"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 10,
    },
    configFields: [],
    buildRequest: () => ({
      name: "json-transform",
      description: "Transform and query JSON data using JSONata expressions",
      endpoint_url: "https://api.jsonata.org/evaluate",
      http_method: "POST",
      parameters: { type: "object", properties: { expression: { type: "string", description: "JSONata expression" }, data: { type: "object", description: "Input JSON data" } }, required: ["expression", "data"] },
      headers: { "Content-Type": "application/json" },
      timeout_sec: 10,
    }),
  },
  {
    id: "custom-tool",
    name: "Custom Tool",
    description: "Create a fully custom HTTP tool with all fields exposed",
    category: "utilities",
    iconName: "Wrench",
    authType: "none",
    defaults: {
      endpoint_url: "",
      http_method: "GET",
      parameters: {},
      timeout_sec: 30,
    },
    configFields: [
      { key: "name", label: "Tool Name", type: "text", required: true, placeholder: "my-custom-tool" },
      { key: "description", label: "Description", type: "text", required: true, placeholder: "What does this tool do?" },
      { key: "endpoint_url", label: "Endpoint URL", type: "url", required: true, placeholder: "https://api.example.com/endpoint" },
      { key: "http_method", label: "HTTP Method", type: "select", required: true, defaultValue: "GET", options: [{ label: "GET", value: "GET" }, { label: "POST", value: "POST" }, { label: "PUT", value: "PUT" }, { label: "DELETE", value: "DELETE" }, { label: "PATCH", value: "PATCH" }] },
      { key: "parameters", label: "Parameters (JSON Schema)", type: "json", required: false, placeholder: '{"type": "object", "properties": {"query": {"type": "string"}}}' },
      { key: "headers", label: "Headers (JSON)", type: "json", required: false, placeholder: '{"Content-Type": "application/json"}' },
      { key: "auth_type", label: "Auth Type", type: "select", required: false, options: [{ label: "None", value: "" }, { label: "Bearer Token", value: "bearer" }, { label: "API Key", value: "api_key" }] },
      { key: "auth_config", label: "Auth Config (JSON)", type: "json", required: false, placeholder: '{"token": "your-token"}', helpText: "Bearer: {\"token\": \"...\"} | API Key: {\"key\": \"header-name\", \"value\": \"...\", \"in\": \"header\"}" },
      { key: "timeout_sec", label: "Timeout (seconds)", type: "number", required: false, defaultValue: "30" },
    ],
    buildRequest: (v) => {
      const req: CreateToolRequest = {
        name: v.name || "custom-tool",
        description: v.description || "Custom tool",
        endpoint_url: v.endpoint_url,
        http_method: v.http_method || "GET",
        timeout_sec: parseInt(v.timeout_sec) || 30,
      };
      if (v.parameters) {
        try { req.parameters = JSON.parse(v.parameters); } catch { /* skip */ }
      }
      if (v.headers) {
        try { req.headers = JSON.parse(v.headers); } catch { /* skip */ }
      }
      if (v.auth_type) {
        req.auth_type = v.auth_type;
        if (v.auth_config) {
          try { req.auth_config = JSON.parse(v.auth_config); } catch { /* skip */ }
        }
      }
      return req;
    },
  },
];

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

export function getTemplateById(id: string): ToolTemplate | undefined {
  return TOOL_TEMPLATES.find((t) => t.id === id);
}

export function getTemplatesByCategory(category: ToolCategory): ToolTemplate[] {
  return TOOL_TEMPLATES.filter((t) => t.category === category);
}

export function searchTemplates(query: string): ToolTemplate[] {
  const q = query.toLowerCase();
  return TOOL_TEMPLATES.filter(
    (t) =>
      t.name.toLowerCase().includes(q) ||
      t.description.toLowerCase().includes(q) ||
      t.id.toLowerCase().includes(q)
  );
}
