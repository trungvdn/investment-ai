# Activity Model

**Status:** Draft

**Version:** 1.0

---

# Purpose

Activity Model định nghĩa cấu trúc và hành vi của một Activity Definition.

Activity Definition là đơn vị thực thi kỹ thuật nhỏ nhất trong Execution Layer.

Activity mô tả một hành động cụ thể cần được thực hiện để hoàn thành một Task.

Activity Definition không đại diện cho quá trình thực thi.

Khi Task được thực thi, Execution Layer sẽ tạo Activity Instance từ Activity Definition.

---

# Objectives

Activity Model hướng đến các mục tiêu sau.

- Chuẩn hóa các hành động thực thi.
- Tách biệt Business Process và Technical Execution.
- Hỗ trợ nhiều loại Activity.
- Cho phép mở rộng Execution Engine.
- Tăng khả năng tái sử dụng.

---

# Activity Definition

Activity Definition mô tả:

- Mục đích của Activity.
- Loại Activity.
- Input.
- Output.
- Execution Rules.
- Completion Criteria.

Activity Definition không lưu:

- Runtime State.
- Runtime Context.
- Runtime Result.

---

# Activity Structure

```
Activity Definition

├── Metadata
├── Type
├── Objective
├── Input Definition
├── Output Definition
├── Execution Rules
├── Completion Criteria
└── Constraints
```

---

# Activity Components

## Metadata

Thông tin mô tả Activity.

Ví dụ:

- Activity ID
- Name
- Description
- Version

---

## Type

Loại Activity.

Ví dụ:

- Capability Activity
- API Activity
- Database Activity
- Cache Activity
- Event Activity
- LLM Activity
- Human Activity
- Utility Activity

---

## Objective

Mục tiêu kỹ thuật của Activity.

Ví dụ:

- Call Capability
- Read Database
- Publish Event
- Invoke LLM
- Validate Data

---

## Input Definition

Định nghĩa dữ liệu đầu vào.

---

## Output Definition

Định nghĩa dữ liệu đầu ra.

---

## Execution Rules

Quy tắc thực hiện Activity.

Ví dụ:

- Synchronous
- Asynchronous
- Idempotent
- Transaction Required

---

## Completion Criteria

Điều kiện hoàn thành Activity.

---

## Constraints

Các ràng buộc của Activity.

Ví dụ:

- Required Permissions
- Maximum Duration
- Retry Policy Reference

---

# Activity Hierarchy

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

Activity Definition là mức thấp nhất trong mô hình định nghĩa.

---

# Activity Instance

Activity Definition không được thực thi trực tiếp.

Khi Task được thực thi:

```
Activity Definition

↓

Execution Layer

↓

Activity Instance
```

Activity Instance được quản lý trong Execution Model.

---

# Activity Types

Execution Layer hỗ trợ nhiều loại Activity.

---

## Capability Activity

Thực thi một Business Capability.

Ví dụ:

- Market Intelligence
- Data Acquisition
- Valuation

---

## API Activity

Gọi dịch vụ bên ngoài.

Ví dụ:

- Market Data Provider
- News Provider

---

## Database Activity

Đọc hoặc ghi dữ liệu.

Ví dụ:

- Read Portfolio
- Save Report

---

## Cache Activity

Đọc hoặc cập nhật Cache.

---

## Event Activity

Publish hoặc Consume Event.

---

## LLM Activity

Gọi Large Language Model.

Ví dụ:

- Summarize News
- Analyze Earnings Call
- Generate Report

---

## Human Activity

Yêu cầu sự can thiệp của con người.

Ví dụ:

- Manual Approval
- Review Recommendation

---

## Utility Activity

Các tác vụ kỹ thuật hỗ trợ.

Ví dụ:

- Validate Input
- Transform Data
- Merge Results
- Format Output

---

# Activity Relationships

Activity Definition có quan hệ với:

## Task Model

Activity thuộc về một Task Definition.

---

## Execution Model

Execution Layer tạo Activity Instance.

---

## Infrastructure Layer

Một số Activity sử dụng các dịch vụ hạ tầng như:

- Database
- Queue
- Cache
- Object Storage
- Event Bus
- LLM Gateway

---

## Business Capability

Capability Activity tham chiếu đến một Business Capability.

---

# Activity Boundaries

Activity chịu trách nhiệm:

- Thực hiện một hành động kỹ thuật cụ thể.
- Tạo ra Output cho Task.
- Tuân thủ Execution Rules.

Activity không chịu trách nhiệm:

- Business Workflow.
- Business Milestone.
- Business Decision.
- Runtime Scheduling.

---

# Design Principles

## Atomic

Mỗi Activity chỉ thực hiện một hành động.

---

## Technology Independent

Activity mô tả hành động, không phụ thuộc công nghệ cụ thể.

---

## Reusable

Activity có thể được sử dụng trong nhiều Task.

---

## Extensible

Có thể bổ sung loại Activity mới mà không ảnh hưởng Task hoặc Workflow.

---

## Composable

Nhiều Activity có thể kết hợp để hoàn thành một Task.

---

# Example

```
Task

Acquire Market Data

│
├── Cache Activity
│      Read Cached Data
│
├── API Activity
│      Fetch Market Data
│
├── Utility Activity
│      Normalize Data
│
├── Utility Activity
│      Validate Data
│
└── Database Activity
       Save Canonical Data
```

Hoặc:

```
Task

Generate Investment Report

│
├── Capability Activity
│      Company Research
│
├── Capability Activity
│      Valuation
│
├── LLM Activity
│      Generate Report
│
└── Database Activity
       Save Report
```

---

# Related Documents

- execution-model.md
- workflow-model.md
- stage-model.md
- task-model.md
- orchestration-model.md
- state-machine.md
