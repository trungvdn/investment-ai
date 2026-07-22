# Data Acquisition Information Model

**Status:** Draft  
**Version:** 1.0  
**Owner:** Investment AI Team

---

# 1. Purpose

Tài liệu này định nghĩa Information Model của Data Acquisition Capability.

Mục tiêu:

- Chuẩn hóa luồng dữ liệu.
- Định nghĩa các Business Object.
- Xác định quan hệ giữa các thành phần.
- Làm nền tảng cho Capability API và Implementation.

Information Model mô tả **dữ liệu nghiệp vụ**, không mô tả Database Schema hay Event Schema.

---

# 2. Information Flow

```
External Provider
        │
        ▼
Raw Data
        │
        ▼
Validation
        │
        ▼
Normalized Data
        │
        ▼
Metadata Enrichment
        │
        ▼
Quality Assessment
        │
        ▼
Published Data
```

---

# 3. Business Objects

Data Acquisition quản lý các Business Object sau.

---

## Data Source

Đại diện cho nguồn dữ liệu.

Ví dụ:

- Stock Exchange
- Financial Provider
- Government Agency
- News Provider
- Alternative Data Provider

Thuộc tính:

- Source ID
- Source Name
- Source Type
- Provider
- Status

---

## Raw Data

Dữ liệu nhận trực tiếp từ Data Source.

Ví dụ:

- JSON
- CSV
- XML
- API Response

Đặc điểm:

- Chưa kiểm tra
- Chưa chuẩn hóa
- Không dùng trực tiếp cho Business Capability

---

## Validation Result

Kết quả kiểm tra dữ liệu.

Bao gồm:

- Validation Status
- Validation Errors
- Missing Fields
- Invalid Values
- Duplicate Detection

---

## Normalized Data

Dữ liệu đã được chuẩn hóa.

Đặc điểm:

- Field Name thống nhất
- Data Type thống nhất
- Timezone thống nhất
- Currency thống nhất
- Symbol thống nhất

Đây là dữ liệu được sử dụng bởi các Capability khác.

---

## Metadata

Thông tin mô tả dữ liệu.

Ví dụ:

- Source
- Timestamp
- Version
- Currency
- Exchange
- Timezone
- Data Frequency
- Refresh Time

---

## Data Quality Report

Đánh giá chất lượng dữ liệu.

Bao gồm:

- Completeness
- Accuracy
- Freshness
- Consistency
- Timeliness

---

## Published Data

Dữ liệu cuối cùng được cung cấp cho các Capability.

Ví dụ:

- Market Data
- Financial Data
- Macro Data
- News Data

---

# 4. Business Relationships

```
Data Source
        │
        ▼
Raw Data
        │
        ▼
Validation Result
        │
        ▼
Normalized Data
        │
        ▼
Metadata
        │
        ▼
Quality Report
        │
        ▼
Published Data
```

---

# 5. Capability Interaction

```
External Provider
        │
        ▼
Data Acquisition
        │
        ├── Collect
        ├── Validate
        ├── Normalize
        ├── Enrich Metadata
        ├── Assess Quality
        └── Publish
        │
        ▼
Business Capabilities
```

---

# 6. Data Categories

Capability quản lý các nhóm dữ liệu sau.

## Market Data

Ví dụ:

- OHLCV
- Order Book
- Trade History
- Index Data
- Sector Data

---

## Company Data

Ví dụ:

- Company Profile
- Corporate Actions
- Shareholder Information
- Insider Trading

---

## Financial Data

Ví dụ:

- Income Statement
- Balance Sheet
- Cash Flow
- Financial Ratios

---

## Macro Data

Ví dụ:

- CPI
- GDP
- PMI
- Interest Rate
- Credit Growth
- Exchange Rate

---

## News Data

Ví dụ:

- Company News
- Market News
- Economic News
- Regulatory News

---

# 7. Processing States

Một Data Object có thể trải qua các trạng thái sau:

```
Collected

↓

Validated

↓

Normalized

↓

Enriched

↓

Quality Checked

↓

Published

↓

Archived
```

Nếu Validation thất bại:

```
Collected

↓

Validation Failed

↓

Rejected
```

---

# 8. Data Lifecycle

```
Acquire

↓

Validate

↓

Normalize

↓

Enrich

↓

Quality Check

↓

Publish

↓

Consume

↓

Archive
```

---

# 9. Ownership

| Object | Owner |
|----------|--------|
| Data Source | Data Acquisition |
| Raw Data | Data Acquisition |
| Validation Result | Data Acquisition |
| Normalized Data | Data Acquisition |
| Metadata | Data Acquisition |
| Data Quality Report | Data Acquisition |
| Published Data | Data Acquisition |

---

# 10. Related Documents

- README.md
- specification.md
- knowledge-contracts.md
- business-rules.md
