import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Routes, Route } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Layout } from "../components/layout";
import Research from "./research";
import Macro from "./macro";
import { fetchStockReport } from "../data/stock-api-client.js";
vi.mock("../data/stock-api-client.js", () => ({ fetchStockReport: vi.fn() }));
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
  it("filters macro indicators and opens history detail", async () => {
    const user = userEvent.setup();
    renderWorkspace("/macro");
    expect(screen.getByText("Hiển thị 15 / 15 chỉ báo")).toBeInTheDocument();
    await user.selectOptions(
      screen.getByRole("combobox", { name: "Nhóm chỉ báo" }),
      "Inflation",
    );
    expect(screen.getByText("Hiển thị 3 / 15 chỉ báo")).toBeInTheDocument();
    await user.type(
      screen.getByRole("textbox", { name: "Tìm chỉ báo" }),
      "Core",
    );
    expect(screen.getByText("Hiển thị 1 / 15 chỉ báo")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: /^Core CPI$/ }));
    expect(
      screen.getByRole("region", { name: "Chi tiết Core CPI" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Lịch sử chỉ báo")).toBeInTheDocument();
  });
});
