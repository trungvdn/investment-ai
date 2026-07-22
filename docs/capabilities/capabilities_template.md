# Capability Specification Template

Version: 2.0

---

# 1. Overview

## Capability Name

Tên chính thức của Capability.

---

## Purpose

Capability này tồn tại để giải quyết vấn đề gì?

Mục tiêu nghiệp vụ.

---

## Business Goal

Business Outcome mà Capability mang lại.

Ví dụ

- Generate Market Analysis
- Assess Financial Health
- Produce Investment Recommendation

---

## Scope

### In Scope

Những gì Capability chịu trách nhiệm.

### Out of Scope

Những gì Capability không chịu trách nhiệm.

---

# 2. Responsibilities

Liệt kê toàn bộ Responsibility.

Ví dụ

- Analyze Market Trend
- Analyze Liquidity
- Evaluate Macro Environment
- Generate Market Score

---

# 3. Inputs

## Input Contracts

Knowledge Contracts đầu vào.

Ví dụ

- RawData
- MarketData
- MacroData

---

## Required Context

Context phải được Workflow chuẩn bị trước.

Ví dụ

- Market
- Exchange
- Analysis Period

---

# 4. Outputs

## Output Contract

Knowledge Contract đầu ra.

Ví dụ

MarketAnalysisResult

---

## Produced Knowledge

Capability tạo ra loại tri thức nào.

Ví dụ

Evidence

Signals

Analysis

Recommendation

---

# 5. Knowledge Transformation

Mô tả quá trình chuyển đổi tri thức.

```text
Raw Data

↓

Evidence

↓

Signals

↓

Analysis

↓

Knowledge Contract
```

Giải thích ngắn gọn từng bước.

---

# 6. Domain Model

## Entities

Các Entity được sử dụng.

Ví dụ

Market

Sector

Index

---

## Value Objects

Ví dụ

Trend

Confidence

Score

DateRange

---

## Domain Services

Ví dụ

MarketAnalysisService

---

## Domain Events

Ví dụ

MarketAnalysisCompleted

---

# 7. Business Rules

Liệt kê Rule được sử dụng.

Ví dụ

- BR-001
- BR-007
- BR-013

Không mô tả chi tiết Rule.

---

# 8. Execution Pipeline

Capability phải tuân theo Execution Model chuẩn.

```text
Validate Input

↓

Load Context

↓

Normalize Data

↓

Generate Evidence

↓

Detect Signals

↓

Analyze

↓

Generate Output

↓

Validate Output
```

---

# 9. Activities

Liệt kê Activity nội bộ.

Ví dụ

Load Market Data

Normalize

Generate Evidence

Calculate Score

Generate Summary

---

# 10. Algorithms / Reasoning

Liệt kê phương pháp được sử dụng.

Ví dụ

Rule Engine

Statistical Model

Machine Learning

LLM

Hybrid

Nếu dùng AI:

- Prompt Strategy
- Structured Output
- Confidence Estimation

---

# 11. Dependencies

## Depends On

Ví dụ

Business Rules

Knowledge Contracts

Domain Services

---

## Independent Of

Capability không được phụ thuộc vào:

Go

Database

LLM Provider

Workflow Engine

---

# 12. Quality Attributes

Ví dụ

Accuracy

Consistency

Deterministic (nếu áp dụng)

Explainability

Traceability

Latency

Availability

---

# 13. Error Model

## Business Errors

Ví dụ

Missing Financial Statement

Unsupported Exchange

---

## Technical Errors

API Timeout

Storage Failure

Model Unavailable

---

## Data Quality Errors

Missing Data

Outdated Data

Invalid Data

---

# 14. Observability

Capability phải ghi nhận tối thiểu:

Execution ID

Capability Version

Input Contract Version

Output Contract Version

Execution Time

Evidence Count

Signal Count

Confidence

Warnings

Errors

---

# 15. Security & Governance

Định nghĩa các yêu cầu phi chức năng.

Ví dụ

Read-only Access

Audit Log

Versioned Output

Rule Traceability

Explainability

---

# 16. Extension Points

Các điểm mở rộng trong tương lai.

Ví dụ

Thay Rule Engine bằng AI

Thêm Model Scoring

Thêm External Data Provider

Thêm Domain Service

---

# 17. Performance Targets

Ví dụ

Average Latency

Maximum Latency

Maximum Memory

Maximum Token Usage

Maximum External API Calls

---

# 18. Capability Lifecycle

```text
Design

↓

Implement

↓

Test

↓

Deploy

↓

Observe

↓

Improve
```

---

# 19. Example

## Input

Ví dụ Input Contract.

---

## Processing

Ví dụ quá trình chuyển đổi.

---

## Output

Ví dụ Output Contract.

---

# 20. Appendix

## Related Capabilities

Liên kết tới các Capability liên quan.

Ví dụ

Market Intelligence

Risk Assessment

Reporting

---

## Related Business Rules

Danh sách Rule ID.

---

## Related Knowledge Contracts

Danh sách Contract.

---

## References

Các tài liệu liên quan.

- knowledge-model.md
- knowledge-contracts.md
- workflow-architecture.md
- business-rules.md
