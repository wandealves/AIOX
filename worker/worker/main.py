import asyncio
import logging
import tracemalloc

from .config import Config
from .client import WorkerClient


def main():
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s [%(name)s] %(message)s",
    )
    tracemalloc.start()

    config = Config()
    logger = logging.getLogger(__name__)
    logger.info("Starting AIOX worker: %s", config.worker_id)
    logger.info("gRPC target: %s", config.grpc_target)
    logger.info("Supported providers: %s", config.supported_providers)
    logger.info("Max concurrent tasks: %d", config.max_concurrent)

    client = WorkerClient(config)
    asyncio.run(client.run())


if __name__ == "__main__":
    main()
