# Data Quality

**Status:** Draft  
**Version:** 2.1

---

# 1. Purpose

## Objective

Data Quality chịu trách nhiệm đánh giá chất lượng của dữ liệu đã được Validation và Normalization nhằm xác định dữ liệu có đủ điều kiện để công bố hay không.

Module hoạt động như **Quality Assessment Engine**, thực hiện việc đo lường các khía cạnh chất lượng của dữ liệu, đánh giá dữ liệu theo các Quality Policy và tạo ra Quality Result cho các bước xử lý tiếp theo.

Module không thực hiện:

- Data Validation
- Data Normalization
- Data Cleansing
- Business Analysis
- Data Publishing

---

# 2. Responsibilities

Data Quality chịu trách nhiệm:

- Assess Data Quality
- Calculate Quality Metrics
- Evaluate Quality Policies
- Determine Quality Status
- Generate Quality Result

---

# 3. Business Questions

Module trả lời các câu hỏi sau:

- Dữ liệu có đạt tiêu chuẩn chất lượng không?
- Dữ liệu có thể publish không?
- Dữ liệu đang vi phạm tiêu chuẩn nào?
- Chất lượng dữ liệu hiện tại như thế nào?

---

# 4. Scope

## In Scope

- Quality Assessment
- Quality Evaluation
- Quality Classification
- Quality Decision

---

## Out of Scope

- Validation Rules
- Normalization Rules
- Quality Policy Definition
- Metric Definition
- Report Formatting
- Data Repair

---

# 5. Inputs

Module nhận:

- Canonical Data
- Validation Result
- Quality Policies
- Quality Metrics Definition

---

# 6. Assessment Dimensions

Assessment Engine đánh giá dữ liệu trên các chiều chất lượng được định nghĩa bởi hệ thống.

Ví dụ:

- Completeness
- Accuracy
- Consistency
- Freshness
- Timeliness
- Uniqueness
- Traceability

Chi tiết từng Dimension được định nghĩa trong:

- quality-metrics.md

---

# 7. Assessment Pipeline

```text
Canonical Data
        │
        ▼
Load Quality Policies
        │
        ▼
Measure Quality Metrics
        │
        ▼
Evaluate Policies
        │
        ▼
Determine Quality Status
        │
        ▼
Generate Quality Result
```

---

# 8. Quality Evaluation

Assessment Engine sử dụng:

- Quality Policies
- Quality Metrics

để đánh giá dữ liệu.

Engine không chứa bất kỳ threshold nào.

Mọi threshold đều được định nghĩa trong:

quality-policies.md

---

# 9. Quality Decision

Sau khi đánh giá, Engine đưa ra một trong các kết quả:

- Pass
- Pass with Warning
- Reject

Decision chỉ phản ánh kết quả đánh giá.

Engine không quyết định dữ liệu sẽ được publish như thế nào.

---

# 10. Quality Result

Module tạo ra:

- Quality Status
- Quality Score
- Violated Policies
- Assessment Metadata

Quality Result được sử dụng bởi các module tiếp theo.

---

# 11. Error Handling

Nếu Assessment Engine gặp lỗi:

- Policy Loading Failure
- Metric Calculation Failure
- Unsupported Metric
- Invalid Policy
- Internal Processing Error

Module trả về Assessment Failure.

---

# 12. Dependencies

## Consumes

- Validation Result
- Canonical Data
- Quality Policies
- Quality Metrics

## Produces

- Quality Result

---

# 13. Related Documents

- specification.md
- information-model.md
- business-rules.md
- quality-policies.md
- quality-metrics.md
- quality-report.md
