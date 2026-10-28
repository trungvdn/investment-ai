import { MacroScorePanel } from "../components/macro-score-panel";
import { useEffect, useMemo, useRef, useState } from "react";
import { useLocation } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  Area,
  ComposedChart,
  Line,
  Legend,
  CartesianGrid,
  ResponsiveContainer,
  ReferenceLine,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import {
  Activity,
  Download,
  ArrowDownRight,
  ArrowUpRight,
  ChevronDown,
  Database,
  ExternalLink,
  Globe2,
  Info,
  Landmark,
  LoaderCircle,
  RefreshCw,
  Search,
  ShieldCheck,
  SlidersHorizontal,
  TrendingUp,
  TriangleAlert,
  Wallet,
} from "lucide-react";
import { Button } from "../components/ui/button";
import {
  fetchMacroDashboard,
  fetchMacroOperations,
} from "../data/macro-api-client";
import {
  groupLabels,
  historyPoints,
  observationValue,
  previousChange,
  statusLabels,
  changeUnit,
  sourceMessage,
  sourceErrorLabels,
} from "../domain/macro-live";
import { cn, formatDate, formatNumber } from "../lib/utils";
import type { MacroIndicator } from "../lib/types";
function Status({ indicator }: { indicator: MacroIndicator }) {
  return (
    <span
      className={cn(
        "live-status",
        indicator.status === "AVAILABLE"
          ? "positive"
          : indicator.status === "NOT_IMPLEMENTED"
            ? "muted"
            : "neutral",
      )}
    >
      {statusLabels[indicator.status] || indicator.status}
    </span>
  );
}
export default function Macro() {
  const query = useQuery({
    queryKey: ["macro-dashboard"],
    queryFn: async ({ signal }) => {
      const timeout = AbortSignal.timeout(20000);
      try {
        return await fetchMacroDashboard({
          signal: AbortSignal.any([signal, timeout]),
        });
      } catch (error) {
        if (timeout.aborted)
          throw new Error(
            "Macro API phản hồi quá thời gian chờ. Vui lòng thử lại.",
          );
        throw error;
      }
    },
    staleTime: 60000,
    refetchInterval: 300000,
  });
  const operations = useQuery({
    queryKey: ["macro-operations"],
    queryFn: ({ signal }) =>
      fetchMacroOperations({
        signal: AbortSignal.any([signal, AbortSignal.timeout(20000)]),
      }),
    refetchInterval: 60000,
    staleTime: 30000,
  });
  const [group, setGroup] = useState("all"),
    [search, setSearch] = useState(""),
    [filter, setFilter] = useState("all"),
    [sort, setSort] = useState("default"),
    [selectedCode, setSelectedCode] = useState("VN_CPI_YOY"),
    [range, setRange] = useState("12");
  const dashboard = query.data;
  const { hash, key: locationKey } = useLocation();
  const scrolledLocation = useRef("");
  useEffect(() => {
    if (!hash || !dashboard || scrolledLocation.current === locationKey) return;
    const target = document.getElementById(hash.slice(1));
    if (target) { target.scrollIntoView({ behavior: "smooth", block: "start" }); scrolledLocation.current = locationKey; }
  }, [hash, locationKey, dashboard]);
  const highlights = [
    { code: "VN_DEPOSIT_RATE", label: "Lãi suất huy động", icon: Landmark },
    { code: "VN_CPI_YOY", label: "Lạm phát CPI", icon: Activity },
    { code: "VN_PMI", label: "PMI sản xuất", icon: TrendingUp },
    { code: "VN_CREDIT_GROWTH", label: "Tăng trưởng tín dụng", icon: Wallet },
  ];
  const exportData = () => {
    if (!dashboard) return;
    const escape = (v: unknown) => `"${String(v ?? "").replaceAll('"', '""')}"`;
    const records = [["Chỉ báo", "Kỳ", "Giá trị", "Cận trên", "Đơn vị", "Nguồn", "URL", "Trạng thái"], ...rows.map(i => [i.name, i.latest?.period, i.latest?.value, i.latest?.upperValue, i.unit, i.latest?.source, i.latest?.sourceUrl, i.status])];
    const url = URL.createObjectURL(new Blob(["\ufeff" + records.map(r => r.map(escape).join(",")).join("\r\n")], { type: "text/csv;charset=utf-8" }));
    const link = document.createElement("a"); link.href = url; link.download = "investment-ai-macro.csv"; link.click(); URL.revokeObjectURL(url);
  };
  const selected =
    dashboard?.indicators.find((i) => i.code === selectedCode) ||
    dashboard?.indicators.find((i) => i.latest) ||
    dashboard?.indicators[0];
  const points = useMemo(
    () => (selected ? historyPoints(selected, Number(range)) : []),
    [selected, range],
  );
  const rows = (dashboard?.indicators || [])
    .filter(
      (i) =>
        (group === "all" || i.group === group) &&
        `${i.name} ${i.code}`.toLowerCase().includes(search.toLowerCase()) &&
        (filter === "all" ||
          (filter === "available" && i.latest !== null) ||
          (filter === "attention" && i.enabled && i.status !== "AVAILABLE")),
    )
    .sort((a, b) =>
      sort === "name"
        ? a.name.localeCompare(b.name, "vi")
        : sort === "period"
          ? (b.latest?.period || "").localeCompare(a.latest?.period || "")
          : 0,
    );
  const alerts = (dashboard?.indicators || []).filter(
    (i) => i.enabled && i.status !== "AVAILABLE",
  );
  return (
    <div className="page macro-page macro-lens-page" id="macro-overview">
      <div className="page-heading">
        <div>
          <p className="eyebrow">VIETNAM MACRO INTELLIGENCE</p>
          <h1>Bức tranh kinh tế</h1>
          <p className="page-description">Theo dõi những chuyển động lớn, từ một góc nhìn rõ ràng.</p>
        </div>
        <div className="macro-heading-actions"><span className="macro-country"><Globe2 size={14}/> Việt Nam</span>
        <Button variant="outline" size="sm" disabled={!dashboard} onClick={exportData}><Download size={13}/> Xuất dữ liệu</Button>
        <Button
          variant="outline"
          size="sm"
          disabled={query.isFetching}
          onClick={() => void query.refetch()}
        >
          {query.isFetching ? (
            <LoaderCircle className="spin" size={13} />
          ) : (
            <RefreshCw size={13} />
          )}{" "}
          Tải lại
        </Button></div>
      </div>
      {query.isPending && (
        <div className="loading-report" role="status">
          <div className="loading-title">
            <LoaderCircle className="spin" size={20} />
            <div>
              <strong>Đang tải dữ liệu từ Macro API</strong>
              <p>Đọc các quan sát đã được kiểm tra và lưu trong PostgreSQL…</p>
            </div>
          </div>
          <div className="skeleton-grid">
            <div />
            <div />
          </div>
        </div>
      )}
      {query.isError && (
        <div role="alert" className="notice error">
          <TriangleAlert size={18} />
          <div>
            <strong>Không tải được dữ liệu vĩ mô</strong>
            <p>{query.error.message}</p>
            {dashboard && (
              <p>
                Đang giữ bản dữ liệu đã tải gần nhất. Thời điểm thu thập và kỳ
                thống kê vẫn được hiển thị.
              </p>
            )}
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
      {dashboard && (
        <>
          <div className="macro-ribbon"><span className="macro-ribbon-icon"><Globe2 size={23}/></span><div><strong>Dữ liệu tạo nên góc nhìn.</strong><p>Khám phá lạm phát, sản xuất và dòng vốn trong cùng một không gian.</p></div><div className="macro-data-health"><span className="health-dot"/>{dashboard.available}/{dashboard.enabled} chỉ báo có dữ liệu<small>Cập nhật {new Date(dashboard.generatedAt).toLocaleTimeString("vi-VN", { timeZone: "Asia/Ho_Chi_Minh", hour: "2-digit", minute: "2-digit" })} · {formatDate(dashboard.generatedAt)}</small></div></div>
          <div className="macro-highlights" aria-label="Các chỉ báo trọng tâm">{highlights.map(({code,label,icon:Icon}) => {
            const indicator = dashboard.indicators.find(i => i.code === code);
            const change = indicator ? previousChange(indicator) : null;
            return <button className="macro-highlight-card" key={code} onClick={() => {setSelectedCode(code); document.getElementById("macro-history")?.scrollIntoView({behavior:"smooth"});}} disabled={!indicator}>
              <div className="highlight-label"><span>{label}</span><Icon size={16}/></div>
              <div className="highlight-value">{indicator?.latest ? observationValue(indicator.latest) : "—"}<small>{indicator?.unit ?? (code === "VN_PMI" ? "điểm" : "%")}</small></div>
              <div className="highlight-bottom"><span className="highlight-change">{change === null ? "Chưa đủ kỳ so sánh" : <>{change < 0 ? <ArrowDownRight size={12}/> : <ArrowUpRight size={12}/>} {change > 0 ? "+" : ""}{formatNumber(change)} {indicator ? changeUnit(indicator) : ""}</>}<small>{indicator?.latest ? `Kỳ ${indicator.latest.period}` : "Chưa có dữ liệu"}</small></span><MiniTrend indicator={indicator}/></div>
              {indicator && indicator.status !== "AVAILABLE" && <span className="highlight-status">{statusLabels[indicator.status] || indicator.status}</span>}
            </button>;
          })}</div>
          <MacroScorePanel />
          <section className="card indicator-section" id="macro-indicators">
            <div className="card-heading">
              <div>
                <span className="section-kicker">INDICATOR MONITOR</span>
                <h2>
                  Các chỉ số kinh tế{" "}
                  <span className="tab-count">
                    {dashboard.indicators.length}
                  </span>
                </h2>
              </div>
              <span className="badge green">Dữ liệu thực tế</span>
            </div>
            <div className="indicator-filters">
              <div className="table-search">
                <Search size={16} />
                <input
                  aria-label="Tìm chỉ báo"
                  placeholder="Tìm tên hoặc mã chỉ báo…"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                />
              </div>
              <label className="filter-select">
                <SlidersHorizontal size={14} />
                <select
                  aria-label="Nhóm chỉ báo"
                  value={group}
                  onChange={(e) => setGroup(e.target.value)}
                >
                  <option value="all">Tất cả nhóm</option>
                  {Object.entries(groupLabels).map(([key, name]) => (
                    <option key={key} value={key}>
                      {name}
                    </option>
                  ))}
                </select>
              </label>
              <label className="filter-select">
                <select
                  aria-label="Trạng thái dữ liệu"
                  value={filter}
                  onChange={(e) => setFilter(e.target.value)}
                >
                  <option value="all">Tất cả trạng thái</option>
                  <option value="available">Đã có quan sát</option>
                  <option value="attention">Cần chú ý</option>
                </select>
              </label>
              <label className="filter-select">
                <select
                  aria-label="Sắp xếp chỉ báo"
                  value={sort}
                  onChange={(e) => setSort(e.target.value)}
                >
                  <option value="default">Thứ tự mặc định</option>
                  <option value="name">Theo tên</option>
                  <option value="period">Kỳ mới nhất</option>
                </select>
              </label>
            </div>
            <div className="table-scroll">
              <table className="indicator-table">
                <thead>
                  <tr>
                    <th>Chỉ báo / Nguồn</th>
                    <th>Giá trị</th>
                    <th>Kỳ thống kê</th>
                    <th>Thay đổi</th>
                    <th>Trạng thái</th>
                    <th>Thu thập gần nhất</th>
                  </tr>
                </thead>
                <tbody>
                  {rows.map((i) => {
                    const change = previousChange(i);
                    return (
                      <tr key={i.code}>
                        <td>
                          <button
                            className="indicator-name"
                            onClick={() => { setSelectedCode(i.code); document.getElementById("macro-history")?.scrollIntoView({behavior:"smooth"}); }}
                          >
                            {i.name}
                          </button>
                          <div className="source-cell">
                            {i.latest ? (
                              <a
                                href={i.latest.sourceUrl}
                                target="_blank"
                                rel="noreferrer"
                              >
                                {i.latest.source} <ExternalLink size={9} />
                              </a>
                            ) : (
                              i.description
                            )}
                          </div>
                        </td>
                        <td className="indicator-value">
                          {i.latest ? observationValue(i.latest) : "—"}{" "}
                          <small>{i.unit}</small>
                        </td>
                        <td>{i.latest?.period || "—"}</td>
                        <td>
                          {change === null ? (
                            "—"
                          ) : (
                            <span className="observation-change">
                              {change < 0 ? (
                                <ArrowDownRight size={12} />
                              ) : (
                                <ArrowUpRight size={12} />
                              )}{" "}
                              {change > 0 ? "+" : ""}
                              {formatNumber(change)}
                              <small>{changeUnit(i)}</small>
                            </span>
                          )}
                        </td>
                        <td>
                          <Status indicator={i} />
                        </td>
                        <td>
                          {i.latest ? formatDate(i.latest.collectedAt) : "—"}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
              {!rows.length && (
                <div className="tab-empty">
                  Không có chỉ báo phù hợp với bộ lọc.
                </div>
              )}
            </div>
            <div className="table-footer">
              <span>
                Hiển thị {rows.length} / {dashboard.indicators.length} chỉ báo
              </span>
              <span>
                Biến động tăng/giảm không tự động đồng nghĩa với tốt/xấu.
              </span>
            </div>
          </section>
          <div className="macro-chart-grid" id="macro-history">
            <section className="card macro-chart">
              <div className="card-heading">
                <div>
                  <span className="section-kicker">
                    LỊCH SỬ QUAN SÁT THỰC TẾ
                  </span>
                  <h2>{selected?.name || "Chọn chỉ báo"}</h2>
                </div>
                <div className="segmented-control">
                  {["6", "12", "36"].map((v) => (
                    <button
                      key={v}
                      className={range === v ? "active" : ""}
                      aria-pressed={range === v}
                      onClick={() => setRange(v)}
                    >
                      {v}T
                    </button>
                  ))}
                </div>
              </div>
              <label className="chart-indicator-select">
                <span className="sr-only">Chỉ báo trên biểu đồ</span>
                <select
                  aria-label="Chỉ báo trên biểu đồ"
                  value={selected?.code || ""}
                  onChange={(e) => setSelectedCode(e.target.value)}
                >
                  {dashboard.indicators
                    .filter((i) => i.enabled)
                    .map((i) => (
                      <option key={i.code} value={i.code}>
                        {i.name}
                      </option>
                    ))}
                </select>
                <ChevronDown size={13} />
              </label>
              {selected?.latest ? (
                <>
                  <div className="macro-chart-summary">
                    <strong>
                      {observationValue(selected.latest)}
                      <small>{selected.unit}</small>
                    </strong>
                    <span className="subtle-copy">
                      Kỳ {selected.latest.period} · {selected.comparisonBasis === "RANGE" ? "Khoảng công bố" : selected.comparisonBasis}
                    </span>
                  </div>
                  <div className="macro-area-chart">
                    <ResponsiveContainer width="100%" height="100%">
                      <ComposedChart
                        data={points}
                        margin={{ top: 10, right: 12, bottom: 0, left: -12 }}
                      >
                        <defs>
                          <linearGradient
                            id="liveMacroFill"
                            x1="0"
                            y1="0"
                            x2="0"
                            y2="1"
                          >
                            <stop
                              offset="0%"
                              stopColor="#1f9d84"
                              stopOpacity={0.2}
                            />
                            <stop
                              offset="100%"
                              stopColor="#1f9d84"
                              stopOpacity={0}
                            />
                          </linearGradient>
                        </defs>
                        <CartesianGrid
                          vertical={false}
                          stroke="var(--grid)"
                          strokeDasharray="3 4"
                        />
                        <XAxis
                          dataKey="period"
                          minTickGap={40}
                          axisLine={false}
                          tickLine={false}
                          tick={{ fill: "var(--muted)", fontSize: 10 }}
                        />
                        <YAxis
                          domain={["auto", "auto"]}
                          tickFormatter={(v) => formatNumber(Number(v))}
                          axisLine={false}
                          tickLine={false}
                          tick={{ fill: "var(--muted)", fontSize: 10 }}
                        />
                        <Tooltip
                          contentStyle={{
                            background: "var(--card)",
                            border: "1px solid var(--border)",
                            borderRadius: 9,
                            color: "var(--text)",
                            fontSize: 11,
                          }}
                        />
                        {selected.code === "VN_PMI" && (
                          <ReferenceLine
                            ifOverflow="extendDomain"
                            y={50}
                            stroke="#94a3b8"
                            strokeDasharray="4 4"
                            label="50 · Không thay đổi"
                          />
                        )}
                        {selected.comparisonBasis === "RANGE" && <Legend wrapperStyle={{ fontSize: 11 }} />}
                        {selected.comparisonBasis === "RANGE" && <Line
                              type="linear" dataKey="value" name="Cận dưới (%/năm)"
                              stroke="#1f9d84" strokeWidth={3} dot={{ r: 3 }} activeDot={{ r: 5 }}
                              connectNulls isAnimationActive={false}
                            />}
                        {selected.comparisonBasis === "RANGE" && <Line
                              type="linear" dataKey="upperValue" name="Cận trên (%/năm)"
                              stroke="#d49732" strokeWidth={3} dot={{ r: 3 }} activeDot={{ r: 5 }}
                              connectNulls isAnimationActive={false}
                            />}
                        {selected.comparisonBasis !== "RANGE" && <Area
                            type="linear" connectNulls={false} dataKey="value" name={selected.unit}
                            stroke="#1f9d84" strokeWidth={2} fill="url(#liveMacroFill)"
                            isAnimationActive={false}
                          />}
                      </ComposedChart>
                    </ResponsiveContainer>
                  </div>
                  <div className="chart-foot">
                    <span>
                      {selected.comparisonBasis === "RANGE"
                        ? "Nối các kỳ có số liệu · tháng thiếu chưa có quan sát"
                        : selected.frequency === "MONTHLY"
                        ? "Chuỗi theo tháng · giữ nguyên khoảng thiếu"
                        : selected.frequency === "QUARTERLY"
                          ? "Chuỗi theo quý · giữ nguyên khoảng thiếu"
                          : "Chuỗi theo ngày áp dụng/giao dịch"}
                    </span>
                    <Status indicator={selected} />
                  </div>
                </>
              ) : (
                <div className="tab-empty">
                  <Database size={26} />
                  <h3>Chưa có quan sát cho chỉ báo này</h3>
                  <p>
                    {selected
                      ? sourceMessage(selected)
                      : "Chờ thu thập từ nguồn chính thức."}
                  </p>
                </div>
              )}
            </section>
            <div className="macro-source-stack" id="macro-operations"><section className="card alerts-card" aria-label="Vận hành thu thập">
 <div className="card-heading"><h2>Lịch thu thập & trạng thái nguồn</h2></div>
 {operations.isError ? <p role="alert">Không tải được trạng thái vận hành nguồn.</p> : operations.data?.sources.length ? (
 <div className="macro-alerts" style={{maxHeight:360,overflowY:"auto"}}>
 {operations.data.sources.map(source=><div className="macro-alert" key={source.provider}><div>
 <strong>{source.provider}</strong>
 <p>{source.errorKind ? `${sourceErrorLabels[source.errorKind] || source.errorKind} · ${source.consecutiveFailures} lần lỗi liên tiếp` : source.leaseUntil && new Date(source.leaseUntil).getTime()>Date.now() ? "Đang thu thập" : source.lastChecked ? "Lần kiểm tra gần nhất thành công" : "Chưa chạy"}</p>
 <p>Lần tiếp theo: {new Date(source.nextDue).toLocaleString("vi-VN",{timeZone:"Asia/Ho_Chi_Minh"})} (giờ Việt Nam)</p>
 {source.errorKind && <p>{source.message}</p>}
 {source.lastSuccess && <p>Kiểm tra thành công: {new Date(source.lastSuccess).toLocaleString("vi-VN",{timeZone:"Asia/Ho_Chi_Minh"})}</p>}
 </div></div>)}
 </div>
 ) : <p>Chờ worker đăng ký lịch nguồn.</p>}
 {operations.data?.alerts.some(a=>a.status==="SCHEDULER_OVERDUE") && <p role="alert">Có nguồn đã quá lịch thu thập. Kiểm tra worker.</p>}
 </section>
 <section className="card alerts-card">
              <div className="card-heading">
                <div>
                  <span className="section-kicker">
                    CHẤT LƯỢNG & KHẢ NĂNG TRUY CẬP
                  </span>
                  <h2>Nguồn cần chú ý</h2>
                </div>
                <span className="alert-counter">{alerts.length}</span>
              </div>
              {alerts.length ? (
                <div className="macro-alerts">
                  {alerts.map((i) => (
                    <button
                      key={i.code}
                      className="macro-alert"
                      onClick={() => { setSelectedCode(i.code); document.getElementById("macro-history")?.scrollIntoView({behavior:"smooth"}); }}
                    >
                      <span className="alert-dot" />
                      <div>
                        <strong>{i.name}</strong>
                        <p>{statusLabels[i.status]}</p><p>{sourceMessage(i)}</p>
                        {i.latest && <p>Kỳ gần nhất: {i.latest.period}</p>}
                      </div>
                      <Info size={14} />
                    </button>
                  ))}
                </div>
              ) : (
                <p className="detail-assessment">
                  Các chỉ báo đang tích hợp đều có dữ liệu trong ngưỡng độ mới
                  đã cấu hình.
                </p>
              )}
              <p className="alert-note">
                <Info size={13} /> Độ mới dựa trên kỳ thống kê, không dựa vào
                việc vừa tải lại. Chỉ báo chưa sẵn sàng được giữ trống, không
                thay bằng dữ liệu mẫu.
              </p>
            </section></div>
          </div>
          {selected && (
            <section
              className="card indicator-detail"
              aria-label={`Chi tiết ${selected.name}`}
            >
              <div className="card-heading">
                <div>
                  <span className="section-kicker">
                    NGUỒN GỐC & PHƯƠNG PHÁP
                  </span>
                  <h2>{selected.name}</h2>
                </div>
                <Status indicator={selected} />
              </div>
              <p className="detail-assessment">{selected.description}</p>
              {selected.latest && selected.status !== "AVAILABLE" && (
                <p className="detail-assessment" role="status">
                  {sourceMessage(selected)} Giá trị bên dưới thuộc kỳ{" "}
                  {selected.latest.period}.
                </p>
              )}
              {selected.latest ? (
                <dl className="data-list provenance-list">
                  <div>
                    <dt>Kỳ thống kê / Cơ sở so sánh</dt>
                    <dd>
                      {selected.latest.period} / {selected.comparisonBasis === "RANGE" ? "Khoảng công bố" : selected.comparisonBasis}
                    </dd>
                  </div>
                  <div>
                    <dt>Tần suất dữ liệu</dt>
                    <dd>
                      {(
                        {
                          DAILY: "Ngày",
                          MONTHLY: "Tháng",
                          QUARTERLY: "Quý",
                        } as Record<string, string>
                      )[selected.frequency] || selected.frequency}
                    </dd>
                  </div>
                  <div>
                    <dt>Nguồn / Mã chuỗi</dt>
                    <dd>
                      {selected.latest.source} /{" "}
                      {selected.latest.sourceSeriesId}
                    </dd>
                  </div>
                  <div>
                    <dt>Ngày công bố</dt>
                    <dd>
                      {selected.latest.publishedAt
                        ? formatDate(selected.latest.publishedAt)
                        : "Nguồn chưa cung cấp ngày công bố cho quan sát này"}
                    </dd>
                  </div>
                  <div>
                    <dt>Phát hiện lần đầu</dt>
                    <dd>{formatDate(selected.latest.firstSeenAt)}</dd>
                  </div>
                  <div>
                    <dt>Thu thập gần nhất</dt>
                    <dd>
                      {new Date(selected.latest.collectedAt).toLocaleString(
                        "vi-VN",
                      )}
                    </dd>
                  </div>
                  <div>
                    <dt>Phiên bản / Tài liệu gốc</dt>
                    <dd>
                      r{selected.latest.revision} / #{selected.latest.rawId}
                    </dd>
                  </div>
                  <div>
                    <dt>Phương pháp</dt>
                    <dd>{selected.latest.method}</dd>
                  </div>
                  <div>
                    <dt>Tài liệu nguồn</dt>
                    <dd>
                      <a
                        className="source-link"
                        href={selected.latest.sourceUrl}
                        target="_blank"
                        rel="noreferrer"
                      >
                        Mở nguồn chính thức <ExternalLink size={12} />
                      </a>
                    </dd>
                  </div>
                </dl>
              ) : (
                <p className="detail-assessment">{sourceMessage(selected)}</p>
              )}
            </section>
          )}
          <p className="disclaimer">
            <ShieldCheck size={13} /> Quan sát có thể được nguồn điều chỉnh.
            Ngày thu thập khác ngày công bố. Broad USD Index của FRED khác ICE
            DXY. Chỉ báo chưa tích hợp vẫn để trống; không dùng dữ liệu mẫu để
            lấp chỗ thiếu.
          </p>
        </>
      )}
    </div>
  );
}

function MiniTrend({ indicator }: { indicator?: MacroIndicator }) {
  const observations = indicator?.history.filter(o => Number.isFinite(o.value)).slice(-12) ?? [];
  if (observations.length < 2) return <span className="mini-trend-empty">—</span>;
  const values = observations.map(o => o.value);
  const lo = Math.min(...values), span = Math.max(...values) - lo || 1;
  const path = values.map((v,i) => `${i ? "L" : "M"}${3+i*80/(values.length-1)},${30-(v-lo)/span*25}`).join(" ");
  return <svg className="highlight-sparkline" viewBox="0 0 86 36" role="img" aria-label={`${indicator?.name}: xu hướng ${values.length} kỳ có dữ liệu${indicator?.comparisonBasis === "RANGE" ? ", cận dưới" : ""}`}><path d={path} stroke="currentColor" strokeWidth="2" fill="none"/><circle cx="83" cy={30-(values.at(-1)!-lo)/span*25} r="2.5" fill="currentColor"/></svg>;
}
