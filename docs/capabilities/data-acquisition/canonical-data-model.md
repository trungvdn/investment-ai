# Canonical Data Model

**Status:** Draft

**Version:** 1.0

---

# 1. Purpose

## Objective

Canonical Data Model định nghĩa các Business Object chuẩn được sử dụng trong toàn bộ Investment AI.

Mọi dữ liệu sau khi hoàn thành Validation, Normalization và Quality Assessment đều phải được chuyển đổi sang Canonical Data Model trước khi được Publish.

Business Capability không được phụ thuộc vào định dạng dữ liệu của Data Provider.

---

# 2. Design Principles

## Single Canonical Model

Mỗi Business Concept chỉ có một Data Model duy nhất.

Ví dụ:

- MarketData
- FinancialStatement
- NewsArticle

---

## Provider Independence

Canonical Model không phụ thuộc:

- HOSE
- HNX
- FiinPro
- CafeF
- AlphaVantage
- FRED

---

## Stable Contract

Canonical Model là hợp đồng giữa Data Acquisition và Business Capability.

Thay đổi Provider không làm thay đổi Business Capability.

---

## Business Oriented

Model phản ánh Business Concept thay vì cấu trúc của Database hoặc API.

---

# 3. Business Objects

Data Acquisition publish các Business Object sau.

---

## MarketData

Đại diện cho dữ liệu giao dịch của một mã chứng khoán.

Thuộc tính chính:

- Symbol
- Exchange
- Trading Date
- Open
- High
- Low
- Close
- Volume
- Value
- Metadata

---

## FinancialStatement

Đại diện cho báo cáo tài chính.

Bao gồm:

- Company
- Period
- Income Statement
- Balance Sheet
- Cash Flow
- Metadata

---

## CompanyProfile

Đại diện thông tin doanh nghiệp.

Bao gồm:

- Company Name
- Symbol
- Industry
- Sector
- Exchange
- Listing Date
- Shares Outstanding

---

## MacroIndicator

Đại diện chỉ báo kinh tế.

Ví dụ:

- GDP
- CPI
- PMI
- Interest Rate
- Exchange Rate
- Credit Growth

---

## NewsArticle

Đại diện một bản tin.

Bao gồm:

- News ID
- Title
- Summary
- Content
- Published Time
- Categories
- Companies
- Source
- Metadata

---

## CorporateAction

Đại diện sự kiện doanh nghiệp.

Ví dụ:

- Dividend
- Buyback
- Stock Split
- Rights Issue
- ESOP
- Listing
- Delisting

---

# 4. Common Metadata

Mọi Business Object đều có Metadata.

Bao gồm:

- Provider
- Source
- Version
- Acquisition Time
- Last Updated
- Quality Status
- Trace ID

---

# 5. Object Lifecycle

```
Raw Data
      │
      ▼
Validated Data
      │
      ▼
Normalized Data
      │
      ▼
Quality Assessed Data
      │
      ▼
Canonical Business Object
      │
      ▼
Published Data
```

---

# 6. Published Data

Chỉ Canonical Business Object mới được Publish.

Business Capability không được sử dụng:

- Raw Data
- Provider Format
- Intermediate Data

---

# 7. Consumers

Canonical Data được sử dụng bởi:

- Market Intelligence
- Fundamental Analysis
- Company Research
- Valuation
- Opportunity Discovery
- Portfolio Management
- Risk Management
- Monitoring

---

# 8. Related Documents

- specification.md
- information-model.md
- knowledge-contracts.md
- business-rules.md
- data-validation.md
- data-normalization.md
- data-quality.md
