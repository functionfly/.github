"""Type definitions for the FunctionFly vault SDK.

These mirror the JSON shapes returned by the vault API. The SDK
exposes them as plain dicts at runtime — the type aliases document
the expected shape and serve as a useful reference.
"""

from __future__ import annotations

from enum import Enum
from typing import Any, Dict, List, Optional


class SecretType(str, Enum):
    API_KEY = "api_key"
    OAUTH_TOKEN = "oauth_token"
    PASSWORD = "password"
    CERTIFICATE = "certificate"


class DynamicDBType(str, Enum):
    POSTGRES = "postgres"
    MYSQL = "mysql"


EncryptedData = Dict[str, Any]
"""A dict with keys: ciphertext, iv, salt, tag, key_version (all base64
strings except key_version, which is an int)."""

Secret = Dict[str, Any]
SecretCreate = Dict[str, Any]
SecretList = Dict[str, Any]
SecretListOptions = Dict[str, Any]
SecretUpdate = Dict[str, Any]
SecretRotate = Dict[str, Any]

Token = Dict[str, Any]
TokenCreate = Dict[str, Any]
TokenList = Dict[str, Any]

DynamicTarget = Dict[str, Any]
DynamicTargetCreate = Dict[str, Any]
DynamicCredential = Dict[str, Any]
DynamicCredentialCreate = Dict[str, Any]
GeneratedCredential = Dict[str, Any]
GenerateOptions = Dict[str, Any]
RenewOptions = Dict[str, Any]

AuditEntry = Dict[str, Any]
AuditList = Dict[str, Any]
AuditListOptions = Dict[str, Any]
