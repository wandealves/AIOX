"""Tavily Search — real-time web search tool for LLM agents."""

import json
import logging
import os

import aiohttp

from ..base import BuiltinTool

logger = logging.getLogger(__name__)

TAVILY_API_URL = "https://api.tavily.com/search"

TAVILY_SEARCH_PARAMETERS = {
    "type": "object",
    "properties": {
        "query": {
            "type": "string",
            "description": "The search query to look up on the web",
        },
        "search_depth": {
            "type": "string",
            "enum": ["basic", "advanced"],
            "description": "Search depth: 'basic' for fast results, 'advanced' for more thorough search",
            "default": "basic",
        },
        "max_results": {
            "type": "integer",
            "description": "Maximum number of results to return (1-20)",
            "default": 5,
            "minimum": 1,
            "maximum": 20,
        },
        "include_answer": {
            "type": "boolean",
            "description": "Whether to include a short AI-generated answer summarizing the results",
            "default": True,
        },
    },
    "required": ["query"],
}


class TavilySearchTool(BuiltinTool):
    """Web search tool powered by Tavily API.

    Performs real-time web searches and returns structured results
    including titles, URLs, content snippets, and an optional
    AI-generated summary answer.
    """

    name = "tavily_search"
    description = (
        "Search the web for real-time information. Returns relevant web pages "
        "with titles, URLs, and content snippets. Useful for finding current "
        "events, facts, documentation, and up-to-date information."
    )
    parameters = TAVILY_SEARCH_PARAMETERS
    version = "1.0.0"
    idempotent = True
    side_effect_level = "read"

    @staticmethod
    def _resolve_api_key(arguments: dict, context: dict | None) -> str:
        """Extract Tavily API key with priority:

        1. arguments["api_key"] — sent by the LLM or set as parameter default
        2. context["auth_config"]["key"] — from tool auth configuration (api_key auth)
        3. context["auth_config"]["token"] — from tool auth configuration (bearer auth)
        4. TAVILY_API_KEY env var — fallback
        """
        # 1. From arguments (LLM may fill it, or it may carry the schema default)
        if arguments.get("api_key"):
            return arguments["api_key"]

        # 2/3. From tool auth_config
        if context:
            auth_config = context.get("auth_config", {})
            if isinstance(auth_config, dict):
                key = auth_config.get("key", "")
                if key:
                    return key
                token = auth_config.get("token", "")
                if token:
                    return token

        # 4. Env var fallback
        return os.getenv("TAVILY_API_KEY", "")

    async def execute(self, arguments: dict, context: dict | None = None) -> dict:
        api_key = self._resolve_api_key(arguments, context)
        if not api_key:
            return {
                "status": "error",
                "data": None,
                "error": "Tavily API key not configured (set via tool auth_config or TAVILY_API_KEY env var)",
            }

        query = arguments.get("query", "").strip()
        if not query:
            return {
                "status": "error",
                "data": None,
                "error": "query parameter is required",
            }

        search_depth = arguments.get("search_depth", "basic")
        if search_depth not in ("basic", "advanced"):
            search_depth = "basic"

        max_results = arguments.get("max_results", 5)
        if not isinstance(max_results, int) or max_results < 1:
            max_results = 5
        max_results = min(max_results, 20)

        include_answer = arguments.get("include_answer", True)

        payload = {
            "api_key": api_key,
            "query": query,
            "search_depth": search_depth,
            "max_results": max_results,
            "include_answer": bool(include_answer),
        }

        timeout = aiohttp.ClientTimeout(total=30)
        try:
            async with aiohttp.ClientSession(timeout=timeout) as session:
                async with session.post(
                    TAVILY_API_URL,
                    json=payload,
                    headers={"Content-Type": "application/json"},
                ) as resp:
                    body = await resp.text()

                    if resp.status != 200:
                        return {
                            "status": "error",
                            "data": None,
                            "error": f"Tavily API error (HTTP {resp.status}): {body[:500]}",
                        }

                    data = json.loads(body)

                    results = []
                    for r in data.get("results", []):
                        results.append(
                            {
                                "title": r.get("title", ""),
                                "url": r.get("url", ""),
                                "content": r.get("content", ""),
                                "score": r.get("score", 0),
                            }
                        )

                    return {
                        "status": "success",
                        "data": {
                            "query": query,
                            "results": results,
                            "answer": data.get("answer", ""),
                        },
                        "error": None,
                    }

        except aiohttp.ClientError as e:
            logger.error("Tavily API request failed: %s", e)
            return {
                "status": "error",
                "data": None,
                "error": f"Request failed: {e}",
            }
        except json.JSONDecodeError:
            return {
                "status": "error",
                "data": None,
                "error": "Failed to parse Tavily API response",
            }
