# Domain Building Blocks

Version: 1.0

---

# 1. Purpose

Tài liệu này định nghĩa các thành phần xây dựng (Building Blocks) của Domain Model.

Investment AI áp dụng Domain-Driven Design (DDD).

Business Logic phải được biểu diễn thông qua các Building Blocks thay vì phụ thuộc vào framework hoặc database.

---

# 2. Overview

```text
Entity
│
├── Value Object
├── Domain Service
├── Domain Event
├── Repository
└── Factory
```

---

# 3. Entity

Entity là đối tượng có định danh (Identity).

Identity không thay đổi trong suốt vòng đời của Entity.

Ví dụ:

* Stock
* Company
* Sector
* Market
* Investment Report

Entity có thể thay đổi thuộc tính nhưng Identity không thay đổi.

---

# 4. Value Object

Value Object không có Identity.

Hai Value Object bằng nhau nếu toàn bộ giá trị của chúng bằng nhau.

Value Object nên immutable.

---

## Price

```text
Price

- value
- currency
```

Ví dụ:

100.25 VND

---

## Percentage

```text
Percentage

- value
```

Ví dụ:

15%

---

## Score

```text
Score

- value
- max
```

Ví dụ:

82 / 100

---

## Confidence

```text
Confidence

- score

- level
```

Level

* Low
* Medium
* High

---

## Money

```text
Money

- amount
- currency
```

---

## DateRange

```text
start

end
```

---

## Recommendation

Enum Value Object

```text
Strong Buy

Buy

Hold

Sell

Strong Sell
```

---

## Trend

```text
Bullish

Neutral

Bearish
```

---

# 5. Domain Service

Domain Service chứa Business Logic không thuộc riêng một Entity.

Không lưu trạng thái.

---

## MacroAnalysisService

Input

Macro Data

Output

Macro Analysis

---

## TechnicalAnalysisService

Input

Price History

Output

Technical Analysis

---

## FundamentalAnalysisService

Input

Financial Statement

Output

Fundamental Analysis

---

## ValuationService

Input

Financial Data

Output

Fair Value

---

## RiskAssessmentService

Input

All Analyses

Output

Risk Assessment

---

## RecommendationService

Input

Investment Thesis

Risk

Output

Recommendation

---

# 6. Domain Event

Domain Event mô tả một sự kiện đã xảy ra trong Business Domain.

Event luôn ở thì quá khứ.

---

## MarketDataUpdated

Market Data đã được cập nhật.

---

## FinancialStatementLoaded

Báo cáo tài chính đã được tải.

---

## MacroAnalysisCompleted

Phân tích vĩ mô đã hoàn thành.

---

## TechnicalAnalysisCompleted

Phân tích kỹ thuật đã hoàn thành.

---

## InvestmentThesisGenerated

Luận điểm đầu tư đã được tạo.

---

## RecommendationGenerated

Khuyến nghị đầu tư đã được tạo.

---

## ReportGenerated

Báo cáo đầu tư đã được sinh.

---

# 7. Repository

Repository là abstraction để truy cập Entity.

Repository không chứa Business Logic.

Ví dụ:

StockRepository

CompanyRepository

FinancialRepository

MarketRepository

NewsRepository

---

# 8. Factory

Factory chịu trách nhiệm tạo các Aggregate phức tạp.

Ví dụ:

InvestmentReportFactory

InvestmentThesisFactory

AnalysisFactory

Factory giúp tách logic khởi tạo khỏi Business Logic.

---

# 9. Design Principles

* Entity có Identity.
* Value Object không có Identity.
* Value Object nên immutable.
* Domain Service không lưu trạng thái.
* Domain Event chỉ mô tả sự kiện đã xảy ra.
* Repository chỉ truy cập dữ liệu.
* Factory chỉ tạo đối tượng.
* Business Logic không nằm trong Repository.
* Business Logic không nằm trong Database.
* Business Logic không nằm trong API Layer.

```
```
