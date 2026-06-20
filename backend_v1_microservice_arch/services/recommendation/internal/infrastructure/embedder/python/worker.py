#!/usr/bin/env python3
"""Recommendation embedder worker.

Long-lived subprocess managed by the Go recommendation service. Loads a
sentence-transformers model once at startup and serves newline-delimited JSON
embedding requests over a Unix domain socket.

Protocol (one request/response per line, UTF-8):
    request : {"texts": ["hello", "world"]}
    success : {"embeddings": [[...], [...]]}
    failure : {"error": "message"}

Startup handshake: prints "READY" to stdout once the model is loaded and the
socket is accepting connections. The Go client blocks on that line.

Environment:
    RECOMMENDATION_EMBEDDER_SOCKET  required, Unix socket path
    RECOMMENDATION_EMBEDDER_MODEL   optional, model name (default all-MiniLM-L6-v2)
"""

from __future__ import annotations

import json
import os
import socket
import sys
import traceback

DEFAULT_MODEL = "sentence-transformers/all-MiniLM-L6-v2"


def log(msg: str) -> None:
    print(msg, file=sys.stderr, flush=True)


def main() -> int:
    socket_path = os.environ.get("RECOMMENDATION_EMBEDDER_SOCKET", "").strip()
    if not socket_path:
        log("RECOMMENDATION_EMBEDDER_SOCKET is required")
        return 2

    model_name = os.environ.get("RECOMMENDATION_EMBEDDER_MODEL", DEFAULT_MODEL).strip() or DEFAULT_MODEL

    try:
        from sentence_transformers import SentenceTransformer
    except ImportError as exc:
        log(f"import sentence-transformers failed: {exc}")
        return 2

    try:
        model = SentenceTransformer(model_name)
    except Exception as exc:  # noqa: BLE001
        log(f"load model {model_name!r} failed: {exc}")
        return 2

    try:
        os.unlink(socket_path)
    except FileNotFoundError:
        pass

    srv = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    srv.bind(socket_path)
    srv.listen(1)

    print("READY", flush=True)

    try:
        while True:
            conn, _ = srv.accept()
            try:
                serve(conn, model)
            except Exception:  # noqa: BLE001
                log("connection handler crashed:\n" + traceback.format_exc())
            finally:
                conn.close()
    finally:
        srv.close()
        try:
            os.unlink(socket_path)
        except FileNotFoundError:
            pass

    return 0


def serve(conn: socket.socket, model) -> None:
    rfile = conn.makefile("rb")
    wfile = conn.makefile("wb")

    for raw in rfile:
        resp = handle(raw, model)
        wfile.write((json.dumps(resp) + "\n").encode("utf-8"))
        wfile.flush()


def handle(raw: bytes, model) -> dict:
    try:
        req = json.loads(raw)
        texts = req.get("texts") or []
        if not isinstance(texts, list):
            return {"error": "texts must be a list"}
        if not texts:
            return {"embeddings": []}
        vectors = model.encode(texts, normalize_embeddings=True, convert_to_numpy=True)
        return {"embeddings": vectors.tolist()}
    except Exception as exc:  # noqa: BLE001
        return {"error": f"{type(exc).__name__}: {exc}"}


if __name__ == "__main__":
    sys.exit(main())
