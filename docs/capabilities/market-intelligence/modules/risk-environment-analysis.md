# RFC-011 — Risk Environment Analysis Module

**Status:** Draft
**Version:** 1.0
**Owner:** Investment AI Team
**Parent Capability:** Market Intelligence

---

# 1. Purpose

## Objective

Risk Environment Analysis đánh giá mức độ rủi ro tổng thể của môi trường đầu tư.

Module trả lời các câu hỏi:

- Môi trường hiện tại có mức rủi ro như thế nào?
- Nhà đầu tư nên ưu tiên Risk On hay Risk Off?
- Những yếu tố nào đang làm gia tăng rủi ro?
- Những yếu tố nào đang hỗ trợ thị trường?

Module không:

- Dự báo giá
- Khuyến nghị mua bán
- Định giá doanh nghiệp

---

# 2. Business Questions

Module phải trả lời các câu hỏi sau:

- Risk Level hiện tại là gì?
- Môi trường đang Risk On hay Risk Off?
- Những yếu tố rủi ro lớn nhất là gì?
- Mức độ tin cậy của đánh giá là bao nhiêu?

---

# 3. Scope

## In Scope

- Market Risk
- Macro Risk
- Liquidity Risk
- Volatility Risk
- Policy Risk
- Systemic Risk
- Market Confidence

## Out of Scope

- Company Risk
- Credit Risk của doanh nghiệp
- Portfolio Optimization
- Trading Strategy

---

# 4. Inputs

Module sử dụng kết quả từ:

## Core Inputs

- Market Trend Analysis
- Liquidity Analysis
- Breadth Analysis
- Capital Flow Analysis
- Macro Analysis
- Market Cycle Analysis

## Supporting Inputs

- Volatility Index
- Geopolitical Events
- Monetary Policy Events
- Economic Calendar

---

# 5. Information Model

```
Analysis Results
        │
        ▼
Risk Evidence
        │
        ▼
Risk Signals
        │
        ▼
Risk Environment Result
```

---

# 6. Analysis Rules

Risk Environment Analysis sử dụng các nhóm Rule sau.

## Market Risk

Đánh giá:

- Low Risk
- Medium Risk
- High Risk

---

## Liquidity Risk

Đánh giá:

- Healthy Liquidity
- Tight Liquidity

---

## Macro Risk

Đánh giá:

- Supportive
- Neutral
- Adverse

---

## Volatility Risk

Đánh giá:

- Stable
- Elevated
- Extreme

---

## Overall Risk Environment

Đánh giá:

- Risk On
- Neutral
- Risk Off

---

# 7. Evidence Generation

Ví dụ Evidence.

## Risk On

Điều kiện:

- Xu hướng tăng được xác nhận.
- Breadth mở rộng.
- Dòng tiền tích cực.
- Vĩ mô thuận lợi.

---

## Risk Off

Điều kiện:

- Xu hướng suy yếu.
- Thanh khoản giảm.
- Dòng tiền rút khỏi thị trường.
- Vĩ mô bất lợi.

---

## Elevated Risk

Điều kiện:

- Biến động tăng.
- Các tín hiệu mâu thuẫn.
- Sự kiện bất thường.

---

# 8. Signal Generation

Ví dụ Signal.

## Low Risk

Evidence:

- Risk On
- Stable Environment

---

## Medium Risk

Evidence:

- Mixed Signals
- Neutral Environment

---

## High Risk

Evidence:

- Risk Off
- Elevated Volatility

---

# 9. Analysis Result

Risk Environment Result bao gồm:

- Summary
- Overall Risk Level
- Market Regime
- Confidence
- Key Risk Factors
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

- Khủng hoảng tài chính.
- Thiên tai hoặc dịch bệnh.
- Xung đột địa chính trị.
- Chính sách bất ngờ từ ngân hàng trung ương.
- Các Analysis Module đưa ra tín hiệu trái ngược nhau.

Module phải phản ánh mức độ bất định thay vì đưa ra kết luận tuyệt đối.

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
- Market Cycle Analysis
