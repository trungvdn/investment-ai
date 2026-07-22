# Knowledge Contracts

Version: 1.0

---

# 1. Purpose

Knowledge Contracts định nghĩa các đối tượng chuẩn được trao đổi giữa các Capability trong Investment AI.

Mục tiêu:

* Chuẩn hóa giao tiếp giữa các Capability.
* Tách biệt Workflow khỏi implementation.
* Đảm bảo khả năng mở rộng và kiểm thử.
* Hỗ trợ Explainable AI thông qua khả năng truy vết.

Knowledge Contract không phụ thuộc:

* Go Struct
* Database
* Message Queue
* REST
* gRPC

Đây là mô hình nghiệp vụ, không phải mô hình truyền tải.

---

# 2. Knowledge Flow

```text
RawData
    │
    ▼
Evidence
    │
    ▼
Signal
    │
    ▼
AnalysisResult
    │
    ▼
InvestmentThesis
    │
    ▼
Recommendation
    │
    ▼
InvestmentReport
```

Mỗi Capability nhận một hoặc nhiều Knowledge Contract và trả về một Knowledge Contract mới.

---

# 3. RawData Contract

## Purpose

Đại diện cho dữ liệu gốc từ các nguồn bên ngoài.

## Required Fields

* Source
* Timestamp
* DataType
* Payload
* Metadata

## Characteristics

* Immutable.
* Không chứa Business Logic.
* Có thể truy vết đến nguồn dữ liệu.

---

# 4. Evidence Contract

## Purpose

Đại diện cho một bằng chứng phục vụ phân tích.

## Required Fields

* Evidence ID
* Source
* Description
* Category
* Confidence
* Supporting RawData

## Characteristics

* Có thể truy vết.
* Có thể được nhiều Analysis sử dụng.
* Không chứa Recommendation.

---

# 5. Signal Contract

## Purpose

Đại diện cho một tín hiệu được phát hiện từ Evidence.

## Required Fields

* Signal ID
* Type
* Description
* Related Evidence
* Confidence

## Examples

* Golden Cross
* Spring
* SOS
* No Supply
* EPS Growth

---

# 6. AnalysisResult Contract

## Purpose

Output chuẩn của mọi Capability phân tích.

## Required Fields

* Analysis ID
* Capability
* Summary
* Score
* Confidence
* Evidence List
* Signal List
* Risks

## Characteristics

* Mỗi Capability chỉ sinh một AnalysisResult trong một lần thực thi.
* Không sinh Recommendation.

---

# 7. InvestmentThesis Contract

## Purpose

Tổng hợp nhiều AnalysisResult thành một luận điểm đầu tư.

## Required Fields

* Thesis ID
* Opportunities
* Risks
* Catalysts
* Key Drivers
* Supporting Analyses

## Characteristics

* Là đầu vào duy nhất của Recommendation.
* Không phụ thuộc vào nguồn dữ liệu cụ thể.

---

# 8. Recommendation Contract

## Purpose

Khuyến nghị đầu tư cuối cùng.

## Required Fields

* Recommendation
* Confidence
* Supporting Thesis
* Generated Time

## Enum

* Strong Buy
* Buy
* Hold
* Sell
* Strong Sell

---

# 9. InvestmentReport Contract

## Purpose

Đầu ra cuối cùng dành cho người dùng.

## Required Fields

* Executive Summary
* Recommendation
* Investment Thesis
* Analysis Results
* Evidence References
* Appendices

Report chỉ trình bày kết quả, không sinh tri thức mới.

---

# 10. Contract Evolution Rules

* Không xóa trường bắt buộc khi nâng phiên bản.
* Chỉ bổ sung trường mới theo hướng tương thích ngược.
* Mỗi Contract phải có Version.
* Mỗi Contract phải có ID duy nhất.
* Mọi Capability phải khai báo Input Contract và Output Contract.

---

# 11. Capability Contract Matrix

| Capability           | Input                           | Output           |
| -------------------- | ------------------------------- | ---------------- |
| Market Intelligence  | RawData                         | AnalysisResult   |
| Company Intelligence | RawData                         | AnalysisResult   |
| Financial Analysis   | RawData                         | AnalysisResult   |
| Technical Analysis   | RawData                         | AnalysisResult   |
| Valuation            | AnalysisResult                  | AnalysisResult   |
| Risk Assessment      | AnalysisResult                  | AnalysisResult   |
| Investment Decision  | AnalysisResult                  | InvestmentThesis |
| Recommendation       | InvestmentThesis                | Recommendation   |
| Reporting            | Recommendation + AnalysisResult | InvestmentReport |

---

# 12. Design Principles

* Capability chỉ giao tiếp qua Knowledge Contract.
* Workflow không phụ thuộc vào implementation của Capability.
* Mọi Contract đều phải truy vết được về Knowledge Pipeline.
* Không truyền dữ liệu không cần thiết giữa các Capability.
* Contract phải ổn định hơn implementation.
