# RFC-003 — Market Intelligence Knowledge Contracts

Status: Draft

Version: 1.0

Owner: Investment AI Team

Parent Capability: Market Intelligence

---

# 1. Purpose

## Objective

Knowledge Contract định nghĩa cách tri thức được trao đổi giữa các Capability.

Contract đảm bảo:

- Khả năng tương thích.
- Tính ổn định.
- Khả năng mở rộng.
- Khả năng version.
- Khả năng tái sử dụng.

Knowledge Contract không định nghĩa:

- API
- REST
- GraphQL
- Database
- Kafka Message
- Go Struct

Một Capability chỉ giao tiếp với Capability khác thông qua Knowledge Contract.

---

# 2. Design Principles

## Stable Interface

Capability có thể thay đổi implementation.

Knowledge Contract phải ổn định.

---

## Immutable

Một Contract sau khi publish không bị sửa.

Nếu nội dung thay đổi sẽ tạo Contract mới.

---

## Versioned

Mọi Contract đều có Version.

Ví dụ

MarketContext v1

MarketContext v2

---

## Self-descriptive

Contract phải đủ thông tin để Capability khác hiểu mà không cần đọc implementation.

---

## Traceable

Mọi tri thức đều phải truy vết được.

Summary

↓

Evidence

↓

Indicator

---

## Technology Independent

Không phụ thuộc:

- Go
- Python
- JSON
- YAML
- Proto

---

# 3. Contract Hierarchy

Market Intelligence sinh ra các Contract sau.

Indicator Contract

↓

Evidence Contract

↓

Signal Contract

↓

Analysis Result Contract

↓

Market Context Contract

↓

Market Analysis Result Contract
