import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { ResponsiveContainer, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip } from "recharts";
import { fetchMacroScore } from "../data/macro-score-client";
import { formatNumber } from "../lib/utils";
const groupNames: Record<string, string> = { Liquidity: "Thanh khoản & lãi suất", Production: "Sản xuất", Inflation: "Lạm phát" };
const reasons: Record<string, string> = { MISSING: "Chưa có dữ liệu", STALE: "Dữ liệu đã cũ", INSUFFICIENT_HISTORY: "Thiếu lịch sử để tính xu hướng", INCOMPARABLE_HISTORY: "Khác chuỗi hoặc phương pháp", RANGE_REQUIRED: "Thiếu cận lãi suất", INVALID_CREDIT_BASE: "Mốc tín dụng không hợp lệ", ELIGIBLE: "Đủ dữ liệu" };
const roles: Record<string, string> = { VN_DEPOSIT_RATE: "Điều kiện tiền tệ / tác động có độ trễ", VN_PMI: "Khảo sát sớm về hoạt động sản xuất", VN_CPI_YOY: "Giá tiêu dùng / thường có độ trễ", VN_CREDIT_GROWTH: "Truyền dẫn tín dụng / độ trễ công bố", VN_IIP_YOY: "Hoạt động sản xuất thực tế", US_EFFR: "Điều kiện tiền tệ quốc tế", VN_PPI_YOY: "Giá đầu vào / áp lực giá sớm", VN_M2_GROWTH: "Thanh khoản / tác động có độ trễ" };
const sourceHref = (value: string) => { try { const u = new URL(value); return u.protocol === "https:" && !u.username && !u.password ? u.href : undefined; } catch { return undefined; } };
const num = (v: number | null | undefined) => v == null ? "—" : formatNumber(v, 1);
export function MacroScorePanel() {
  const [frequency, setFrequency] = useState<"daily" | "weekly">("daily");
  const query = useQuery({ queryKey: ["macro-score", frequency], queryFn: ({ signal }) => fetchMacroScore(frequency, signal), refetchInterval: 60_000, retry: 1 });
  const current = query.data?.current;
  const history = query.data?.history ?? [];
  const previous = [...history].reverse().find(s => current && new Date(s.asOf) < new Date(current.asOf));
  const delta = current?.score != null && previous?.score != null ? current.score - previous.score : null;
  return <section className="card macro-score-panel" aria-label="Điểm vĩ mô">
    <div className="macro-score-heading">
      <div><span className="section-kicker">MACRO SCORE · THỬ NGHIỆM</span><h2>Mức hỗ trợ của bối cảnh vĩ mô</h2></div>
      <div className="score-frequency" role="group" aria-label="Kỳ lịch sử điểm">
        {(["daily", "weekly"] as const).map(f => <button key={f} aria-pressed={frequency === f} onClick={() => setFrequency(f)}>{f === "daily" ? "Ngày" : "Tuần"}</button>)}
      </div>
    </div>
    {query.isPending && <p role="status">Đang tải điểm vĩ mô…</p>}
    {query.isError && <p role="alert">{query.error.message}</p>}
    {query.data && !current && <p>Chưa có snapshot. Bộ tính điểm chạy mỗi phút sau khi Macro API khởi động.</p>}
    {current && <>
      <div className="score-summary">
        <div className="macro-score"><strong>{num(current.score)}</strong><span>/100 · {current.status === "READY" ? "Đủ điều kiện" : current.status === "PROVISIONAL" ? "Tạm tính" : "Chưa đủ dữ liệu"}</span></div>
        <div><strong>Độ phủ trọng số {num(current.coverage)}%</strong><p>Cần ≥80% và đủ lãi suất huy động, CPI, PMI để đạt trạng thái đủ điều kiện.</p></div>
      </div>
      {previous && <p className="score-notice">So với snapshot {new Date(previous.asOf).toLocaleDateString("vi-VN")}: {delta == null ? "chưa đủ điểm để so sánh" : `${delta >= 0 ? "+" : ""}${num(delta)} điểm`}. {previous.coverage !== current.coverage && "Độ phủ thay đổi; biến động bao gồm thay đổi cơ cấu trọng số."}</p>}
      {current.status === "PROVISIONAL" && <p className="score-notice">Điểm tạm tính được chuẩn hóa trên các chỉ báo đủ dữ liệu. Khi độ phủ thay đổi, cơ cấu đóng góp cũng thay đổi.</p>}
      <div className="score-groups">{current.groups.map(g => <div key={g.name}><span>{groupNames[g.name]}</span><strong>{num(g.score)}<small> /100</small></strong><small>Trọng số {g.weight}% · Độ phủ {num(g.coverage)}%</small></div>)}</div>
      <p className="subtle-copy">Tính lúc {new Date(current.asOf).toLocaleString("vi-VN", { timeZone: "Asia/Ho_Chi_Minh" })} (giờ Việt Nam) · {current.modelVersion}</p>
      <div className="score-history">
        {history.length > 0 ? <ResponsiveContainer width="100%" height={240}><LineChart data={history} margin={{ top: 12, right: 18, left: -20, bottom: 6 }}>
          <CartesianGrid strokeDasharray="3 3" vertical={false} /><XAxis dataKey="asOf" tickFormatter={d => new Date(d).toLocaleDateString("vi-VN", { timeZone: "Asia/Ho_Chi_Minh" })} /><YAxis domain={[0, 100]} />
          <Tooltip labelFormatter={d => new Date(String(d)).toLocaleString("vi-VN", { timeZone: "Asia/Ho_Chi_Minh" })} />
          <Line name="Điểm vĩ mô" type="stepAfter" dataKey="score" stroke="#38bfa1" strokeWidth={2.5} dot={{ r: 3 }} connectNulls={false} />
        </LineChart></ResponsiveContainer> : <p>Chưa có lịch sử {frequency === "daily" ? "ngày" : "tuần"}. Snapshot ngày lưu sau 18:00; snapshot tuần lưu thứ Sáu sau 18:00 (giờ Việt Nam).</p>}
      </div>
      <div className="score-component-table"><table><thead><tr><th>Chỉ báo / dữ liệu đầu vào</th><th>Trọng số</th><th>Điểm</th><th>Đóng góp</th><th>Thay đổi đóng góp</th></tr></thead><tbody>
        {current.components.map(c => <tr key={c.code}><td><details><summary>{c.name}{c.critical ? " *" : ""}<small>{reasons[c.reason] ?? c.reason}</small></summary>
          <p>{roles[c.code]}. Phân loại theo nền kinh tế; không cố định thời điểm dẫn dắt giá cổ phiếu.</p>
          <p>Trọng số hiệu dụng: {num(c.effectiveWeight)}%. {c.warnings.includes("USER_SUPPLIED_UNVERIFIED") && "Có dữ liệu nhập thủ công, chưa kiểm chứng độc lập."}</p>
          {c.inputs.map((o, i) => <p key={`${o.period}-${i}`}>{o.period}: {num(o.value)}{o.upperValue != null ? `–${num(o.upperValue)}` : ""} {o.unit} · <a href={sourceHref(o.sourceUrl)} target="_blank" rel="noreferrer">{o.source}</a> · {o.sourceSeriesId} · bản {o.revision}<br /><small>{o.method}</small><br /><small>Raw #{o.rawId} · Ghi nhận {new Date(o.firstSeenAt).toLocaleString("vi-VN")}</small></p>)}
        </details></td><td>{c.weight}%</td><td>{num(c.score)}</td><td>{c.score === null ? "—" : `+${num(c.contribution)}`}</td><td>{previous ? num(c.contribution - (previous.components.find(p => p.code === c.code)?.contribution ?? 0)) : "—"}</td></tr>)}
      </tbody></table></div>
      <p className="subtle-copy">* Chỉ báo bắt buộc. Đóng góp dùng trọng số hiệu dụng và cộng thành điểm tổng. Biểu đồ giữ nguyên điểm khi không có thông tin mới; dữ liệu quá cũ sẽ giảm độ phủ. Không dựng lại điểm quá khứ từ dữ liệu backfill.</p>
    </>}
    <details className="score-method"><summary>Phương pháp và giới hạn</summary><p>Mô hình v1 dùng mức chỉ báo và xu hướng: PMI quanh ngưỡng 50; biến động riêng hai cận lãi suất; tín dụng quy đổi YTD sang tăng trưởng tháng, xử lý lại mốc đầu năm. CPI/PPI/M2 dùng vùng tham chiếu thử nghiệm, không coi giảm phát là luôn tích cực. IIP làm mượt ba tháng; Fed dùng EFFR.</p><p>Thiếu kỳ đối chiếu sẽ không chấm điểm thành phần. Nguồn đang lỗi vẫn có thể dùng quan sát đã lưu nếu còn mới. Trọng số và ngưỡng chưa được kiểm chứng bằng backtest; điểm không phải dự báo lợi nhuận hoặc tín hiệu mua bán.</p></details>
  </section>;
}
