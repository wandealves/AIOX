import asyncio
import json
import logging
import tracemalloc

import grpc
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

# Import generated protobuf modules
from . import worker_pb2, worker_pb2_grpc
from .budget import BudgetTracker, ExecutionBudget
from .config import Config
from .embedding import EmbeddingService
from .errors import AgentEngineError, BudgetExceededError
from .memory import MemoryConfig, MemoryContext
from .providers import (
    AnthropicProvider,
    DeepSeekProvider,
    GeminiProvider,
    LLMProvider,
    LLMResponse,
    OllamaProvider,
    OpenAIProvider,
)
from .retry import RetryConfig, retry_with_backoff
from .tools import ToolExecutor, proto_tools_to_openai_format

logger = logging.getLogger(__name__)


def _setup_tracing(config: Config):
    """Initialize OpenTelemetry tracing if enabled."""
    if not config.tracing_enabled:
        return
    try:
        resource = Resource.create({"service.name": f"aiox-agent-{config.worker_id}"})
        exporter = OTLPSpanExporter(
            endpoint=config.tracing_otlp_endpoint, insecure=True
        )
        provider = TracerProvider(resource=resource)
        provider.add_span_processor(BatchSpanProcessor(exporter))
        trace.set_tracer_provider(provider)
        logger.info("Tracing initialized: endpoint=%s", config.tracing_otlp_endpoint)
    except ImportError:
        logger.warning("OpenTelemetry packages not installed, tracing disabled")


