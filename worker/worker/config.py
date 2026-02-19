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
        self.ollama_base_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")

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
        providers.append("ollama")  # always available (local)
        return providers
