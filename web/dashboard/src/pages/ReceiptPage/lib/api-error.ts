// Small typed error class for the receipt hooks. Kept here (rather than
// re-using a global one) so the receipt feature is self-contained and
// doesn't depend on the rest of the API layer's evolving error types.

export class ApiError extends Error {
  public readonly status: number;
  public readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    // Maintain proper prototype chain for `instanceof` to work after
    // compilation to ES5 targets.
    Object.setPrototypeOf(this, ApiError.prototype);
  }
}
