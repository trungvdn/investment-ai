# Architecture Decision Records (ADR)

Version: 1.0

---

# 1. Purpose

Architecture Decision Records (ADR) lưu lại các quyết định kiến trúc quan trọng của dự án Investment AI.

Mỗi ADR trả lời bốn câu hỏi:

1. Vấn đề là gì?
2. Những phương án đã được cân nhắc?
3. Vì sao chọn phương án hiện tại?
4. Hệ quả của quyết định này là gì?

ADR giúp:

* Hiểu lịch sử thiết kế.
* Tránh lặp lại các cuộc thảo luận cũ.
* Giải thích kiến trúc cho thành viên mới.
* Đánh giá tác động khi thay đổi kiến trúc.

---

# ADR-001

## Title

Business Capability First

## Status

Accepted

## Context

Có hai hướng thiết kế:

* Thiết kế theo Agent.
* Thiết kế theo Business Capability.

## Decision

Toàn bộ hệ thống được thiết kế theo Business Capability.

Agent chỉ là implementation.

## Consequences

Ưu điểm

* Không phụ thuộc AI.
* Dễ thay đổi công nghệ.
* Capability có thể tái sử dụng.

Nhược điểm

* Cần đầu tư nhiều thời gian vào phân tích Domain.

---

# ADR-002

## Title

Workflow-Centric Architecture

## Status

Accepted

## Context

Có hai lựa chọn:

Agent điều phối lẫn nhau.

Hoặc

Workflow điều phối tất cả capability.

## Decision

Workflow là trung tâm của hệ thống.

Capability không gọi trực tiếp capability khác.

## Consequences

Ưu điểm

* Dễ quan sát.
* Dễ kiểm thử.
* Dễ retry.
* Dễ mở rộng.

---

# ADR-003

## Title

Structured Output Only

## Status

Accepted

## Context

LLM có thể sinh:

* Natural Language
* Structured Output

## Decision

Các capability chỉ trao đổi bằng Domain Model hoặc JSON có cấu trúc.

Natural Language chỉ được sinh ở lớp Reporting.

## Consequences

* Dễ kiểm thử.
* Dễ tích hợp.
* Giảm lỗi khi thay đổi prompt.

---

# ADR-004

## Title

Evidence Before Recommendation

## Status

Accepted

## Context

Một số hệ thống AI sinh khuyến nghị trực tiếp từ LLM.

Điều này khó kiểm chứng và khó giải thích.

## Decision

Recommendation chỉ được tạo sau khi:

Raw Data

↓

Evidence

↓

Analysis

↓

Investment Thesis

↓

Recommendation

## Consequences

Ưu điểm

* Minh bạch.
* Có thể truy vết.
* Phù hợp với Explainable AI.

Nhược điểm

* Workflow dài hơn.

---

# ADR-005

## Title

Stateless Capability

## Status

Accepted

## Context

Có thể lưu trạng thái bên trong capability hoặc để workflow quản lý.

## Decision

Capability không lưu trạng thái.

Workflow chịu trách nhiệm quản lý trạng thái thực thi.

## Consequences

* Dễ scale.
* Dễ retry.
* Dễ kiểm thử.
* Có thể chạy song song.

---

# ADR-006

## Title

Data Is The Source of Truth

## Status

Accepted

## Context

LLM có thể bổ sung hoặc suy diễn thông tin khi dữ liệu không đầy đủ.

## Decision

LLM không được tạo dữ liệu mới.

Nếu thiếu dữ liệu phải trả về:

* Unknown
* Insufficient Evidence

## Consequences

Ưu điểm

* Giảm hallucination.
* Tăng độ tin cậy.

Nhược điểm

* Báo cáo có thể thiếu thông tin khi dữ liệu không đủ.

---

# ADR-007

## Title

Incremental Evolution

## Status

Accepted

## Context

Có thể xây dựng đầy đủ Multi-Agent ngay từ đầu hoặc phát triển từng giai đoạn.

## Decision

Ưu tiên hoàn thành MVP trước.

Sau đó mới bổ sung:

* Memory
* Reflection
* RAG
* Multi-Agent
* Portfolio
* Monitoring

## Consequences

Ưu điểm

* Giảm độ phức tạp.
* Có giá trị sớm.
* Kiểm chứng thiết kế bằng sản phẩm thực tế.

---

# ADR Template

Mọi ADR mới phải tuân theo mẫu sau:

```markdown
# ADR-XXX

Title

Status

Context

Options

Decision

Consequences

Related Documents
```

---

# Relationship With Other Documents

Vision

↓

Architecture Principles

↓

Quality Attributes

↓

Architecture Decision Records

↓

Business Capabilities

↓

Workflow

↓

Implementation

ADR chỉ được tạo khi quyết định có ảnh hưởng lâu dài đến kiến trúc hoặc quy trình phát triển.
