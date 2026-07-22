# RFC-002 — Market Intelligence Information Model

Status: Draft

Version: 1.0

Owner: Investment AI Team

Parent Capability: Market Intelligence

---

# 1. Purpose

## Objective

Information Model định nghĩa toàn bộ Business Objects được sử dụng trong Market Intelligence Capability.

Tài liệu này mô tả:

- Ý nghĩa nghiệp vụ của từng đối tượng.
- Quan hệ giữa các đối tượng.
- Luồng chuyển đổi tri thức.
- Ngôn ngữ chung (Ubiquitous Language).

Information Model **không định nghĩa**:

- JSON Schema
- Go Struct
- Database Schema
- API Contract
- YAML
- Implementation

---

# 2. Design Principles

Information Model tuân theo các nguyên tắc sau.

## Business First

Mọi đối tượng đều phản ánh khái niệm nghiệp vụ.

Ví dụ:

✓ Market Trend

✓ Liquidity

✓ Risk Environment

Không sử dụng tên kỹ thuật như:

× DTO

× Response

× Payload

× Entity

---

## Immutable Knowledge

Mỗi Business Object đại diện cho tri thức tại một thời điểm phân tích.

Sau khi được tạo, nội dung không bị chỉnh sửa.

Nếu dữ liệu thay đổi, hệ thống tạo một phiên bản mới.

---

## Traceability

Mọi kết luận đều phải truy vết được.

Market Summary

↓

Market Context

↓

Evidence

↓

Indicators

---

## Explainability

Không có kết luận nào tồn tại nếu không có Evidence hỗ trợ.

---

# 3. Information Hierarchy

Market Intelligence tạo ra tri thức theo mô hình sau.

Raw Market Data

↓

Indicators

↓

Evidence

↓

Signals

↓

Analysis Result

↓

Market Context

↓

Market Analysis Result

Trong đó:

Indicators là dữ liệu quan sát.

Evidence là diễn giải dữ liệu.

Signals là kết luận cục bộ.

Analysis Result là kết quả của từng Analysis Module.

Market Context là bức tranh tổng hợp.

Market Analysis Result là sản phẩm cuối cùng của Capability.

---

# 4. Core Business Objects
