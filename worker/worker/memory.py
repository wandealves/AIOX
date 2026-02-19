import json
import logging
from dataclasses import dataclass, field

logger = logging.getLogger(__name__)


@dataclass
class ConversationEntry:
    role: str
    content: str
    timestamp: str = ""


@dataclass
class RelevantMemory:
    content: str
    memory_type: str = "long_term"
    similarity: float = 0.0


@dataclass
class MemoryContext:
    recent_messages: list[ConversationEntry] = field(default_factory=list)
    relevant_memories: list[RelevantMemory] = field(default_factory=list)

    @classmethod
    def from_json(cls, data: str) -> "MemoryContext":
        """Parse the memory context JSON sent by the Go dispatcher."""
        if not data:
            return cls()
        try:
            raw = json.loads(data)
        except json.JSONDecodeError:
            logger.warning("Failed to parse memory context JSON")
            return cls()

        recent = []
        for msg in raw.get("recent_messages") or []:
            recent.append(ConversationEntry(
                role=msg.get("role", "user"),
                content=msg.get("content", ""),
                timestamp=msg.get("timestamp", ""),
            ))

        memories = []
        for mem in raw.get("relevant_memories") or []:
            memories.append(RelevantMemory(
                content=mem.get("content", ""),
                memory_type=mem.get("memory_type", "long_term"),
                similarity=mem.get("similarity", 0.0),
            ))

        return cls(recent_messages=recent, relevant_memories=memories)

    def build_messages_for_llm(
        self, system_prompt: str, user_message: str
    ) -> list[dict]:
        """Build a full messages array for the LLM with memory context.

        Structure:
        1. System message (with relevant memories appended if any)
        2. Recent conversation history (from short-term memory)
        3. Current user message
        """
        # Build system content with relevant memories
        system_content = system_prompt
        if self.relevant_memories:
            memory_section = "\n\n--- Relevant memories from past interactions ---"
            for mem in self.relevant_memories:
                memory_section += f"\n[{mem.memory_type}] {mem.content}"
            system_content += memory_section

        messages = [{"role": "system", "content": system_content}]

        # Add recent conversation history
        for entry in self.recent_messages:
            messages.append({"role": entry.role, "content": entry.content})

        # Add current user message
        messages.append({"role": "user", "content": user_message})

        return messages


@dataclass
class MemoryConfig:
    enabled: bool = False
    short_term_enabled: bool = True
    long_term_enabled: bool = True
    max_short_term_msgs: int = 20
    short_term_ttl_sec: int = 3600
    max_long_term_results: int = 5
    similarity_threshold: float = 0.7

    @classmethod
    def from_json(cls, data: str) -> "MemoryConfig":
        """Parse memory config JSON from the Go dispatcher."""
        if not data:
            return cls()
        try:
            raw = json.loads(data)
        except json.JSONDecodeError:
            return cls()
        return cls(
            enabled=raw.get("enabled", False),
            short_term_enabled=raw.get("short_term_enabled", True),
            long_term_enabled=raw.get("long_term_enabled", True),
            max_short_term_msgs=raw.get("max_short_term_msgs", 20),
            short_term_ttl_sec=raw.get("short_term_ttl_sec", 3600),
            max_long_term_results=raw.get("max_long_term_results", 5),
            similarity_threshold=raw.get("similarity_threshold", 0.7),
        )
