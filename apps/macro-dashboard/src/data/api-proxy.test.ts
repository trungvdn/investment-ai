import { expect, it } from "vitest";
import { apiProxy } from "../../config/api-proxy";

it("defaults to Docker's IPv4 loopback and routes macro before stock APIs", () => {
  const proxy = apiProxy();
  expect(Object.keys(proxy)).toEqual(["/api/macro", "/api"]);
  expect(proxy["/api/macro"].target).toBe("http://127.0.0.1:3000");
  expect(proxy["/api"].target).toBe("http://127.0.0.1:8081");
});

it("uses env-file targets but gives explicit shell settings precedence", () => {
  const file = { MACRO_API_URL: "http://macro-api:3000", STOCK_API_URL: "http://stock:8081" };
  expect(apiProxy({}, file)["/api/macro"].target).toBe("http://macro-api:3000");
  expect(apiProxy({ MACRO_API_URL: "http://127.0.0.1:3300" }, file)["/api/macro"].target)
    .toBe("http://127.0.0.1:3300");
  expect(apiProxy({}, file)["/api"].target).toBe("http://stock:8081");
});
