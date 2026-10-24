import { useEffect, useMemo, useRef, useState } from "react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import {
  Activity,
  ArrowDownRight,
  ArrowRight,
  ArrowUpRight,
  ChevronDown,
  Globe2,
  Info,
  Landmark,
  Search,
  ShieldCheck,
  SlidersHorizontal,
  TrendingUp,
  TriangleAlert,
  Wallet,
  X,
} from "lucide-react";
import { Button } from "../components/ui/button";
import { buildMacroDashboard } from "../domain/macro-dashboard.js";
import {
  asOf,
  mockIndicators,
  mockScoreHistory,
} from "../data/mock-macro-data.js";
import { cn, formatNumber } from "../lib/utils";
const groups: Record<string, { name: string; icon: typeof Activity }> = {
  Liquidity: { name: "Thanh khoản", icon: Wallet },
  Credit: { name: "Tín dụng", icon: Landmark },
  Inflation: { name: "Lạm phát", icon: Activity },
  Growth: { name: "Tăng trưởng", icon: TrendingUp },
  External: { name: "Quốc tế", icon: Globe2 },
};
const regimes: Record<string, string> = {
  Supportive: "Thuận lợi",
  Balanced: "Cân bằng",
  Cautious: "Thận trọng",
  Restrictive: "Thắt chặt",
};
type Indicator = {
  name: string;
  group: string;
  value: number;
  unit: string;
  score: number;
  history: number[];
  assessment: string;
  trend: string;
};
type MacroDashboard = {
  macroScore: number;
  regime: string;
  groupScores: { name: string; score: number; count: number }[];
  indicators: Indicator[];
  scoreHistory: { label: string; score: number }[];
  macroTrend: string;
};
const tone = (score: number) =>
  score >= 65 ? "positive" : score >= 50 ? "neutral" : "negative";
