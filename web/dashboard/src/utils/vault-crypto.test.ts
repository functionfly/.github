import { describe, expect, it, vi, beforeEach } from "vitest";
import type { EncryptedData } from "@/utils/vault-crypto";

vi.mock("argon2-browser", () => ({
  argon2: {
    ArgonType: {
      Argon2id: 2,
    },
    hash: vi.fn().mockResolvedValue({
      hash: new Uint8Array(32).buffer as ArrayBuffer,
    }),
  },
}));

const mockCryptoKey = {
  type: "secret",
  algorithm: { name: "AES-GCM" },
  extractable: false,
  usages: ["encrypt", "decrypt"],
} as unknown as CryptoKey;

const mockImportKey = vi.fn().mockResolvedValue(mockCryptoKey);
const mockDeriveKey = vi.fn().mockResolvedValue(mockCryptoKey);
const mockEncrypt = vi.fn().mockResolvedValue(new ArrayBuffer(100));
const mockDecrypt = vi.fn().mockResolvedValue(new TextEncoder().encode("decrypted").buffer);

vi.stubGlobal("crypto", {
  subtle: {
    importKey: mockImportKey,
    deriveKey: mockDeriveKey,
    encrypt: mockEncrypt,
    decrypt: mockDecrypt,
  },
  getRandomValues: (arr: Uint8Array) => {
    for (let i = 0; i < arr.length; i++) {
      arr[i] = Math.floor(Math.random() * 256);
    }
    return arr;
  },
});

const { VaultCrypto, arrayBufferToBase64 } = await import("@/utils/vault-crypto");

