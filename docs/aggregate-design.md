# Aggregate Design

Version: 1.0

---

# 1. Purpose

Tài liệu này định nghĩa Aggregate Boundary của Investment AI.

Aggregate là đơn vị nhất quán (Consistency Boundary) trong Domain Model.

Mỗi Aggregate đại diện cho một Business Capability hoặc một Decision Stage.

Aggregate không được thiết kế dựa trên Database Table.

Aggregate không được thiết kế dựa trên API.

Aggregate phải phản ánh đúng quy trình đầu tư.

---

# 2. Design Principles

Aggregate phải:

* Có một Aggregate Root.
* Có ranh giới rõ ràng.
* Có trách nhiệm duy nhất.
* Có thể hoạt động độc lập.
* Không tham chiếu trực tiếp tới Aggregate khác.
* Giao tiếp thông qua Domain Event hoặc Domain Reference.

---

# 3. Aggregate Overview

```text
Investment AI

├── Market Aggregate
├── Company Aggregate
├── Financial Aggregate
├── Analysis Aggregate
├── Decision Aggregate
└── Report Aggregate
```

---

# 4. Market Aggregate

## Aggregate Root

Market

---

## Children

* Macro Indicator
* Liquidity
* Interest Rate
* Inflation
* Economic Cycle
* Market Breadth
* Market Sentiment

---

## Responsibilities

* Mô tả trạng thái thị trường.
* Không đưa ra Recommendation.
* Không đánh giá doanh nghiệp.

---

# 5. Company Aggregate

## Aggregate Root

Company

---

## Children

* Stock
* Sector
* Business Model
* Management
* Shareholders
* Competitive Advantage

---

## Responsibilities

* Mô tả doanh nghiệp.
* Không chứa Financial Analysis.
* Không chứa Technical Analysis.

---

# 6. Financial Aggregate

## Aggregate Root

Financial Statement

---

## Children

* Income Statement
* Balance Sheet
* Cash Flow
* Financial Metrics
* Financial Ratios
* Valuation

---

## Responsibilities

* Chuẩn hóa dữ liệu tài chính.
* Tính toán các chỉ số tài chính.
* Sinh Fair Value.

---

# 7. Analysis Aggregate

## Aggregate Root

Analysis Bundle

---

## Children

* Macro Analysis
* Sector Analysis
* Fundamental Analysis
* Technical Analysis
* Wyckoff Analysis
* VSA Analysis
* Risk Analysis

---

## Responsibilities

* Chuẩn hóa mọi kết quả phân tích.
* Không tạo Recommendation.
* Không tạo Report.

---

# 8. Decision Aggregate

## Aggregate Root

Investment Thesis

---

## Children

* Opportunities
* Risks
* Catalysts
* Recommendation
* Confidence

---

## Responsibilities

* Tổng hợp tất cả Analysis.
* Sinh Recommendation.
* Đánh giá mức độ tin cậy.

---

# 9. Report Aggregate

## Aggregate Root

Investment Report

---

## Children

* Executive Summary
* Analysis Results
* Investment Thesis
* Recommendation
* Appendix

---

## Responsibilities

* Biểu diễn kết quả.
* Không thực hiện Business Logic.

---

# 10. Aggregate Relationship

```text
Market Aggregate
            │
            ▼
Company Aggregate
            │
            ▼
Financial Aggregate
            │
            ▼
Analysis Aggregate
            │
            ▼
Decision Aggregate
            │
            ▼
Report Aggregate
```

Mỗi Aggregate chỉ phụ thuộc vào Aggregate trước đó trong quy trình phân tích.

---

# 11. Aggregate Ownership

| Aggregate           | Owner Capability     |
| ------------------- | -------------------- |
| Market Aggregate    | Market Intelligence  |
| Company Aggregate   | Company Intelligence |
| Financial Aggregate | Financial Analysis   |
| Analysis Aggregate  | Analysis Engine      |
| Decision Aggregate  | Investment Decision  |
| Report Aggregate    | Reporting            |

---

# 12. Aggregate Rules

* Aggregate chỉ có một Aggregate Root.
* Chỉ Aggregate Root được phép thay đổi trạng thái của Aggregate.
* Aggregate không truy cập trực tiếp vào Aggregate khác.
* Aggregate chỉ trao đổi thông qua Domain Model hoặc Domain Event.
* Mỗi Aggregate phải có Repository riêng.
* Business Rule không được vượt qua Aggregate Boundary.

---

# 13. Evolution Strategy

MVP sử dụng các Aggregate ở mức đơn giản.

Trong các phiên bản tiếp theo có thể mở rộng thêm:

* Portfolio Aggregate
* Watchlist Aggregate
* Strategy Aggregate
* Alert Aggregate
* User Preference Aggregate

Việc mở rộng Aggregate không được làm thay đổi Aggregate hiện có.
