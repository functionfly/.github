"""FunctionFly vault Python SDK.

Quick start::

    from functionfly_vault import VaultClient, SecretType

    client = VaultClient(token="fnly_xxx", base_url="https://api.functionfly.com")

    # Caller performs encryption locally (zero-knowledge).
    ct, iv, salt, tag = encrypt_my_value("super-secret")

    secret = client.secrets.create(
        name="STRIPE_API_KEY",
        secret_type=SecretType.API_KEY,
        encrypted_data={
            "ciphertext": ct,
            "iv": iv,
            "salt": salt,
            "tag": tag,
            "key_version": 2,  # 1=PBKDF2, 2=Argon2id
        },
    )
"""

from .client import VaultClient
from .errors import VaultAPIError
from .types import (
    SecretType,
    DynamicDBType,
    EncryptedData,
    Secret,
    SecretCreate,
    SecretList,
    SecretListOptions,
    SecretUpdate,
    SecretRotate,
    Token,
    TokenCreate,
    TokenList,
    DynamicTarget,
    DynamicTargetCreate,
    DynamicCredential,
    DynamicCredentialCreate,
    GeneratedCredential,
    GenerateOptions,
    RenewOptions,
    AuditEntry,
    AuditList,
    AuditListOptions,
)

__version__ = "0.1.0"

__all__ = [
    "VaultClient",
    "VaultAPIError",
    "SecretType",
    "DynamicDBType",
    "EncryptedData",
    "Secret",
    "SecretCreate",
    "SecretList",
    "SecretListOptions",
    "SecretUpdate",
    "SecretRotate",
    "Token",
    "TokenCreate",
    "TokenList",
    "DynamicTarget",
    "DynamicTargetCreate",
    "DynamicCredential",
    "DynamicCredentialCreate",
    "GeneratedCredential",
    "GenerateOptions",
    "RenewOptions",
    "AuditEntry",
    "AuditList",
    "AuditListOptions",
    "__version__",
]
