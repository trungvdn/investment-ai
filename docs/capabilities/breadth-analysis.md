# RFC-007 — Breadth Analysis Module

**Status:** Draft  
**Version:** 1.0  
**Owner:** Investment AI Team  
**Parent Capability:** Market Intelligence

---

# 1. Purpose

## Objective

Breadth Analysis đánh giá mức độ lan tỏa của xu hướng thị trường.

Module trả lời các câu hỏi:

- Bao nhiêu cổ phiếu đang tham gia vào xu hướng?
- Xu hướng hiện tại có sự đồng thuận của toàn thị trường không?
- Thị trường đang mở rộng hay thu hẹp?
- Breadth có xác nhận xu hướng hiện tại không?

Module không đánh giá:

- Thanh khoản
- Dòng tiền
- Vĩ mô
- Giá trị doanh nghiệp

---

# 2. Business Questions

Module phải trả lời các câu hỏi sau:

- Breadth hiện tại mạnh hay yếu?
- Bao nhiêu cổ phiếu tăng và giảm?
- Breadth có đang cải thiện không?
- Breadth có xác nhận xu hướng không?
- Có dấu hiệu phân kỳ giữa chỉ số và Breadth không?

---

# 3. Scope

## In Scope

- Advance / Decline
- New High / New Low
- Market Participation
- Sector Participation
- Breadth Trend
- Breadth Confirmation

## Out of Scope

- Liquidity
- Capital Flow
- Macro
- Company Fundamentals

---

# 4. Indicators

Module sử dụng các Indicator sau.

## Core Indicators

- Advance Count
- Decline Count
- Unchanged Count
- New High Count
- New Low Count

## Supporting Indicators

- Advance/Decline Ratio
- Advance/Decline Line
- Up Volume
- Down Volume
- Percentage of Stocks Above Moving Average

---

# 5. Information Model

```
Indicators
      │
      ▼
Breadth Evidence
      │
      ▼
Breadth Signals
      │
      ▼
Breadth Analysis Result
```

---

# 6. Analysis Rules

Breadth Analysis sử dụng các nhóm Rule sau.

## Market Participation

Đánh giá:

- Broad Participation
- Narrow Participation

---

## Breadth Strength

Đánh giá:

- Strong Breadth
- Neutral Breadth
- Weak Breadth

---

## Breadth Trend

Đánh giá:

- Improving
- Stable
- Deteriorating

---

## Breadth Confirmation

Đánh giá:

- Trend Confirmed
- Trend Not Confirmed

---

## Breadth Divergence

Phát hiện:

- Bullish Divergence
- Bearish Divergence

---

# 7. Evidence Generation

Ví dụ Evidence.

## Strong Breadth

Điều kiện:

- Số lượng cổ phiếu tăng chiếm ưu thế.
- Nhiều ngành cùng tham gia xu hướng.
- New High mở rộng.

---

## Weak Breadth

Điều kiện:

- Chỉ một số ít cổ phiếu dẫn dắt.
- New Low gia tăng.
- Độ lan tỏa suy giảm.

---

## Breadth Divergence

Điều kiện:

- Chỉ số và Breadth di chuyển theo hai hướng khác nhau.

---

# 8. Signal Generation

Ví dụ Signal.

## Bullish Breadth

Evidence:

- Strong Breadth
- Trend Confirmed

---

## Bearish Breadth

Evidence:

- Weak Breadth
- Breadth Deteriorating

---

## Neutral Breadth

Evidence trái chiều hoặc chưa đủ dữ liệu.

---

# 9. Analysis Result

Breadth Analysis Result bao gồm:

- Summary
- Breadth Status
- Breadth Strength
- Confidence
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

- Chỉ số tăng nhờ một vài cổ phiếu vốn hóa lớn.
- Thanh khoản thấp trong kỳ nghỉ lễ.
- Thị trường biến động mạnh do sự kiện bất thường.
- Thiếu dữ liệu của một số mã hoặc ngành.

Module phải trả về trạng thái phù hợp thay vì suy diễn.

---

# 12. Related Documents

- RFC-001 — Market Intelligence Specification
- RFC-002 — Information Model
- RFC-003 — Knowledge Contracts
- RFC-004 — Business Rules
