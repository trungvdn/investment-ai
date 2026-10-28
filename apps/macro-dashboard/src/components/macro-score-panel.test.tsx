import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MacroScorePanel } from "./macro-score-panel";
import { fetchMacroScore } from "../data/macro-score-client";
vi.mock("../data/macro-score-client", () => ({ fetchMacroScore: vi.fn() }));
vi.mock("recharts", () => ({
  ResponsiveContainer: ({ children }: any) => <div>{children}</div>,
  LineChart: ({ data, children }: any) => <div data-testid="score-chart" data-points={JSON.stringify(data)}>{children}</div>,
  Line: () => null, XAxis: () => null, YAxis: () => null, CartesianGrid: () => null, Tooltip: () => null,
}));
function scoreData(history: any[] = []) {
  return {
    current: {
      asOf: "2026-10-28T12:00:00Z", modelVersion: "macro-v1.0.0", score: 65, coverage: 100, status: "READY",
      groups: [{ name: "Liquidity", weight: 50, coverage: 100, score: 65 }, { name: "Production", weight: 30, coverage: 100, score: 65 }, { name: "Inflation", weight: 20, coverage: 100, score: 65 }],
      components: [{ code: "VN_PMI", name: "PMI sản xuất", weight: 20, critical: true, score: 65, contribution: 13, effectiveWeight: 20, reason: "ELIGIBLE", inputs: [], warnings: [] }],
    }, history,
  } as any;
}
afterEach(cleanup);
function renderPanel() { render(<QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}><MacroScorePanel /></QueryClientProvider>); }
describe("macro score panel", () => {
  it("shows no fabricated history and requests the selected frequency", async () => {
    vi.mocked(fetchMacroScore).mockResolvedValue({ current: null, history: [] } as any);
    renderPanel(); await screen.findByText(/Chưa có snapshot/);
    await userEvent.click(screen.getByRole("button", { name: "Tuần" }));
    await waitFor(() => expect(fetchMacroScore).toHaveBeenCalledWith("weekly", expect.any(AbortSignal)));
  });
  it("isolates an unavailable score endpoint", async () => {
    vi.mocked(fetchMacroScore).mockRejectedValue(new Error("Kho điểm không sẵn sàng"));
    renderPanel(); expect(await screen.findByRole("alert", {}, { timeout: 3000 })).toHaveTextContent("Kho điểm không sẵn sàng");
  });
  it("keeps the chart visible and reveals indicators and groups only on request", async () => {
    vi.mocked(fetchMacroScore).mockResolvedValue(scoreData());
    renderPanel(); await screen.findByRole("region", { name: "Biểu đồ biến động Macro Score" });
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(screen.queryByText("Thanh khoản & lãi suất")).not.toBeInTheDocument();
    const user = userEvent.setup();
    const indicators = screen.getByRole("button", { name: "Xem chi tiết điểm từng chỉ báo" });
    expect(indicators).toHaveAttribute("aria-expanded", "false");
    await user.click(indicators);
    expect(screen.getByRole("table")).toBeInTheDocument();
    expect(screen.getByText("PMI sản xuất *")).toBeInTheDocument();
    expect(screen.queryByText("Thanh khoản & lãi suất")).not.toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Xem điểm theo nhóm" }));
    expect(screen.getByText("Thanh khoản & lãi suất")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Ẩn chi tiết điểm từng chỉ báo" }));
    expect(screen.queryByRole("table")).not.toBeInTheDocument();
    expect(screen.getByText("Thanh khoản & lãi suất")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Ẩn điểm theo nhóm" }));
    expect(screen.queryByText("Thanh khoản & lãi suất")).not.toBeInTheDocument();
    expect(screen.getByTestId("score-chart")).toBeInTheDocument();
  });
  it("plots current score before the first close without fabricating history", async () => {
    vi.mocked(fetchMacroScore).mockResolvedValue(scoreData());
    renderPanel(); const chart = await screen.findByTestId("score-chart");
    const points = JSON.parse(chart.getAttribute("data-points")!);
    expect(points).toHaveLength(1);
    expect(points[0]).toMatchObject({ score: 65, asOf: "2026-10-28T12:00:00Z", isCurrent: true });
    expect(screen.getByText(/Chưa có lịch sử để thể hiện biến động/)).toBeInTheDocument();
  });
  it("preserves historical gaps and appends the explicitly labelled current point", async () => {
    vi.mocked(fetchMacroScore).mockResolvedValue(scoreData([
      { asOf: "2026-10-26T11:00:00Z", score: 60 },
      { asOf: "2026-10-27T11:00:00Z", score: null },
    ]));
    renderPanel(); const chart = await screen.findByTestId("score-chart");
    const points = JSON.parse(chart.getAttribute("data-points")!);
    expect(points.map((p: any) => p.score)).toEqual([60, null, 65]);
    expect(points.map((p: any) => p.isCurrent)).toEqual([false, false, true]);
  });

});
