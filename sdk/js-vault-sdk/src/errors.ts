/**
 * VaultAPIError is the structured error returned by the SDK when the
 * FunctionFly vault API returns a non-2xx status code.
 */
export class VaultAPIError extends Error {
  public readonly status: number;
  public readonly code: string;
  public readonly details: Record<string, unknown>;

  constructor(
    status: number,
    code: string,
    label: string,
    messageOrDetails?: string | Record<string, unknown>,
    details?: Record<string, unknown>,
  ) {
    let msg: string;
    let det: Record<string, unknown> = {};
    if (typeof messageOrDetails === "string") {
      msg = messageOrDetails;
      det = details ?? {};
    } else {
      msg = label;
      det = messageOrDetails ?? {};
    }
    super(`${code}: ${msg}`);
    this.name = "VaultAPIError";
    this.status = status;
    this.code = code;
    this.details = det;
  }
}
