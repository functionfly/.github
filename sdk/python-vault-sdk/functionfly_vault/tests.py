"""Tests for the FunctionFly vault Python SDK.

Run with::

    python -m unittest discover functionfly_vault

or::

    python -m unittest functionfly_vault.tests
"""

from __future__ import annotations

import json
import unittest
from http.server import BaseHTTPRequestHandler, HTTPServer
from threading import Thread
from typing import Any, Dict, Tuple

from functionfly_vault import VaultAPIError
from functionfly_vault.client import VaultClient


class _StubHandler(BaseHTTPRequestHandler):
    """A minimal HTTP handler that captures the last request and returns
    a pre-programmed response."""

    captured: Dict[str, Any] = {}
    response: Tuple[int, bytes, str] = (200, b"{}", "application/json")

    def log_message(self, *_args: Any) -> None:  # silence stderr noise
        return

    def do_GET(self) -> None:  # noqa: N802
        type(self).captured = {"method": "GET", "path": self.path, "headers": dict(self.headers)}
        self._respond()

    def do_POST(self) -> None:  # noqa: N802
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length) if length else b""
        type(self).captured = {
            "method": "POST",
            "path": self.path,
            "headers": dict(self.headers),
            "body": body,
        }
        self._respond()

    def do_PATCH(self) -> None:  # noqa: N802
        length = int(self.headers.get("Content-Length", "0"))
        body = self.rfile.read(length) if length else b""
        type(self).captured = {
            "method": "PATCH",
            "path": self.path,
            "headers": dict(self.headers),
            "body": body,
        }
        self._respond()

    def do_DELETE(self) -> None:  # noqa: N802
        type(self).captured = {"method": "DELETE", "path": self.path, "headers": dict(self.headers)}
        self._respond()

    def _respond(self) -> None:
        status, body, ctype = self.response
        self.send_response(status)
        self.send_header("Content-Type", ctype)
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)


class _Server:
    """Context manager that starts a local HTTP server for the duration
    of a test and tears it down afterwards."""

    def __enter__(self) -> Tuple[HTTPServer, _StubHandler]:
        self._server = HTTPServer(("127.0.0.1", 0), _StubHandler)
        self._thread = Thread(target=self._server.serve_forever, daemon=True)
        self._thread.start()
        return self._server, _StubHandler

    def __exit__(self, *_args: Any) -> None:
        self._server.shutdown()
        self._server.server_close()
        self._thread.join(timeout=2)


class ClientTestCase(unittest.TestCase):
    def setUp(self) -> None:
        _StubHandler.captured = {}
        _StubHandler.response = (200, b"{}", "application/json")

    def test_token_is_required(self) -> None:
        with self.assertRaises(ValueError):
            VaultClient(token="")  # type: ignore[arg-type]

    def test_sends_bearer_token(self) -> None:
        with _Server() as (_server, handler):
            client = VaultClient(token="fnly_xxx", base_url=f"http://{_server.server_address[0]}:{_server.server_address[1]}")
            client.request("GET", "/v1/vault/secrets")
            self.assertEqual(handler.captured["headers"]["Authorization"], "Bearer fnly_xxx")
            self.assertEqual(handler.captured["method"], "GET")
            self.assertEqual(handler.captured["path"], "/v1/vault/secrets")

    def test_posts_json_body(self) -> None:
        with _Server() as (_server, handler):
            client = VaultClient(token="fnly_xxx", base_url=f"http://{_server.server_address[0]}:{_server.server_address[1]}")
            client.request("POST", "/v1/vault/secrets", body={"name": "X"})
            self.assertEqual(handler.captured["method"], "POST")
            self.assertEqual(handler.captured["headers"]["Content-Type"], "application/json")
            self.assertEqual(json.loads(handler.captured["body"]), {"name": "X"})

    def test_handles_api_error(self) -> None:
        with _Server() as (_server, handler):
            handler.response = (
                403,
                json.dumps({"error": "forbidden", "message": "nope", "code": "FORBIDDEN"}).encode(),
                "application/json",
            )
            client = VaultClient(token="fnly_xxx", base_url=f"http://{_server.server_address[0]}:{_server.server_address[1]}")
            with self.assertRaises(VaultAPIError) as ctx:
                client.request("GET", "/v1/vault/secrets")
            self.assertEqual(ctx.exception.status, 403)
            self.assertEqual(ctx.exception.code, "FORBIDDEN")
            self.assertEqual(ctx.exception.message, "nope")

    def test_quoting(self) -> None:
        with _Server() as (_server, handler):
            client = VaultClient(token="t", base_url=f"http://{_server.server_address[0]}:{_server.server_address[1]}")
            client.secrets.get("a/b c")
            self.assertEqual(handler.captured["path"], "/v1/vault/secrets/a%2Fb%20c")


if __name__ == "__main__":
    unittest.main()
