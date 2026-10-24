import { it, expect } from "vitest";
import {
  fetchMacroDashboard,
  validateMacroDashboard,
} from "./macro-api-client";
const empty = {
  mode: "live",
  generatedAt: "2026-10-24T00:00:00Z",
  available: 0,
  enabled: 8,
  indicators: [],
  score: null,
  scoreStatus: "NOT_CONFIGURED",
};
it("accepts a partial live dashboard and rejects mock scores and HTML fallbacks", async () => {
  expect(validateMacroDashboard(empty).available).toBe(0);
  expect(() => validateMacroDashboard({ ...empty, score: 62 })).toThrow();
  expect(() => validateMacroDashboard({ ...empty, mode: "mock" })).toThrow();
  await expect(
    fetchMacroDashboard({
      fetchImpl: async () =>
        new Response("<html>", { headers: { "Content-Type": "text/html" } }),
    }),
  ).rejects.toThrow("chưa được cấu hình");
});
it("does not replace an API outage with mock observations", async () => {
  await expect(
    fetchMacroDashboard({
      fetchImpl: async () => new Response("", { status: 503 }),
    }),
  ).rejects.toThrow("Kho dữ liệu");
  await expect(
    fetchMacroDashboard({
      fetchImpl: async () => {
        throw new Error("offline");
      },
    }),
  ).rejects.toThrow("không được dùng thay thế");
});
