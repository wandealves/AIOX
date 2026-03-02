"""Policy engine — validates tool URLs and provider access."""

from dataclasses import dataclass, field
from urllib.parse import urlparse


@dataclass
class Policy:
    """Per-task policy derived from the governance permissions sent by Go."""

    allowed_providers: list[str] = field(default_factory=list)
    allowed_domains: list[str] = field(default_factory=list)
    blocked: bool = False

    @classmethod
    def from_permissions(cls, permissions: dict | None) -> "Policy":
        if not permissions:
            return cls()
        return cls(
            allowed_providers=permissions.get("allowed_providers") or [],
            allowed_domains=permissions.get("allowed_domains") or [],
            blocked=permissions.get("blocked", False),
        )

    def check_provider(self, provider: str) -> bool:
        """Return True if the provider is allowed (empty list = all allowed)."""
        if not self.allowed_providers:
            return True
        return provider in self.allowed_providers

    def check_tool_url(self, url: str) -> bool:
        """Return True if the URL's hostname is allowed (empty list = all allowed).

        Supports subdomain matching: ``api.example.com`` matches ``example.com``.
        """
        if not self.allowed_domains:
            return True
        try:
            hostname = urlparse(url).hostname or ""
        except Exception:
            return False
        hostname = hostname.lower()
        for domain in self.allowed_domains:
            domain = domain.lower()
            if hostname == domain or hostname.endswith("." + domain):
                return True
        return False
