import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import {
  Activity,
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
import { fetchMacroDashboard } from "../data/macro-api-client";
import {
  groupLabels,
  historyPoints,
  previousChange,
  statusLabels,
  changeUnit,
  sourceMessage,
} from "../domain/macro-live";
import { cn, formatDate, formatNumber } from "../lib/utils";
import type { MacroIndicator } from "../lib/types";
const icons: Record<string, typeof Activity> = {
  Liquidity: Wallet,
  Credit: Landmark,
  Inflation: Activity,
  Growth: TrendingUp,
  External: Globe2,
};
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
  const [group, setGroup] = useState("all"),
    [search, setSearch] = useState(""),
    [filter, setFilter] = useState("all"),
    [sort, setSort] = useState("default"),
    [selectedCode, setSelectedCode] = useState("VN_CPI_YOY"),
    [range, setRange] = useState("12");
  const dashboard = query.data;
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
    <div className="page macro-page">
      <div className="page-heading">
        <div>
          <p className="eyebrow">VIETNAM MACRO INTELLIGENCE</p>
          <h1>
            Bối cảnh lớn. <span>Dữ liệu có nguồn.</span>
          </h1>
          <p className="page-description">
            Theo dõi các kỳ công bố chính thức, độ mới và lịch sử điều chỉnh dữ
            liệu.
          </p>
        </div>
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
        </Button>
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
          <div className="live-source-banner">
            <span>
              <Database size={14} />
              <strong>Nguồn thực tế: FRED + NSO + NHNN</strong>
            </span>
            <span>
              API cập nhật lúc{" "}
              {new Date(dashboard.generatedAt).toLocaleString("vi-VN")} · Lịch
              công bố tùy chỉ báo
            </span>
          </div>
          <div className="live-overview">
            <section className="macro-score-card">
              <div>
                <span className="section-kicker">ĐỘ PHỦ DỮ LIỆU</span>
                <div className="macro-score">
                  <strong>{dashboard.available}</strong>
                  <span>/ {dashboard.enabled} chỉ báo đang tích hợp</span>
                </div>
                <p>
                  Có quan sát được lưu · Độ mới hiển thị riêng theo từng kỳ.
                </p>
              </div>
              <Database size={43} className="coverage-icon" />
            </section>
            <section className="card score-disabled">
              <ShieldCheck size={22} />
              <div>
                <span className="section-kicker">MACRO SCORE TẠM ẨN</span>
                <h2>Ưu tiên dữ liệu có thể kiểm chứng</h2>
                <p>
                  Chưa chấm điểm hoặc gán trạng thái thị trường cho dữ liệu
                  thật. Quy tắc tổng hợp và trọng số sẽ được xây dựng, kiểm thử
                  riêng.
                </p>
              </div>
            </section>
          </div>
          <div className="section-title">
            <div>
              <span className="section-kicker">CÁC NHÓM CHỈ BÁO</span>
              <h2>Độ phủ theo từng động lực thị trường</h2>
            </div>
            <span className="subtle-copy">
              Không thay dữ liệu thiếu bằng số minh họa
            </span>
          </div>
          <div className="group-grid">
            {Object.entries(groupLabels).map(([key, name]) => {
              const Icon = icons[key];
              const members = dashboard.indicators.filter(
                (i) => i.group === key,
              );
              return (
                <button
                  key={key}
                  className={cn("card group-card", group === key && "selected")}
                  aria-pressed={group === key}
                  onClick={() => setGroup(group === key ? "all" : key)}
                >
                  <div className="group-card-top">
                    <span className="group-icon">
                      <Icon size={18} />
                    </span>
                    <ArrowUpRight size={15} />
                  </div>
                  <span className="group-name">{name}</span>
                  <div className="group-score">
                    <strong>{members.filter((i) => i.latest).length}</strong>
                    <span>/ {members.length} có dữ liệu</span>
                  </div>
                  <span className="group-count">
                    {members.filter((i) => i.enabled).length} chỉ báo đã có
                    connector
                  </span>
                </button>
              );
            })}
          </div>
          <div className="macro-chart-grid">
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
                      {formatNumber(selected.latest.value)}
                      <small>{selected.unit}</small>
                    </strong>
                    <span className="subtle-copy">
                      Kỳ {selected.latest.period} · {selected.comparisonBasis}
                    </span>
                  </div>
                  <div className="macro-area-chart">
                    <ResponsiveContainer width="100%" height="100%">
                      <AreaChart
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
                        <Area
                          type="linear"
                          connectNulls={false}
                          dataKey="value"
                          name={selected.unit}
                          stroke="#1f9d84"
                          strokeWidth={2}
                          fill="url(#liveMacroFill)"
                          isAnimationActive={false}
                        />
                      </AreaChart>
                    </ResponsiveContainer>
                  </div>
                  <div className="chart-foot">
                    <span>
                      {selected.frequency === "MONTHLY"
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
                      onClick={() => setSelectedCode(i.code)}
                    >
                      <span className="alert-dot" />
                      <div>
                        <strong>{i.name}</strong>
                        <p>{statusLabels[i.status]}</p>
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
            </section>
          </div>
          <section className="card indicator-section">
            <div className="card-heading">
              <div>
                <span className="section-kicker">INDICATOR MONITOR</span>
                <h2>
                  Bảng theo dõi chỉ báo{" "}
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
                            onClick={() => setSelectedCode(i.code)}
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
                          {i.latest ? formatNumber(i.latest.value) : "—"}{" "}
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
              {selected.latest ? (
                <dl className="data-list provenance-list">
                  <div>
                    <dt>Kỳ thống kê / Cơ sở so sánh</dt>
                    <dd>
                      {selected.latest.period} / {selected.comparisonBasis}
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
