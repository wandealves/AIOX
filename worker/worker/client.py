import asyncio
import json
import logging
import time
import tracemalloc

import grpc

from .config import Config
from .llm.base import LLMProvider, LLMResponse
from .llm.openai import OpenAIProvider
from .llm.anthropic import AnthropicProvider

# Import generated protobuf modules
from . import worker_pb2
from . import worker_pb2_grpc

logger = logging.getLogger(__name__)


class WorkerClient:
    """gRPC client that connects to the AIOX server, receives tasks, and returns results."""

    def __init__(self, config: Config):
        self.config = config
        self.providers: dict[str, LLMProvider] = {}
        self._setup_providers()
        self.semaphore = asyncio.Semaphore(config.max_concurrent)

    def _setup_providers(self):
        if self.config.openai_api_key:
            self.providers["openai"] = OpenAIProvider(self.config.openai_api_key)
            logger.info("OpenAI provider configured")
        if self.config.anthropic_api_key:
            self.providers["anthropic"] = AnthropicProvider(self.config.anthropic_api_key)
            logger.info("Anthropic provider configured")

    def _get_provider(self, provider_name: str) -> LLMProvider | None:
        return self.providers.get(provider_name)

    async def run(self):
        """Main loop: connect, register, process tasks. Reconnects on failure."""
        while True:
            try:
                await self._connect_and_process()
            except Exception as e:
                logger.error("Connection error: %s", e)

            logger.info(
                "Reconnecting in %d seconds...", self.config.reconnect_delay
            )
            await asyncio.sleep(self.config.reconnect_delay)

    async def _connect_and_process(self):
        metadata = []
        if self.config.grpc_api_key:
            metadata.append(("x-api-key", self.config.grpc_api_key))

        channel = grpc.aio.insecure_channel(self.config.grpc_target)
        stub = worker_pb2_grpc.WorkerServiceStub(channel)

        try:
            stream = stub.TaskStream(metadata=metadata)

            # Register
            register_msg = worker_pb2.WorkerMessage(
                register=worker_pb2.RegisterWorker(
                    worker_id=self.config.worker_id,
                    max_concurrent=self.config.max_concurrent,
                    supported_providers=self.config.supported_providers,
                )
            )
            await stream.write(register_msg)

            # Wait for ack
            server_msg = await stream.read()
            ack = server_msg.register_ack
            if not ack.accepted:
                logger.error("Registration rejected: %s", ack.message)
                return

            logger.info("Registered as %s", self.config.worker_id)

            # Start heartbeat task
            heartbeat_task = asyncio.create_task(
                self._heartbeat_loop(stub, metadata)
            )

            # Process incoming tasks
            try:
                async for server_msg in stream:
                    task_req = server_msg.task_request
                    if task_req and task_req.request_id:
                        asyncio.create_task(
                            self._process_task(stream, task_req)
                        )
            finally:
                heartbeat_task.cancel()
                try:
                    await heartbeat_task
                except asyncio.CancelledError:
                    pass

        finally:
            await channel.close()

    async def _process_task(self, stream, task_req):
        """Process a single task with concurrency limiting."""
        async with self.semaphore:
            logger.info(
                "Processing task %s for agent %s",
                task_req.request_id,
                task_req.agent_id,
            )

            response = await self._call_llm(task_req)

            result_msg = worker_pb2.WorkerMessage(
                task_response=worker_pb2.TaskResponse(
                    request_id=task_req.request_id,
                    worker_id=self.config.worker_id,
                    response_text=response.text,
                    tokens_used=response.tokens_used,
                    duration_ms=response.duration_ms,
                    model_used=response.model_used,
                    error_message=response.error,
                )
            )
            await stream.write(result_msg)

            logger.info(
                "Task %s completed: %d tokens, %dms",
                task_req.request_id,
                response.tokens_used,
                response.duration_ms,
            )

    async def _call_llm(self, task_req) -> LLMResponse:
        """Call the appropriate LLM provider based on agent's llm_config."""
        try:
            llm_config = json.loads(task_req.llm_config_json) if task_req.llm_config_json else {}
        except json.JSONDecodeError:
            llm_config = {}

        provider_name = llm_config.get("provider", "openai")
        model = llm_config.get("model", "")
        temperature = llm_config.get("temperature", 0.7)
        max_tokens = llm_config.get("max_tokens", 1024)

        provider = self._get_provider(provider_name)
        if provider is None:
            return LLMResponse(
                text="",
                tokens_used=0,
                model_used="",
                duration_ms=0,
                error=f"LLM provider '{provider_name}' not configured on this worker",
            )

        return await provider.generate(
            system_prompt=task_req.system_prompt,
            user_message=task_req.user_message,
            model=model,
            temperature=temperature,
            max_tokens=max_tokens,
        )

    async def _heartbeat_loop(self, stub, metadata):
        """Periodically send heartbeat to the server."""
        while True:
            await asyncio.sleep(self.config.heartbeat_interval)
            try:
                mem_mb = 0
                if tracemalloc.is_started():
                    current, _ = tracemalloc.get_traced_memory()
                    mem_mb = current // (1024 * 1024)

                await stub.Heartbeat(
                    worker_pb2.HeartbeatRequest(
                        worker_id=self.config.worker_id,
                        active_tasks=self.config.max_concurrent - self.semaphore._value,
                        memory_usage_mb=mem_mb,
                    ),
                    metadata=metadata,
                )
            except Exception as e:
                logger.warning("Heartbeat failed: %s", e)
