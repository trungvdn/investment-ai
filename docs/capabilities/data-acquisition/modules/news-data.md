# News Data Acquisition

**Status:** Draft  
**Version:** 2.0

---

# 1. Purpose

## Objective

News Data Acquisition chịu trách nhiệm thu thập, kiểm tra, chuẩn hóa và phân phối dữ liệu tin tức liên quan đến thị trường tài chính.

Module đảm bảo mọi dữ liệu tin tức được cung cấp cho các Business Capability đều đầy đủ, nhất quán và có thể truy vết.

Module không thực hiện phân tích cảm xúc (Sentiment Analysis), đánh giá tác động hay đưa ra quyết định đầu tư.

---

# 2. Business Questions

Module trả lời các câu hỏi sau:

- Có tin tức mới cần thu thập không?
- Tin tức có hợp lệ không?
- Tin tức đã được chuẩn hóa chưa?
- Tin tức có thể liên kết với doanh nghiệp, ngành hoặc thị trường không?
- Tin tức có đạt chất lượng để publish không?

---

# 3. Scope

## In Scope

Thu thập và xử lý:

- Company News
- Market News
- Macro News
- Industry News
- Regulatory News
- Corporate Announcement
- Exchange Announcement

---

## Out of Scope

Không bao gồm:

- Sentiment Analysis
- Event Impact Analysis
- News Summarization
- Fake News Detection
- Investment Recommendation

---

# 4. Data Sources

Module có thể thu thập dữ liệu từ nhiều nguồn.

Ví dụ:

- Financial News Provider
- Stock Exchange
- Company Announcement
- Regulatory Agency
- RSS Feed
- Internal Data Feed

Việc lựa chọn Provider không thuộc trách nhiệm của module.

---

# 5. Acquired Data

Module thu thập các nhóm dữ liệu sau.

## News Content

Ví dụ:

- Title
- Summary
- Content
- Publication Time
- Author

---

## Company News

Ví dụ:

- Earnings Release
- Dividend Announcement
- M&A
- Share Buyback
- Management Change

---

## Market News

Ví dụ:

- Market Overview
- Trading Activity
- Index Movement

---

## Macro News

Ví dụ:

- Interest Rate Decision
- Inflation Report
- GDP Release
- Economic Policy

---

## Regulatory News

Ví dụ:

- New Regulation
- Listing Rule
- Disclosure Requirement

---

# 6. Processing Pipeline

```
Collect News
        │
        ▼
Validate
        │
        ▼
Normalize
        │
        ▼
Classify
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
- Duplicate News
- Invalid Content
- Missing Fields
- Publication Time
- Source Availability

Nếu Validation thất bại:

- Không publish dữ liệu.
- Ghi nhận Validation Error.

---

# 8. Normalization

Module chuẩn hóa:

- Company Symbol
- Company Name
- Industry
- Market
- Country
- Language
- Timezone
- Date Format

Sau bước này, mọi Consumer đều nhận cùng một định dạng dữ liệu tin tức.

---

# 9. Classification

Module phân loại tin tức theo các nhóm.

Ví dụ:

- Company
- Industry
- Market
- Macro
- Regulation
- Corporate Action

Một tin tức có thể thuộc nhiều Category.

---

# 10. Metadata Enrichment

Module bổ sung Metadata.

Bao gồm:

- Source
- Provider
- Version
- Timestamp
- Category
- Company
- Industry
- Market
- Language
- Acquisition Time

---

# 11. Quality Assessment

Module đánh giá chất lượng dữ liệu dựa trên:

- Completeness
- Freshness
- Consistency
- Traceability
- Duplicate Rate

Chỉ dữ liệu đạt yêu cầu mới được publish.

---

# 12. Published Knowledge

Module publish:

- News Data
- News Metadata
- News Category
- Data Quality Information

Các Business Capability chỉ sử dụng Published Knowledge.

---

# 13. Consumers

Published News Data được sử dụng bởi:

- Company Research
- Market Intelligence
- Opportunity Discovery
- Monitoring
- Risk Management

---

# 14. Error Handling

Nếu xảy ra lỗi:

- Validation Failure
- Duplicate News
- Missing Content
- Provider Failure
- Timeout
- Invalid Format

Module không publish dữ liệu chưa đạt yêu cầu.

---

# 15. Quality Metrics

Theo dõi các chỉ số:

- News Coverage
- Data Availability
- Freshness
- Validation Success Rate
- Duplicate Rate
- Publish Success Rate

---

# 16. Edge Cases

Ví dụ:

- Một tin xuất hiện từ nhiều Provider.
- Tin được cập nhật sau khi đã publish.
- Tin bị gỡ khỏi nguồn.
- Tin liên quan nhiều doanh nghiệp.
- Tin đa ngôn ngữ.
- Tin đính chính.
- Tin trùng nội dung nhưng khác tiêu đề.

---

# 17. Dependencies

## Consumes

- News Provider

## Publishes

- News Data Contract

---

# 18. Related Documents

- specification.md
- information-model.md
- knowledge-contracts.md
- business-rules.md