class WorkerClient:
    """gRPC client that connects to the AIOX server, receives tasks, and returns results."""

    def __init__(self, config: Config):
        self.config = config
        self.providers: dict[str, LLMProvider] = {}
        self._setup_providers()
        self.semaphore = asyncio.Semaphore(config.max_concurrent)
        self.embedding_svc = EmbeddingService()
        _setup_tracing(config)
        self._tracer = None
        try:
            from opentelemetry import trace

            self._tracer = trace.get_tracer("aiox-agent")
        except ImportError:
            pass

        self._retry_config = RetryConfig(
            max_retries=config.retry_max_retries,
            base_delay_sec=config.retry_base_delay_sec,
            max_delay_sec=config.retry_max_delay_sec,
        )

    def _setup_providers(self):
        if self.config.openai_api_key:
            self.providers["openai"] = OpenAIProvider(self.config.openai_api_key)
            logger.info("OpenAI provider configured")
        if self.config.anthropic_api_key:
            self.providers["anthropic"] = AnthropicProvider(
                self.config.anthropic_api_key
            )
            logger.info("Anthropic provider configured")
        if self.config.deepseek_api_key:
            self.providers["deepseek"] = DeepSeekProvider(self.config.deepseek_api_key)
            logger.info("DeepSeek provider configured")
        if self.config.gemini_api_key:
            self.providers["gemini"] = GeminiProvider(self.config.gemini_api_key)
            logger.info("Gemini provider configured")
        self.providers["ollama"] = OllamaProvider(
            model=self.config.ollama_default_model,
            base_url=self.config.ollama_base_url,
        )
        logger.info(
            "Ollama provider configured (model=%s)", self.config.ollama_default_model
        )

    def _get_provider(self, provider_name: str) -> LLMProvider | None:
        return self.providers.get(provider_name)

    def _make_budget_tracker(self) -> BudgetTracker:
        """Create a BudgetTracker from the current config."""
        return BudgetTracker(
            ExecutionBudget(
                max_steps=self.config.budget_max_steps,
                max_tokens=self.config.budget_max_tokens,
                max_time_sec=self.config.budget_max_time_sec,
                max_cost_usd=self.config.budget_max_cost_usd,
            )
        )

    async def run(self):
        """Main loop: connect, register, process tasks. Reconnects on failure."""
        while True:
            try:
                await self._connect_and_process()
            except Exception as e:
                logger.error("Connection error: %s", e)

            logger.info("Reconnecting in %d seconds...", self.config.reconnect_delay)
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

            # Wait for ack — use stream.read() consistently.
            # NOTE: mixing stream.read() with "async for stream" on the same
            # RPC raises "iterator and read/write APIs may not be mixed".
            # We therefore use stream.read() throughout.
            server_msg = await stream.read()
            if server_msg == grpc.aio.EOF:
                logger.error("Server closed stream before sending RegisterAck")
                return

            ack = server_msg.register_ack
            if not ack.accepted:
                logger.error("Registration rejected: %s", ack.message)
                return

            logger.info("Registered as %s", self.config.worker_id)

            # Start heartbeat task
            heartbeat_task = asyncio.create_task(self._heartbeat_loop(stub, metadata))

            # Process incoming tasks using stream.read() (not async for)
            try:
                while True:
                    server_msg = await stream.read()
                    if server_msg == grpc.aio.EOF:
                        logger.info("Server closed stream (EOF)")
                        break
                    task_req = server_msg.task_request
                    if task_req and task_req.request_id:
                        asyncio.create_task(self._process_task(stream, task_req))
            finally:
                heartbeat_task.cancel()
                try:
                    await heartbeat_task
                except asyncio.CancelledError:
                    pass

        finally:
            await channel.close()

    async def _process_task(self, stream, task_req):
        """Process a single task with concurrency limiting and memory support."""
        async with self.semaphore:
            # Start a tracing span if available
            span = None
            if self._tracer:
                span = self._tracer.start_span(
                    "worker.process_task",
                    attributes={
                        "request_id": task_req.request_id,
                        "agent_id": task_req.agent_id,
                        "worker_id": self.config.worker_id,
                    },
                )

            logger.info(
                "Processing task %s for agent %s",
                task_req.request_id,
                task_req.agent_id,
            )

            # Parse memory context and config
            mem_config = MemoryConfig.from_json(task_req.memory_config_json)
            mem_context = MemoryContext.from_json(task_req.memory_context_json)

            # Build messages array with memory context if enabled
            messages = None
            if mem_config.enabled and (
                mem_context.recent_messages or mem_context.relevant_memories
            ):
                messages = mem_context.build_messages_for_llm(
                    task_req.system_prompt, task_req.user_message
                )

            # Convert attachments to multimodal content parts
            attachment_parts = []
            if task_req.attachments:
                for att in task_req.attachments:
                    if att.content_type.startswith("image/"):
                        import base64

                        b64 = base64.b64encode(att.content).decode("utf-8")
                        attachment_parts.append(
                            {
                                "type": "image",
                                "filename": att.filename,
                                "content_type": att.content_type,
                                "data_b64": b64,
                            }
                        )
                    else:
                        try:
                            text_content = att.content.decode("utf-8")
                        except UnicodeDecodeError:
                            text_content = f"[Binary file: {att.filename}]"
                        attachment_parts.append(
                            {
                                "type": "text",
                                "filename": att.filename,
                                "content": text_content,
                            }
                        )

            # Convert proto tools to LLM-specific format
            tools_for_llm = None
            tool_executor = None
            if task_req.tools:
                # Build tool context from proto ToolContext
                tool_ctx = {}
                if task_req.tool_context:
                    tc = task_req.tool_context
                    tool_ctx = {
                        "agent_id": tc.agent_id,
                        "user_id": tc.user_id,
                        "session_id": tc.session_id,
                        "correlation_id": tc.correlation_id,
                    }
                    if tc.permissions_json:
                        try:
                            tool_ctx["permissions"] = json.loads(tc.permissions_json)
                        except json.JSONDecodeError:
                            pass
                    if tc.metadata_json:
                        try:
                            tool_ctx["metadata"] = json.loads(tc.metadata_json)
                        except json.JSONDecodeError:
                            pass

                tool_executor = ToolExecutor(tool_context=tool_ctx)
                # All providers receive OpenAI-formatted tools with private execution
                # metadata (_endpoint_url, _http_method, etc.). Each provider converts
                # to its own API format internally before calling the LLM.
                tools_for_llm = proto_tools_to_openai_format(task_req.tools)

            # Create per-task budget tracker
            budget_tracker = self._make_budget_tracker()

            # Sanitize input (Phase 2 integration)
            user_message = task_req.user_message
            try:
                from .sanitize import (
                    MAX_USER_MESSAGE_LEN,
                    check_length,
                    detect_injection,
                )

                warning = check_length(
                    user_message, MAX_USER_MESSAGE_LEN, "user_message"
                )
                if warning:
                    logger.warning("Input validation: %s", warning)
                    user_message = user_message[:MAX_USER_MESSAGE_LEN]

                injection_patterns = detect_injection(user_message)
                if injection_patterns:
                    logger.warning(
                        "Potential injection detected in task %s: %s",
                        task_req.request_id,
                        injection_patterns,
                    )
            except ImportError:
                pass  # sanitize module not yet available

            # Policy engine (Phase 2 integration)
            policy = None
            try:
                from .policy import Policy

                if task_req.tool_context and task_req.tool_context.permissions_json:
                    try:
                        perms = json.loads(task_req.tool_context.permissions_json)
                        policy = Policy.from_permissions(perms)
                    except (json.JSONDecodeError, Exception):
                        pass
                if tool_executor and policy:
                    tool_executor.policy = policy
            except ImportError:
                pass  # policy module not yet available

            # Planner (Phase 3 integration)
            try:
                from .planner import ActionType, Planner

                if self.config.planner_enabled and tools_for_llm:
                    llm_config = (
                        json.loads(task_req.llm_config_json)
                        if task_req.llm_config_json
                        else {}
                    )
                    provider_name = llm_config.get("provider", "openai")
                    planner_provider = self._get_provider(provider_name)
                    if planner_provider:
                        planner = Planner(
                            provider=planner_provider,
                            planning_model=self.config.planner_model,
                        )
                        tool_names = [
                            t.get("function", {}).get("name", "") for t in tools_for_llm
                        ]
                        plan_result = await planner.plan(
                            user_message=user_message,
                            system_prompt=task_req.system_prompt,
                            available_tools=tool_names,
                            budget_tracker=budget_tracker,
                        )
                        if plan_result.action == ActionType.RESPOND:
                            tools_for_llm = None
                            tool_executor = None
                        elif plan_result.action == ActionType.CLARIFY:
                            # Return clarification directly
                            result_msg = worker_pb2.WorkerMessage(
                                task_response=worker_pb2.TaskResponse(
                                    request_id=task_req.request_id,
                                    worker_id=self.config.worker_id,
                                    response_text=plan_result.clarification
                                    or plan_result.reasoning,
                                    tokens_used=0,
                                    duration_ms=0,
                                    model_used="",
                                )
                            )
                            await stream.write(result_msg)
                            if span:
                                span.end()
                            return
            except ImportError:
                pass  # planner module not yet available

            response = await self._call_llm(
                task_req,
                messages=messages,
                tools=tools_for_llm,
                tool_executor=tool_executor,
                attachments=attachment_parts if attachment_parts else None,
                budget_tracker=budget_tracker,
            )

            # Convert tools_called from LLM response to proto ToolCall messages
            proto_tools_called = []
            for tc in response.tools_called:
                proto_tools_called.append(
                    worker_pb2.ToolCall(
                        tool_name=tc.tool_name,
                        arguments_json=tc.arguments_json,
                        result_json=tc.result_json,
                        duration_ms=tc.duration_ms,
                        error=tc.error,
                    )
                )

            # Generate embedding for user message if long-term memory is enabled
            new_memories = []
            if (
                mem_config.enabled
                and mem_config.long_term_enabled
                and not response.error
            ):
                try:
                    embedding = self.embedding_svc.embed(task_req.user_message)
                    new_memories.append(
                        worker_pb2.MemoryEntry(
                            content=task_req.user_message,
                            embedding=embedding,
                            memory_type="conversation",
                            metadata_json=json.dumps(
                                {
                                    "source": "user_message",
                                    "request_id": task_req.request_id,
                                    "agent_id": task_req.agent_id,
                                }
                            ),
                        )
                    )
                except Exception as e:
                    logger.warning("Failed to generate embedding: %s", e)

            result_msg = worker_pb2.WorkerMessage(
                task_response=worker_pb2.TaskResponse(
                    request_id=task_req.request_id,
                    worker_id=self.config.worker_id,
                    response_text=response.text,
                    tokens_used=response.tokens_used,
                    duration_ms=response.duration_ms,
                    model_used=response.model_used,
                    error_message=response.error,
                    new_memories=new_memories,
                    tools_called=proto_tools_called,
                )
            )
            await stream.write(result_msg)

            if span:
                span.set_attribute("tokens_used", response.tokens_used)
                span.set_attribute("duration_ms", response.duration_ms)
                span.set_attribute("model_used", response.model_used)
                if response.error:
                    span.set_attribute("error", True)
                    span.set_attribute("error.message", response.error)
                span.end()

            logger.info(
                "Task %s completed: %d tokens, %dms, %d new memories, %d tool calls",
                task_req.request_id,
                response.tokens_used,
                response.duration_ms,
                len(new_memories),
                len(proto_tools_called),
            )

    async def _call_llm(
        self,
        task_req,
        messages: list[dict] | None = None,
        tools: list[dict] | None = None,
        tool_executor: ToolExecutor | None = None,
        attachments: list[dict] | None = None,
        budget_tracker: BudgetTracker | None = None,
    ) -> LLMResponse:
        """Call the appropriate LLM provider based on agent's llm_config."""
        try:
            llm_config = (
                json.loads(task_req.llm_config_json) if task_req.llm_config_json else {}
            )
        except json.JSONDecodeError:
            llm_config = {}

        provider_name = llm_config.get("provider", "openai")
        model = llm_config.get("model", "")
        temperature = llm_config.get("temperature", 0.7)
        max_tokens = llm_config.get("max_tokens", 1024)

        # Model router (Phase 4 integration)
        try:
            from .model_router import route_model

            if self.config.model_router_enabled and not model:
                model = route_model(
                    provider_name,
                    task_req.user_message,
                    has_tools=bool(tools),
                    num_tools=len(tools) if tools else 0,
                    has_memory=messages is not None,
                )
        except ImportError:
            pass

        provider = self._get_provider(provider_name)
        if provider is None:
            return LLMResponse(
                text="",
                tokens_used=0,
                model_used="",
                duration_ms=0,
                error=f"LLM provider '{provider_name}' not configured on this worker",
            )

        # Wrap provider call with retry
        async def _do_generate():
            if budget_tracker:
                budget_tracker.record_step()
                budget_tracker.check_all()
            return await provider.generate(
                system_prompt=task_req.system_prompt,
                user_message=task_req.user_message,
                model=model,
                temperature=temperature,
                max_tokens=max_tokens,
                messages=messages,
                tools=tools,
                tool_executor=tool_executor,
                attachments=attachments,
            )

        try:
            response = await retry_with_backoff(
                _do_generate,
                config=self._retry_config,
            )
            # Track tokens in budget
            if budget_tracker and response.tokens_used:
                budget_tracker.record_tokens(response.tokens_used, response.model_used)
            return response
        except BudgetExceededError as e:
            logger.warning("Budget exceeded for task %s: %s", task_req.request_id, e)
            return LLMResponse(
                text="",
                tokens_used=budget_tracker.tokens if budget_tracker else 0,
                model_used=model,
                duration_ms=int(budget_tracker.elapsed_sec * 1000)
                if budget_tracker
                else 0,
                error=str(e),
                error_type="budget_exceeded",
            )
        except AgentEngineError as e:
            logger.error("Provider error for task %s: %s", task_req.request_id, e)

            # Fallback chain (Phase 4 integration)
            try:
                from .fallback import FallbackChain

                if self.config.fallback_providers:
                    fallback_chain = []
                    for fb_name in self.config.fallback_providers:
                        if fb_name != provider_name:
                            fb_provider = self._get_provider(fb_name)
                            if fb_provider:
                                fallback_chain.append((fb_name, fb_provider))
                    if fallback_chain:
                        chain = FallbackChain(fallback_chain)
                        logger.info(
                            "Trying fallback chain: %s",
                            [n for n, _ in fallback_chain],
                        )
                        return await chain.generate(
                            system_prompt=task_req.system_prompt,
                            user_message=task_req.user_message,
                            model="",
                            temperature=temperature,
                            max_tokens=max_tokens,
                            messages=messages,
                            tools=tools,
                            tool_executor=tool_executor,
                            attachments=attachments,
                        )
            except ImportError:
                pass

            return LLMResponse(
                text="",
                tokens_used=0,
                model_used=model,
                duration_ms=0,
                error=str(e),
                error_type=type(e).__name__,
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
