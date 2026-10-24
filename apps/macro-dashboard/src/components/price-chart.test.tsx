import { it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { PriceChart } from "./price-chart";
import type { StockReport } from "../lib/types";
const fake = vi.hoisted(() => {
  const candleData = vi.fn(),
    volumeData = vi.fn(),
    markers = vi.fn(),
    remove = vi.fn(),
    focus = vi.fn();
  let click: (event: { time: string }) => void = () => {};
  const bars = {
    setData: candleData,
    createPriceLine: vi.fn(),
    priceToCoordinate: () => 60,
  };
  const volume = {
    setData: volumeData,
    priceScale: () => ({ applyOptions: vi.fn() }),
  };
  const chart = {
    addSeries: (type: string) => (type === "candle" ? bars : volume),
    subscribeClick: (cb: typeof click) => {
      click = cb;
    },
    subscribeCrosshairMove: vi.fn(),
    timeScale: () => ({
      subscribeVisibleLogicalRangeChange: vi.fn(),
      fitContent: vi.fn(),
      setVisibleLogicalRange: focus,
    }),
    remove,
  };
  return {
    candleData,
    volumeData,
    markers,
    remove,
    focus,
    createChart: vi.fn(() => chart),
    click: (time: string) => click({ time }),
  };
});
vi.mock("lightweight-charts", () => ({
  CandlestickSeries: "candle",
  HistogramSeries: "volume",
  ColorType: { Solid: "solid" },
  CrosshairMode: { Normal: 0 },
  createChart: fake.createChart,
  createSeriesMarkers: fake.markers,
}));
const event = {
  type: "SPRING",
  date: "2026-10-20T00:00:00Z",
  confidence: 0.7,
  evidence: "Spring evidence",
};
const report = {
  symbol: "FPT",
  timeframe: "1D",
  candles: [
    {
      date: event.date,
      open: 100,
      high: 102,
      low: 98,
      close: 101,
      volume: 200,
    },
  ],
  events: [event],
  tradingRange: { support: 99, resistance: 103, status: "PROVISIONAL" },
} as StockReport;
it("connects exact candles, volume and dated event selection to the chart; removes chart on unmount", () => {
  const select = vi.fn();
  const view = render(
    <PriceChart
      report={report}
      dark={false}
      selectedEvent={null}
      onSelectEvent={select}
    />,
  );
  expect(fake.candleData).toHaveBeenCalledWith([
    expect.objectContaining({ time: "2026-10-20", open: 100, close: 101 }),
  ]);
  expect(fake.volumeData).toHaveBeenCalledWith([
    expect.objectContaining({ time: "2026-10-20", value: 200 }),
  ]);
  expect(fake.markers).toHaveBeenCalledWith(expect.anything(), [
    expect.objectContaining({ time: "2026-10-20", text: "SPRING" }),
  ]);
  fake.click("2026-10-20");
  expect(select).toHaveBeenCalledWith(event);
  view.rerender(
    <PriceChart
      report={report}
      dark={false}
      selectedEvent={event}
      onSelectEvent={select}
    />,
  );
  expect(fake.focus).toHaveBeenCalled();
  view.unmount();
  expect(fake.remove).toHaveBeenCalled();
});
it("recreates the chart with dark colors and toggles volume", async () => {
  const select = vi.fn(),
    user = userEvent.setup();
  render(
    <PriceChart
      report={report}
      dark
      selectedEvent={null}
      onSelectEvent={select}
    />,
  );
  expect(fake.createChart).toHaveBeenLastCalledWith(
    expect.anything(),
    expect.objectContaining({
      layout: expect.objectContaining({
        background: { type: "solid", color: "#142232" },
      }),
    }),
  );
  const toggle = screen.getByRole("button", { name: "Khối lượng" });
  expect(toggle).toHaveAttribute("aria-pressed", "true");
  await user.click(toggle);
  expect(toggle).toHaveAttribute("aria-pressed", "false");
});
