from .openai import OpenAIProvider

DEEPSEEK_API_BASE = "https://api.deepseek.com"
DEFAULT_MODEL = "deepseek-chat"


class DeepSeekProvider(OpenAIProvider):
    """DeepSeek provider via its OpenAI-compatible API.

    Inherits the full OpenAI implementation (tool calling, multimodal, memory)
    and only overrides the base URL and default model.

    Supported models:
    - deepseek-chat    (DeepSeek-V3 — general purpose, supports tool calling)
    - deepseek-reasoner (DeepSeek-R1 — reasoning, does NOT support tool calling)
    """

    _default_model: str = DEFAULT_MODEL

    def __init__(self, api_key: str):
        super().__init__(api_key, base_url=DEEPSEEK_API_BASE)
