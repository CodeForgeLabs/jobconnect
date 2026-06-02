# Recommendation service image with an embedded Python worker for semantic
# embeddings. Use this in compose when RECOMMENDATION_EMBEDDER_BACKEND=python.
# For deployments that only need the token-cosine ranker, the generic
# services/Dockerfile produces a much smaller image.
#
# Build from the backend - deprecated/ root:
#   docker build -f services/recommendation/Dockerfile -t recommendation .

FROM golang:1.26.1-bookworm AS builder
WORKDIR /src
COPY . .
WORKDIR /src/services/recommendation
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/service ./cmd/recommendationd

FROM python:3.12-slim
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1 \
    PIP_NO_CACHE_DIR=1 \
    HF_HUB_DISABLE_TELEMETRY=1 \
    TRANSFORMERS_OFFLINE=0

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install CPU-only PyTorch first (GPU wheels are ~2GB, CPU wheel is ~200MB).
RUN pip install --extra-index-url https://download.pytorch.org/whl/cpu \
    "torch==2.4.*" \
    "sentence-transformers==3.*"

COPY services/recommendation/internal/infrastructure/embedder/python/worker.py /app/embedder/worker.py

ARG EMBEDDER_MODEL=sentence-transformers/all-MiniLM-L6-v2
ENV RECOMMENDATION_EMBEDDER_MODEL=${EMBEDDER_MODEL}

# Pre-download model weights so container start is offline and fast.
RUN python -c "from sentence_transformers import SentenceTransformer; SentenceTransformer('${EMBEDDER_MODEL}')"

COPY --from=builder /out/service /app/service

ENV RECOMMENDATION_EMBEDDER_BACKEND=python \
    RECOMMENDATION_EMBEDDER_PYTHON=python3 \
    RECOMMENDATION_EMBEDDER_WORKER_SCRIPT=/app/embedder/worker.py \
    RECOMMENDATION_EMBEDDER_SOCKET=/tmp/recommendation-embedder.sock

EXPOSE 50064 50164
ENTRYPOINT ["/app/service"]
