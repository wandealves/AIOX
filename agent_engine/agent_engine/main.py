import asyncio
import json
import logging
import tracemalloc

from .config import Config
from .client import WorkerClient


class JSONFormatter(logging.Formatter):
    """Structured JSON log formatter."""

    def format(self, record: logging.LogRecord) -> str:
        entry = {
            "timestamp": self.formatTime(record, self.datefmt),
            "level": record.levelname,
            "logger": record.name,
            "message": record.getMessage(),
        }
        # Propagate extra fields (e.g. request_id, agent_id)
        for key in ("request_id", "agent_id"):
            val = getattr(record, key, None)
            if val:
                entry[key] = val
        if record.exc_info and record.exc_info[0]:
            entry["exception"] = self.formatException(record.exc_info)
        return json.dumps(entry, default=str)


def main():
    config = Config()

    if config.log_json:
        handler = logging.StreamHandler()
        handler.setFormatter(JSONFormatter())
        logging.root.handlers = [handler]
        logging.root.setLevel(logging.INFO)
    else:
        logging.basicConfig(
            level=logging.INFO,
            format="%(asctime)s %(levelname)s [%(name)s] %(message)s",
        )

    tracemalloc.start()

    logger = logging.getLogger(__name__)
    logger.info("Starting AIOX worker: %s", config.worker_id)
    logger.info("gRPC target: %s", config.grpc_target)
    logger.info("Supported providers: %s", config.supported_providers)
    logger.info("Max concurrent tasks: %d", config.max_concurrent)

    client = WorkerClient(config)
    asyncio.run(client.run())


if __name__ == "__main__":
    main()
