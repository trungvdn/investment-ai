# Domain Model

Version: 1.0

---

# 1. Purpose

Domain Model định nghĩa các thực thể nghiệp vụ cốt lõi của Investment AI.

Đây là ngôn ngữ chung giữa:

* Business
* Workflow
* Capability
* API
* Database
* AI

Domain Model không phụ thuộc:

* Database
* Go Struct
* JSON
* LLM

---

# 2. Domain Overview

```text
Investment AI

├── Market Domain
├── Company Domain
├── Financial Domain
├── Analysis Domain
├── Decision Domain
└── Reporting Domain
```

Mỗi Domain chịu trách nhiệm cho một nhóm nghiệp vụ riêng biệt.

---

# 3. Market Domain

## Stock

Đại diện cho một mã cổ phiếu.

Thuộc tính chính:

* Symbol
* Exchange
* Company
* Sector
* Industry

Quan hệ:

Stock

↓

Company

↓

Sector

---

## Market

Đại diện cho toàn bộ thị trường.

Ví dụ:

* VNINDEX
* HNX
* UPCOM

Bao gồm:

* Market Breadth
* Liquidity
* Market Trend
* Market Sentiment

---

## Sector

Một nhóm doanh nghiệp hoạt động trong cùng ngành.

Một Sector có nhiều Company.

Một Company chỉ thuộc một Sector trong MVP.

---

## Price History

Dữ liệu lịch sử giao dịch.

Bao gồm:

* Open
* High
* Low
* Close
* Volume
* Trading Value

---

# 4. Company Domain

## Company

Thực thể đại diện cho doanh nghiệp.

Bao gồm:

* Company Profile
* Business Model
* Management
* Shareholders
* Competitive Advantage

Quan hệ:

Company

↓

Financial Statements

↓

Valuation

---

## Business Model

Mô tả cách doanh nghiệp tạo ra doanh thu.

Ví dụ:

* Revenue Streams
* Cost Structure
* Growth Drivers

---

## Management

Thông tin về ban lãnh đạo.

Bao gồm:

* CEO
* Board
* Ownership
* Governance

---

# 5. Financial Domain

## Financial Statement

Bao gồm:

* Income Statement
* Balance Sheet
* Cash Flow

---

## Financial Metrics

Các chỉ số được tính toán từ Financial Statement.

Ví dụ:

* ROE
* ROA
* Gross Margin
* Net Margin
* Debt Ratio
* EPS
* BVPS

---

## Valuation

Kết quả định giá.

Bao gồm:

* Fair Value
* Margin of Safety
* Valuation Method
* Assumptions

---

# 6. Analysis Domain

## Analysis

Đơn vị kết quả chuẩn của mọi Capability.

Mỗi Analysis luôn có:

* Summary
* Score
* Confidence
* Evidence
* Signals

Đây là Output chuẩn của toàn bộ hệ thống.

---

## Evidence

Thông tin hỗ trợ cho một kết luận.

Nguồn có thể là:

* Market Data
* Financial Data
* News
* Macro Data
* Technical Indicator

---

## Signal

Một phát hiện được sinh ra từ dữ liệu.

Ví dụ:

* Golden Cross
* Spring
* SOS
* No Supply

Signal chưa phải Recommendation.

---

# 7. Decision Domain

## Investment Thesis

Luận điểm đầu tư.

Được tạo bằng cách tổng hợp nhiều Analysis.

Bao gồm:

* Opportunities
* Risks
* Catalysts
* Key Drivers

---

## Recommendation

Khuyến nghị cuối cùng.

Enum:

* Strong Buy
* Buy
* Hold
* Sell
* Strong Sell

---

## Risk Assessment

Đánh giá rủi ro.

Bao gồm:

* Business Risk
* Financial Risk
* Market Risk
* Valuation Risk

---

# 8. Reporting Domain

## Investment Report

Đầu ra cuối cùng của hệ thống.

Bao gồm:

* Executive Summary
* Analyses
* Investment Thesis
* Recommendation
* Confidence
* Appendices

---

# 9. Domain Relationships

```text
Market
    │
    ▼
Sector
    │
    ▼
Company
    │
    ▼
Financial Statement
    │
    ▼
Financial Metrics
    │
    ▼
Valuation

Price History ─────────────┐
                            │
Macro Indicators ───────────┤
                            ▼
                    Analysis
                            │
                            ▼
                  Investment Thesis
                            │
                            ▼
                  Recommendation
                            │
                            ▼
                  Investment Report
```

---

# 10. Design Rules

* Domain Model không chứa Business Logic.
* Domain Model không phụ thuộc Database.
* Domain Model không phụ thuộc AI.
* Mọi Capability chỉ trao đổi bằng Domain Model.
* Mọi Workflow đều xoay quanh Domain Model.
* Mọi API đều ánh xạ từ Domain Model.
