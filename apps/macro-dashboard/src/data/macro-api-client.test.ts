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

it("accepts official NHNN provenance and quarterly PPI periods", () => {
  const latest = {
    indicatorCode: "VN_PPI_YOY",
    value: 4.75,
    unit: "% YoY",
    period: "2026-Q3",
    periodEnd: "2026-09-30",
    sourceUrl: "https://nsdp.nso.gov.vn/GSO-chung/SDMXFiles/GSO/PPIVNM.xml",
  };
  const indicator = {
    code: "VN_PPI_YOY",
    name: "PPI công nghiệp",
    enabled: true,
    unit: "% YoY",
    latest,
    history: [latest],
  };
  expect(
    validateMacroDashboard({ ...empty, indicators: [indicator] }).indicators[0]
      .latest?.period,
  ).toBe("2026-Q3");
  const sbv = {
    ...latest,
    indicatorCode: "VN_USD_VND",
    unit: "VND/USD",
    period: "2026-10-23",
    periodEnd: "2026-10-23",
    sourceUrl: "https://sbv.gov.vn/vi/tỷ-giá",
  };
  expect(
    validateMacroDashboard({
      ...empty,
      indicators: [
        {
          ...indicator,
          code: "VN_USD_VND",
          unit: "VND/USD",
          latest: sbv,
          history: [sbv],
        },
      ],
    }).indicators[0].latest?.sourceUrl,
  ).toContain("sbv.gov.vn");
  expect(() =>
    validateMacroDashboard({
      ...empty,
      indicators: [
        {
          ...indicator,
          latest: { ...latest, sourceUrl: "https://sbv.gov.vn.evil.example/" },
        },
      ],
    }),
  ).toThrow();
});
