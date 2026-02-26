export interface ApiResponse<T> {
  data?: T;
  message?: string;
  error?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  total_count: number;
  page: number;
  page_size: number;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface User {
  id: string;
  email: string;
  role?: string;
  disabled_at?: string;
  created_at: string;
}

export interface AgentProfile {
  name: string;
  description: string;
  system_prompt: string;
  personality_traits?: string[];
  encrypted: boolean;
}

export interface Agent {
  id: string;
  owner_user_id: string;
  jid: string;
  profile: AgentProfile;
  llm_config: Record<string, unknown>;
  capabilities: Record<string, unknown>;
  memory_config: Record<string, unknown>;
  governance: Record<string, unknown>;
  visibility: string;
  org_id?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateAgentRequest {
  name: string;
  description: string;
  system_prompt: string;
  personality_traits?: string[];
  llm_config?: Record<string, unknown>;
  capabilities?: Record<string, unknown>;
  memory_config?: Record<string, unknown>;
  governance?: Record<string, unknown>;
  visibility?: string;
}

export interface UpdateAgentRequest {
  name?: string;
  description?: string;
  system_prompt?: string;
  personality_traits?: string[];
  llm_config?: Record<string, unknown>;
  capabilities?: Record<string, unknown>;
  memory_config?: Record<string, unknown>;
  governance?: Record<string, unknown>;
  visibility?: string;
}

export interface Conversation {
  id: string;
  agent_id: string;
  input: string;
  output: string;
  status: string;
  error_message?: string;
  tokens_used: number;
  duration_ms: number;
  created_at: string;
}

export interface SendMessageResponse {
  request_id: string;
  status: string;
}

export interface QuotaInfo {
  user_id: string;
  tokens_used_today: number;
  requests_today: number;
  max_tokens_per_day: number;
  max_requests_per_day: number;
  rate_limit_requests_per_minute: number;
  rate_limit_current_minute: number;
}

export interface AuditLog {
  id: string;
  owner_user_id: string;
  event_type: string;
  severity: string;
  resource_type: string;
  resource_id: string;
  details: string;
  created_at: string;
}

export interface Memory {
  id: string;
  agent_id: string;
  content: string;
  memory_type: string;
  metadata: Record<string, unknown>;
  created_at: string;
}

export interface AdminStats {
  total_users: number;
  total_agents: number;
  total_executions: number;
  active_users_today: number;
}

export interface Organization {
  id: string;
  name: string;
  slug: string;
  description: string;
  settings: Record<string, unknown>;
  quota_config: Record<string, unknown>;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface OrgMember {
  id: string;
  org_id: string;
  user_id: string;
  role: string;
  email?: string;
  joined_at: string;
}

export interface OrgInvite {
  id: string;
  org_id: string;
  email: string;
  role: string;
  token: string;
  invited_by: string;
  expires_at: string;
  accepted_at?: string;
  created_at: string;
}

export interface ToolDefinition {
  id: string;
  agent_id: string;
  name: string;
  description: string;
  parameters: Record<string, unknown> | null;
  endpoint_url: string;
  http_method: string;
  headers: Record<string, string> | null;
  auth_type: string;
  auth_config: Record<string, unknown> | null;
  is_active: boolean;
  timeout_sec: number;
  created_at: string;
  updated_at: string;
}

export interface CreateToolRequest {
  name: string;
  description: string;
  parameters?: Record<string, unknown>;
  endpoint_url: string;
  http_method: string;
  headers?: Record<string, string>;
  auth_type?: string;
  auth_config?: Record<string, unknown>;
  timeout_sec?: number;
}

export interface UpdateToolRequest {
  name?: string;
  description?: string;
  parameters?: Record<string, unknown>;
  endpoint_url?: string;
  http_method?: string;
  headers?: Record<string, string>;
  auth_type?: string;
  auth_config?: Record<string, unknown>;
  is_active?: boolean;
  timeout_sec?: number;
}

// WebSocket message types
export interface WSServerMessage {
  type: "message" | "error";
  agent_id?: string;
  body?: string;
  request_id?: string;
  error?: string;
}

export interface WSClientMessage {
  type: "message";
  agent_id: string;
  body: string;
}
