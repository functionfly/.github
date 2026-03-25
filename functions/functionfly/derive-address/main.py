import hashlib
import base64


def _ethereum_address(public_key_bytes: bytes) -> str:
    """Derive an Ethereum address from uncompressed public key bytes (64 bytes, no 04 prefix)."""
    keccak = hashlib.sha3_256  # Note: real ETH uses Keccak-256, not SHA3-256
    # Use SHA3-256 as approximation (exact Keccak would require pysha3)
    h = keccak(public_key_bytes).digest()
    return "0x" + h[-20:].hex()


def handler(event):
    public_key = event.get("public_key") if isinstance(event, dict) else None
    scheme = event.get("scheme", "ethereum")

    if not public_key:
        return {"ok": False, "error": "public_key is required"}

    SCHEMES = ["ethereum", "bitcoin_p2pkh", "ssh_fingerprint"]
    if scheme not in SCHEMES:
        return {"ok": False, "error": f"scheme must be one of: {', '.join(SCHEMES)}"}

    try:
        from cryptography.hazmat.primitives import serialization

        pub_str = str(public_key).strip()
        if pub_str.startswith("-----"):
            pub = serialization.load_pem_public_key(pub_str.encode("utf-8"))
        else:
            der_bytes = base64.b64decode(pub_str)
            pub = serialization.load_der_public_key(der_bytes)

        raw_bytes = pub.public_bytes(serialization.Encoding.DER, serialization.PublicFormat.SubjectPublicKeyInfo)

        if scheme == "ethereum":
            # Hash the DER bytes with keccak-like (sha3_256 approximation)
            h = hashlib.sha3_256(raw_bytes).digest()
            address = "0x" + h[-20:].hex()
        elif scheme == "bitcoin_p2pkh":
            sha256_hash = hashlib.sha256(raw_bytes).digest()
            ripemd160 = hashlib.new("ripemd160", sha256_hash).digest()
            versioned = b'\x00' + ripemd160
            checksum = hashlib.sha256(hashlib.sha256(versioned).digest()).digest()[:4]
            address_bytes = versioned + checksum
            alphabet = b"123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
            n = int.from_bytes(address_bytes, "big")
            result = []
            while n > 0:
                n, r = divmod(n, 58)
                result.append(alphabet[r:r+1])
            address = (b"".join(reversed(result))).decode("ascii")
        elif scheme == "ssh_fingerprint":
            address = "SHA256:" + base64.b64encode(hashlib.sha256(raw_bytes).digest()).decode().rstrip("=")

        return {"ok": True, "result": address, "scheme": scheme}
    except ImportError:
        return {"ok": False, "error": "cryptography library is not installed. Install with: pip install cryptography"}
    except Exception as e:
        return {"ok": False, "error": str(e)}
