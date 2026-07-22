# RFC-006 — Market Trend Analysis Module

**Status:** Draft  
**Version:** 1.0  
**Owner:** Investment AI Team  
**Parent Capability:** Market Intelligence

---

# 1. Purpose

## Objective

Market Trend Analysis đánh giá xu hướng hiện tại của thị trường.

Module trả lời các câu hỏi:

- Thị trường đang trong xu hướng tăng, giảm hay đi ngang?
- Xu hướng hiện tại mạnh hay yếu?
- Xu hướng có đang thay đổi không?
- Xu hướng có được xác nhận hay không?

Module không đánh giá:

- Thanh khoản
- Dòng tiền
- Môi trường vĩ mô
- Khuyến nghị đầu tư

---

# 2. Business Questions

Module phải trả lời các câu hỏi sau:

- Xu hướng hiện tại là gì?
- Xu hướng đang mạnh lên hay yếu đi?
- Có dấu hiệu đảo chiều không?
- Xu hướng có được xác nhận không?
- Mức độ tin cậy của xu hướng là bao nhiêu?

---

# 3. Scope

## In Scope

- Primary Trend
- Secondary Trend
- Minor Trend
- Trend Strength
- Trend Confirmation
- Trend Reversal

## Out of Scope

- Liquidity
- Breadth
- Capital Flow
- Macro
- Company Fundamentals

---

# 4. Indicators

Module sử dụng các Indicator sau.

## Core Indicators

- Price
- High
- Low
- Close

## Supporting Indicators

- Moving Average
- Trendline
- Swing High
- Swing Low
- Higher High
- Higher Low
- Lower High
- Lower Low

---

# 5. Information Model

```
Indicators
      │
      ▼
Trend Evidence
      │
      ▼
Trend Signals
      │
      ▼
Trend Analysis Result
```

---

# 6. Analysis Rules

Market Trend Analysis sử dụng các nhóm Rule sau.

## Trend Identification

Xác định:

- Uptrend
- Downtrend
- Sideway

---

## Trend Strength Analysis

Đánh giá:

- Strong Trend
- Normal Trend
- Weak Trend

---

## Trend Confirmation

Đánh giá:

- Confirmed Trend
- Unconfirmed Trend

---

## Trend Reversal Detection

Phát hiện:

- Bullish Reversal
- Bearish Reversal

---

## Trend Continuation

Đánh giá khả năng:

- Trend Continuation
- Trend Exhaustion

---

# 7. Evidence Generation

Ví dụ Evidence.

## Uptrend Evidence

Điều kiện:

- Higher High
- Higher Low
- Giá duy trì trên vùng hỗ trợ chính

---

## Downtrend Evidence

Điều kiện:

- Lower High
- Lower Low
- Giá duy trì dưới vùng kháng cự chính

---

## Sideway Evidence

Điều kiện:

- Giá dao động trong vùng giao dịch
- Không hình thành xu hướng rõ ràng

---

## Trend Weakening

Điều kiện:

- Động lượng xu hướng suy giảm
- Xuất hiện dấu hiệu mất cân bằng

---

# 8. Signal Generation

Ví dụ Signal.

## Bullish Trend

Evidence:

- Uptrend Evidence
- Confirmed Trend

---

## Bearish Trend

Evidence:

- Downtrend Evidence
- Confirmed Trend

---

## Neutral Trend

Evidence:

- Sideway Evidence
- Unconfirmed Trend

---

## Trend Reversal Watch

Evidence:

- Trend Weakening
- Reversal Evidence

---

# 9. Analysis Result

Trend Analysis Result bao gồm:

- Summary
- Current Trend
- Trend Strength
- Trend Confidence
- Supporting Evidence
- Supporting Signals
- Warnings

---

# 10. Quality Metrics

Module phải đảm bảo:

- Explainability
- Traceability
- Consistency
- Deterministic Result

---

# 11. Edge Cases

Ví dụ:

- Sideway kéo dài
- False Breakout
- False Breakdown
- Gap Up / Gap Down
- Dữ liệu thiếu
- Biến động bất thường do sự kiện

Module phải trả về trạng thái phù hợp thay vì suy diễn.

---

# 12. Related Documents

- RFC-001 — Market Intelligence Specification
- RFC-002 — Information Model
- RFC-003 — Knowledge Contracts
- RFC-004 — Business Rules
