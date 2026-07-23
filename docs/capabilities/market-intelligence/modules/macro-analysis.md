# RFC-009 — Macro Analysis Module

**Status:** Draft  
**Version:** 1.0  
**Owner:** Investment AI Team  
**Parent Capability:** Market Intelligence

---

# 1. Purpose

## Objective

Macro Analysis đánh giá môi trường kinh tế vĩ mô và tác động của nó đến thị trường tài chính.

Module trả lời các câu hỏi:

- Môi trường vĩ mô hiện tại thuận lợi hay bất lợi?
- Chính sách tiền tệ đang nới lỏng hay thắt chặt?
- Nền kinh tế đang mở rộng hay suy giảm?
- Các yếu tố vĩ mô có hỗ trợ xu hướng thị trường không?

Module không đánh giá:

- Xu hướng giá
- Thanh khoản
- Dòng tiền
- Giá trị doanh nghiệp

---

# 2. Business Questions

Module phải trả lời các câu hỏi sau:

- Môi trường vĩ mô hiện tại là gì?
- Chính sách tiền tệ đang thay đổi theo hướng nào?
- Lạm phát có đang được kiểm soát không?
- Tăng trưởng kinh tế mạnh hay yếu?
- Bối cảnh vĩ mô có hỗ trợ thị trường không?

---

# 3. Scope

## In Scope

- Monetary Policy
- Interest Rate
- Inflation
- Economic Growth
- Manufacturing Activity
- Credit Growth
- Money Supply
- Exchange Rate

## Out of Scope

- Company Financial Analysis
- Market Liquidity
- Capital Flow
- Market Breadth

---

# 4. Indicators

Module sử dụng các Indicator sau.

## Core Indicators

- Interest Rate
- CPI
- GDP Growth
- PMI
- Credit Growth
- Money Supply (M2)

## Supporting Indicators

- Exchange Rate
- Unemployment Rate
- PPI
- FDI
- Export Growth
- Import Growth
- Government Bond Yield

---

# 5. Information Model

```
Indicators
      │
      ▼
Macro Evidence
      │
      ▼
Macro Signals
      │
      ▼
Macro Analysis Result
```

---

# 6. Analysis Rules

Macro Analysis sử dụng các nhóm Rule sau.

## Monetary Policy Analysis

Đánh giá:

- Easing
- Neutral
- Tightening

---

## Inflation Analysis

Đánh giá:

- Inflation Rising
- Stable Inflation
- Inflation Cooling

---

## Economic Growth Analysis

Đánh giá:

- Expansion
- Stable Growth
- Slowdown
- Contraction

---

## Liquidity Environment

Đánh giá:

- Abundant Liquidity
- Normal Liquidity
- Tight Liquidity

---

## Macro Environment

Đánh giá:

- Favorable
- Neutral
- Unfavorable

---

# 7. Evidence Generation

Ví dụ Evidence.

## Favorable Macro

Điều kiện:

- Chính sách tiền tệ hỗ trợ.
- Tăng trưởng kinh tế ổn định.
- Lạm phát trong tầm kiểm soát.

---

## Tight Monetary Policy

Điều kiện:

- Lãi suất tăng.
- Thanh khoản bị thu hẹp.
- Điều kiện tín dụng chặt chẽ hơn.

---

## Economic Slowdown

Điều kiện:

- GDP tăng trưởng chậm.
- PMI suy yếu.
- Hoạt động sản xuất giảm.

---

## Inflation Pressure

Điều kiện:

- CPI tăng nhanh.
- PPI tăng kéo dài.
- Áp lực chi phí gia tăng.

---

# 8. Signal Generation

Ví dụ Signal.

## Bullish Macro

Evidence:

- Favorable Macro
- Easing Policy

---

## Bearish Macro

Evidence:

- Tight Monetary Policy
- Economic Slowdown

---

## Neutral Macro

Evidence trái chiều hoặc chưa đủ dữ liệu.

---

# 9. Analysis Result

Macro Analysis Result bao gồm:

- Summary
- Macro Environment
- Monetary Policy Status
- Inflation Status
- Economic Growth Status
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

- Dữ liệu kinh tế được công bố trễ.
- Chính sách thay đổi đột ngột.
- Biến động do sự kiện địa chính trị.
- Số liệu bị điều chỉnh sau khi công bố.
- Chỉ báo mâu thuẫn nhau.

Module phải phản ánh mức độ không chắc chắn thay vì đưa ra kết luận tuyệt đối.

---

# 12. Related Documents

- RFC-001 — Market Intelligence Specification
- RFC-002 — Information Model
- RFC-003 — Knowledge Contracts
- RFC-004 — Business Rules