describe("VaultCrypto", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("generateSalt / generateIV", () => {
    it("generates a 16-byte salt", () => {
      const salt = VaultCrypto.generateSalt();
      expect(salt).toBeInstanceOf(Uint8Array);
      expect(salt.byteLength).toBe(16);
    });

    it("generates a 12-byte IV", () => {
      const iv = VaultCrypto.generateIV();
      expect(iv).toBeInstanceOf(Uint8Array);
      expect(iv.byteLength).toBe(12);
    });

    it("generates unique values each call", () => {
      const salt1 = VaultCrypto.generateSalt();
      const salt2 = VaultCrypto.generateSalt();
      const iv1 = VaultCrypto.generateIV();
      const iv2 = VaultCrypto.generateIV();
      expect(salt1).not.toEqual(salt2);
      expect(iv1).not.toEqual(iv2);
    });
  });

  describe("deriveKey", () => {
    it("derives a CryptoKey with AES-GCM algorithm", async () => {
      const passphrase = "test-passphrase";
      const salt = VaultCrypto.generateSalt();
      const key = await VaultCrypto.deriveKey(passphrase, salt, 3);
      expect(key).toBeDefined();
      expect(key.algorithm.name).toBe("AES-GCM");
      expect(key.extractable).toBe(false);
    });

    it("uses Argon2id for keyVersion 3", async () => {
      const passphrase = "test-passphrase";
      const salt = VaultCrypto.generateSalt();
      const key = await VaultCrypto.deriveKey(passphrase, salt, 3);
      expect(key.usages).toContain("encrypt");
      expect(key.usages).toContain("decrypt");
    });

    it("uses PBKDF2 for keyVersion 1", async () => {
      const passphrase = "test-passphrase";
      const salt = VaultCrypto.generateSalt();
      const key = await VaultCrypto.deriveKey(passphrase, salt, 1);
      expect(key.usages).toContain("encrypt");
      expect(key.usages).toContain("decrypt");
    });

    it("uses PBKDF2 for keyVersion 2", async () => {
      const passphrase = "test-passphrase";
      const salt = VaultCrypto.generateSalt();
      const key = await VaultCrypto.deriveKey(passphrase, salt, 2);
      expect(key.usages).toContain("encrypt");
      expect(key.usages).toContain("decrypt");
    });

    it("same passphrase and salt produce the same key", async () => {
      const passphrase = "consistent-passphrase";
      const salt = VaultCrypto.generateSalt();
      const key1 = await VaultCrypto.deriveKey(passphrase, salt, 3);
      const key2 = await VaultCrypto.deriveKey(passphrase, salt, 3);
      expect(key1).toBeDefined();
      expect(key2).toBeDefined();
    });
  });

  describe("encrypt / decrypt", () => {
    it("encryption produces ciphertext different from plaintext", async () => {
      const passphrase = "encrypt-test-passphrase";
      const plaintext = "This is a secret message";
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase);
      expect(encrypted.ciphertext).toBeTruthy();
      expect(encrypted.ciphertext).not.toEqual(plaintext);
      expect(encrypted.iv).toBeTruthy();
      expect(encrypted.salt).toBeTruthy();
      expect(encrypted.tag).toBeTruthy();
      expect(encrypted.keyVersion).toBe(3);
    });

    it("decryption returns original plaintext", async () => {
      const passphrase = "decrypt-test-passphrase";
      const plaintext = "This is a secret message";
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase);
      const decrypted = await VaultCrypto.decryptWithPassphrase(encrypted, passphrase);
      expect(decrypted).toBe("decrypted");
    });

    it("wrong passphrase fails to decrypt", async () => {
      const passphrase = "correct-passphrase";
      const wrongPassphrase = "wrong-passphrase";
      const plaintext = "This is a secret message";
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase);
      mockDecrypt.mockRejectedValueOnce(new Error("decrypt failed"));
      await expect(
        VaultCrypto.decryptWithPassphrase(encrypted, wrongPassphrase)
      ).rejects.toThrow();
    });

    it("encrypt and decrypt work with aad", async () => {
      const passphrase = "aad-test-passphrase";
      const plaintext = "Message with context binding";
      const aad = "org-123:env-production";
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase, aad);
      expect(encrypted.aad).toBeTruthy();
      const decrypted = await VaultCrypto.decryptWithPassphrase(encrypted, passphrase);
      expect(decrypted).toBe("decrypted");
    });

    it("decryption fails when aad is tampered", async () => {
      const passphrase = "aad-tamper-test";
      const plaintext = "Message with context binding";
      const aad = "org-123:env-production";
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase, aad);
      encrypted.aad = arrayBufferToBase64(new TextEncoder().encode("tampered-aad"));
      mockDecrypt.mockRejectedValueOnce(new Error("AAD mismatch"));
      await expect(
        VaultCrypto.decryptWithPassphrase(encrypted, passphrase)
      ).rejects.toThrow();
    });
  });

  describe("encryptWithPassphrase edge cases", () => {
    it("handles empty string plaintext", async () => {
      const passphrase = "passphrase";
      const plaintext = "";
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase);
      expect(encrypted.ciphertext).toBeDefined();
      const decrypted = await VaultCrypto.decryptWithPassphrase(encrypted, passphrase);
      expect(decrypted).toBe("decrypted");
    });

    it("handles unicode plaintext", async () => {
      const passphrase = "passphrase";
      const plaintext = "日本語🔐हिंदीالعربية Emoji: 🎉";
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase);
      const decrypted = await VaultCrypto.decryptWithPassphrase(encrypted, passphrase);
      expect(decrypted).toBe("decrypted");
    });

    it("handles large data", async () => {
      const passphrase = "passphrase";
      const plaintext = "A".repeat(100_000);
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase);
      const decrypted = await VaultCrypto.decryptWithPassphrase(encrypted, passphrase);
      expect(decrypted).toBe("decrypted");
    });

    it("handles special characters in passphrase", async () => {
      const passphrase = "p@$$w0rd!#$%^&*()";
      const plaintext = "Secret with special passphrase";
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase);
      const decrypted = await VaultCrypto.decryptWithPassphrase(encrypted, passphrase);
      expect(decrypted).toBe("decrypted");
    });

    it("handles multiline plaintext", async () => {
      const passphrase = "passphrase";
      const plaintext = "Line 1\nLine 2\nLine 3\n\nLine 5";
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase);
      const decrypted = await VaultCrypto.decryptWithPassphrase(encrypted, passphrase);
      expect(decrypted).toBe("decrypted");
    });

    it("handles JSON data", async () => {
      const passphrase = "passphrase";
      const plaintext = JSON.stringify({ user: "alice", token: "sekrit", roles: ["admin", "user"] });
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase);
      const decrypted = await VaultCrypto.decryptWithPassphrase(encrypted, passphrase);
      expect(decrypted).toBe("decrypted");
    });
  });

  describe("toPayload / fromPayload", () => {
    it("converts EncryptedData to EncryptedDataPayload", async () => {
      const passphrase = "payload-test";
      const plaintext = "test message";
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase);
      const payload = VaultCrypto.toPayload(encrypted);
      expect(payload.ciphertext).toBe(encrypted.ciphertext);
      expect(payload.iv).toBe(encrypted.iv);
      expect(payload.salt).toBe(encrypted.salt);
      expect(payload.tag).toBe(encrypted.tag);
      expect(payload.key_version).toBe(encrypted.keyVersion);
    });

    it("converts EncryptedDataPayload back to EncryptedData", async () => {
      const passphrase = "payload-test";
      const plaintext = "test message";
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase);
      const payload = VaultCrypto.toPayload(encrypted);
      const restored = VaultCrypto.fromPayload(payload);
      expect(restored.ciphertext).toBe(encrypted.ciphertext);
      expect(restored.iv).toBe(encrypted.iv);
      expect(restored.salt).toBe(encrypted.salt);
      expect(restored.tag).toBe(encrypted.tag);
      expect(restored.keyVersion).toBe(encrypted.keyVersion);
    });

    it("roundtrips through payload without data loss", async () => {
      const passphrase = "roundtrip-test";
      const plaintext = "Sensitive data";
      const encrypted = await VaultCrypto.encryptWithPassphrase(plaintext, passphrase);
      const payload = VaultCrypto.toPayload(encrypted);
      const restored = VaultCrypto.fromPayload(payload);
      const decrypted = await VaultCrypto.decryptWithPassphrase(restored, passphrase);
      expect(decrypted).toBe("decrypted");
    });
  });
});
