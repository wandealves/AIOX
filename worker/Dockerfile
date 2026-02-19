FROM python:3.12-slim

WORKDIR /app

# Install dependencies
COPY worker/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy proto and generate Python code
COPY proto/worker/v1/worker.proto /tmp/worker.proto
RUN python -m grpc_tools.protoc \
    -I/tmp \
    --python_out=/app/worker \
    --grpc_python_out=/app/worker \
    /tmp/worker.proto

# Copy worker source
COPY worker/worker/ /app/worker/

# Fix generated import to be relative
RUN sed -i 's/import worker_pb2/from . import worker_pb2/' /app/worker/worker_pb2_grpc.py

CMD ["python", "-m", "worker.main"]
