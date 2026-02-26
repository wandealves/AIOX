FROM python:3.12-slim

WORKDIR /app

# Install dependencies
COPY agent/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy proto and generate Python code
COPY proto/worker/v1/worker.proto /tmp/worker.proto
RUN mkdir -p /app/agent
RUN python -m grpc_tools.protoc \
    -I/tmp \
    --python_out=/app/agent \
    --grpc_python_out=/app/agent \
    /tmp/worker.proto

# Pre-download embedding model at build time
RUN python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('sentence-transformers/all-MiniLM-L6-v2')"

# Copy agent source
COPY agent/agent/ /app/agent/

# Fix generated import to be relative
RUN sed -i 's/^import worker_pb2/from . import worker_pb2/' /app/agent/worker_pb2_grpc.py

CMD ["python", "-m", "agent.main"]
