# RFC-001 — Market Intelligence Capability

Status: Draft

Version: 1.0

Owner: Investment AI Team

Category: Business Capability

---

# 1. Problem Statement

## Background

Mọi quyết định đầu tư đều chịu ảnh hưởng bởi bối cảnh thị trường.

Một doanh nghiệp tốt vẫn có thể tạo ra kết quả đầu tư kém nếu thị trường đang ở giai đoạn suy giảm mạnh hoặc môi trường vĩ mô bất lợi.

Nếu mỗi Capability tự phân tích bối cảnh thị trường thì sẽ dẫn đến:

- Trùng lặp logic.
- Kết quả không nhất quán.
- Khó bảo trì Business Rules.
- Không thể tái sử dụng kết quả phân tích.

Hệ thống cần một Capability chuyên trách tạo ra **Market Context** thống nhất để toàn bộ các Capability khác sử dụng.

---

## Problem

Investment AI hiện chưa có một cơ chế chuẩn để:

- Đánh giá trạng thái của thị trường.
- Tổng hợp dữ liệu vĩ mô và dữ liệu thị trường.
- Chuẩn hóa Market Context.
- Cung cấp kết quả phân tích có khả năng truy vết.

Điều này làm giảm tính nhất quán của toàn bộ hệ thống.

---

# 2. Business Goal

Market Intelligence cung cấp một đánh giá toàn diện về môi trường đầu tư hiện tại.

Capability này trả lời các câu hỏi:

- Thị trường đang mạnh hay yếu?
- Thanh khoản có cải thiện không?
- Dòng tiền đang dịch chuyển như thế nào?
- Chu kỳ kinh tế hiện tại là gì?
- Mức độ Risk-On hay Risk-Off ra sao?
- Bối cảnh hiện tại hỗ trợ hay cản trở quyết định đầu tư?

Capability không trả lời:

- Có nên mua cổ phiếu X?
- Giá mục tiêu là bao nhiêu?

---

# 3. Scope

## In Scope

Capability chịu trách nhiệm:

- Phân tích xu hướng thị trường.
- Phân tích thanh khoản.
- Phân tích Market Breadth.
- Phân tích Capital Flow.
- Phân tích Macro Environment.
- Phân tích Market Cycle.
- Đánh giá Risk Environment.
- Tổng hợp Market Context.
- Sinh Market Analysis Result.

## Out of Scope

Capability không thực hiện:

- Company Analysis.
- Financial Analysis.
- Technical Analysis của từng cổ phiếu.
- Valuation.
- Portfolio Management.
- Investment Recommendation.

---

# 4. Business Outcome

Sau khi hoàn thành, Capability tạo ra:

- Market Context.
- Market Score.
- Market Summary.
- Supporting Evidence.
- Supporting Signals.
- Confidence.
- MarketAnalysisResult.

Đầu ra này sẽ được sử dụng bởi:

- Company Intelligence.
- Financial Analysis.
- Technical Analysis.
- Risk Assessment.
- Investment Decision.
- Reporting.

---

# 5. Architecture Position

```text
Workflow

↓

Task

↓

Market Intelligence Capability

↓

Analysis Modules

↓

Knowledge Contracts
```

Market Intelligence là một Business Capability.

Không phải Workflow.

Không phải AI Agent.

Không phải Microservice.

---

# 6. Analysis Modules

Capability được chia thành các Analysis Module.

## Macro Analysis

Đánh giá môi trường kinh tế.

Ví dụ:

- Interest Rate
- Inflation
- GDP
- PMI
- Credit Growth
- M2

---

## Market Trend Analysis

Đánh giá:

- Primary Trend
- Secondary Trend
- Long-term Trend

---

## Liquidity Analysis

Đánh giá:

- Trading Volume
- Trading Value
- Turnover

---

## Breadth Analysis

Đánh giá:

- Advance/Decline
- New High/New Low
- Up Volume/Down Volume

---

## Capital Flow Analysis

Đánh giá:

- Foreign Flow
- ETF Flow
- Proprietary Trading
- Sector Rotation

---

## Market Cycle Analysis

Đánh giá:

- Expansion
- Slowdown
- Recession
- Recovery

---

## Risk Environment Analysis

Đánh giá:

- Risk-On
- Neutral
- Risk-Off

---

## Market Context Synthesis

Tổng hợp toàn bộ kết quả thành:

- Summary
- Score
- Confidence
- Market Context

---

# 7. Knowledge Transformation

```text
Raw Market Data
        │
        ▼
Indicators
        │
        ▼
Evidence
        │
        ▼
Signals
        │
        ▼
Module Analysis
        │
        ▼
Market Context
        │
        ▼
MarketAnalysisResult
```

Mọi kết luận phải truy vết được đến Indicator và Evidence.

---

# 8. Execution Pipeline

Capability tuân theo Capability Execution Model chuẩn.

```text
Validate Input
        │
        ▼
Load Context
        │
        ▼
Execute Analysis Modules
        │
        ▼
Generate Evidence
        │
        ▼
Generate Signals
        │
        ▼
Synthesize Market Context
        │
        ▼
Generate MarketAnalysisResult
        │
        ▼
Validate Output
```

---

# 9. Dependencies

## Depends On

- Knowledge Contracts
- Business Rules
- Domain Model
- Market Data Sources
- Macro Data Sources

## Independent Of

- Go
- Database
- AI Provider
- Workflow Engine
- UI

---

# 10. Quality Attributes

Capability phải đảm bảo:

- Accuracy.
- Consistency.
- Explainability.
- Traceability.
- Reproducibility.
- Extensibility.

---

# 11. Error Model

## Business Errors

- Unsupported Market.
- Missing Trading Session.
- Invalid Analysis Period.

## Data Quality Errors

- Missing Indicator.
- Outdated Market Data.
- Inconsistent Macro Data.

## Technical Errors

- Provider Timeout.
- External API Failure.
- Storage Failure.

---

# 12. Observability

Capability phải ghi nhận:

- Execution ID.
- Capability Version.
- Input Contract Version.
- Output Contract Version.
- Execution Duration.
- Executed Modules.
- Evidence Count.
- Signal Count.
- Confidence.
- Warning Messages.
- Error Messages.

---

# 13. Security & Governance

Capability phải:

- Chỉ đọc dữ liệu.
- Không thay đổi dữ liệu nguồn.
- Có Audit Trail.
- Có Versioning.
- Có Rule Traceability.

---

# 14. Extension Points

Có thể mở rộng bằng cách:

- Thêm Analysis Module.
- Thay Rule Engine bằng AI.
- Thêm Market mới.
- Thêm Indicator mới.
- Thêm Data Provider mới.

Không cần thay đổi Workflow hoặc Capability Interface.

---

# 15. Success Criteria

Capability được xem là hoàn thành khi:

- Tạo đúng MarketAnalysisResult.
- Có khả năng giải thích mọi kết luận.
- Có khả năng truy vết về dữ liệu nguồn.
- Được Capability khác tái sử dụng mà không cần xử lý lại.
- Có thể thay thế implementation mà không thay đổi hợp đồng đầu vào/đầu ra.

---

# 16. Related Documents

- capability-specification-template.md
- capability-execution-model.md
- workflow-architecture.md
- knowledge-model.md
- knowledge-contracts.md
- business-rules.md
