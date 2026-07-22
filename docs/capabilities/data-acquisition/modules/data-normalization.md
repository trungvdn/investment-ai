# Data Normalization

**Status:** Draft  
**Version:** 2.0

---

# 1. Purpose

## Objective

Data Normalization chịu trách nhiệm chuyển đổi dữ liệu hợp lệ từ nhiều Data Provider sang Canonical Data Model của Investment AI.

Module đảm bảo mọi Business Capability sử dụng cùng một định dạng dữ liệu, bất kể dữ liệu được thu thập từ nguồn nào.

Module không thực hiện phân tích dữ liệu hay đánh giá chất lượng dữ liệu.

---

# 2. Business Questions

Module trả lời các câu hỏi sau:

- Dữ liệu đã tuân theo Canonical Data Model chưa?
- Dữ liệu có cần chuyển đổi định dạng không?
- Dữ liệu đã được chuẩn hóa để các Business Capability sử dụng chưa?
- Dữ liệu có thể publish được chưa?

---

# 3. Scope

## In Scope

Chuẩn hóa:

- Field Naming
- Data Type
- Date & Time
- Timezone
- Currency
- Number Format
- Symbol
- Exchange
- Country
- Unit of Measurement
- Enumeration Values

---

## Out of Scope

Không bao gồm:

- Data Validation
- Data Cleansing
- Data Quality Assessment
- Business Analysis
- Data Enrichment

---

# 4. Normalization Inputs

Module áp dụng cho:

- Market Data
- Financial Data
- Macro Data
- News Data
- Corporate Actions

Chỉ dữ liệu đã vượt qua Validation mới được chuẩn hóa.

---

# 5. Normalization Categories

## Naming Normalization

Chuẩn hóa tên trường theo Canonical Data Model.

Ví dụ:

- Company Symbol
- Exchange
- Industry
- Indicator Name

---

## Data Type Normalization

Chuẩn hóa kiểu dữ liệu.

Ví dụ:

- Number
- String
- Boolean
- Date
- Timestamp

---

## Date & Time Normalization

Chuẩn hóa:

- Date Format
- Timestamp Format
- Timezone

---

## Currency Normalization

Chuẩn hóa:

- Currency Code
- Currency Precision

---

## Symbol Normalization

Chuẩn hóa:

- Stock Symbol
- Index Symbol
- ETF Symbol

---

## Unit Normalization

Chuẩn hóa:

- Percentage
- Currency
- Volume
- Value
- Basis Point

---

## Enumeration Normalization

Chuẩn hóa các giá trị phân loại.

Ví dụ:

- Market Status
- Trading Session
- Event Type
- Report Type

---

# 6. Normalization Pipeline

```
Validated Data
        │
        ▼
Field Mapping
        │
        ▼
Format Conversion
        │
        ▼
Canonical Mapping
        │
        ▼
Normalized Data
```

---

# 7. Canonical Data Model

Sau khi chuẩn hóa:

- Một Business Object chỉ có một định dạng chuẩn.
- Các Business Capability không phụ thuộc vào Data Provider.
- Mọi Published Data đều tuân theo Canonical Data Model.

---

# 8. Normalization Result

Module tạo ra:

- Normalized Data
- Transformation Metadata
- Mapping Information

---

# 9. Published Knowledge

Module publish:

- Canonical Data
- Transformation Metadata

Các module phía sau chỉ sử dụng Canonical Data.

---

# 10. Consumers

Normalized Data được sử dụng bởi:

- Data Quality
- Market Intelligence
- Fundamental Analysis
- Company Research
- Valuation
- Monitoring

---

# 11. Error Handling

Nếu xảy ra lỗi:

- Mapping Failure
- Unsupported Format
- Unknown Enumeration
- Unsupported Unit
- Conversion Failure

Module không publish dữ liệu chưa chuẩn hóa thành công.

---

# 12. Quality Metrics

Theo dõi:

- Normalization Success Rate
- Mapping Success Rate
- Conversion Failure Rate
- Unsupported Value Rate
- Canonical Compliance Rate

---

# 13. Edge Cases

Ví dụ:

- Hai Provider dùng tên trường khác nhau cho cùng một khái niệm.
- Hai Provider sử dụng đơn vị đo khác nhau.
- Khác múi giờ.
- Khác định dạng ngày.
- Khác chuẩn mã tiền tệ.
- Một Provider bổ sung trường mới.
- Một Provider ngừng cung cấp trường cũ.

---

# 14. Dependencies

## Consumes

- Validation Result

## Publishes

- Canonical Data

---

# 15. Related Documents

- specification.md
- information-model.md
- knowledge-contracts.md
- business-rules.md
