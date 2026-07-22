# Execution Layer Architecture

**Status:** Draft

**Version:** 1.0

---

# Purpose

Tài liệu này mô tả kiến trúc tổng thể của Execution Layer.

Execution Layer chịu trách nhiệm chuyển đổi Business Process Definitions thành Runtime Execution thông qua một tập hợp các thành phần độc lập và có trách nhiệm rõ ràng.

---

# Architectural Goals

Execution Layer được thiết kế với các mục tiêu sau.

- Tách biệt Business Logic khỏi Technical Execution.
- Chuẩn hóa Runtime Execution.
- Đảm bảo khả năng mở rộng.
- Đảm bảo khả năng bảo trì.
- Tuân thủ Single Responsibility Principle.
- Hỗ trợ bổ sung Execution Strategy trong tương lai.

---

# Architecture Overview

Execution Layer bao gồm bốn nhóm thành phần.

```
                   Execution Layer

        +--------------------------------------+
        |          Definition Models           |
        +--------------------------------------+
        | Workflow → Stage → Task → Activity   |
        +-------------------+------------------+
                            |
                            ▼
        +--------------------------------------+
        |           Runtime Models             |
        +--------------------------------------+
        | Execution → Workflow → Stage         |
        |            → Task → Activity         |
        +-------------------+------------------+
                            |
                            ▼
        +--------------------------------------+
        |          Execution Engine            |
        +--------------------------------------+
        | Orchestrator                         |
        | State Machine                        |
        | Activity Executor                    |
        | Execution Policy                     |
        +-------------------+------------------+
                            |
                            ▼
        +--------------------------------------+
        |        Supporting Components         |
        +--------------------------------------+
        | Context Model                        |
        | Business Rules                       |
        +--------------------------------------+
```

---

# Definition Models

Definition Models mô tả Business Process.

Definition chỉ chứa thông tin mô tả và không được thực thi trực tiếp.

```
Workflow

↓

Stage

↓

Task

↓

Activity
```

Definition Models trả lời câu hỏi:

> Hệ thống cần thực hiện điều gì?

---

# Runtime Models

Runtime Models đại diện cho một lần thực thi của Definition Models.

```
Execution Runtime

↓

Workflow Instance

↓

Stage Instance

↓

Task Instance

↓

Activity Instance
```

Runtime Models trả lời câu hỏi:

> Phiên thực thi hiện tại đang ở trạng thái nào?

---

# Execution Engine

Execution Engine chịu trách nhiệm điều khiển Runtime.

Execution Engine bao gồm bốn thành phần.

---

## Orchestrator

Orchestrator quyết định Runtime Instance nào sẽ được thực thi tiếp theo.

Không quản lý:

- Context
- State
- Business Logic

---

## State Machine

State Machine quản lý Runtime State.

State Machine là thành phần duy nhất được phép thay đổi Runtime State.

---

## Activity Executor

Activity Executor thực thi Activity Instance.

Executor không chứa Business Logic.

Executor chỉ gọi Activity Handler tương ứng.

---

## Execution Policy

Execution Policy định nghĩa cách Runtime được thực thi.

Ví dụ.

- Sequential
- Retry
- Timeout
- Concurrency

Execution Policy không quyết định Business Process.

---

# Supporting Components

---

## Context Model

Context Model quản lý Runtime Data.

```
Execution Context

↓

Workflow Context

↓

Stage Context

↓

Task Context

↓

Activity Context
```

Context được truyền theo Runtime Hierarchy.

---

## Business Rules

Business Rules định nghĩa các điều kiện nghiệp vụ.

Business Rules trả lời câu hỏi:

> Runtime có được phép tiếp tục hay không?

Business Rules hoàn toàn độc lập với Execution Engine.

---

# Component Interaction

Quan hệ giữa các thành phần.

```
Workflow Definition

↓

Execution Runtime

↓

Orchestrator

↓

State Machine

↓

Activity Executor

↓

Activity Handler

↓

Execution Result
```

Trong quá trình thực thi.

- Runtime cung cấp Runtime Objects.
- Orchestrator quyết định bước tiếp theo.
- State Machine chuyển Runtime State.
- Activity Executor thực thi Activity.
- Context Model cung cấp dữ liệu.
- Business Rules xác thực nghiệp vụ.

---

# Separation of Responsibilities

| Component | Responsibility |
|-----------|----------------|
| Definition Models | Define Business Process |
| Runtime Models | Represent Runtime Execution |
| Orchestrator | Decide Execution Flow |
| State Machine | Manage Runtime State |
| Activity Executor | Execute Activities |
| Execution Policy | Define Technical Execution Strategy |
| Context Model | Manage Runtime Data |
| Business Rules | Validate Business Constraints |

---

# Dependency Direction

Các thành phần phụ thuộc theo hướng sau.

```
Definitions

↓

Runtime

↓

Execution Engine

↓

Activity Handler
```

Supporting Components được sử dụng bởi nhiều thành phần nhưng không tạo phụ thuộc vòng.

```
Context Model

▲

Execution Engine

▼

Business Rules
```

---

# Design Principles

## Separation of Concerns

Business Logic và Technical Execution được tách biệt hoàn toàn.

---

## Single Responsibility

Mỗi thành phần chỉ có một trách nhiệm.

---

## Definition Driven

Business Process được xác định bởi Definition Models.

---

## Runtime Oriented

Execution Engine chỉ làm việc với Runtime.

---

## Stateless Orchestrator

Orchestrator không lưu Runtime Data.

---

## Centralized State Management

State chỉ được quản lý bởi State Machine.

---

## Hierarchical Context

Context được tổ chức theo Runtime Hierarchy.

---

## Policy Driven Execution

Execution Policy quyết định hành vi kỹ thuật của Runtime.

---

# Future Evolution

Kiến trúc hiện tại được thiết kế để hỗ trợ mở rộng trong tương lai.

Ví dụ.

Execution Engine có thể bổ sung.

- Parallel Execution
- Retry Engine
- Timeout Engine
- Compensation Engine
- Distributed Execution

Activity Executor có thể mở rộng.

- Activity Registry
- Plugin System
- Middleware Pipeline

Execution Policy có thể mở rộng.

- Adaptive Retry
- Dynamic Scheduling
- Rate Limiting
- Circuit Breaker

Những mở rộng này không làm thay đổi kiến trúc cốt lõi của Execution Layer.

---

# Related Documents

- README.md
- specification.md
- workflow-model.md
- stage-model.md
- task-model.md
- activity-model.md
- execution-model.md
- context-model.md
- state-machine.md
- orchestration-model.md
- activity-executor.md
- execution-policy.md
- business-rules.md
