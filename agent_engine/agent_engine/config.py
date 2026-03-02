import os


class Config:
    """Worker configuration loaded from environment variables."""

    def __init__(self):
        self.worker_id = os.getenv("WORKER_ID", f"worker-{os.getpid()}")
        self.grpc_host = os.getenv("GRPC_HOST", "localhost")
        self.grpc_port = int(os.getenv("GRPC_PORT", "50051"))
        self.grpc_api_key = os.getenv("GRPC_WORKER_API_KEY", "")
        self.max_concurrent = int(os.getenv("MAX_CONCURRENT", "4"))
        self.heartbeat_interval = int(os.getenv("HEARTBEAT_INTERVAL", "30"))
        self.reconnect_delay = int(os.getenv("RECONNECT_DELAY", "5"))

        # LLM API keys
        self.openai_api_key = os.getenv("OPENAI_API_KEY", "")
        self.anthropic_api_key = os.getenv("ANTHROPIC_API_KEY", "")
        self.deepseek_api_key = os.getenv("DEEPSEEK_API_KEY", "")
        self.gemini_api_key = os.getenv("GEMINI_API_KEY", "")

        # Ollama (always available as a local provider)
        self.ollama_base_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
        self.ollama_default_model = os.getenv("OLLAMA_DEFAULT_MODEL", "llama3.2")

        # Tool API keys
        self.tavily_api_key = os.getenv("TAVILY_API_KEY", "")

        # Tracing
        self.tracing_enabled = os.getenv("TRACING_ENABLED", "").lower() in ("true", "1")
        self.tracing_otlp_endpoint = os.getenv(
            "TRACING_OTLP_ENDPOINT", "localhost:4317"
        )

        # Budget limits
        self.budget_max_steps = int(os.getenv("BUDGET_MAX_STEPS", "15"))
        self.budget_max_tokens = int(os.getenv("BUDGET_MAX_TOKENS", "50000"))
        self.budget_max_time_sec = float(os.getenv("BUDGET_MAX_TIME_SEC", "90"))
        self.budget_max_cost_usd = float(os.getenv("BUDGET_MAX_COST_USD", "1.0"))

        # Retry
        self.retry_max_retries = int(os.getenv("RETRY_MAX_RETRIES", "3"))
        self.retry_base_delay_sec = float(os.getenv("RETRY_BASE_DELAY_SEC", "1.0"))
        self.retry_max_delay_sec = float(os.getenv("RETRY_MAX_DELAY_SEC", "30.0"))

        # Planner
        self.planner_enabled = os.getenv("PLANNER_ENABLED", "").lower() in ("true", "1")
        self.planner_model = os.getenv("PLANNER_MODEL", "")

        # Fallback
        self.fallback_providers = [
            p.strip()
            for p in os.getenv("FALLBACK_PROVIDERS", "").split(",")
            if p.strip()
        ]

        # Model router
        self.model_router_enabled = os.getenv("MODEL_ROUTER_ENABLED", "").lower() in (
            "true",
            "1",
        )

        # Logging
        self.log_json = os.getenv("LOG_JSON", "true").lower() in ("true", "1")

    @property
    def grpc_target(self) -> str:
        return f"{self.grpc_host}:{self.grpc_port}"

    @property
    def supported_providers(self) -> list[str]:
        providers = []
        if self.openai_api_key:
            providers.append("openai")
        if self.anthropic_api_key:
            providers.append("anthropic")
        if self.deepseek_api_key:
            providers.append("deepseek")
        if self.gemini_api_key:
            providers.append("gemini")
        providers.append("ollama")  # always available (local)
        return providers
