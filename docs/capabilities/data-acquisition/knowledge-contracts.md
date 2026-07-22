# Data Acquisition Knowledge Contracts

**Status:** Draft  
**Version:** 1.0  
**Owner:** Investment AI Team

---

# 1. Purpose

Tài liệu này định nghĩa các Knowledge Contract được cung cấp bởi Data Acquisition Capability.

Knowledge Contract là giao diện nghiệp vụ giữa Data Acquisition và các Capability khác.

Mục tiêu:

- Chuẩn hóa dữ liệu đầu ra.
- Đảm bảo tính nhất quán giữa các Capability.
- Giảm coupling giữa Data Acquisition và Consumer.
- Làm nền tảng cho Capability API.

---

# 2. Principles

Mọi Knowledge Contract phải:

- Có ý nghĩa nghiệp vụ rõ ràng.
- Độc lập với Database Schema.
- Độc lập với API Schema.
- Có Version.
- Có Metadata.
- Có Data Quality.

---

# 3. Published Knowledge

Data Acquisition publish các Knowledge sau.

```
Data Acquisition

├── Market Data
├── Company Data
├── Financial Data
├── Macro Data
├── News Data
├── Corporate Actions
└── Data Quality
```

---

# 4. Market Data Contract

## Purpose

Cung cấp dữ liệu giao dịch thị trường.

## Contains

- Trading Information
- Price Information
- Volume Information
- Market Statistics
- Index Information
- Sector Information

## Consumers

- Market Intelligence
- Opportunity Discovery
- Monitoring

---

# 5. Company Data Contract

## Purpose

Cung cấp thông tin doanh nghiệp.

## Contains

- Company Profile
- Industry
- Listing Information
- Share Information
- Management Information

## Consumers

- Fundamental Analysis
- Company Research
- Valuation

---

# 6. Financial Data Contract

## Purpose

Cung cấp dữ liệu báo cáo tài chính.

## Contains

- Income Statement
- Balance Sheet
- Cash Flow Statement
- Financial Ratios

## Consumers

- Fundamental Analysis
- Valuation

---

# 7. Macro Data Contract

## Purpose

Cung cấp dữ liệu kinh tế vĩ mô.

## Contains

- Interest Rate
- CPI
- GDP
- PMI
- Credit Growth
- Exchange Rate
- Money Supply

## Consumers

- Market Intelligence
- Risk Management

---

# 8. News Data Contract

## Purpose

Cung cấp dữ liệu tin tức.

## Contains

- Company News
- Market News
- Macro News
- Regulatory News

## Consumers

- Company Research
- Monitoring
- Opportunity Discovery

---

# 9. Corporate Actions Contract

## Purpose

Cung cấp các sự kiện doanh nghiệp.

## Contains

- Dividend
- Stock Split
- Bonus Share
- Rights Issue
- Buyback
- Share Issuance

## Consumers

- Company Research
- Portfolio Management
- Monitoring

---

# 10. Data Quality Contract

## Purpose

Đánh giá chất lượng dữ liệu.

## Contains

- Completeness
- Accuracy
- Freshness
- Consistency
- Timeliness

## Consumers

Tất cả Capability.

---

# 11. Common Metadata

Mọi Knowledge Contract đều phải có Metadata.

## Required Metadata

- Contract Version
- Source
- Provider
- Timestamp
- Timezone
- Exchange
- Data Frequency

---

# 12. Contract Lifecycle

```
Create

↓

Validate

↓

Version

↓

Publish

↓

Consume

↓

Deprecate

↓

Archive
```

---

# 13. Compatibility Rules

Knowledge Contract phải đảm bảo:

- Backward Compatible khi có thể.
- Không thay đổi ý nghĩa nghiệp vụ của dữ liệu.
- Không xóa trường bắt buộc trong cùng Major Version.
- Thay đổi Breaking Change phải tạo Major Version mới.

---

# 14. Ownership

| Contract | Owner |
|----------|--------|
| Market Data | Data Acquisition |
| Company Data | Data Acquisition |
| Financial Data | Data Acquisition |
| Macro Data | Data Acquisition |
| News Data | Data Acquisition |
| Corporate Actions | Data Acquisition |
| Data Quality | Data Acquisition |

---

# 15. Related Documents

- README.md
- specification.md
- information-model.md
- business-rules.md
