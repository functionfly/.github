import { VaultAPIError } from "./errors";

/**
 * Options for configuring a {@link VaultClient}.
 */
export interface VaultClientOptions {
  /** FunctionFly API base URL. Default: https://api.functionfly.com */
  baseUrl?: string;
  /** Bearer token used for authentication. */
  token: string;
  /** Request timeout in milliseconds. Default: 30_000. */
  timeoutMs?: number;
  /** Optional override for the underlying fetch function (for testing). */
  fetchImpl?: typeof fetch;
  /** Custom User-Agent suffix. Default: @functionfly/vault/<version>. */
  userAgent?: string;
}

export const SDK_VERSION = "0.1.0";
export const DEFAULT_BASE_URL = "https://api.functionfly.com";

/**
 * VaultClient is the entry point for the FunctionFly vault SDK.
 * The client itself is stateless; each sub-service has its own
 * methods that map 1:1 to the REST API.
 */
export class VaultClient {
  public readonly baseUrl: string;
  public readonly token: string;
  private readonly timeoutMs: number;
  private readonly fetchImpl: typeof fetch;
  private readonly userAgent: string;

  public readonly secrets: import("./services/secrets").SecretsService;
  public readonly tokens: import("./services/tokens").TokensService;
  public readonly dynamicCredentials: import("./services/dynamic").DynamicCredentialsService;
  public readonly dynamicTargets: import("./services/dynamic").DynamicTargetsService;
  public readonly leases: import("./services/dynamic").LeasesService;
  public readonly audit: import("./services/audit").AuditService;

  constructor(opts: VaultClientOptions) {
    if (!opts.token) {
      throw new Error("token is required");
    }
    this.baseUrl = (opts.baseUrl ?? DEFAULT_BASE_URL).replace(/\/+$/, "");
    this.token = opts.token;
    this.timeoutMs = opts.timeoutMs ?? 30_000;
    this.fetchImpl = opts.fetchImpl ?? globalThis.fetch.bind(globalThis);
    this.userAgent = opts.userAgent ?? `@functionfly/vault/${SDK_VERSION}`;

    // Lazily-required to break circular type imports in declarations.
    // eslint-disable-next-line @typescript-eslint/no-var-requires
    const { SecretsService } = require("./services/secrets");
    // eslint-disable-next-line @typescript-eslint/no-var-requires
    const { TokensService } = require("./services/tokens");
    // eslint-disable-next-line @typescript-eslint/no-var-requires
    const { DynamicCredentialsService, DynamicTargetsService, LeasesService } =
      require("./services/dynamic");
    // eslint-disable-next-line @typescript-eslint/no-var-requires
    const { AuditService } = require("./services/audit");

    this.secrets = new SecretsService(this);
    this.tokens = new TokensService(this);
    this.dynamicCredentials = new DynamicCredentialsService(this);
    this.dynamicTargets = new DynamicTargetsService(this);
    this.leases = new LeasesService(this);
    this.audit = new AuditService(this);
  }

  /**
   * Internal: performs an HTTP request and decodes the JSON response.
   * Throws {@link VaultAPIError} on non-2xx status codes.
   */
  public async request<T>(
    method: string,
    path: string,
    body?: unknown,
  ): Promise<T> {
    const url = this.baseUrl + path;
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.timeoutMs);
    const headers: Record<string, string> = {
      Accept: "application/json",
      "User-Agent": this.userAgent,
    };
    let payload: string | undefined;
    if (body !== undefined) {
      payload = JSON.stringify(body);
      headers["Content-Type"] = "application/json";
    }
    let response: Response;
    try {
      response = await this.fetchImpl(url, {
        method,
        headers: {
          ...headers,
          Authorization: `Bearer ${this.token}`,
        },
        body: payload,
        signal: controller.signal,
      });
    } catch (err) {
      clearTimeout(timeout);
      throw new VaultAPIError(0, "network_error", "network", String(err));
    }
    clearTimeout(timeout);
    const text = await response.text();
    if (!response.ok) {
      let errBody: Record<string, unknown> = {};
      try {
        errBody = text ? JSON.parse(text) : {};
      } catch {
        // ignore
      }
      const message =
        (typeof errBody.message === "string" && errBody.message) ||
        (typeof errBody.error === "string" && errBody.error) ||
        `HTTP ${response.status}`;
      const code = typeof errBody.code === "string" ? errBody.code : "HTTP_ERROR";
      throw new VaultAPIError(response.status, code, message, errBody);
    }
    if (text.length === 0) {
      return undefined as unknown as T;
    }
    try {
      return JSON.parse(text) as T;
    } catch (err) {
      throw new VaultAPIError(
        response.status,
        "decode_error",
        "decode",
        `could not parse response: ${String(err)}`,
      );
    }
  }
}
