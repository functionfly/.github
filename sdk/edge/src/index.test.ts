import { describe, expect, it, vi } from "vitest";
import { EdgeSDKError, FunctionFlyClient, createClient } from "./index";

describe("FunctionFlyClient", () => {
  it("sends bearer auth and tenant headers", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ states: [], total: 0 }),
    });

    const client = new FunctionFlyClient(
      {
        apiKey: "test-key",
        tenantId: "tenant-1",
      },
      fetchMock,
    );

    await client.listStates();

    expect(fetchMock).toHaveBeenCalledOnce();
    const [url, init] = fetchMock.mock.calls[0];
    expect(String(url)).toContain("/v1/state");
    expect(init.headers.Authorization).toBe("Bearer test-key");
    expect(init.headers["X-Tenant-ID"]).toBe("tenant-1");
  });

  it("encodes state paths in requests", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ path: "acme/users", value: {} }),
    });

    const client = createClient({ apiKey: "test-key" }, fetchMock);
    await client.getState("acme/users");

    const [url] = fetchMock.mock.calls[0];
    expect(String(url)).toContain("/state/acme%2Fusers");
  });

  it("sets and gets state values with encoded paths", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ path: "acme/cart", value: { items: ["sku-1"] } }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({ path: "acme/cart", value: { items: ["sku-1"] } }),
      });

    const client = createClient({ apiKey: "test-key" }, fetchMock);

    await client.setValue("acme/cart", { items: ["sku-1"] });
    const value = await client.getValue("acme/cart");

    expect(value.value).toEqual({ items: ["sku-1"] });
    expect(fetchMock).toHaveBeenCalledTimes(2);

    const setCall = fetchMock.mock.calls[0];
    expect(String(setCall[0])).toContain("/state/acme%2Fcart/value");
    expect(setCall[1]?.method).toBe("PUT");
    expect(JSON.parse(String(setCall[1]?.body))).toEqual({
      value: { items: ["sku-1"] },
    });
  });

  it("throws EdgeSDKError on API failures", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: false,
      status: 403,
      statusText: "Forbidden",
      json: async () => ({ error: "forbidden", message: "Access denied" }),
    });

    const client = createClient({ apiKey: "test-key" }, fetchMock);

    await expect(client.getState("secret/data")).rejects.toBeInstanceOf(
      EdgeSDKError,
    );
  });
});

describe("EdgeSDKError", () => {
  it("captures status code and response body", () => {
    const err = new EdgeSDKError("failed", 503, { retryable: true });
    expect(err.statusCode).toBe(503);
    expect(err.body).toEqual({ retryable: true });
  });
});
