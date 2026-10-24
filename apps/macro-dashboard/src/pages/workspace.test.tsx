import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Layout } from "../components/layout";
import Research from "./research";
import Macro from "./macro";
import { fetchStockReport } from "../data/stock-api-client.js";
import { fetchMacroDashboard } from "../data/macro-api-client";
import type { MacroDashboard } from "../lib/types";
vi.mock("../data/stock-api-client.js", () => ({ fetchStockReport: vi.fn() }));
vi.mock("../data/macro-api-client", () => ({ fetchMacroDashboard: vi.fn() }));
vi.mock("../components/price-chart", () => ({
  PriceChart: ({ onSelectEvent, report }: any) => (
    <button onClick={() => onSelectEvent(report.events[0])}>
      Biểu đồ kiểm thử
    </button>
  ),
}));
vi.mock("recharts", () => ({
  ResponsiveContainer: ({ children }: any) => <div>{children}</div>,
  AreaChart: () => <div>Biểu đồ vĩ mô</div>,
  LineChart: () => <div>Lịch sử chỉ báo</div>,
  Area: () => null,
  Line: () => null,
  CartesianGrid: () => null,
  Tooltip: () => null,
  XAxis: () => null,
  YAxis: () => null,
}));
const report = {
  schemaVersion: "1.0",
  symbol: "FPT",
  asOf: "2026-10-23",
  structure: "UNCLEAR",
  phase: "UNKNOWN",
  confidence: 0.2,
  tradingRange: {
    support: 98,
    resistance: 102,
    status: "PROVISIONAL",
    widthPct: 4,
  },
  events: [],
  evidence: ["Bằng chứng thử nghiệm"],
  risks: ["Điểm không phải xác suất"],
  watchlist: ["Chờ xác nhận"],
  dataQuality: {
    mode: "demo",
    status: "DEMO",
    dataPoints: 60,
    notes: ["Dữ liệu tổng hợp"],
  },
  disclaimer: "Chỉ dùng nghiên cứu",
  summary: "Unclear",
  candles: [
    {
      date: "2026-10-23",
      open: 100,
      high: 102,
      low: 99,
      close: 101,
      volume: 10000,
    },
  ],
  timeframe: "1D",
};

