# Task Model

**Status:** Draft

**Version:** 2.0

---

# Purpose

Task Model định nghĩa cấu trúc và hành vi của một Task Definition.

Task Definition đại diện cho một đơn vị công việc nghiệp vụ (Business Work Unit) trong một Stage.

Task xác định mục tiêu cần đạt được, dữ liệu đầu vào, dữ liệu đầu ra và các Activity cần thiết để hoàn thành công việc.

Task Definition là một phần của Stage Definition.

Task Definition không đại diện cho quá trình thực thi.

---

# Objectives

Task Model hướng đến các mục tiêu sau.

- Chia nhỏ Stage thành các đơn vị công việc.
- Mỗi Task chỉ phục vụ một mục tiêu nghiệp vụ.
- Tách biệt mục tiêu nghiệp vụ và cách thực hiện.
- Hỗ trợ tái sử dụng Task.
- Cho phép mở rộng cách thực hiện mà không thay đổi Task.

---

# Task Definition

Task Definition mô tả:

- Business Objective
- Input Definition
- Output Definition
- Activity Definitions
- Completion Criteria
- Constraints

Task Definition không lưu:

- Runtime State
- Runtime Context
- Runtime Result

---

# Task Structure

```
Task Definition

├── Metadata
├── Business Objective
├── Input Definition
├── Output Definition
├── Activity Definitions
├── Completion Criteria
└── Constraints
```

---

# Task Components

## Metadata

Thông tin mô tả Task.

Ví dụ:

- Task ID
- Name
- Description
- Version

---

## Business Objective

Mục tiêu nghiệp vụ của Task.

Ví dụ:

- Acquire Market Data
- Analyze Market Trend
- Evaluate Financial Health
- Estimate Intrinsic Value

Business Objective mô tả **điều cần đạt được**, không mô tả cách thực hiện.

---

## Input Definition

Định nghĩa dữ liệu đầu vào cần thiết.

Ví dụ:

- Stock Symbol
- Market Data
- Financial Statements

---

## Output Definition

Định nghĩa dữ liệu đầu ra.

Ví dụ:

- Market Analysis
- Financial Analysis
- Valuation Result

---

## Activity Definitions

Task bao gồm một hoặc nhiều Activity Definition.

Activity mô tả cách thực hiện Task.

Ví dụ:

Task

Acquire Market Data

Activities

- Read Cache
- Call Market Data API
- Normalize Data
- Validate Data
- Save Canonical Data

---

## Completion Criteria

Điều kiện để Task được xem là hoàn thành.

Ví dụ:

- Required Output Produced
- All Activities Completed

---

## Constraints

Các ràng buộc của Task.

Ví dụ:

- Required Inputs
- Required Permissions
- Business Rules

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

Task là cầu nối giữa Business Process và Technical Execution.

---

# Task Instance

Task Definition không được thực thi trực tiếp.

Khi Workflow được thực thi:

```
Task Definition

↓

Execution Layer

↓

Task Instance
```

Task Instance được quản lý trong Execution Model.

---

# Task Relationships

## Stage Model

Task thuộc về một Stage Definition.

---

## Activity Model

Task được thực hiện thông qua một hoặc nhiều Activity Definition.

---

## Execution Model

Execution Layer tạo và quản lý Task Instance.

---

# Task Boundaries

Task chịu trách nhiệm:

- Xác định một mục tiêu nghiệp vụ.
- Định nghĩa Input và Output.
- Liệt kê các Activity cần thực hiện.
- Xác định điều kiện hoàn thành.

Task không chịu trách nhiệm:

- Runtime Scheduling.
- Runtime State.
- Runtime Context.
- Retry.
- Timeout.
- Logging.
- Monitoring.

Các trách nhiệm này thuộc Execution Layer.

---

# Design Principles

## Business Oriented

Task phản ánh một đơn vị công việc nghiệp vụ.

---

## Single Responsibility

Một Task chỉ phục vụ một mục tiêu nghiệp vụ.

---

## Activity Driven

Task được thực hiện thông qua các Activity.

---

## Independent

Task không phụ thuộc vào Runtime.

---

## Reusable

Task có thể được sử dụng trong nhiều Workflow khác nhau.

---

## Extensible

Có thể thay đổi hoặc bổ sung Activity mà không làm thay đổi mục tiêu của Task.

---

# Example

```
Workflow

Analyze Stock

│
└── Stage

      Market Analysis

      │

      ├── Task
      │      Analyze Trend
      │
      │      Activities
      │      ├── Read Market Data
      │      ├── Call Market Intelligence Capability
      │      └── Generate Trend Result
      │
      ├── Task
      │      Analyze Liquidity
      │
      └── Task
             Analyze Volume
```

Trong ví dụ trên:

- Workflow mô tả quy trình phân tích cổ phiếu.
- Stage đại diện cho giai đoạn "Market Analysis".
- Mỗi Task giải quyết một mục tiêu nghiệp vụ cụ thể.
- Các Activity mô tả cách thực hiện từng Task.

---

# Related Documents

- execution-model.md
- workflow-model.md
- stage-model.md
- activity-model.md
- orchestration-model.md
- state-machine.md
