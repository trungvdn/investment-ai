# RFC-008 — Capital Flow Analysis Module

**Status:** Draft  
**Version:** 1.0  
**Owner:** Investment AI Team  
**Parent Capability:** Market Intelligence

---

# 1. Purpose

## Objective

Capital Flow Analysis đánh giá sự dịch chuyển của dòng tiền trong thị trường.

Module trả lời các câu hỏi:

- Dòng tiền đang vào hay rời khỏi thị trường?
- Dòng tiền đang tập trung vào nhóm tài sản nào?
- Có sự luân chuyển giữa các nhóm ngành không?
- Dòng tiền có xác nhận xu hướng hiện tại không?

Module không đánh giá:

- Xu hướng giá
- Thanh khoản
- Chu kỳ kinh tế
- Định giá doanh nghiệp

---

# 2. Business Questions

Module phải trả lời các câu hỏi sau:

- Dòng tiền hiện tại mạnh hay yếu?
- Dòng tiền đang vào hay ra khỏi thị trường?
- Ngành nào đang thu hút dòng tiền?
- Có hiện tượng Sector Rotation không?
- Dòng tiền có xác nhận xu hướng không?

---

# 3. Scope

## In Scope

- Market Capital Flow
- Sector Rotation
- Capital Concentration
- Institutional Flow
- Foreign Flow
- Proprietary Trading Flow
- ETF Flow

## Out of Scope

- Company Financial Analysis
- Macro Indicators
- Liquidity Analysis
- Market Breadth

---

# 4. Indicators

Module sử dụng các Indicator sau.

## Core Indicators

- Foreign Net Buy/Sell
- ETF Net Flow
- Proprietary Trading
- Sector Trading Value
- Sector Volume
- Capitalization Flow

## Supporting Indicators

- Sector Performance
- Large Cap Performance
- Mid Cap Performance
- Small Cap Performance
- Relative Strength by Sector

---

# 5. Information Model

```
Indicators
      │
      ▼
Capital Flow Evidence
      │
      ▼
Capital Flow Signals
      │
      ▼
Capital Flow Analysis Result
```

---

# 6. Analysis Rules

Capital Flow Analysis sử dụng các nhóm Rule sau.

## Market Flow Analysis

Đánh giá:

- Capital Inflow
- Capital Outflow
- Stable Flow

---

## Sector Rotation Analysis

Đánh giá:

- Rotation Into Sector
- Rotation Out Of Sector
- Balanced Rotation

---

## Institutional Flow Analysis

Đánh giá:

- Institutional Buying
- Institutional Selling
- Neutral Activity

---

## Foreign Flow Analysis

Đánh giá:

- Foreign Accumulation
- Foreign Distribution
- Neutral Foreign Activity

---

## Capital Concentration

Đánh giá:

- Broad Capital Distribution
- Concentrated Capital
- Narrow Leadership

---

# 7. Evidence Generation

Ví dụ Evidence.

## Capital Inflow

Điều kiện:

- Dòng tiền tăng trên toàn thị trường.
- Giá trị giao dịch mở rộng.
- Nhiều nhóm ngành được phân bổ vốn.

---

## Sector Rotation

Điều kiện:

- Dòng tiền rút khỏi một nhóm ngành.
- Đồng thời tăng vào nhóm ngành khác.

---

## Institutional Accumulation

Điều kiện:

- Dòng tiền tổ chức duy trì mua ròng.
- Có tính liên tục.

---

## Foreign Distribution

Điều kiện:

- Khối ngoại bán ròng kéo dài.
- Áp lực bán xuất hiện trên nhiều nhóm ngành.

---

# 8. Signal Generation

Ví dụ Signal.

## Bullish Capital Flow

Evidence:

- Capital Inflow
- Institutional Accumulation

---

## Bearish Capital Flow

Evidence:

- Capital Outflow
- Institutional Distribution

---

## Sector Rotation

Evidence:

- Rotation Into Sector
- Rotation Out Of Sector

---

## Neutral Capital Flow

Evidence chưa đủ mạnh hoặc trái chiều.

---

# 9. Analysis Result

Capital Flow Analysis Result bao gồm:

- Summary
- Market Flow Status
- Sector Rotation Status
- Institutional Activity
- Foreign Activity
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

- ETF cơ cấu danh mục.
- Đáo hạn phái sinh.
- Khối ngoại giao dịch thỏa thuận lớn.
- Tin tức bất thường làm thay đổi dòng tiền trong ngắn hạn.
- Một ngành bị chi phối bởi một cổ phiếu vốn hóa lớn.

Module phải phân biệt dòng tiền ngắn hạn và xu hướng dòng tiền bền vững.

---

# 12. Related Documents

- RFC-001 — Market Intelligence Specification
- RFC-002 — Information Model
- RFC-003 — Knowledge Contracts
- RFC-004 — Business Rules
