# Data Quality

**Status:** Draft  
**Version:** 2.0

---

# 1. Purpose

## Objective

Data Quality chịu trách nhiệm đánh giá chất lượng của dữ liệu sau khi dữ liệu đã hoàn thành Validation và Normalization.

Module đảm bảo chỉ những dữ liệu đáp ứng tiêu chuẩn chất lượng của hệ thống mới được công bố cho các Business Capability.

Module hoạt động như một **Quality Assessment Engine**, đánh giá dữ liệu dựa trên các tiêu chí chất lượng đã được định nghĩa.

---

# 2. Responsibilities

Data Quality chịu trách nhiệm:

- Assess Data Quality
- Measure Quality Metrics
- Evaluate Quality Policies
- Determine Quality Status
- Generate Quality Report
- Produce Published Data

Module không chịu trách nhiệm:

- Data Collection
- Data Validation
- Data Normalization
- Data Cleansing
- Business Analysis

---

# 3. Business Questions

Module trả lời các câu hỏi sau:

- Dữ liệu có đủ chất lượng để sử dụng không?
- Dữ liệu có thể publish không?
- Dữ liệu đang vi phạm tiêu chuẩn nào?
- Chất lượng dữ liệu hiện tại là bao nhiêu?
- Có thể tin cậy dữ liệu này không?

---

# 4. Scope

## In Scope

- Data Quality Assessment
- Quality Metrics
- Quality Policy Evaluation
- Quality Classification
- Quality Reporting
- Publish Decision

---

## Out of Scope

- Validation Rules
- Normalization Rules
- Data Repair
- Business Decision
- Investment Analysis

---

# 5. Inputs

Module nhận dữ liệu từ:

- Data Validation
- Data Normalization

Áp dụng cho:

- Market Data
- Financial Data
- Macro Data
- News Data
- Corporate Actions

---

# 6. Quality Dimensions

Data Quality đánh giá dữ liệu theo các chiều chất lượng sau.

## Completeness

Đánh giá mức độ đầy đủ của dữ liệu.

Ví dụ:

- Required Fields
- Missing Values
- Optional Fields

---

## Accuracy

Đánh giá dữ liệu có phản ánh đúng dữ liệu từ nguồn gốc.

Ví dụ:

- Source Accuracy
- Value Integrity

---

## Consistency

Đánh giá tính nhất quán.

Ví dụ:

- Internal Consistency
- Cross Data Consistency

---

## Freshness

Đánh giá độ mới của dữ liệu.

Ví dụ:

- Market Data Delay
- News Delay
- Financial Report Age

---

## Timeliness

Đánh giá dữ liệu có được xử lý đúng SLA.

---

## Uniqueness

Đánh giá dữ liệu có bị trùng lặp.

---

## Traceability

Đánh giá khả năng truy vết.

Ví dụ:

- Source
- Provider
- Version
- Acquisition Time

---

# 7. Quality Assessment Pipeline

```
Canonical Data
        │
        ▼
Measure Quality Dimensions
        │
        ▼
Calculate Quality Metrics
        │
        ▼
Evaluate Quality Policies
        │
        ▼
Determine Quality Status
        │
        ▼
Generate Quality Report
        │
        ▼
Publish Decision
```

---

# 8. Quality Policies

Quality Policy định nghĩa các tiêu chuẩn chất lượng mà dữ liệu phải đáp ứng.

Policy được cấu hình độc lập với Business Logic.

Mỗi loại dữ liệu có thể sử dụng Policy khác nhau.

Ví dụ:

### Market Data

- Maximum Delay
- Minimum Completeness
- Maximum Duplicate Rate

---

### Financial Data

- Required Financial Statements
- Reporting Period Consistency
- Currency Consistency

---

### Macro Data

- Reporting Frequency
- Publication Date
- Indicator Availability

---

### News Data

- Maximum Delay
- Required Metadata
- Trusted Source

---

### Corporate Actions

- Required Event Date
- Required Symbol
- Announcement Completeness

---

# 9. Quality Status

Sau khi đánh giá, dữ liệu được phân loại thành:

## Excellent

Đáp ứng toàn bộ tiêu chuẩn chất lượng.

---

## Good

Đáp ứng phần lớn tiêu chuẩn.

Có thể publish.

---

## Acceptable

Đủ điều kiện publish nhưng cần theo dõi.

---

## Warning

Có vấn đề về chất lượng.

Có thể publish nếu Policy cho phép.

---

## Rejected

Không đạt tiêu chuẩn.

Không được publish.

---

# 10. Publish Decision

Sau khi đánh giá chất lượng, Module đưa ra quyết định:

```
Pass
        │
        ├── Publish
        │
        ▼
Pass with Warning
        │
        ├── Publish with Warning
        │
        ▼
Reject
        │
        └── Do Not Publish
```

Publish Decision chỉ dựa trên kết quả Quality Assessment.

---

# 11. Quality Report

Module tạo báo cáo chất lượng bao gồm:

- Quality Score
- Quality Status
- Failed Dimensions
- Policy Violations
- Assessment Time
- Source
- Provider

Báo cáo được sử dụng cho:

- Monitoring
- Audit
- Troubleshooting
- Data Governance

---

# 12. Quality Metrics

Module theo dõi các chỉ số:

- Overall Quality Score
- Completeness Rate
- Accuracy Rate
- Consistency Rate
- Freshness Rate
- Timeliness Rate
- Duplicate Rate
- Policy Compliance Rate
- Publish Success Rate

Các chỉ số này phục vụ:

- Theo dõi chất lượng dữ liệu
- Phân tích xu hướng
- Đánh giá Data Provider
- Cải thiện hệ thống

---

# 13. Error Handling

Nếu Assessment gặp lỗi:

- Metric Calculation Failure
- Policy Evaluation Failure
- Internal Processing Error
- Unsupported Metric

Module trả về Quality Assessment Failure.

---

# 14. Edge Cases

Ví dụ:

- Market Data hợp lệ nhưng chậm 15 phút.
- Financial Report thiếu Notes.
- Hai Provider có dữ liệu khác nhau.
- Tin tức bị cập nhật sau khi đã publish.
- Corporate Action bị điều chỉnh.
- Không thể truy vết Provider.
- Dữ liệu trùng từ nhiều nguồn.

---

# 15. Outputs

Module tạo ra:

- Published Data
- Quality Status
- Quality Score
- Quality Report
- Quality Metadata

Đây là đầu ra cuối cùng của Data Acquisition Capability.

---

# 16. Dependencies

## Consumes

- Validation Result
- Canonical Data

## Publishes

- Published Data

---

# 17. Related Documents

- specification.md
- information-model.md
- knowledge-contracts.md
- business-rules.md
- data-validation.md
- data-normalization.md
