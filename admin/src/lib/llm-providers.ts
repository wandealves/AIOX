export interface ModelOption {
  value: string;
  label: string;
}

export interface ProviderConfig {
  value: string;
  label: string;
  defaultModel: string;
  models: ModelOption[];
}

export const LLM_PROVIDERS: ProviderConfig[] = [
  {
    value: "openai",
    label: "OpenAI",
    defaultModel: "gpt-4o-mini",
    models: [
      { value: "gpt-4o", label: "GPT-4o" },
      { value: "gpt-4o-mini", label: "GPT-4o Mini" },
      { value: "gpt-4-turbo", label: "GPT-4 Turbo" },
      { value: "gpt-3.5-turbo", label: "GPT-3.5 Turbo" },
    ],
  },
  {
    value: "anthropic",
    label: "Anthropic",
    defaultModel: "claude-sonnet-4-20250514",
    models: [
      { value: "claude-opus-4-5", label: "Claude Opus 4.5" },
      { value: "claude-sonnet-4-20250514", label: "Claude Sonnet 4" },
      { value: "claude-haiku-4-5-20251001", label: "Claude Haiku 4.5" },
      { value: "claude-3-5-sonnet-20241022", label: "Claude 3.5 Sonnet" },
    ],
  },
  {
    value: "deepseek",
    label: "DeepSeek",
    defaultModel: "deepseek-chat",
    models: [
      { value: "deepseek-chat", label: "DeepSeek Chat (V3)" },
      { value: "deepseek-reasoner", label: "DeepSeek Reasoner (R1)" },
    ],
  },
  {
    value: "gemini",
    label: "Google Gemini",
    defaultModel: "gemini-2.0-flash",
    models: [
      { value: "gemini-2.0-flash", label: "Gemini 2.0 Flash" },
      { value: "gemini-2.0-flash-lite", label: "Gemini 2.0 Flash Lite" },
      { value: "gemini-1.5-pro", label: "Gemini 1.5 Pro" },
      { value: "gemini-1.5-flash", label: "Gemini 1.5 Flash" },
    ],
  },
  {
    value: "ollama",
    label: "Ollama (Local)",
    defaultModel: "llama3.2",
    models: [
      { value: "llama3.2", label: "Llama 3.2" },
      { value: "llama3.1", label: "Llama 3.1" },
      { value: "mistral", label: "Mistral" },
      { value: "qwen2.5", label: "Qwen 2.5" },
      { value: "phi4", label: "Phi-4" },
      { value: "codellama", label: "Code Llama" },
    ],
  },
];

export function getProviderConfig(value: string): ProviderConfig | undefined {
  return LLM_PROVIDERS.find((p) => p.value === value);
}

export function getProviderLabel(value: string): string {
  return getProviderConfig(value)?.label ?? value;
}

/** Tailwind classes for the provider badge in agent-card */
export const PROVIDER_BADGE_STYLES: Record<string, string> = {
  openai:    "bg-sky-500/10 text-sky-600",
  anthropic: "bg-amber-500/10 text-amber-600",
  deepseek:  "bg-indigo-500/10 text-indigo-600",
  gemini:    "bg-emerald-500/10 text-emerald-600",
  ollama:    "bg-violet-500/10 text-violet-600",
};

export function providerBadgeStyle(value: string): string {
  return (
    PROVIDER_BADGE_STYLES[value] ??
    "bg-[var(--background-tertiary)] text-[var(--foreground-muted)]"
  );
}
