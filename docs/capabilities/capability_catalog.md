# Capability Catalog

Version: 1.0

---

# 1. Purpose

Capability Catalog định nghĩa toàn bộ Business Capability của Investment AI.

Đây là danh mục chuẩn để:

* Thiết kế Workflow.
* Phân chia nhóm phát triển.
* Xây dựng AI Agent.
* Quản lý phạm vi hệ thống.

Capability Catalog không mô tả implementation.

Capability Catalog chỉ mô tả trách nhiệm nghiệp vụ.

---

# 2. Capability Hierarchy

```text
Investment AI

├── Data Acquisition
│
├── Market Intelligence
│
├── Company Intelligence
│
├── Financial Analysis
│
├── Technical Analysis
│
├── Valuation
│
├── Risk Assessment
│
├── Investment Decision
│
├── Reporting
│
└── Portfolio (Future)
```

---

# 3. Data Acquisition

## Purpose

Thu thập dữ liệu từ các nguồn bên ngoài.

### Responsibilities

* Market Data
* Financial Statements
* Macro Data
* News
* Corporate Actions

### Output

RawData

---

# 4. Market Intelligence

## Purpose

Đánh giá bối cảnh thị trường.

### Responsibilities

* Market Trend
* Liquidity
* Macro
* Economic Cycle
* Market Breadth

### Output

AnalysisResult

---

# 5. Company Intelligence

## Purpose

Đánh giá doanh nghiệp.

### Responsibilities

* Business Model
* Competitive Advantage
* Management
* Sector Position
* Shareholders

### Output

AnalysisResult

---

# 6. Financial Analysis

## Purpose

Đánh giá sức khỏe tài chính.

### Responsibilities

* Financial Metrics
* Profitability
* Leverage
* Growth
* Cash Flow

### Output

AnalysisResult

---

# 7. Technical Analysis

## Purpose

Đánh giá hành vi giá và khối lượng.

### Responsibilities

* Trend
* Momentum
* Wyckoff
* VSA
* Market Structure

### Output

AnalysisResult

---

# 8. Valuation

## Purpose

Ước tính giá trị nội tại.

### Responsibilities

* Fair Value
* Margin of Safety
* Valuation Method
* Assumptions

### Output

AnalysisResult

---

# 9. Risk Assessment

## Purpose

Đánh giá rủi ro đầu tư.

### Responsibilities

* Business Risk
* Financial Risk
* Market Risk
* Liquidity Risk

### Output

AnalysisResult

---

# 10. Investment Decision

## Purpose

Tổng hợp tất cả các phân tích thành quyết định đầu tư.

### Responsibilities

* Investment Thesis
* Recommendation
* Confidence
* Decision Explanation

### Output

InvestmentThesis

Recommendation

---

# 11. Reporting

## Purpose

Trình bày kết quả cho người dùng.

### Responsibilities

* Executive Summary
* Investment Report
* Charts
* Evidence References

### Output

InvestmentReport

---

# 12. Future Capabilities

Các Capability dự kiến bổ sung sau MVP:

* Portfolio Management
* Watchlist
* Alert Engine
* Strategy Backtesting
* Screening Engine
* Market Scanner
* Knowledge Graph
* Reflection Engine
* Learning Engine

---

# 13. Dependency Matrix

| Capability           | Depends On                                                  |
| -------------------- | ----------------------------------------------------------- |
| Data Acquisition     | None                                                        |
| Market Intelligence  | Data Acquisition                                            |
| Company Intelligence | Data Acquisition                                            |
| Financial Analysis   | Data Acquisition                                            |
| Technical Analysis   | Data Acquisition                                            |
| Valuation            | Financial Analysis                                          |
| Risk Assessment      | Market Intelligence, Financial Analysis, Technical Analysis |
| Investment Decision  | All Analysis Capabilities                                   |
| Reporting            | Investment Decision                                         |

---

# 14. Ownership

Mỗi Capability phải có:

* Một Business Owner.
* Một Technical Owner.
* Một Capability Specification.
* Một Knowledge Contract.
* Một tập KPI.
* Một bộ Test Cases.

---

# 15. Design Rules

* Không có hai Capability cùng trách nhiệm.
* Capability không điều phối Capability khác.
* Workflow chịu trách nhiệm orchestration.
* Capability chỉ giao tiếp bằng Knowledge Contracts.
* Capability phải có thể thay thế implementation mà không thay đổi Workflow.