function Trend({ trend }: { trend: string }) {
  return (
    <span
      className={cn(
        "trend",
        trend === "Rising"
          ? "positive"
          : trend === "Falling"
            ? "negative"
            : "neutral",
      )}
    >
      {trend === "Rising" ? (
        <ArrowUpRight size={14} />
      ) : trend === "Falling" ? (
        <ArrowDownRight size={14} />
      ) : (
        <ArrowRight size={14} />
      )}{" "}
      {trend === "Rising" ? "Tăng" : trend === "Falling" ? "Giảm" : "Ổn định"}
    </span>
  );
}
export default function Macro() {
  const dashboard = useMemo(
    () =>
      buildMacroDashboard(mockIndicators, mockScoreHistory) as MacroDashboard,
    [],
  );
  const [group, setGroup] = useState("all"),
    [search, setSearch] = useState(""),
    [sort, setSort] = useState("default"),
    [range, setRange] = useState("6"),
    [selected, setSelected] = useState<Indicator | null>(null);
  const detailRef = useRef<HTMLElement>(null);
  useEffect(() => {
    if (selected)
      detailRef.current?.scrollIntoView({ behavior: "smooth", block: "start" });
  }, [selected]);
  const indicators = dashboard.indicators
    .filter(
      (i) =>
        (group === "all" || i.group === group) &&
        `${i.name} ${i.assessment}`
          .toLowerCase()
          .includes(search.toLowerCase()),
    )
    .sort((a, b) =>
      sort === "asc"
        ? a.score - b.score
        : sort === "desc"
          ? b.score - a.score
          : 0,
    );
  const alerts = dashboard.indicators
    .filter((i) => i.score < 55)
    .sort((a, b) => a.score - b.score);
  const chartData = dashboard.scoreHistory.slice(-Number(range));
  return (
    <div className="page macro-page">
      <div className="page-heading">
        <div>
          <p className="eyebrow">VIETNAM MACRO INTELLIGENCE</p>
          <h1>
            Bối cảnh lớn. <span>Góc nhìn rõ hơn.</span>
          </h1>
          <p className="page-description">
            Theo dõi thanh khoản, chu kỳ kinh tế và những lực tác động lên thị
            trường.
          </p>
        </div>
        <span className="heading-tag">
          <Globe2 size={14} /> Vietnam market
        </span>
      </div>
      <div className="macro-data-notice">
        <span>
          <span className="demo-dot" />
          <strong>Dữ liệu minh họa</strong> · {asOf}
        </span>
        <span>15 chỉ báo · 5 nhóm · Không phải dữ liệu hiện tại</span>
      </div>
      <div className="macro-overview">
        <section className="macro-score-card">
          <div>
            <span className="section-kicker">ĐIỂM VĨ MÔ TỔNG HỢP</span>
            <div className="macro-score">
              <strong>{dashboard.macroScore}</strong>
              <span>/ 100</span>
            </div>
            <div className="macro-score-trend">
              <ArrowDownRight size={15} />
              <span>Giảm so với tháng trước</span>
            </div>
            <p>Bình quân điểm của 5 nhóm chỉ báo.</p>
          </div>
          <div
            className="score-ring"
            style={
              { "--score": `${dashboard.macroScore}%` } as React.CSSProperties
            }
          >
            <span>
              {dashboard.macroScore}
              <small>MACRO SCORE</small>
            </span>
          </div>
        </section>
        <section className="card regime-card">
          <span className="section-kicker">TRẠNG THÁI MÔI TRƯỜNG</span>
          <div className="regime-title">
            <span className="regime-dot" />
            <h2>{regimes[dashboard.regime]}</h2>
            <span className="badge">{dashboard.regime}</span>
          </div>
          <p>
            Tăng trưởng và thanh khoản tạo nền hỗ trợ; điều kiện bên ngoài vẫn
            cần được theo dõi.
          </p>
          <div className="regime-scale">
            <span />
            <span />
            <span className="active" />
            <span />
          </div>
          <div className="scale-labels">
            <span>Thắt chặt</span>
            <span>Thuận lợi</span>
          </div>
        </section>
      </div>
      <div className="section-title">
        <div>
          <span className="section-kicker">CÁC ĐỘNG LỰC CHÍNH</span>
          <h2>Nhìn thị trường qua 5 lăng kính</h2>
        </div>
        <span className="subtle-copy">Chọn nhóm để lọc chỉ báo bên dưới</span>
      </div>
      <div className="group-grid">
        {dashboard.groupScores.map((g) => {
          const Icon = groups[g.name].icon;
          return (
            <button
              key={g.name}
              className={cn("card group-card", group === g.name && "selected")}
              aria-pressed={group === g.name}
              onClick={() => setGroup(group === g.name ? "all" : g.name)}
            >
              <div className="group-card-top">
                <span className="group-icon">
                  <Icon size={18} />
                </span>
                <ArrowUpRight size={15} />
              </div>
              <span className="group-name">{groups[g.name].name}</span>
              <div className="group-score">
                <strong className={tone(g.score)}>{g.score}</strong>
                <span>/100</span>
              </div>
              <div className="group-progress">
                <i className={tone(g.score)} style={{ width: `${g.score}%` }} />
              </div>
              <span className="group-count">
                {g.count} chỉ báo đang theo dõi
              </span>
            </button>
          );
        })}
      </div>
      <div className="macro-chart-grid">
        <section className="card macro-chart">
          <div className="card-heading">
            <div>
              <span className="section-kicker">THEO DÕI CHUYỂN ĐỘNG</span>
              <h2>Xu hướng điểm vĩ mô</h2>
            </div>
            <div className="segmented-control">
              {["3", "6"].map((v) => (
                <button
                  key={v}
                  aria-pressed={range === v}
                  className={range === v ? "active" : ""}
                  onClick={() => setRange(v)}
                >
                  {v} tháng
                </button>
              ))}
            </div>
          </div>
          <div className="macro-chart-summary">
            <strong>
              {dashboard.scoreHistory.at(-1)?.score}
              <small>/100</small>
            </strong>
            <span className="negative">−1 điểm trong tháng gần nhất</span>
          </div>
          <div className="macro-area-chart">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart
                data={chartData}
                margin={{ top: 12, right: 12, bottom: 0, left: -24 }}
              >
                <defs>
                  <linearGradient id="macroFill" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#1f9d84" stopOpacity={0.22} />
                    <stop offset="100%" stopColor="#1f9d84" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid
                  vertical={false}
                  stroke="var(--grid)"
                  strokeDasharray="3 4"
                />
                <XAxis
                  dataKey="label"
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: "var(--muted)", fontSize: 11 }}
                  dy={8}
                />
                <YAxis
                  domain={[40, 80]}
                  axisLine={false}
                  tickLine={false}
                  tick={{ fill: "var(--muted)", fontSize: 10 }}
                />
                <Tooltip
                  contentStyle={{
                    background: "var(--card)",
                    border: "1px solid var(--border)",
                    borderRadius: 10,
                    color: "var(--text)",
                    fontSize: 12,
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="score"
                  name="Điểm vĩ mô"
                  stroke="#1f9d84"
                  strokeWidth={2.5}
                  fill="url(#macroFill)"
                  dot={{
                    r: 3,
                    fill: "#1f9d84",
                    strokeWidth: 3,
                    stroke: "var(--card)",
                  }}
                  activeDot={{ r: 5 }}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
          <div className="chart-foot">
            <span>Minh họa · Tháng 3–8/2026</span>
            <span>Thang điểm 0–100</span>
          </div>
        </section>
        <section className="card alerts-card">
          <div className="card-heading">
            <div>
              <span className="section-kicker">ĐIỂM CẦN CHÚ Ý</span>
              <h2>Áp lực lên thị trường</h2>
            </div>
            <span className="alert-counter">{alerts.length}</span>
          </div>
          <div className="macro-alerts">
            {alerts.map((a) => (
              <button
                key={a.name}
                className="macro-alert"
                onClick={() => setSelected(a)}
              >
                <span className={cn("alert-dot", tone(a.score))} />
                <div>
                  <strong>{a.name}</strong>
                  <p>{a.assessment}</p>
                </div>
                <span className={cn("alert-score", tone(a.score))}>
                  {a.score}
                </span>
              </button>
            ))}
          </div>
          <p className="alert-note">
            <Info size={13} /> Các chỉ báo có điểm dưới 55 trong bộ dữ liệu minh
            họa.
          </p>
        </section>
      </div>
      <section className="card indicator-section">
        <div className="card-heading">
          <div>
            <span className="section-kicker">INDICATOR MONITOR</span>
            <h2>
              Bảng theo dõi chỉ báo <span className="tab-count">15</span>
            </h2>
          </div>
          <span className="badge amber">Dữ liệu minh họa</span>
        </div>
        <div className="indicator-filters">
          <div className="table-search">
            <Search size={16} />
            <input
              aria-label="Tìm chỉ báo"
              placeholder="Tìm tên chỉ báo…"
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
              {Object.entries(groups).map(([key, g]) => (
                <option key={key} value={key}>
                  {g.name}
                </option>
              ))}
            </select>
          </label>
          <label className="filter-select">
            <select
              aria-label="Sắp xếp chỉ báo"
              value={sort}
              onChange={(e) => setSort(e.target.value)}
            >
              <option value="default">Thứ tự mặc định</option>
              <option value="asc">Điểm thấp trước</option>
              <option value="desc">Điểm cao trước</option>
            </select>
            <ChevronDown size={13} />
          </label>
        </div>
        <div className="table-scroll">
          <table className="indicator-table">
            <thead>
              <tr>
                <th>Chỉ báo</th>
                <th>Nhóm</th>
                <th>Giá trị gần nhất</th>
                <th>Xu hướng</th>
                <th>Điểm</th>
                <th>Đánh giá</th>
                <th>
                  <span className="sr-only">Chi tiết</span>
                </th>
              </tr>
            </thead>
            <tbody>
              {indicators.map((i) => (
                <tr key={i.name}>
                  <td>
                    <button
                      className="indicator-name"
                      onClick={() => setSelected(i)}
                    >
                      {i.name}
                    </button>
                  </td>
                  <td>
                    <span className="group-chip">{groups[i.group].name}</span>
                  </td>
                  <td className="indicator-value">
                    {formatNumber(i.value)} <small>{i.unit}</small>
                  </td>
                  <td>
                    <Trend trend={i.trend} />
                  </td>
                  <td>
                    <span className={cn("score-badge", tone(i.score))}>
                      {i.score}
                    </span>
                  </td>
                  <td className="assessment">{i.assessment}</td>
                  <td>
                    <button
                      className="row-detail"
                      aria-label={`Xem chi tiết ${i.name}`}
                      onClick={() => setSelected(i)}
                    >
                      <ArrowUpRight size={15} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {!indicators.length && (
            <div className="tab-empty">
              Không có chỉ báo phù hợp. Hãy thay đổi bộ lọc.
            </div>
          )}
        </div>
        <div className="table-footer">
          <span>Hiển thị {indicators.length} / 15 chỉ báo</span>
          <span>
            Chọn chỉ báo để xem lịch sử 6 tháng <ArrowRight size={12} />
          </span>
        </div>
      </section>
      {selected && (
        <section
          ref={detailRef}
          className="card indicator-detail"
          aria-label={`Chi tiết ${selected.name}`}
        >
          <div className="card-heading">
            <div>
              <span className="section-kicker">LỊCH SỬ CHỈ BÁO · MINH HỌA</span>
              <h2>
                {selected.name}{" "}
                <span className="subtle-tag">{selected.unit}</span>
              </h2>
            </div>
            <Button
              variant="ghost"
              size="icon"
              aria-label="Đóng chi tiết chỉ báo"
              onClick={() => setSelected(null)}
            >
              <X size={18} />
            </Button>
          </div>
          <p className="detail-assessment">{selected.assessment}</p>
          <div className="indicator-history">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart
                data={selected.history.map((value, i) => ({
                  label: mockScoreHistory[i].label,
                  value,
                }))}
              >
                <CartesianGrid vertical={false} stroke="var(--grid)" />
                <XAxis
                  dataKey="label"
                  tick={{ fill: "var(--muted)", fontSize: 11 }}
                  axisLine={false}
                  tickLine={false}
                />
                <YAxis
                  domain={["auto", "auto"]}
                  tick={{ fill: "var(--muted)", fontSize: 11 }}
                  axisLine={false}
                  tickLine={false}
                  width={65}
                />
                <Tooltip
                  contentStyle={{
                    background: "var(--card)",
                    borderColor: "var(--border)",
                    borderRadius: 8,
                    color: "var(--text)",
                  }}
                />
                <Line
                  dataKey="value"
                  name={selected.name}
                  stroke="#249d83"
                  strokeWidth={2.5}
                  dot={{ r: 4 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </section>
      )}
      <p className="disclaimer">
        <ShieldCheck size={13} /> Điểm số là đánh giá dựa trên bộ dữ liệu minh
        họa có sẵn, không phải xác suất hoặc khuyến nghị đầu tư. Chưa kết nối
        nguồn vĩ mô trực tiếp.
      </p>
    </div>
  );
}
