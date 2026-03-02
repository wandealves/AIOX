"""Input sanitization and injection detection (heuristic, log-only)."""

import re

MAX_USER_MESSAGE_LEN = 32_000
MAX_TOOL_RESULT_LEN = 100_000

# Heuristic patterns for common prompt injection attempts.
# These are advisory — they log warnings but do NOT block.
_INJECTION_PATTERNS: list[tuple[str, re.Pattern]] = [
    ("ignore_previous", re.compile(r"ignore\s+(all\s+)?previous\s+instructions?", re.IGNORECASE)),
    ("system_override", re.compile(r"you\s+are\s+now\s+(a|an)\s+", re.IGNORECASE)),
    ("role_switch", re.compile(r"<\|?(system|assistant|im_start)\|?>", re.IGNORECASE)),
    ("prompt_leak", re.compile(r"(repeat|print|show|reveal)\s+(your|the)\s+(system\s+)?prompt", re.IGNORECASE)),
    ("jailbreak_marker", re.compile(r"(DAN|developer\s+mode|jailbreak)", re.IGNORECASE)),
]


def check_length(text: str, max_len: int, field_name: str = "input") -> str | None:
    """Return a warning string if *text* exceeds *max_len*, else None."""
    if len(text) > max_len:
        return f"{field_name} exceeds max length ({len(text)} > {max_len})"
    return None


def detect_injection(text: str) -> list[str]:
    """Return a list of matched heuristic pattern names. Empty = clean."""
    matches = []
    for name, pattern in _INJECTION_PATTERNS:
        if pattern.search(text):
            matches.append(name)
    return matches


def sanitize_tool_result(result: str) -> str:
    """Truncate and clean a tool result for inclusion in the LLM context."""
    if len(result) > MAX_TOOL_RESULT_LEN:
        return result[:MAX_TOOL_RESULT_LEN] + "\n[... truncated]"
    return result
