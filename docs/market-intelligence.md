# Capability Execution Model

Version: 1.0

---

# 1. Purpose

Capability Execution Model định nghĩa vòng đời chuẩn của mọi Capability trong Investment AI.

Mọi Capability, bất kể được triển khai bằng:

* Go
* AI Agent
* Rule Engine
* Machine Learning
* Hybrid

đều phải tuân theo cùng một Execution Model.

Điều này giúp Workflow không cần biết implementation phía sau.

---

# 2. Execution Pipeline

```text
Input Contract
        │
        ▼
Validate
        │
        ▼
Load Context
        │
        ▼
Normalize
        │
        ▼
Generate Evidence
        │
        ▼
Detect Signals
        │
        ▼
Analyze
        │
        ▼
Generate Output Contract
        │
        ▼
Validate Output
        │
        ▼
Return
```

---

# 3. Stage 1 — Validate

Mục tiêu:

Kiểm tra Input Contract.

Bao gồm:

* Schema.
* Required Fields.
* Business Preconditions.
* Version Compatibility.

Nếu thất bại:

Trả về Business Error.

---

# 4. Stage 2 — Load Context

Capability có thể cần:

* Market Context.
* Company Context.
* Financial Context.
* Previous Analysis.

Context chỉ được đọc.

Không được thay đổi.

---

# 5. Stage 3 — Normalize

Chuẩn hóa dữ liệu để mọi bước phía sau hoạt động trên cùng một định dạng.

Ví dụ:

* Currency.
* Date.
* Timezone.
* Exchange Symbol.
* Missing Values.

---

# 6. Stage 4 — Generate Evidence

Biến dữ liệu thành các Evidence có ý nghĩa.

Ví dụ:

Raw Data:

ROE = 23%

↓

Evidence:

"ROE duy trì trên 20% trong ba năm liên tiếp."

Evidence luôn phải truy vết được về Raw Data.

---

# 7. Stage 5 — Detect Signals

Phát hiện các tín hiệu.

Ví dụ:

* Golden Cross.
* Spring.
* SOS.
* Earnings Growth.
* Margin Expansion.

Signal không phải kết luận.

---

# 8. Stage 6 — Analyze

Đánh giá toàn bộ Evidence và Signal.

Sinh:

* Summary.
* Score.
* Confidence.
* Risks.

Đây là nơi có thể sử dụng:

* Rule Engine.
* AI.
* Statistical Model.

---

# 9. Stage 7 — Generate Output Contract

Sinh đúng một Knowledge Contract.

Ví dụ:

AnalysisResult

hoặc

InvestmentThesis

hoặc

Recommendation

---

# 10. Stage 8 — Validate Output

Kiểm tra:

* Schema.
* Traceability.
* Confidence.
* Required Evidence.

Không Output nào được trả về nếu thiếu khả năng truy vết.

---

# 11. Execution Rules

* Capability chỉ sinh một Output Contract.
* Capability không điều phối Capability khác.
* Capability không ghi trực tiếp vào Database.
* Capability không gọi Workflow.
* Capability không sinh Report.

---

# 12. Extension Points

Mỗi Stage có thể thay thế implementation.

Ví dụ:

Generate Evidence

* Rule-based
* AI
* Hybrid

Analyze

* AI
* Quantitative Model
* Human Review

Nhờ đó hệ thống có thể phát triển dần mà không thay đổi kiến trúc.

---

# 13. Benefits

* Chuẩn hóa mọi Capability.
* Độc lập với công nghệ.
* Dễ kiểm thử từng giai đoạn.
* Dễ quan sát (Observability).
* Hỗ trợ Explainable AI.
* Hỗ trợ Multi-Agent.
* Hỗ trợ Workflow Engine.
