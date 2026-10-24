import { useEffect, useRef, useState } from "react";
import {
  CandlestickSeries,
  HistogramSeries,
  ColorType,
  CrosshairMode,
  createChart,
  createSeriesMarkers,
  type IChartApi,
  type Time,
} from "lightweight-charts";
import { Expand, Layers, LocateFixed, MousePointer2 } from "lucide-react";
import { Button } from "./ui/button";
import type { StockReport, WyckoffEvent, Candle } from "../lib/types";
import { formatDate, formatNumber } from "../lib/utils";
export function chartData(candles: Candle[]) {
  return candles.map((c) => ({ ...c, time: c.date.slice(0, 10) as Time }));
}
export function PriceChart({
  report,
  dark,
  selectedEvent,
  onSelectEvent,
}: {
  report: StockReport;
  dark: boolean;
  selectedEvent: WyckoffEvent | null;
  onSelectEvent: (event: WyckoffEvent) => void;
}) {
  const container = useRef<HTMLDivElement>(null),
    chartRef = useRef<IChartApi | null>(null);
  const [events, setEvents] = useState(true),
    [range, setRange] = useState(true),
    [volume, setVolume] = useState(true);
  const [hover, setHover] = useState<Candle | null>(null),
    [shade, setShade] = useState({ top: 0, height: 0 });
  useEffect(() => {
    if (!container.current || !report.candles.length) return;
    const chart = createChart(container.current, {
      autoSize: true,
      height: 440,
      layout: {
        background: {
          type: ColorType.Solid,
          color: dark ? "#142232" : "#ffffff",
        },
        textColor: dark ? "#8b9db0" : "#8291a2",
        fontSize: 11,
        attributionLogo: true,
      },
      grid: {
        vertLines: { color: dark ? "#203145" : "#f1f4f7" },
        horzLines: { color: dark ? "#203145" : "#edf1f5" },
      },
      crosshair: { mode: CrosshairMode.Normal },
      timeScale: {
        borderColor: dark ? "#2a3b4e" : "#e7edf2",
        rightOffset: 5,
        timeVisible: false,
      },
      rightPriceScale: {
        borderVisible: false,
        scaleMargins: { top: 0.12, bottom: volume ? 0.26 : 0.08 },
      },
    });
    chartRef.current = chart;
    const bars = chart.addSeries(CandlestickSeries, {
      upColor: "#19977a",
      downColor: "#cf746b",
      wickUpColor: "#19977a",
      wickDownColor: "#cf746b",
      borderVisible: false,
      priceFormat: { type: "price", precision: 2, minMove: 0.01 },
    });
    bars.setData(chartData(report.candles));
    if (volume) {
      const histogram = chart.addSeries(HistogramSeries, {
        priceFormat: { type: "volume" },
        priceScaleId: "volume",
        lastValueVisible: false,
        priceLineVisible: false,
      });
      histogram
        .priceScale()
        .applyOptions({ scaleMargins: { top: 0.82, bottom: 0 } });
      histogram.setData(
        report.candles.map((c) => ({
          time: c.date.slice(0, 10) as Time,
          value: c.volume,
          color: c.close >= c.open ? "#19977a55" : "#cf746b55",
        })),
      );
    }
    if (range && report.tradingRange) {
      for (const [price, title] of [
        [report.tradingRange.support, "Hỗ trợ"],
        [report.tradingRange.resistance, "Kháng cự"],
      ] as const)
        bars.createPriceLine({
          price,
          color: "#448e7e",
          lineWidth: 1,
          lineStyle: 2,
          axisLabelVisible: true,
          title,
        });
    }
    if (events) {
      const dates = new Set(report.candles.map((c) => c.date.slice(0, 10)));
      createSeriesMarkers(
        bars,
        report.events
          .filter((ev) => dates.has(ev.date.slice(0, 10)))
          .map((ev) => ({
            time: ev.date.slice(0, 10) as Time,
            position: "aboveBar",
            color: "#537a91",
            shape: "circle",
            text: ev.type,
          })),
      );
    }
    chart.subscribeCrosshairMove((param) => {
      const c = report.candles.find((c) => c.date.slice(0, 10) === param.time);
      setHover(c || null);
    });
    chart.subscribeClick((param) => {
      const event = report.events.find(
        (ev) => ev.date.slice(0, 10) === param.time,
      );
      if (event) onSelectEvent(event);
    });
    const updateShade = () => {
      if (!range || !report.tradingRange) {
        setShade({ top: 0, height: 0 });
        return;
      }
      const y1 = bars.priceToCoordinate(report.tradingRange.resistance),
        y2 = bars.priceToCoordinate(report.tradingRange.support);
      if (y1 !== null && y2 !== null) {
        const top = Math.max(0, y1),
          bottom = Math.min(440, y2);
        setShade({ top, height: Math.max(0, bottom - top) });
      }
    };
    chart.timeScale().subscribeVisibleLogicalRangeChange(updateShade);
    chart.timeScale().fitContent();
    const animation = requestAnimationFrame(updateShade);
    const observer = new ResizeObserver(updateShade);
    observer.observe(container.current);
    return () => {
      cancelAnimationFrame(animation);
      observer.disconnect();
      chart.remove();
      chartRef.current = null;
    };
  }, [report, dark, events, range, volume, onSelectEvent]);
  useEffect(() => {
    if (!selectedEvent || !chartRef.current) return;
    const i = report.candles.findIndex(
      (c) => c.date.slice(0, 10) === selectedEvent.date.slice(0, 10),
    );
    if (i >= 0)
      chartRef.current
        .timeScale()
        .setVisibleLogicalRange({
          from: Math.max(0, i - 24),
          to: Math.min(report.candles.length + 2, i + 24),
        });
  }, [selectedEvent, report]);
  const candle = hover || report.candles.at(-1);
  return (
    <section className="card chart-card">
      <div className="card-heading">
        <div>
          <span className="section-kicker">PRICE ACTION</span>
          <h2>
            Giá & khối lượng{" "}
            <span className="subtle-tag">{report.timeframe}</span>
          </h2>
        </div>
        <div className="chart-tools">
          <Button
            variant="ghost"
            size="icon"
            title="Xem toàn bộ dữ liệu"
            aria-label="Xem toàn bộ dữ liệu"
            onClick={() => chartRef.current?.timeScale().fitContent()}
          >
            <Expand size={16} />
          </Button>
        </div>
      </div>
      <div className="chart-toggles">
        <Layers size={14} />
        {[
          [events, setEvents, "Sự kiện Wyckoff"],
          [range, setRange, "Vùng giá"],
          [volume, setVolume, "Khối lượng"],
        ].map(([active, setter, label]) => (
          <button
            key={String(label)}
            type="button"
            aria-pressed={Boolean(active)}
            className={active ? "selected" : ""}
            onClick={() => (setter as (v: boolean) => void)(!active)}
          >
            {String(label)}
          </button>
        ))}
      </div>
      <div className="ohlc">
        <strong>{report.symbol}</strong>
        <span>{candle ? formatDate(candle.date) : "—"}</span>
        {candle &&
          (["open", "high", "low", "close"] as const).map((key, i) => (
            <span key={key}>
              {["O", "H", "L", "C"][i]}{" "}
              <b className={key === "close" ? "text-emerald" : ""}>
                {formatNumber(candle[key])}
              </b>
            </span>
          ))}
        <span>
          KL <b>{formatNumber(candle?.volume, 0)}</b>
        </span>
      </div>
      <div className="chart-stage">
        <div
          ref={container}
          className="financial-chart"
          role="img"
          aria-label={`Biểu đồ nến và khối lượng ${report.symbol}, khung ${report.timeframe}`}
        />
        {range && (
          <div
            className="range-shade"
            style={{ top: shade.top, height: shade.height }}
          />
        )}
      </div>
      <div className="chart-foot">
        <span>
          <MousePointer2 size={12} /> Kéo để di chuyển · Cuộn để phóng to
        </span>
        <span>Giá: nghìn đồng · KL: cổ phiếu</span>
      </div>
      {selectedEvent && (
        <div className="selected-event">
          <LocateFixed size={17} />
          <div>
            <strong>
              {selectedEvent.type} · {formatDate(selectedEvent.date)}
            </strong>
            <p>{selectedEvent.evidence}</p>
          </div>
        </div>
      )}
    </section>
  );
}
