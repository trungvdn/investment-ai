import { useId, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { ResponsiveContainer, AreaChart, Area, ReferenceLine, XAxis, YAxis, CartesianGrid, Tooltip } from "recharts";
import { ArrowDownRight, ArrowUpRight, ChevronDown, Info, ShieldCheck } from "lucide-react";
import { fetchMacroScore } from "../data/macro-score-client";
import { formatNumber } from "../lib/utils";
const groupNames: Record<string, string> = { Liquidity: "Thanh khoản & lãi suất", Production: "Sản xuất", Inflation: "Lạm phát" };
const reasons: Record<string, string> = { MISSING: "Chưa có dữ liệu", STALE: "Dữ liệu đã cũ", INSUFFICIENT_HISTORY: "Thiếu lịch sử để tính xu hướng", INCOMPARABLE_HISTORY: "Khác chuỗi hoặc phương pháp", RANGE_REQUIRED: "Thiếu cận lãi suất", INVALID_CREDIT_BASE: "Mốc tín dụng không hợp lệ", ELIGIBLE: "Đủ dữ liệu" };
const roles: Record<string, string> = { VN_DEPOSIT_RATE: "Điều kiện tiền tệ / tác động có độ trễ", VN_PMI: "Khảo sát sớm về hoạt động sản xuất", VN_CPI_YOY: "Giá tiêu dùng / thường có độ trễ", VN_CREDIT_GROWTH: "Truyền dẫn tín dụng / độ trễ công bố", VN_IIP_YOY: "Hoạt động sản xuất thực tế", US_EFFR: "Điều kiện tiền tệ quốc tế", VN_PPI_YOY: "Giá đầu vào / áp lực giá sớm", VN_M2_GROWTH: "Thanh khoản / tác động có độ trễ" };
const sourceHref = (value: string) => { try { const u = new URL(value); return u.protocol === "https:" && !u.username && !u.password ? u.href : undefined; } catch { return undefined; } };
const num = (v: number | null | undefined) => v == null ? "—" : formatNumber(v, 1);
export function MacroScorePanel() {
  const [showIndicators, setShowIndicators] = useState(false);
  const [showGroups, setShowGroups] = useState(false);
  const detailsId = useId();
  const [selectedAsOf, setSelectedAsOf] = useState("");
  const [frequency, setFrequency] = useState<"daily" | "weekly">("daily");
  const query = useQuery({ queryKey: ["macro-score", frequency], queryFn: ({ signal }) => fetchMacroScore(frequency, signal), refetchInterval: 60_000, retry: 1 });
  const current = query.data?.current;
  const history = (query.data?.history ?? []).slice(frequency === "daily" ? -7 : -52);
  const chartPoints = history.map(s => ({ asOf: s.asOf, score: s.score, timestamp: Date.parse(s.asOf), reconstructed: !!s.reconstructed, isCurrent: false }));
  const lastPoint = chartPoints.at(-1);
  const includeCurrent = !!current && (!lastPoint || Date.parse(current.asOf) > lastPoint.timestamp);
  if (includeCurrent) chartPoints.push({ asOf: current!.asOf, score: current!.score, timestamp: Date.parse(current!.asOf), reconstructed: false, isCurrent: true });
  const values = chartPoints.flatMap(p => p.score == null ? [] : [p.score]);
  const range = values.length ? Math.max(...values) - Math.min(...values) : null;
  const firstScore = chartPoints[0]?.score;
  const finalScore = chartPoints.at(-1)?.score;
  const change = chartPoints.length > 1 && firstScore != null && finalScore != null ? finalScore - firstScore : null;
  const previous = [...history].reverse().find(s => current && new Date(s.asOf) < new Date(current.asOf));
  const snapshots = includeCurrent && current ? [...history, { ...current, kind: frequency === "daily" ? "DAILY" as const : "WEEKLY" as const, reconstructed: false }] : history;
  const focused = snapshots.find(s => s.asOf === selectedAsOf) ?? current;
  const focusedPoint = snapshots.find(s => s.asOf === focused?.asOf);
  const focusIndex = snapshots.findIndex(s => s.asOf === focused?.asOf);
  const focusPrevious = focusIndex > 0 ? snapshots[focusIndex - 1] : undefined;
  const focusDelta = focused?.score != null && focusPrevious?.score != null ? focused.score - focusPrevious.score : null;
  const date = (value: string) => new Date(value).toLocaleDateString("vi-VN", { timeZone: "Asia/Ho_Chi_Minh", day: "2-digit", month: "2-digit" });
  return <section id="macro-score" className="card macro-score-panel" aria-label="Điểm vĩ mô">
    <div className="macro-score-heading">
      <div><span className="section-kicker">MACRO SCORE · THỬ NGHIỆM</span><h2>Nhịp đập vĩ mô</h2><p className="subtle-copy">Mức hỗ trợ của bối cảnh kinh tế đối với thị trường cổ phiếu Việt Nam.</p></div>
      <div className="score-frequency" role="group" aria-label="Kỳ lịch sử điểm">
        {(["daily", "weekly"] as const).map(f => <button key={f} aria-pressed={frequency === f} onClick={() => { setFrequency(f); setSelectedAsOf(""); }}>{f === "daily" ? "Ngày" : "Tuần"}</button>)}
      </div>
    </div>
    {query.isPending && <div className="score-chart-empty" role="status">Đang tải điểm vĩ mô…</div>}
    {query.isError && <p role="alert" className="score-notice">{query.error.message}</p>}
    {query.data && !current && <div className="score-chart-empty"><ActivityPlaceholder /><p>Chưa có snapshot. Bộ tính điểm chạy mỗi phút sau khi Macro API khởi động.</p></div>}
    {current && <>
      <div className="score-layout">
        <section className="score-history" aria-label="Biểu đồ biến động Macro Score">
          <div className="score-summary">
            <div className="macro-score"><strong>{num(current.score)}</strong><span>/ 100 điểm<small>{current.status === "READY" ? "Đủ điều kiện" : current.status === "PROVISIONAL" ? "Tạm tính" : "Chưa đủ dữ liệu"}</small></span></div>
            <div className="score-period-delta"><span className="subtle-copy">{frequency === "daily" ? "7 ngày gần nhất" : "Lịch sử theo tuần"}</span><strong>{change == null ? "—" : <>{change < 0 ? <ArrowDownRight size={15}/> : <ArrowUpRight size={15}/>} {change >= 0 ? "+" : ""}{num(change)} điểm</>}</strong><small>Thay đổi trong kỳ</small></div>
          </div>
          <div className="score-chart-key"><span><i/>Macro Score</span><span>0–100 điểm</span></div>
          {chartPoints.some(p => p.score != null) ? <ResponsiveContainer width="100%" height={250}><AreaChart data={chartPoints} margin={{ top: 14, right: 14, left: -25, bottom: 4 }}>
            <defs><linearGradient id={`${detailsId}-fill`} x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stopColor="#168368" stopOpacity={0.18}/><stop offset="100%" stopColor="#168368" stopOpacity={0.01}/></linearGradient></defs>
            <CartesianGrid stroke="var(--grid)" strokeDasharray="4 5" vertical={false} />
            {focused && <ReferenceLine x={Date.parse(focused.asOf)} stroke="var(--green)" strokeOpacity={0.35} strokeDasharray="3 5" />}
            <ReferenceLine y={50} stroke="var(--border)" strokeDasharray="5 5" />
            <XAxis dataKey="timestamp" type="number" scale="time" domain={["dataMin", "dataMax"]} axisLine={false} tickLine={false} minTickGap={25} tick={{ fill: "var(--muted)", fontSize: 11 }} tickFormatter={d => date(new Date(d).toISOString())} />
            <YAxis domain={[0, 100]} ticks={[0,25,50,75,100]} axisLine={false} tickLine={false} tick={{ fill: "var(--muted)", fontSize: 11 }} />
            <Tooltip contentStyle={{ background: "var(--card)", border: "1px solid var(--border)", borderRadius: 10, fontSize: 12, color: "var(--text)" }} labelFormatter={(d, payload) => `${date(new Date(Number(d)).toISOString())}${payload?.[0]?.payload?.isCurrent ? " · Hiện tại" : payload?.[0]?.payload?.reconstructed ? " · Tính lại" : ""}`} formatter={value => [num(value as number), "Macro Score"]} />
            <Area fill={`url(#${detailsId}-fill)`} name="Macro Score" type="stepAfter" dataKey="score" stroke="var(--green)" strokeWidth={2.5} dot={{ r: 3, fill: "var(--card)", strokeWidth: 2 }} activeDot={{ r: 6, strokeWidth: 3, stroke: "var(--card)" }} connectNulls={false} isAnimationActive={false} />
          </AreaChart></ResponsiveContainer> : <p className="score-chart-empty">Chưa đủ dữ liệu để vẽ Macro Score. Biểu đồ sẽ xuất hiện khi có điểm hợp lệ.</p>}
          <div className="score-chart-footer"><span>Biên độ cao–thấp <b>{num(range)} điểm</b></span><span>Thấp nhất / Cao nhất <b>{values.length ? `${num(Math.min(...values))} / ${num(Math.max(...values))}` : "—"}</b></span></div>
          {includeCurrent && <p className="subtle-copy score-caption">Điểm cuối là điểm hiện tại, chưa phải snapshot chốt {frequency === "daily" ? "ngày" : "tuần"}.{history.length === 0 ? " Chưa có lịch sử để thể hiện biến động." : ""}</p>}
        </section>
        {focused && <aside className="score-insight" aria-label="Thông tin kỳ điểm đang chọn">
          <span className="section-kicker">TẠI KỲ ĐANG CHỌN</span>
          <label className="score-period-select"><span className="sr-only">Kỳ điểm vĩ mô</span><select aria-label="Kỳ điểm vĩ mô" value={snapshots.some(s => s.asOf === focused.asOf) ? focused.asOf : ""} onChange={e => setSelectedAsOf(e.target.value)}>
            {snapshots.length ? snapshots.map(s => <option key={s.asOf} value={s.asOf}>{new Date(s.asOf).toLocaleDateString("vi-VN", { timeZone: "Asia/Ho_Chi_Minh" })} · {num(s.score)} điểm{s.reconstructed ? " · Tính lại" : ""}</option>) : <option value="">Hiện tại</option>}
          </select><ChevronDown size={14}/></label>
          <div className="score-focus-number">{num(focused.score)}<small>/100</small></div>
          <p>{focusDelta == null ? "Chưa đủ dữ liệu để so sánh kỳ trước." : `${focusDelta >= 0 ? "+" : ""}${num(focusDelta)} điểm so với kỳ trước.`}</p>
          <div className="score-coverage"><div><ShieldCheck size={15}/><strong>Độ phủ trọng số {num(focused.coverage)}%</strong></div><div className="score-coverage-track"><i style={{width:`${focused.coverage}%`}}/></div><p>{focused.components.filter(c => c.score != null).length}/{focused.components.length} chỉ báo đủ dữ liệu</p></div>
          <p className="score-insight-note">{focused.status === "READY" ? "Đủ ngưỡng 80% và các chỉ báo bắt buộc." : "Cần ≥80% và đủ lãi suất huy động, CPI, PMI. Điểm tạm tính chuẩn hóa trên chỉ báo đủ dữ liệu."}</p>
          <span className="score-vintage">{focusedPoint?.reconstructed ? "Điểm tính lại" : "Snapshot hệ thống"} · {focused.modelVersion}</span>
        </aside>}
      </div>
      {history.some(p => p.reconstructed) && <p className="score-reconstruction-note"><Info size={13}/><span>Có điểm tính lại từ dữ liệu hiện lưu theo ngày công bố; có thể chứa bản sửa thu thập sau đó. Không phải lịch sử hệ thống đã biết tại thời điểm đó.</span></p>}
      {previous && previous.coverage !== current.coverage && <p className="score-notice">Độ phủ thay đổi; biến động bao gồm thay đổi cơ cấu trọng số.</p>}
      <div className="score-detail-actions">
        <button type="button" aria-expanded={showIndicators} aria-controls={`${detailsId}-indicators`} onClick={() => setShowIndicators(v => !v)}>{showIndicators ? "Ẩn chi tiết điểm từng chỉ báo" : "Xem chi tiết điểm từng chỉ báo"}</button>
        <button type="button" aria-expanded={showGroups} aria-controls={`${detailsId}-groups`} onClick={() => setShowGroups(v => !v)}>{showGroups ? "Ẩn điểm theo nhóm" : "Xem điểm theo nhóm"}</button>
      </div>
      {showGroups && <section id={`${detailsId}-groups`} aria-label="Điểm theo nhóm">
      <div className="score-groups">{focused!.groups.map(g => <div key={g.name}><span>{groupNames[g.name]}</span><strong>{num(g.score)}<small> /100</small></strong><small>Trọng số {g.weight}% · Độ phủ {num(g.coverage)}%</small></div>)}</div>
      </section>}
      {showIndicators && <section id={`${detailsId}-indicators`} aria-label="Chi tiết điểm từng chỉ báo">
      <div className="score-component-table"><table><thead><tr><th>Chỉ báo / dữ liệu đầu vào</th><th>Trọng số</th><th>Điểm</th><th>Đóng góp</th><th>Thay đổi đóng góp</th></tr></thead><tbody>
        {focused!.components.map(c => <tr key={c.code}><td><details><summary>{c.name}{c.critical ? " *" : ""}<small>{reasons[c.reason] ?? c.reason}</small></summary>
          <p>{roles[c.code]}. Phân loại theo nền kinh tế; không cố định thời điểm dẫn dắt giá cổ phiếu.</p>
          <p>Trọng số hiệu dụng: {num(c.effectiveWeight)}%. {c.warnings.includes("USER_SUPPLIED_UNVERIFIED") && "Có dữ liệu nhập thủ công, chưa kiểm chứng độc lập."}</p>
          {c.inputs.map((o, i) => <p key={`${o.period}-${i}`}>{o.period}: {num(o.value)}{o.upperValue != null ? `–${num(o.upperValue)}` : ""} {o.unit} · <a href={sourceHref(o.sourceUrl)} target="_blank" rel="noreferrer">{o.source}</a> · {o.sourceSeriesId} · bản {o.revision}<br /><small>{o.method}</small><br /><small>Raw #{o.rawId} · Ghi nhận {new Date(o.firstSeenAt).toLocaleString("vi-VN")}</small></p>)}
        </details></td><td>{c.weight}%</td><td>{num(c.score)}</td><td>{c.score === null ? "—" : `+${num(c.contribution)}`}</td><td>{focusPrevious ? num(c.contribution - (focusPrevious.components.find(p => p.code === c.code)?.contribution ?? 0)) : "—"}</td></tr>)}
      </tbody></table></div>
      <p className="subtle-copy">* Chỉ báo bắt buộc. Đóng góp dùng trọng số hiệu dụng và cộng thành điểm tổng. Biểu đồ giữ nguyên điểm khi không có thông tin mới; dữ liệu quá cũ sẽ giảm độ phủ. Điểm tính lại được đánh dấu riêng và không ghi đè snapshot đã lưu.</p>
      </section>}
    </>}
    <details className="score-method"><summary>Phương pháp và giới hạn</summary><p>Mô hình v1 dùng mức chỉ báo và xu hướng: PMI quanh ngưỡng 50; biến động riêng hai cận lãi suất; tín dụng quy đổi YTD sang tăng trưởng tháng, xử lý lại mốc đầu năm. CPI/PPI/M2 dùng vùng tham chiếu thử nghiệm, không coi giảm phát là luôn tích cực. IIP làm mượt ba tháng; Fed dùng EFFR.</p><p>Thiếu kỳ đối chiếu sẽ không chấm điểm thành phần. Nguồn đang lỗi vẫn có thể dùng quan sát đã lưu nếu còn mới. Trọng số và ngưỡng chưa được kiểm chứng bằng backtest; điểm không phải dự báo lợi nhuận hoặc tín hiệu mua bán.</p></details>
  </section>;
}

function ActivityPlaceholder() { return <svg width="48" height="32" viewBox="0 0 48 32" fill="none" aria-hidden="true"><path d="M2 20h11l6-14 10 23 6-13h11" stroke="currentColor" strokeWidth="2"/></svg>; }
