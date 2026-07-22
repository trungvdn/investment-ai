# RFC-010 — Market Cycle Analysis Module

**Status:** Draft  
**Version:** 1.0  
**Owner:** Investment AI Team  
**Parent Capability:** Market Intelligence

---

# 1. Purpose

## Objective

Market Cycle Analysis xác định giai đoạn hiện tại của chu kỳ thị trường.

Module trả lời các câu hỏi:

- Thị trường đang ở giai đoạn nào?
- Chu kỳ đang chuyển sang pha mới hay vẫn tiếp diễn?
- Bối cảnh hiện tại phù hợp với giai đoạn nào của chu kỳ?

Module không đánh giá:

- Giá trị doanh nghiệp
- Khuyến nghị đầu tư
- Giá mục tiêu

---

# 2. Business Questions

Module phải trả lời các câu hỏi sau:

- Thị trường đang ở pha nào?
- Chu kỳ đang mở rộng hay thu hẹp?
- Có dấu hiệu chuyển pha không?
- Mức độ tin cậy của việc xác định chu kỳ là bao nhiêu?

---

# 3. Scope

## In Scope

- Business Cycle
- Market Cycle
- Cycle Transition
- Cycle Confirmation

## Out of Scope

- Company Financial Analysis
- Individual Stock Analysis
- Portfolio Allocation

---

# 4. Indicators

Module sử dụng kết quả từ các Analysis Module khác.

## Core Inputs

- Market Trend Analysis
- Liquidity Analysis
- Breadth Analysis
- Capital Flow Analysis
- Macro Analysis

## Supporting Inputs

- Market Volatility
- Market Sentiment
- Sector Rotation

---

# 5. Information Model

```
Analysis Results
        │
        ▼
Cycle Evidence
        │
        ▼
Cycle Signals
        │
        ▼
Market Cycle Result
```

---

# 6. Analysis Rules

Market Cycle Analysis sử dụng các nhóm Rule sau.

## Cycle Identification

Đánh giá:

- Recovery
- Expansion
- Slowdown
- Recession

---

## Cycle Confirmation

Đánh giá:

- Confirmed Cycle
- Unconfirmed Cycle

---

## Cycle Transition

Đánh giá:

- Early Transition
- Confirmed Transition
- No Transition

---

## Cycle Consistency

Đánh giá:

- Consistent Signals
- Mixed Signals
- Conflicting Signals

---

# 7. Evidence Generation

Ví dụ Evidence.

## Recovery

Điều kiện:

- Xu hướng cải thiện.
- Thanh khoản tăng.
- Dòng tiền quay trở lại.
- Vĩ mô dần ổn định.

---

## Expansion

Điều kiện:

- Xu hướng tăng được xác nhận.
- Breadth mở rộng.
- Dòng tiền tích cực.
- Môi trường vĩ mô thuận lợi.

---

## Slowdown

Điều kiện:

- Xu hướng suy yếu.
- Breadth thu hẹp.
- Dòng tiền giảm.
- Vĩ mô bắt đầu xấu đi.

---

## Recession

Điều kiện:

- Xu hướng giảm rõ ràng.
- Thanh khoản suy giảm.
- Dòng tiền rút khỏi thị trường.
- Môi trường vĩ mô bất lợi.

---

# 8. Signal Generation

Ví dụ Signal.

## Recovery Signal

Evidence:

- Recovery
- Confirmed Transition

---

## Expansion Signal

Evidence:

- Expansion
- Confirmed Cycle

---

## Slowdown Signal

Evidence:

- Slowdown
- Mixed Signals

---

## Recession Signal

Evidence:

- Recession
- Confirmed Cycle

---

# 9. Analysis Result

Market Cycle Result bao gồm:

- Summary
- Current Cycle
- Cycle Confidence
- Supporting Evidence
- Supporting Signals
- Warnings

---

# 10. Quality Metrics

Module phải đảm bảo:

- Explainability
- Traceability
- Consistency
- Reproducibility

---

# 11. Edge Cases

Ví dụ:

- Các Analysis Module đưa ra kết luận mâu thuẫn.
- Chu kỳ chuyển pha nhưng chưa được xác nhận.
- Sự kiện bất thường làm thay đổi chu kỳ trong ngắn hạn.
- Thiếu dữ liệu từ một hoặc nhiều Analysis Module.

Module phải phản ánh mức độ không chắc chắn thay vì ép buộc kết luận.

---

# 12. Related Documents

- RFC-001 — Market Intelligence Specification
- RFC-002 — Information Model
- RFC-003 — Knowledge Contracts
- RFC-004 — Business Rules
- Market Trend Analysis
- Liquidity Analysis
- Breadth Analysis
- Capital Flow Analysis
- Macro Analysis
