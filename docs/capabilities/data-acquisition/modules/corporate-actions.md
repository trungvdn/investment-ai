# Corporate Actions Acquisition

**Status:** Draft  
**Version:** 2.0

---

# 1. Purpose

## Objective

Corporate Actions Acquisition chịu trách nhiệm thu thập, kiểm tra, chuẩn hóa và phân phối dữ liệu về các sự kiện doanh nghiệp.

Module đảm bảo mọi Corporate Action được cung cấp cho các Business Capability đều đầy đủ, chính xác, nhất quán và có thể truy vết.

Module không đánh giá tác động của Corporate Action đối với doanh nghiệp hay thị trường.

---

# 2. Business Questions

Module trả lời các câu hỏi sau:

- Có Corporate Action mới được công bố không?
- Dữ liệu Corporate Action có hợp lệ không?
- Dữ liệu đã được chuẩn hóa chưa?
- Corporate Action có đạt chất lượng để publish không?
- Corporate Action nào có thể cung cấp cho các Business Capability?

---

# 3. Scope

## In Scope

Thu thập và xử lý:

- Dividend
- Stock Split
- Reverse Split
- Bonus Share
- Rights Issue
- Share Buyback
- Share Issuance
- ESOP
- Merger & Acquisition
- Listing
- Delisting
- Trading Suspension

---

## Out of Scope

Không bao gồm:

- Corporate Action Impact Analysis
- Company Valuation
- Investment Recommendation
- Portfolio Rebalancing
- Event Prediction

---

# 4. Data Sources

Module có thể thu thập dữ liệu từ nhiều nguồn.

Ví dụ:

- Stock Exchange
- Company Disclosure
- Regulatory Filing
- Financial Information Provider
- Internal Data Feed

Việc lựa chọn Provider không thuộc trách nhiệm của module.

---

# 5. Acquired Data

Module thu thập các nhóm dữ liệu sau.

## Dividend

Ví dụ:

- Cash Dividend
- Stock Dividend
- Dividend Rate
- Ex-Date
- Record Date
- Payment Date

---

## Capital Changes

Ví dụ:

- Stock Split
- Reverse Split
- Bonus Share
- Rights Issue
- Share Issuance
- ESOP

---

## Corporate Transactions

Ví dụ:

- Share Buyback
- Mergers
- Acquisitions
- Asset Disposal
- Capital Reduction

---

## Listing Status

Ví dụ:

- Initial Listing
- Additional Listing
- Delisting
- Trading Suspension
- Trading Resumption

---

## Event Information

Ví dụ:

- Announcement Date
- Effective Date
- Approval Status
- Event Status

---

# 6. Processing Pipeline

```
Collect Data
        │
        ▼
Validate
        │
        ▼
Normalize
        │
        ▼
Enrich Metadata
        │
        ▼
Assess Quality
        │
        ▼
Publish
```

---

# 7. Validation

Module kiểm tra:

- Required Information
- Company Symbol
- Event Type
- Event Date
- Duplicate Events
- Missing Values
- Invalid Values
- Event Consistency

Nếu Validation thất bại:

- Không publish dữ liệu.
- Ghi nhận Validation Error.

---

# 8. Normalization

Module chuẩn hóa:

- Company Symbol
- Company Name
- Event Type
- Currency
- Percentage Format
- Date Format
- Timezone
- Status

Sau bước này, mọi Consumer đều nhận cùng một định dạng dữ liệu Corporate Action.

---

# 9. Metadata Enrichment

Module bổ sung Metadata cho mọi Data Object.

Bao gồm:

- Source
- Provider
- Version
- Timestamp
- Announcement Date
- Effective Date
- Acquisition Time

---

# 10. Quality Assessment

Module đánh giá chất lượng dữ liệu dựa trên:

- Completeness
- Accuracy
- Freshness
- Consistency
- Timeliness

Chỉ dữ liệu đạt yêu cầu mới được publish.

---

# 11. Published Knowledge

Module publish:

- Corporate Actions
- Event Metadata
- Data Quality Information

Các Business Capability chỉ sử dụng Published Knowledge.

---

# 12. Consumers

Published Corporate Actions được sử dụng bởi:

- Company Research
- Fundamental Analysis
- Valuation
- Portfolio Management
- Monitoring
- Opportunity Discovery

---

# 13. Error Handling

Nếu xảy ra lỗi:

- Validation Failure
- Duplicate Event
- Missing Event Information
- Provider Failure
- Timeout
- Invalid Format

Module không publish dữ liệu chưa đạt yêu cầu.

---

# 14. Quality Metrics

Theo dõi các chỉ số:

- Event Coverage
- Data Availability
- Freshness
- Validation Success Rate
- Publish Success Rate
- Error Rate

---

# 15. Edge Cases

Ví dụ:

- Doanh nghiệp điều chỉnh tỷ lệ cổ tức sau khi công bố.
- Corporate Action bị hủy hoặc hoãn.
- Một sự kiện được công bố nhiều lần với thông tin cập nhật.
- Thay đổi ngày thực hiện (Effective Date).
- Hai Provider cung cấp thông tin khác nhau.
- Doanh nghiệp đổi mã chứng khoán trước hoặc sau Corporate Action.
- Nhiều Corporate Action diễn ra đồng thời.

---

# 16. Dependencies

## Consumes

- Corporate Action Provider

## Publishes

- Corporate Actions Contract

---

# 17. Related Documents

- specification.md
- information-model.md
- knowledge-contracts.md
- business-rules.md
