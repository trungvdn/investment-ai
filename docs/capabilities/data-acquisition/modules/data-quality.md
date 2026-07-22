# Data Quality

**Status:** Draft  
**Version:** 2.0

---

# 1. Purpose

## Objective

Data Quality chịu trách nhiệm đánh giá và quản lý chất lượng của dữ liệu sau khi đã được Validation và Normalization.

Module đảm bảo chỉ những dữ liệu đáp ứng tiêu chuẩn chất lượng của hệ thống mới được công bố cho các Business Capability.

Module không thực hiện chỉnh sửa dữ liệu hoặc phân tích nghiệp vụ.

---

# 2. Business Questions

Module trả lời các câu hỏi sau:

- Dữ liệu có đủ chất lượng để sử dụng không?
- Chất lượng dữ liệu hiện tại là bao nhiêu?
- Dữ liệu có còn mới không?
- Có vấn đề gì ảnh hưởng đến độ tin cậy của dữ liệu không?
- Dữ liệu có thể publish không?

---

# 3. Scope

## In Scope

Đánh giá:

- Completeness
- Accuracy
- Consistency
- Freshness
- Timeliness
- Uniqueness
- Traceability

Quản lý:

- Quality Score
- Quality Status
- Quality Metadata

---

## Out of Scope

Không bao gồm:

- Data Validation
- Data Normalization
- Data Cleansing
- Business Analysis
- Provider Selection

---

# 4. Quality Inputs

Module áp dụng cho:

- Market Data
- Financial Data
- Macro Data
- News Data
- Corporate Actions

Chỉ dữ liệu đã hoàn thành Validation và Normalization mới được đánh giá.

---

# 5. Quality Dimensions

## Completeness

Đánh giá:

- Required Fields
- Optional Fields
- Missing Values

---

## Accuracy

Đánh giá:

- Giá trị có phản ánh đúng dữ liệu từ nguồn gốc.
- Không bị thay đổi ngoài ý muốn trong quá trình xử lý.

---

## Consistency

Đánh giá:

- Cùng một thực thể có dữ liệu nhất quán giữa các bản ghi.
- Không có mâu thuẫn nội bộ trong cùng một Data Object.

---

## Freshness

Đánh giá:

- Dữ liệu có còn mới không.
- Thời gian từ lúc dữ liệu được công bố đến lúc hệ thống tiếp nhận.

---

## Timeliness

Đánh giá:

- Dữ liệu có được xử lý và publish đúng thời gian yêu cầu không.

---

## Uniqueness

Đánh giá:

- Không có bản ghi trùng lặp sau khi publish.

---

## Traceability

Đánh giá:

- Có thể truy vết dữ liệu về Provider và thời điểm thu thập.

---

# 6. Quality Assessment Pipeline

```
Normalized Data
        │
        ▼
Measure Quality
        │
        ▼
Calculate Quality Metrics
        │
        ▼
Determine Quality Status
        │
        ▼
Publish Quality Report
```

---

# 7. Quality Result

Module tạo ra:

- Quality Score
- Quality Status
- Quality Report
- Quality Metadata

---

# 8. Quality Status

Dữ liệu được phân loại thành:

## Excellent

Đạt chất lượng rất cao.

Có thể sử dụng cho mọi Business Capability.

---

## Good

Đạt chất lượng tốt.

Có thể publish.

---

## Acceptable

Đủ điều kiện publish nhưng cần theo dõi.

---

## Poor

Không nên sử dụng cho các Business Capability.

---

## Rejected

Không được publish.

---

# 9. Published Knowledge

Module publish:

- Published Data
- Quality Report
- Quality Metadata

Đây là đầu ra cuối cùng của Data Acquisition Capability.

---

# 10. Consumers

Published Data được sử dụng bởi:

- Market Intelligence
- Fundamental Analysis
- Company Research
- Valuation
- Opportunity Discovery
- Portfolio Management
- Monitoring

---

# 11. Error Handling

Nếu phát hiện dữ liệu không đạt chất lượng:

- Gắn Quality Status phù hợp.
- Không publish nếu dưới ngưỡng quy định.
- Ghi nhận nguyên nhân.
- Chuyển sang quy trình xử lý lỗi nếu cần.

---

# 12. Quality Metrics

Theo dõi:

- Overall Quality Score
- Completeness Rate
- Accuracy Rate
- Consistency Rate
- Freshness Rate
- Timeliness Rate
- Duplicate Rate
- Publish Success Rate

---

# 13. Edge Cases

Ví dụ:

- Dữ liệu hợp lệ nhưng đến quá muộn.
- Thiếu một số trường không bắt buộc.
- Hai Provider cung cấp giá trị khác nhau.
- Dữ liệu đã cũ nhưng chưa có bản cập nhật.
- Một phần dữ liệu không thể truy vết nguồn.

---

# 14. Dependencies

## Consumes

- Canonical Data

## Publishes

- Published Data
- Quality Report

---

# 15. Related Documents

- specification.md
- information-model.md
- knowledge-contracts.md
- business-rules.md
