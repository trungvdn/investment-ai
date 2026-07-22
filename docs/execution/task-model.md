# Task Model

**Status:** Draft

**Version:** 2.0

---

# Purpose

Task Model định nghĩa cấu trúc và hành vi của một Task Definition.

Task Definition là đơn vị công việc nhỏ nhất trong Stage Definition.

Task mô tả **một mục tiêu công việc** cần đạt được.

Task không mô tả chi tiết cách thực hiện.

Việc thực hiện được định nghĩa bởi các Activity.

Task Definition không đại diện cho quá trình thực thi.

---

# Objectives

Task Model hướng đến các mục tiêu sau.

- Chia nhỏ Stage thành các đơn vị công việc.
- Định nghĩa mục tiêu của từng công việc.
- Tách biệt mục tiêu và cách thực hiện.
- Hỗ trợ tái sử dụng Task.
- Hỗ trợ mở rộng Execution Engine.

---

# Task Definition

Task Definition mô tả:

- Objective
- Input
- Output
- Activities
- Completion Criteria

Task Definition không lưu:

- Runtime State
- Runtime Context
- Runtime Result

---

# Task Structure

```
Task Definition

├── Metadata
├── Objective
├── Input Definition
├── Output Definition
├── Activity Definitions
├── Completion Criteria
└── Constraints
```

---

# Task Components

## Metadata

- Task ID
- Name
- Description
- Version

---

## Objective

Mục tiêu của Task.

Ví dụ:

- Acquire Market Data
- Analyze Trend
- Calculate Valuation

---

## Input Definition

Dữ liệu đầu vào.

---

## Output Definition

Dữ liệu đầu ra.

---

## Activity Definitions

Task bao gồm một hoặc nhiều Activity Definition.

Activity mô tả cách thực hiện Task.

---

## Completion Criteria

Điều kiện hoàn thành Task.

---

## Constraints

Các ràng buộc của Task.

---

# Task Hierarchy

```
Workflow Definition
        │
        ▼
Stage Definition
        │
        ▼
Task Definition
        │
        ▼
Activity Definition
```

---

# Task Instance

Task Definition không được thực thi trực tiếp.

Execution Layer tạo Task Instance từ Task Definition.

Task Instance được quản lý trong Execution Model.

---

# Task Relationships

Task Definition thuộc về một Stage Definition.

Task sử dụng Activity để hoàn thành mục tiêu.

---

# Design Principles

## Single Responsibility

Một Task chỉ có một mục tiêu.

---

## Goal Oriented

Task mô tả "làm gì".

---

## Activity Driven

Task được thực hiện thông qua Activity.

---

## Reusable

Task có thể tái sử dụng.

---

## Independent

Task không phụ thuộc Runtime.

---

# Example

Analyze Trend Task

Activities

- Read Market Data
- Call Market Intelligence Capability
- Merge Analysis Result
- Publish Output
