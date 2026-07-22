# Data Quality

**Status:** Draft  
**Version:** 2.0

---

# 1. Purpose

## Objective

Data Quality chịu trách nhiệm đánh giá, quản lý và công bố chất lượng dữ liệu sau khi dữ liệu đã hoàn thành Validation và Normalization.

Module đảm bảo mọi Published Data đều đáp ứng các tiêu chuẩn chất lượng của hệ thống trước khi được sử dụng bởi các Business Capability.

Data Quality không sửa dữ liệu, không chuẩn hóa dữ liệu và không thực hiện phân tích nghiệp vụ.

---

# 2. Business Questions

Module trả lời các câu hỏi sau:

- Dữ liệu có đủ chất lượng để publish không?
- Chất lượng dữ liệu hiện tại là bao nhiêu?
- Dữ liệu vi phạm tiêu chuẩn chất lượng nào?
- Dữ liệu có còn mới không?
- Có thể tin cậy dữ liệu này không?

---

# 3. Scope

## In Scope

- Data Quality Assessment
- Quality Metrics
- Quality Policy Evaluation
- Quality Status Assignment
- Quality Reporting

---

## Out of Scope

- Data Validation
- Data Normalization
- Data Cleansing
- Data Repair
- Business Analysis

---

# 4. Inputs

Module nhận dữ liệu đã hoàn thành:

- Validation
- Normalization

Áp dụng cho:

- Market Data
- Financial Data
- Macro Data
- News Data
- Corporate Actions

---

# 5. Quality Dimensions

Data Quality đánh giá dữ liệu dựa trên các chiều chất lượng sau.

## Completeness

Đánh giá mức độ đầy đủ của dữ liệu.

Ví dụ:

- Required Fields
- Optional Fields
- Missing Values

---

## Accuracy

Đánh giá mức độ phản ánh đúng dữ liệu từ nguồn gốc.

Ví dụ:

- Source Consistency
- Numeric Accuracy
- Value Integrity

---

## Consistency

Đánh giá tính nhất quán của dữ liệu.

Ví dụ:

- Internal Consistency
- Cross Dataset Consistency
- Business Rule Consistency

---

## Freshness

Đánh giá độ mới của dữ liệu.

Ví dụ:

- Market Data Delay
- News Delay
- Financial Report Age

---

## Timeliness

Đánh giá dữ liệu có được xử lý đúng SLA hay không.

---

## Uniqueness

Đánh giá dữ liệu có bị trùng lặp không.

---

## Traceability

Đánh giá khả năng truy vết.

Ví dụ:

- Source
- Provider
- Version
- Acquisition Time

---

# 6. Quality Assessment Pipeline

```
Normalized Data
        │
        ▼
Measure Quality
        │
        ▼
Calculate Metrics
        │
        ▼
Evaluate Policy
        │
        ▼
Assign Quality Status
        │
        ▼
Generate Quality Report
```

---

# 7. Quality Policy

Quality Policy định nghĩa tiêu chuẩn chất lượng tối thiểu mà Published Data phải đáp ứng.

Policy được cấu hình độc lập với Business Logic.

Ví dụ:

- Minimum Completeness
- Maximum Delay
- Minimum Accuracy
- Maximum Duplicate Rate
- Required Traceability

Mỗi loại dữ liệu có thể sử dụng Policy khác nhau.

Ví dụ:

- Market Data
- Financial Data
- Macro Data
- News Data

---

# 8. Quality Evaluation

Module so sánh kết quả Assessment với Quality Policy.

Nếu đạt Policy

↓

Publish

Nếu không đạt

↓

Reject hoặc Warning

---

# 9. Quality Status

Module phân loại dữ liệu thành:

## Excellent

Đáp ứng toàn bộ Quality Policy.

---

## Good

Đáp ứng hầu hết Policy.

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

Không đáp ứng Policy.

Không được publish.

---

# 10. Quality Report

Module tạo báo cáo chất lượng bao gồm:

- Quality Score
- Quality Status
- Failed Dimensions
- Policy Violations
- Assessment Timestamp
- Data Source
- Provider

---

# 11. Published Knowledge

Module publish:

- Published Data
- Quality Report
- Quality Metadata

Đây là đầu ra cuối cùng của Data Acquisition Capability.

---

# 12. Consumers

Published Data được sử dụng bởi:

- Market Intelligence
- Fundamental Analysis
- Company Research
- Valuation
- Opportunity Discovery
- Portfolio Management
- Monitoring

---

# 13. Error Handling

Nếu dữ liệu không đạt Policy:

- Không publish nếu Policy yêu cầu Reject.
- Sinh Quality Report.
- Ghi nhận Policy Violation.
- Chuyển dữ liệu sang trạng thái Failed hoặc Warning.

---

# 14. Quality Metrics

Theo dõi:

- Overall Quality Score
- Completeness Rate
- Accuracy Rate
- Consistency Rate
- Freshness Rate
- Timeliness Rate
- Duplicate Rate
- Policy Compliance Rate
- Publish Success Rate

---

# 15. Edge Cases

Ví dụ:

- Market Data hợp lệ nhưng chậm 10 phút.
- Financial Report thiếu Notes.
- Hai Provider có dữ liệu khác nhau.
- News đến sau khi thị trường đóng cửa.
- Corporate Action bị chỉnh sửa sau khi publish.
- Dữ liệu không thể truy vết nguồn.

---

# 16. Dependencies

## Consumes

- Canonical Data

## Publishes

- Published Data
- Quality Report

---

# 17. Related Documents

- specification.md
- information-model.md
- knowledge-contracts.md
- business-rules.md
