# RFC-005 — Liquidity Analysis Module

**Status:** Draft  
**Version:** 1.0  
**Owner:** Investment AI Team  
**Parent Capability:** Market Intelligence

---

# 1. Purpose

## Objective

Liquidity Analysis đánh giá mức độ thanh khoản của thị trường.

Module trả lời các câu hỏi:

- Thanh khoản hiện tại mạnh hay yếu?
- Thanh khoản đang cải thiện hay suy giảm?
- Thanh khoản có xác nhận xu hướng hiện tại không?
- Dòng tiền có đang mở rộng hay thu hẹp?

Module không đánh giá:

- Giá trị doanh nghiệp
- Xu hướng giá
- Chu kỳ kinh tế
- Khuyến nghị mua bán

---

# 2. Business Questions

Module phải trả lời được các câu hỏi sau.

- Thanh khoản đang tăng hay giảm?
- Thanh khoản có bền vững không?
- Khối lượng có xác nhận xu hướng giá không?
- Thanh khoản có bất thường không?
- Dòng tiền đang mở rộng hay co hẹp?

---

# 3. Scope

## In Scope

- Trading Volume
- Trading Value
- Turnover
- Relative Volume
- Volume Distribution
- Liquidity Trend
- Liquidity Quality

## Out of Scope

- Foreign Flow
- ETF Flow
- Sector Rotation
- Market Breadth
- Macro Indicators

---

# 4. Indicators

Module sử dụng các Indicator sau.

## Core Indicators

- Trading Volume
- Trading Value
- Turnover Ratio
- Average Volume
- Relative Volume

## Supporting Indicators

- Number of Transactions
- Average Trade Size
- Bid/Ask Depth
- Order Imbalance

---

# 5. Information Model

```
Indicators
      │
      ▼
Liquidity Evidence
      │
      ▼
Liquidity Signals
      │
      ▼
Liquidity Analysis Result
```

---

# 6. Analysis Rules

Liquidity Analysis sử dụng các nhóm Rule sau.

## Volume Analysis

Đánh giá:

- Volume Expansion
- Volume Contraction
- Volume Stability

---

## Trading Value Analysis

Đánh giá:

- Value Increasing
- Value Decreasing
- Value Stability

---

## Relative Volume Analysis

So sánh với:

- Historical Average
- Recent Sessions
- Market Average

---

## Liquidity Quality Analysis

Đánh giá:

- Healthy Liquidity
- Weak Liquidity
- Abnormal Liquidity

---

## Liquidity Trend Analysis

Đánh giá:

- Improving
- Stable
- Deteriorating

---

# 7. Evidence Generation

Ví dụ Evidence.

## Liquidity Improving

Điều kiện:

- Volume mở rộng.
- Trading Value mở rộng.
- Relative Volume cải thiện.

---

## Liquidity Weakening

Điều kiện:

- Volume giảm.
- Trading Value giảm.
- Relative Volume thấp.

---

## Abnormal Liquidity

Điều kiện:

- Volume tăng đột biến.
- Trading Value bất thường.
- Không phù hợp với hành vi lịch sử.

---

# 8. Signal Generation

Ví dụ Signal.

## Bullish Liquidity

Evidence:

- Liquidity Improving
- Healthy Liquidity

---

## Bearish Liquidity

Evidence:

- Liquidity Weakening
- Poor Liquidity

---

## Neutral Liquidity

Evidence trái chiều.

---

# 9. Analysis Result

Liquidity Analysis Result bao gồm:

- Summary
- Liquidity Score
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

- Phiên nghỉ lễ.
- Thanh khoản cực thấp.
- Dữ liệu thiếu.
- Thị trường mới mở cửa.
- Phiên ATC bất thường.

Module phải trả về trạng thái phù hợp thay vì suy diễn.

---

# 12. Related Documents

- RFC-001 — Market Intelligence Specification
- RFC-002 — Information Model
- RFC-003 — Knowledge Contracts
- RFC-004 — Business Rules