const latest = {
  indicatorCode: "VN_CORE_CPI_YOY",
  period: "2026-09",
  periodEnd: "2026-09-30T00:00:00Z",
  value: 4.45,
  unit: "% YoY",
  source: "NSO",
  sourceUrl: "https://www.nso.gov.vn/report/",
  sourceSeriesId: "NSO_REPORT",
  method: "EXPLICIT_MONTHLY_YOY",
  publishedAt: "2026-10-03T01:00:00Z",
  firstSeenAt: "2026-10-24T00:00:00Z",
  collectedAt: "2026-10-24T00:00:00Z",
  revision: 1,
  rawId: 1,
};
const macroFixture = {
  mode: "live",
  generatedAt: "2026-10-24T00:00:00Z",
  available: 1,
  enabled: 8,
  score: null,
  scoreStatus: "NOT_CONFIGURED",
  indicators: [
    {
      code: "VN_CORE_CPI_YOY",
      name: "Lạm phát cơ bản",
      group: "Inflation",
      country: "VN",
      unit: "% YoY",
      frequency: "MONTHLY",
      comparisonBasis: "YoY",
      enabled: true,
      maxAgeDays: 60,
      description: "Số tháng so cùng kỳ, không phải lũy kế.",
      status: "AVAILABLE",
      latest,
      history: [latest],
      lastAttempt: null,
    },
    {
      code: "US_EFFR",
      name: "Lãi suất Fed hiệu dụng",
      group: "External",
      country: "US",
      unit: "%",
      frequency: "DAILY",
      comparisonBasis: "LEVEL",
      enabled: true,
      maxAgeDays: 7,
      description: "FRED DFF",
      status: "NOT_CONFIGURED",
      latest: null,
      history: [],
      lastAttempt: {
        message: "FRED_API_KEY is not configured",
        success: false,
        checkedAt: "2026-10-24T00:00:00Z",
        indicatorCode: "US_EFFR",
      },
    },
  ],
} as MacroDashboard;
function renderWorkspace(path = "/") {
  return render(
    <QueryClientProvider
      client={
        new QueryClient({
          defaultOptions: { queries: { retry: false, gcTime: 0 } },
        })
      }
    >
      <MemoryRouter initialEntries={[path]}>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<Research />} />
            <Route path="/macro" element={<Macro />} />
          </Route>
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}
beforeEach(() => {
  vi.mocked(fetchMacroDashboard).mockResolvedValue(macroFixture);

  vi.clearAllMocks();
  localStorage.clear();
  document.documentElement.classList.remove("dark");
});
describe("React workspace", () => {
  it("navigates between React pages and preserves theme", async () => {
    const user = userEvent.setup();
    renderWorkspace();
    expect(
      screen.getByRole("heading", { name: /Đọc dòng tiền/ }),
    ).toBeInTheDocument();
    await user.click(
      screen.getByRole("button", { name: "Chuyển giao diện tối" }),
    );
    expect(document.documentElement).toHaveClass("dark");
    await user.click(screen.getByRole("link", { name: "Bối cảnh vĩ mô" }));
    expect(
      screen.getByRole("heading", { name: /Bối cảnh lớn/ }),
    ).toBeInTheDocument();
    expect(document.documentElement).toHaveClass("dark");
  });
  it("submits a normalized symbol and displays uncertain demo report", async () => {
    vi.mocked(fetchStockReport).mockResolvedValue(report);
    const user = userEvent.setup();
    renderWorkspace();
    await user.type(
      screen.getByRole("textbox", { name: "Mã cổ phiếu" }),
      "fpt",
    );
    await user.click(screen.getByRole("button", { name: /Phân tích/ }));
    await screen.findByText("Bằng chứng thử nghiệm");
    expect(fetchStockReport).toHaveBeenCalledWith(
      expect.objectContaining({ symbol: "FPT", timeframe: "1D", period: "24" }),
    );
    expect(screen.getByText(/Nến tổng hợp phục vụ/)).toBeInTheDocument();
    await user.click(screen.getByRole("tab", { name: /Kịch bản theo dõi/ }));
    expect(
      screen.getByRole("heading", { name: "Vượt kháng cự" }),
    ).toBeInTheDocument();
  });
  it("shows provider failure without fabricated reports", async () => {
    vi.mocked(fetchStockReport).mockRejectedValue(
      new Error("Nguồn dữ liệu không khả dụng"),
    );
    const user = userEvent.setup();
    renderWorkspace();
    await user.click(screen.getByRole("button", { name: "FPT" }));
    await screen.findByText("Nguồn dữ liệu không khả dụng");
    expect(screen.queryByText("Bằng chứng thử nghiệm")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Thử lại" })).toBeInTheDocument();
  });
  it("uses weekly aggregation request and defaults to two years", async () => {
    vi.mocked(fetchStockReport).mockResolvedValue({
      ...report,
      timeframe: "1W",
    });
    const user = userEvent.setup();
    renderWorkspace();
    await user.selectOptions(
      screen.getByRole("combobox", { name: "Khoảng dữ liệu" }),
      "6",
    );
    await user.selectOptions(
      screen.getByRole("combobox", { name: "Khung phân tích" }),
      "1W",
    );
    expect(
      screen.getByRole("combobox", { name: "Khoảng dữ liệu" }),
    ).toHaveValue("24");
    await user.click(screen.getByRole("button", { name: "VCB" }));
    await waitFor(() =>
      expect(fetchStockReport).toHaveBeenCalledWith(
        expect.objectContaining({
          symbol: "VCB",
          timeframe: "1W",
          period: "24",
        }),
      ),
    );
  });

  it("loads live macro data, filters missing sources and shows provenance without a market score", async () => {
    const user = userEvent.setup();
    renderWorkspace("/macro");
    await screen.findByText("Hiển thị 2 / 2 chỉ báo");
    expect(screen.getByText("MACRO SCORE TẠM ẨN")).toBeInTheDocument();
    await user.selectOptions(
      screen.getByRole("combobox", { name: "Nhóm chỉ báo" }),
      "Inflation",
    );
    expect(screen.getByText("Hiển thị 1 / 2 chỉ báo")).toBeInTheDocument();
    await user.type(
      screen.getByRole("textbox", { name: "Tìm chỉ báo" }),
      "Lạm phát",
    );
    expect(
      screen.getByRole("link", { name: "Mở nguồn chính thức" }),
    ).toHaveAttribute("href", "https://www.nso.gov.vn/report/");
    expect(screen.getByText("r1 / #1")).toBeInTheDocument();
  });
  it("shows a macro API outage and never renders the old mock score", async () => {
    vi.mocked(fetchMacroDashboard).mockRejectedValue(
      new Error("Macro API offline"),
    );
    renderWorkspace("/macro");
    await screen.findByText("Macro API offline");
    expect(screen.queryByText("62")).not.toBeInTheDocument();
  });
});
