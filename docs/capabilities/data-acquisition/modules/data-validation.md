# Data Validation

**Status:** Draft  
**Version:** 2.0

---

# 1. Purpose

## Objective

Data Validation chịu trách nhiệm kiểm tra tính hợp lệ của dữ liệu trước khi dữ liệu được chuẩn hóa và publish.

Module đảm bảo mọi dữ liệu được sử dụng trong hệ thống đều đáp ứng các yêu cầu tối thiểu về tính đầy đủ, chính xác và nhất quán.

Module không thực hiện sửa đổi dữ liệu hoặc đánh giá ý nghĩa nghiệp vụ của dữ liệu.

---

# 2. Business Questions

Module trả lời các câu hỏi sau:

- Dữ liệu có hợp lệ không?
- Dữ liệu có đầy đủ không?
- Dữ liệu có nhất quán không?
- Dữ liệu có thể tiếp tục xử lý không?
- Dữ liệu có đủ điều kiện để publish không?

---

# 3. Scope

## In Scope

Kiểm tra:

- Required Information
- Data Type
- Data Format
- Business Constraints
- Duplicate Detection
- Referential Consistency
- Time Consistency
- Range Validation

---

## Out of Scope

Không bao gồm:

- Data Normalization
- Data Cleansing
- Data Enrichment
- Data Quality Scoring
- Business Analysis

---

# 4. Validation Inputs

Module áp dụng cho:

- Market Data
- Financial Data
- Macro Data
- News Data
- Corporate Actions

Mọi Data Object đều phải đi qua Validation Pipeline.

---

# 5. Validation Categories

## Required Information Validation

Kiểm tra:

- Required Fields
- Mandatory Attributes
- Empty Values

---

## Data Type Validation

Kiểm tra:

- Number
- String
- Date
- Boolean
- Enumeration

---

## Format Validation

Kiểm tra:

- Date Format
- Currency Format
- Symbol Format
- Percentage Format

---

## Business Constraint Validation

Kiểm tra:

- Business Rules
- Domain Constraints
- Allowed Values

---

## Duplicate Validation

Kiểm tra:

- Duplicate Record
- Duplicate Event
- Duplicate Publication

---

## Referential Validation

Kiểm tra:

- Company Exists
- Symbol Exists
- Exchange Exists
- Country Exists

---

## Time Validation

Kiểm tra:

- Trading Date
- Reporting Period
- Effective Date
- Timestamp Consistency

---

# 6. Validation Pipeline

```
Receive Data
        │
        ▼
Structural Validation
        │
        ▼
Business Validation
        │
        ▼
Consistency Validation
        │
        ▼
Validation Result
```

---

# 7. Validation Result

Validation có thể trả về:

## Passed

Dữ liệu hợp lệ.

Có thể tiếp tục xử lý.

---

## Warning

Dữ liệu hợp lệ nhưng có cảnh báo.

Có thể tiếp tục xử lý.

---

## Failed

Dữ liệu không hợp lệ.

Không được publish.

---

# 8. Validation Error

Validation Error phải bao gồm:

- Error Type
- Error Code
- Error Message
- Failed Rule
- Detection Time

---

# 9. Published Knowledge

Module publish:

- Validation Result
- Validation Errors
- Validation Metadata

Các module tiếp theo chỉ sử dụng Validation Result.

---

# 10. Consumers

Validation Result được sử dụng bởi:

- Data Normalization
- Data Quality
- Monitoring

---

# 11. Error Handling

Nếu Validation gặp lỗi:

- Dừng Processing Pipeline.
- Không publish dữ liệu.
- Ghi nhận Validation Error.
- Chuyển dữ liệu sang trạng thái Failed.

---

# 12. Quality Metrics

Theo dõi:

- Validation Success Rate
- Validation Failure Rate
- Duplicate Detection Rate
- Missing Data Rate
- Invalid Data Rate

---

# 13. Edge Cases

Ví dụ:

- Thiếu trường bắt buộc.
- Sai kiểu dữ liệu.
- Dữ liệu trùng.
- Sai định dạng ngày.
- Báo cáo thuộc sai kỳ.
- Timestamp trong tương lai.
- Company không tồn tại.
- Symbol đã hủy niêm yết.

---

# 14. Dependencies

## Consumes

- Raw Data

## Publishes

- Validation Result

---

# 15. Related Documents

- specification.md
- information-model.md
- knowledge-contracts.md
- business-rules.md
