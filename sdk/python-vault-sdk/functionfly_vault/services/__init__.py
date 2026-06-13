"""Sub-services that hang off :class:`VaultClient`.

Each service corresponds to one slice of the vault API:

* :class:`SecretsService`              — secret CRUD
* :class:`TokensService`               — runtime access tokens
* :class:`DynamicTargetsService`       — DB target config
* :class:`DynamicCredentialsService`   — credential templates + generation
* :class:`LeasesService`               — lease renew / revoke
* :class:`AuditService`                — audit log query
"""

from .secrets import SecretsService
from .tokens import TokensService
from .dynamic import (
    DynamicCredentialsService,
    DynamicTargetsService,
    LeasesService,
)
from .audit import AuditService

__all__ = [
    "SecretsService",
    "TokensService",
    "DynamicCredentialsService",
    "DynamicTargetsService",
    "LeasesService",
    "AuditService",
]
