import { it, expect, vi } from "vitest";
import {
  fetchMacroDashboard,
  fetchMacroOperations,
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

it("accepts S&P provenance and rejects lookalike domains", () => {
  const latest = {
    indicatorCode: "VN_PMI",
    value: 51.9,
    unit: "index",
    period: "2026-09",
    periodEnd: "2026-09-30",
    sourceUrl:
      "https://www.pmi.spglobal.com/Public/Home/PressRelease/dba8ca68ce13403591b68cb3e07175d9",
  };
  const indicator = {
    code: "VN_PMI",
    name: "PMI",
    enabled: true,
    unit: "index",
    latest,
    history: [latest],
  };
  expect(
    validateMacroDashboard({ ...empty, indicators: [indicator] }).indicators,
  ).toHaveLength(1);
  const bad = {
    ...latest,
    sourceUrl: "https://www.pmi.spglobal.com.evil.test/release",
  };
  expect(() =>
    validateMacroDashboard({
      ...empty,
      indicators: [{ ...indicator, latest: bad, history: [bad] }],
    }),
  ).toThrow();
});

it("keeps ICE DXY and FRED Broad USD as distinct series", () => {
  const latest = {
    indicatorCode: "US_DXY",
    value: 101.2,
    unit: "index Mar 1973=100",
    period: "2026-10-23",
    periodEnd: "2026-10-23",
    sourceUrl:
      "https://www.ice.com/fixed-income-data-services/index-solutions/currency-indices",
  };
  const indicator = {
    code: "US_DXY",
    name: "ICE DXY",
    enabled: true,
    unit: latest.unit,
    latest,
    history: [latest],
  };
  expect(
    validateMacroDashboard({ ...empty, indicators: [indicator] }).indicators[0]
      .code,
  ).toBe("US_DXY");
  expect(() =>
    validateMacroDashboard({
      ...empty,
      indicators: [
        { ...indicator, latest: { ...latest, indicatorCode: "US_BROAD_USD" } },
      ],
    }),
  ).toThrow();
});

it("rejects malformed operations payloads without treating them as healthy sources", async () => {
  try {
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(
          new Response(JSON.stringify({ sources: [null], alerts: [] }), {
            headers: { "Content-Type": "application/json" },
          }),
        ),
    );
    await expect(fetchMacroOperations()).rejects.toThrow("không hợp lệ");
    vi.stubGlobal(
      "fetch",
      vi
        .fn()
        .mockResolvedValue(
          new Response(JSON.stringify({ sources: [], alerts: [null] }), {
            headers: { "Content-Type": "application/json" },
          }),
        ),
    );
    await expect(fetchMacroOperations()).rejects.toThrow("không hợp lệ");
  } finally {
    vi.unstubAllGlobals();
  }
});
