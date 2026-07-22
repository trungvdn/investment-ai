# Capability Specification

Version: 1.0

---

# 1. Purpose

Capability Specification định nghĩa chuẩn thiết kế cho mọi Business Capability trong Investment AI.

Mỗi Capability phải tuân theo cùng một cấu trúc để đảm bảo:

* Tính nhất quán.
* Khả năng thay thế implementation.
* Khả năng kiểm thử.
* Khả năng mở rộng.
* Khả năng orchestration bởi Workflow.

Capability là đơn vị nghiệp vụ, không phải Agent.

---

# 2. Capability Structure

Mỗi Capability phải định nghĩa các thành phần sau:

1. Business Purpose
2. Responsibilities
3. Inputs
4. Outputs
5. Dependencies
6. Business Rules
7. Preconditions
8. Postconditions
9. Error Model
10. Quality Metrics

---

# 3. Business Purpose

Mô tả mục tiêu nghiệp vụ.

Trả lời câu hỏi:

"Tại sao Capability này tồn tại?"

Ví dụ:

Technical Analysis

→ Đánh giá hành vi giá và khối lượng để tạo Technical Analysis.

---

# 4. Responsibilities

Capability chỉ nên có một trách nhiệm chính.

Ví dụ:

Financial Analysis

Có trách nhiệm:

* Phân tích báo cáo tài chính.
* Tính toán Financial Metrics.
* Sinh Financial Analysis.

Không có trách nhiệm:

* Recommendation.
* Reporting.
* Portfolio Management.

---

# 5. Inputs

Input phải sử dụng Knowledge Contracts.

Ví dụ:

* RawData
* Evidence
* AnalysisResult
* InvestmentThesis

Capability không được phụ thuộc vào:

* Database Schema.
* API Request.
* Message Queue Format.

---

# 6. Outputs

Output luôn là Knowledge Contract.

Ví dụ:

Technical Analysis

↓

AnalysisResult

Recommendation Engine

↓

Recommendation

Reporting

↓

InvestmentReport

---

# 7. Dependencies

Capability chỉ được phụ thuộc vào:

* Domain Model.
* Knowledge Contracts.
* Domain Services.

Không phụ thuộc trực tiếp vào Capability khác.

Mọi phối hợp được thực hiện bởi Workflow.

---

# 8. Business Rules

Capability phải khai báo rõ các Business Rules.

Ví dụ:

Valuation:

* Fair Value phải lớn hơn hoặc bằng 0.
* Margin of Safety có thể âm.
* Không định giá nếu thiếu Financial Statement.

Business Rules phải độc lập với implementation.

---

# 9. Preconditions

Điều kiện cần trước khi thực thi.

Ví dụ:

Technical Analysis:

* Có Price History.
* Có Volume History.

Nếu không thỏa mãn:

Capability phải trả về lỗi nghiệp vụ hoặc trạng thái "Insufficient Data".

---

# 10. Postconditions

Điều kiện sau khi hoàn thành.

Ví dụ:

Financial Analysis:

* Sinh đúng một AnalysisResult.
* AnalysisResult có ít nhất một Evidence.
* AnalysisResult có Confidence.

---

# 11. Error Model

Capability phải phân loại lỗi rõ ràng.

## Business Error

Ví dụ:

* Missing Financial Statement.
* Insufficient Market Data.
* Unsupported Exchange.

## Technical Error

Ví dụ:

* Database Timeout.
* API Unavailable.
* Network Failure.

Workflow xử lý Technical Error.

Capability xử lý Business Error.

---

# 12. Quality Metrics

Mỗi Capability phải định nghĩa KPI.

Ví dụ:

* Accuracy.
* Completeness.
* Explainability.
* Traceability.
* Latency.
* Cost.

Các KPI này phục vụ đánh giá chất lượng hệ thống và tối ưu trong tương lai.

---

# 13. Standard Capability Template

```text
Capability Name

Purpose

Responsibilities

Inputs

Outputs

Dependencies

Business Rules

Preconditions

Postconditions

Error Model

Quality Metrics
```

---

# 14. Capability Lifecycle

```text
Workflow
      │
      ▼
Validate Inputs
      │
      ▼
Load Domain Objects
      │
      ▼
Execute Domain Services
      │
      ▼
Generate Knowledge Contract
      │
      ▼
Validate Output
      │
      ▼
Return Result
```

Mọi Capability trong hệ thống phải tuân theo vòng đời này.

---

# 15. Design Principles

* Một Capability chỉ có một mục tiêu nghiệp vụ.
* Capability không lưu trạng thái.
* Capability không điều phối Capability khác.
* Capability không tạo dữ liệu ngoài Domain Model.
* Capability chỉ giao tiếp qua Knowledge Contracts.
* Capability luôn trả về kết quả có thể truy vết.
* Capability phải độc lập với AI Provider.
* Capability phải có thể thay thế implementation mà không ảnh hưởng Workflow.
