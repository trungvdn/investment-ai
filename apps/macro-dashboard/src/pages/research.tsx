import { useCallback, useState } from "react";
import { useOutletContext, useSearchParams, Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  ArrowDownRight,
  ArrowRight,
  ArrowUpRight,
  Check,
  ChevronRight,
  ChartCandlestick,
  Clock3,
  FileSearch,
  FlaskConical,
  Info,
  LoaderCircle,
  Search,
  ShieldCheck,
  Target,
  TriangleAlert,
  TrendingUp,
  Activity,
} from "lucide-react";
import { Button } from "../components/ui/button";
import {
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
} from "../components/ui/tabs";
import { PriceChart } from "../components/price-chart";
import { fetchStockReport } from "../data/stock-api-client.js";
import {
  dateRange,
  normalizeSymbol,
  structureLabels,
  scenarios,
} from "../domain/wyckoff-report.js";
import { formatDate, formatNumber, cn } from "../lib/utils";
import type { StockReport, AnalysisRequest, WyckoffEvent } from "../lib/types";
function initialRequest(params: URLSearchParams): AnalysisRequest | null {
  try {
    const symbol = params.get("symbol");
    if (!symbol) return null;
    const timeframe = params.get("timeframe") === "1W" ? "1W" : "1D";
    const period = ["6", "12", "24"].includes(params.get("period") || "")
      ? params.get("period")!
      : "24";
    return {
      symbol: normalizeSymbol(symbol),
      timeframe,
      period,
      ...dateRange(period),
    };
  } catch {
    return null;
  }
}
function EvidenceList({
  items,
  tone = "positive",
}: {
  items: string[];
  tone?: "positive" | "warning";
}) {
  return (
    <ul className={cn("evidence-list", tone)}>
      {items.length ? (
        items.map((item, i) => (
          <li key={i}>
            {tone === "positive" ? (
              <Check size={15} />
            ) : (
              <TriangleAlert size={15} />
            )}
            <span>{item}</span>
          </li>
        ))
      ) : (
        <li>
          <Info size={15} />
          <span>Chưa có bằng chứng được cung cấp.</span>
        </li>
      )}
    </ul>
  );
}
export default function Research() {
  const { dark } = useOutletContext<{ dark: boolean }>();
  const [params, setParams] = useSearchParams();
  const [request, setRequest] = useState<AnalysisRequest | null>(() =>
    initialRequest(params),
  );
  const [symbol, setSymbol] = useState(params.get("symbol") || ""),
    [timeframe, setTimeframe] = useState<"1D" | "1W">(
      params.get("timeframe") === "1W" ? "1W" : "1D",
    ),
    [period, setPeriod] = useState(request?.period || "24");
  const [validation, setValidation] = useState(""),
    [tab, setTab] = useState("overview"),
    [selectedEvent, setSelectedEvent] = useState<WyckoffEvent | null>(null);
  const onSelectEvent = useCallback(
    (event: WyckoffEvent) => setSelectedEvent(event),
    [],
  );
  const query = useQuery({
    queryKey: ["stock-report", request],
    enabled: !!request,
    staleTime: 60000,
    queryFn: async ({ signal }) => {
      if (!request) throw new Error("Thiếu yêu cầu phân tích.");
      const timeout = AbortSignal.timeout(45000);
      try {
        return (await fetchStockReport({
          ...request,
          signal: AbortSignal.any([signal, timeout]),
        })) as StockReport;
      } catch (error) {
        if (timeout.aborted)
          throw new Error("Yêu cầu quá thời gian chờ. Vui lòng thử lại.");
        throw error;
      }
    },
  });
  function submit(value = symbol) {
    try {
      const valid = normalizeSymbol(value);
      const next = { symbol: valid, timeframe, period, ...dateRange(period) };
      setValidation("");
      setSymbol(valid);
      setSelectedEvent(null);
      setTab("overview");
      if (JSON.stringify(next) === JSON.stringify(request))
        void query.refetch();
      else setRequest(next);
      setParams({ symbol: valid, timeframe, period }, { replace: true });
    } catch (error) {
      setValidation((error as Error).message);
    }
  }
  const r = query.data,
    last = r?.candles.at(-1),
    previous = r?.candles.at(-2);
  const delta = last && previous ? last.close - previous.close : undefined;
  const pct =
    delta !== undefined && previous?.close
      ? (delta / previous.close) * 100
      : undefined;
  const bullish = r?.structure === "ACCUMULATION" || r?.structure === "MARKUP";
  return (
    <div className="page research-page">
      <div className="page-heading">
        <div>
          <p className="eyebrow">WYCKOFF RESEARCH</p>
          <h1>
            Đọc dòng tiền. <span>Hiểu thị trường.</span>
          </h1>
          <p className="page-description">
            Từ cấu trúc giá và khối lượng đến những luận điểm có thể kiểm chứng.
          </p>
        </div>
        <span className="heading-tag">
          <ShieldCheck size={14} /> Evidence-first analysis
        </span>
      </div>
      <section className="search-card">
        <form
          onSubmit={(event) => {
            event.preventDefault();
            submit();
          }}
        >
          <label className="symbol-field">
            <span>MÃ CỔ PHIẾU</span>
            <div className="search-input">
              <Search size={19} />
              <input
                aria-label="Mã cổ phiếu"
                value={symbol}
                onChange={(event) =>
                  setSymbol(event.target.value.toUpperCase())
                }
                placeholder="Tìm mã cổ phiếu…"
                maxLength={10}
                required
              />
              <kbd>↵</kbd>
            </div>
          </label>
          <label>
            <span>KHUNG PHÂN TÍCH</span>
            <select
              aria-label="Khung phân tích"
              value={timeframe}
              onChange={(event) => {
                const frame = event.target.value as "1D" | "1W";
                setTimeframe(frame);
                if (frame === "1W") setPeriod("24");
              }}
            >
              <option value="1D">Ngày · 1D</option>
              <option value="1W">Tuần · 1W</option>
            </select>
          </label>
          <label>
            <span>KHOẢNG DỮ LIỆU</span>
            <select
              aria-label="Khoảng dữ liệu"
              value={period}
              onChange={(event) => setPeriod(event.target.value)}
            >
              <option value="6">6 tháng</option>
              <option value="12">1 năm</option>
              <option value="24">2 năm</option>
            </select>
          </label>
          <Button type="submit" disabled={query.isFetching}>
            {query.isFetching ? (
              <LoaderCircle size={16} className="spin" />
            ) : (
              <FileSearch size={16} />
            )}{" "}
            Phân tích <ArrowUpRight size={16} />
          </Button>
        </form>
        <div className="search-meta">
          <span>Khám phá nhanh</span>
          <div>
            {["FPT", "VCB", "MBB", "VCI"].map((s) => (
              <button
                key={s}
                onClick={() => submit(s)}
                disabled={query.isFetching}
              >
                {s}
                <ArrowUpRight size={11} />
              </button>
            ))}
          </div>
          <span className="search-meta-note">
            <Clock3 size={12} /> Dữ liệu cuối phiên · Tối thiểu 60 nến
          </span>
        </div>
      </section>
      {validation && (
        <div role="alert" className="notice error">
          <TriangleAlert size={18} />
          {validation}
        </div>
      )}
      {query.isFetching && (
        <div className="loading-report" role="status">
          <div className="loading-title">
            <LoaderCircle className="spin" size={20} />
            <div>
              <strong>Đang nghiên cứu {request?.symbol}</strong>
              <p>
                Lấy dữ liệu OHLCV, nhận diện cấu trúc và tổng hợp bằng chứng…
              </p>
            </div>
          </div>
          <div className="skeleton-grid">
            <div />
            <div />
          </div>
        </div>
      )}
      {query.isError && !query.isFetching && (
        <div role="alert" className="notice error">
          <TriangleAlert size={20} />
          <div>
            <strong>Chưa thể tạo báo cáo</strong>
            <p>{query.error.message}</p>
            <Button
              variant="outline"
              size="sm"
              onClick={() => void query.refetch()}
            >
              Thử lại
            </Button>
          </div>
        </div>
      )}
      {!request && (
        <div className="research-welcome">
          <div className="welcome-copy">
            <span className="welcome-icon">
              <ChartCandlestick size={28} />
            </span>
            <p className="eyebrow">BẮT ĐẦU TỪ MỘT CÂU HỎI ĐÚNG</p>
            <h2>
              Giá đang kể
              <br />
              câu chuyện gì?
            </h2>
            <p>
              Chọn một mã cổ phiếu. Khám phá cấu trúc thị trường, đối chiếu bằng
              chứng và xác định điều cần chờ.
            </p>
            <div className="welcome-caption">
              <span /> Không đoán trước. Theo dõi để xác nhận.
            </div>
          </div>
          <div className="research-method">
            <div>
              <span className="step-num">01</span>
              <ChartCandlestick size={21} />
              <h3>Quan sát cấu trúc</h3>
              <p>Đặt diễn biến giá và khối lượng vào bối cảnh thị trường.</p>
            </div>
            <div>
              <span className="step-num">02</span>
              <FileSearch size={21} />
              <h3>Đối chiếu bằng chứng</h3>
              <p>Nhận diện sự kiện Wyckoff và những điểm còn mâu thuẫn.</p>
            </div>
            <div>
              <span className="step-num">03</span>
              <Target size={21} />
              <h3>Theo dõi xác nhận</h3>
              <p>Xây dựng kịch bản và điều kiện làm suy yếu giả thuyết.</p>
            </div>
          </div>
          <Link className="macro-link" to="/macro">
            <Activity size={18} />
            <div>
              <strong>Thị trường không vận động riêng lẻ.</strong>
              <p>Khám phá bối cảnh vĩ mô trước khi đi sâu vào từng cổ phiếu.</p>
            </div>
            <ArrowRight size={19} />
          </Link>
        </div>
      )}
      {r && !query.isFetching && !query.isError && (
        <>
          {r.dataQuality.mode === "demo" && (
            <div className="notice demo">
              <FlaskConical size={17} />
              <span>
                <strong>Dữ liệu minh họa</strong> · Nến tổng hợp phục vụ kiểm
                tra giao diện, không phản ánh cổ phiếu thực tế.
              </span>
            </div>
          )}
          <div className="stock-heading">
            <div className="stock-identity">
              <span className="stock-monogram">{r.symbol.slice(0, 3)}</span>
              <div>
                <div className="stock-title">
                  <h2>{r.symbol}</h2>
                  <span className="badge">WYCKOFF REPORT</span>
                </div>
                <p>
                  Dữ liệu đến {formatDate(r.asOf)}
                  <span>·</span>
                  {r.candles.length} nến{" "}
                  {r.timeframe === "1D" ? "ngày" : "tuần"}
                </p>
              </div>
            </div>
            <div className="stock-price">
              <strong>{formatNumber(last?.close)}</strong>
              <span
                className={cn(
                  "price-change",
                  (delta || 0) < 0 ? "negative" : "positive",
                )}
              >
                {delta !== undefined && delta < 0 ? (
                  <ArrowDownRight size={15} />
                ) : (
                  <ArrowUpRight size={15} />
                )}{" "}
                {delta === undefined
                  ? "—"
                  : `${delta >= 0 ? "+" : ""}${formatNumber(delta)} (${formatNumber(pct)}%)`}
              </span>
              <small>Nghìn đồng · So với nến trước</small>
            </div>
          </div>
          <div
            className={cn(
              "thesis-banner",
              r.structure === "UNCLEAR" && "uncertain",
            )}
          >
            <span className="thesis-icon">
              {r.structure === "UNCLEAR" ? (
                <Info size={23} />
              ) : (
                <TrendingUp size={23} />
              )}
            </span>
            <div>
              <p className="section-kicker">LUẬN ĐIỂM HIỆN TẠI</p>
              <h2>{structureLabels[r.structure]}</h2>
              <p>
                {r.structure === "UNCLEAR"
                  ? "Chưa đủ bằng chứng để kết luận cấu trúc và pha. Tiếp tục quan sát giá, khối lượng và trình tự sự kiện."
                  : `Giả thuyết ${bullish ? "nghiêng về sức mạnh của cầu" : "nghiêng về áp lực cung"}. Cần diễn biến tiếp theo xác nhận trước khi đưa ra kết luận mạnh hơn.`}
              </p>
            </div>
            <span className="badge amber">Cần xác nhận</span>
          </div>
          <div id="chart-anchor" className="analysis-grid">
            <PriceChart
              report={r}
              dark={dark}
              selectedEvent={selectedEvent}
              onSelectEvent={onSelectEvent}
            />
            <aside className="card interpretation">
              <div className="card-heading">
                <div>
                  <span className="section-kicker">THE BIG PICTURE</span>
                  <h2>Cấu trúc thị trường</h2>
                </div>
                <Target size={19} />
              </div>
              <div className="structure-name">
                {structureLabels[r.structure]}
              </div>
              <div className="phase-label">
                <span>Pha giả định</span>
                <strong>
                  {r.phase === "UNKNOWN" ? "Chưa xác định" : `Phase ${r.phase}`}
                </strong>
              </div>
              <div className="phase-steps">
                {["A", "B", "C", "D", "E"].map((p) => (
                  <span key={p} className={r.phase === p ? "active" : ""}>
                    {p}
                  </span>
                ))}
              </div>
              <div className="confidence-heading">
                <span>Mức khớp quy tắc</span>
                <strong>
                  {formatNumber(r.confidence * 100, 0)}
                  <small>/100</small>
                </strong>
              </div>
              <div className="confidence-bar">
                <i style={{ width: `${r.confidence * 100}%` }} />
              </div>
              <p className="metric-help">
                Điểm heuristic, không phải xác suất thành công.
              </p>
              <dl className="range-metrics">
                <div>
                  <dt>Hỗ trợ</dt>
                  <dd>{formatNumber(r.tradingRange?.support)}</dd>
                </div>
                <div>
                  <dt>Kháng cự</dt>
                  <dd>{formatNumber(r.tradingRange?.resistance)}</dd>
                </div>
                <div>
                  <dt>Trạng thái vùng</dt>
                  <dd>
                    {!r.tradingRange
                      ? "Chưa xác định"
                      : r.tradingRange.status === "IDENTIFIED"
                        ? "Được nhận diện"
                        : "Tạm thời"}
                  </dd>
                </div>
              </dl>
              <div className="next-observation">
                <span>
                  <Target size={14} /> ĐIỀU CẦN CHỜ
                </span>
                <p>
                  {r.watchlist[0] ||
                    "Chờ thêm bằng chứng từ giá và khối lượng."}
                </p>
              </div>
              <div className="hypothesis-risk">
                <TriangleAlert size={14} />
                <span>
                  {r.risks[1] ||
                    r.risks[0] ||
                    "Giả thuyết có thể thay đổi khi có dữ liệu mới."}
                </span>
              </div>
            </aside>
          </div>
          <Tabs value={tab} onValueChange={setTab} className="report-tabs">
            <TabsList aria-label="Chi tiết báo cáo">
              <TabsTrigger value="overview">Luận điểm & bằng chứng</TabsTrigger>
              <TabsTrigger value="events">
                Sự kiện Wyckoff{" "}
                <span className="tab-count">{r.events.length}</span>
              </TabsTrigger>
              <TabsTrigger value="scenarios">Kịch bản theo dõi</TabsTrigger>
              <TabsTrigger value="data">Dữ liệu & phương pháp</TabsTrigger>
            </TabsList>
            <TabsContent value="overview">
              <div className="evidence-grid">
                <section className="card">
                  <div className="card-heading">
                    <h2>
                      <ShieldCheck size={18} className="text-emerald" /> Bằng
                      chứng hỗ trợ
                    </h2>
                    <span className="badge green">
                      {r.evidence.length} luận cứ
                    </span>
                  </div>
                  <EvidenceList items={r.evidence} />
                </section>
                <section className="card">
                  <div className="card-heading">
                    <h2>
                      <TriangleAlert size={18} className="text-amber" /> Rủi ro
                      & mâu thuẫn
                    </h2>
                  </div>
                  <EvidenceList items={r.risks} tone="warning" />
                </section>
              </div>
              <section className="card watchlist-card">
                <h2>
                  <Target size={18} /> Các điều kiện cần theo dõi
                </h2>
                <EvidenceList items={r.watchlist} />
              </section>
            </TabsContent>
            <TabsContent value="events">
              <div className="tab-intro">
                <h2>Dòng thời gian sự kiện</h2>
                <p>
                  Chọn sự kiện để đưa biểu đồ tới vùng giá tương ứng. Sự kiện
                  nhận diện là giả thuyết theo quy tắc.
                </p>
              </div>
              {!r.events.length ? (
                <div className="card tab-empty">
                  <FileSearch size={25} />
                  <h3>Chưa nhận diện sự kiện rõ ràng</h3>
                  <p>
                    Không có sự kiện Wyckoff được báo cáo trong cửa sổ phân
                    tích.
                  </p>
                </div>
              ) : (
                <div className="event-timeline">
                  {r.events.map((event, i) => (
                    <button
                      key={`${event.date}-${event.type}-${i}`}
                      className={cn(
                        "card timeline-item",
                        selectedEvent === event && "selected",
                      )}
                      onClick={() => {
                        setSelectedEvent(event);
                        document
                          .getElementById("chart-anchor")
                          ?.scrollIntoView({
                            behavior: "smooth",
                            block: "start",
                          });
                      }}
                    >
                      <span className="timeline-date">
                        {formatDate(event.date)}
                      </span>
                      <span className="event-code">{event.type}</span>
                      <div>
                        <h3>
                          {event.type}{" "}
                          <span className="badge amber">Giả thuyết</span>
                        </h3>
                        <p>{event.evidence}</p>
                        <small>
                          Khớp quy tắc {formatNumber(event.confidence * 100, 0)}
                          /100
                        </small>
                      </div>
                      <ChevronRight size={18} />
                    </button>
                  ))}
                </div>
              )}
            </TabsContent>
            <TabsContent value="scenarios">
              <div className="tab-intro">
                <h2>Chuẩn bị cho nhiều khả năng</h2>
                <p>
                  Điều kiện quan sát dựa trên vùng giá{" "}
                  {r.tradingRange?.status === "IDENTIFIED"
                    ? "được nhận diện"
                    : "tạm thời"}
                  ; không phải dự báo hay lệnh giao dịch.
                </p>
              </div>
              <div className="scenario-grid">
                {scenarios(r.tradingRange).map((s, i) => (
                  <article
                    key={s.title}
                    className={cn("card scenario", s.tone)}
                  >
                    <span className="scenario-icon">
                      {i === 0 ? (
                        <ArrowUpRight />
                      ) : i === 1 ? (
                        <ArrowRight />
                      ) : (
                        <ArrowDownRight />
                      )}
                    </span>
                    <span className="section-kicker">KỊCH BẢN 0{i + 1}</span>
                    <h3>{s.title.split(" / ")[1]}</h3>
                    <div>
                      <h4>Điều kiện kích hoạt</h4>
                      <p>{s.trigger}</p>
                    </div>
                    <div>
                      <h4>Dấu hiệu cần theo dõi</h4>
                      <p>{s.watch}</p>
                    </div>
                    <div className="scenario-invalid">
                      <h4>Vô hiệu khi</h4>
                      <p>{s.invalid}</p>
                    </div>
                  </article>
                ))}
              </div>
              {!r.tradingRange && (
                <div className="card tab-empty">
                  Chưa xác định trading range để xây dựng kịch bản theo mức giá.
                </div>
              )}
            </TabsContent>
            <TabsContent value="data">
              <div className="evidence-grid">
                <section className="card">
                  <h2>Nguồn & chất lượng dữ liệu</h2>
                  <dl className="data-list">
                    <div>
                      <dt>Nguồn</dt>
                      <dd>
                        {r.dataQuality.mode === "demo"
                          ? "Nến tổng hợp (demo)"
                          : "KB Securities · endpoint công khai"}
                      </dd>
                    </div>
                    <div>
                      <dt>Thời điểm</dt>
                      <dd>{formatDate(r.asOf)}</dd>
                    </div>
                    <div>
                      <dt>Số nến / Khung</dt>
                      <dd>
                        {r.dataQuality.dataPoints} / {r.timeframe}
                      </dd>
                    </div>
                    <div>
                      <dt>Chất lượng / Schema</dt>
                      <dd>
                        {r.dataQuality.status} / {r.schemaVersion}
                      </dd>
                    </div>
                    <div>
                      <dt>Khoảng yêu cầu</dt>
                      <dd>
                        {request?.from} → {request?.to}
                      </dd>
                    </div>
                  </dl>
                  <EvidenceList items={r.dataQuality.notes} />
                </section>
                <section className="card">
                  <h2>Phương pháp & giới hạn</h2>
                  <EvidenceList
                    items={[
                      "Phân tích heuristic dựa trên cấu trúc giá, khối lượng và trình tự sự kiện Wyckoff. Không ép nhận định khi bằng chứng chưa đủ.",
                      "Khung tuần tổng hợp OHLCV theo tuần ISO trước khi phân tích. Tuần cuối có thể chưa hoàn tất.",
                      "Số nến đủ không bảo đảm dữ liệu liên tục. Chính sách điều chỉnh cổ tức/quyền của nguồn dữ liệu chưa được bảo đảm.",
                      "Bằng chứng và rủi ro gốc được giữ nguyên từ bộ phân tích để có thể đối chiếu.",
                    ]}
                    tone="warning"
                  />
                  <details className="original-summary">
                    <summary>Nhận định gốc từ bộ phân tích</summary>
                    <p>{r.summary}</p>
                  </details>
                </section>
              </div>
            </TabsContent>
          </Tabs>
          <p className="disclaimer">
            <Info size={13} />
            {r.disclaimer}
          </p>
        </>
      )}
    </div>
  );
}
