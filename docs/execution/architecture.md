# Execution Layer Architecture

**Status:** Draft

**Version:** 2.0

---

# Purpose

Execution Layer chịu trách nhiệm chuyển đổi Business Process Definitions thành Runtime Execution.

Layer này quản lý toàn bộ vòng đời của một Execution thông qua việc:

- Quản lý Runtime Hierarchy.
- Điều phối Runtime Execution.
- Sinh Execution Commands.
- Thực thi Technical Actions.
- Duy trì Runtime State và Context.

Execution Layer không chứa Business Logic.

Business Logic được định nghĩa trong Definition Models và Business Rules.

---

# Architectural Goals

Execution Layer được thiết kế với các mục tiêu sau.

- Separation of Concerns
- Single Responsibility
- Runtime Oriented
- Command Driven Execution
- Extensibility
- Observability
- Maintainability

---

# Architecture Overview

Execution Layer bao gồm năm nhóm thành phần.

```
                     Execution Layer

 ┌────────────────────────────────────────────┐
 │            Definition Models               │
 │ Workflow → Stage → Task → Activity         │
 └──────────────────────┬─────────────────────┘
                        │
                        ▼
 ┌────────────────────────────────────────────┐
 │             Runtime Models                 │
 │ Execution → Workflow → Stage               │
 │            → Task → Activity               │
 └──────────────────────┬─────────────────────┘
                        │
                        ▼
 ┌────────────────────────────────────────────┐
 │            Orchestration Engine            │
 │             (Decision Engine)              │
 │              Orchestrator                  │
 └──────────────────────┬─────────────────────┘
                        │
                        ▼
 ┌────────────────────────────────────────────┐
 │            Execution Commands              │
 │ Transition State                           │
 │ Execute Activity                           │
 └──────────────────────┬─────────────────────┘
                        │
                        ▼
 ┌────────────────────────────────────────────┐
 │             Execution Engine               │
 │ State Machine                             │
 │ Activity Executor                         │
 │ Execution Policy                          │
 └──────────────────────┬─────────────────────┘
                        │
                        ▼
 ┌────────────────────────────────────────────┐
 │         Supporting Components              │
 │ Context Model                             │
 │ Business Rules                            │
 └────────────────────────────────────────────┘
```

---

# Layer Responsibilities

## Definition Models

Definition Models mô tả Business Process.

```
Workflow
    ↓
Stage
    ↓
Task
    ↓
Activity
```

Definition Models không được thực thi trực tiếp.

---

## Runtime Models

Runtime Models đại diện cho một lần thực thi của Definition.

```
Execution Runtime
        ↓
Workflow Runtime
        ↓
Stage Runtime
        ↓
Task Runtime
        ↓
Activity Runtime
```

Runtime Models lưu:

- Runtime State
- Runtime Context Reference
- Definition Reference
- Execution Result

---

## Orchestration Engine

Orchestration Engine chịu trách nhiệm đưa ra quyết định.

Orchestrator liên tục:

- Evaluate Runtime Hierarchy.
- Select Executable Runtime.
- Decide Next Action.
- Generate Execution Command.

Orchestrator không trực tiếp thay đổi Runtime.

---

## Execution Commands

Execution Command là biểu diễn của một quyết định.

Ví dụ:

```yaml
TransitionStateCommand
```

```yaml
ExecuteActivityCommand
```

Command không chứa Business Logic.

Command không tự thực thi.

---

## Execution Engine

Execution Engine chịu trách nhiệm thực thi các Execution Command.

Các thành phần chính.

### State Machine

Quản lý Runtime State.

### Activity Executor

Thực thi Activity Runtime.

### Execution Policy

Áp dụng các chính sách kỹ thuật.

Ví dụ:

- Retry
- Timeout
- Sequential Execution
- Parallel Execution

---

## Supporting Components

### Context Model

Quản lý Runtime Data.

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

---

### Business Rules

Business Rules xác định Runtime có được phép tiếp tục hay không.

Business Rules không điều khiển Execution.

---

# Runtime Flow

Execution được điều khiển bởi Event Loop.

```
Execution Started
        │
        ▼
Evaluate Runtime Hierarchy
        │
        ▼
Generate Execution Command
        │
        ▼
Execute Command
        │
        ▼
Runtime Updated
        │
        └──────────────┐
                       │
                       ▼
        Evaluate Runtime Hierarchy
```

Execution kết thúc khi không còn Runtime nào có thể thực thi.

---

# Component Interaction

```
Client
   │
   ▼
Execution Runtime
   │
   ▼
Orchestrator
   │
   ▼
Execution Command
   │
   ├──────────────┐
   ▼              ▼
State Machine   Activity Executor
   │              │
   └──────┬───────┘
          ▼
    Runtime Updated
          │
          ▼
     Orchestrator
```

---

# Separation of Responsibilities

| Component | Responsibility |
|-----------|----------------|
| Definition Models | Define Business Process |
| Runtime Models | Represent Runtime Execution |
| Orchestrator | Decide Next Action |
| Execution Command | Represent Execution Decision |
| State Machine | Manage Runtime State |
| Activity Executor | Execute Activity |
| Execution Policy | Apply Technical Policies |
| Context Model | Manage Runtime Data |
| Business Rules | Validate Business Constraints |

---

# Design Principles

## Separation of Concerns

Business Logic và Technical Execution được tách biệt.

---

## Single Responsibility

Mỗi thành phần chỉ có một trách nhiệm.

---

## Runtime Oriented

Execution Engine chỉ làm việc với Runtime.

---

## Command Driven

Execution được thực hiện thông qua Execution Commands.

---

## Decision Before Execution

Mọi thay đổi Runtime đều bắt đầu từ một quyết định của Orchestrator.

---

## Hierarchical Runtime

Runtime luôn được điều khiển theo cấu trúc:

```
Workflow
    ↓
Stage
    ↓
Task
    ↓
Activity
```

---

## Continuous Evaluation

Orchestrator liên tục đánh giá Runtime sau mỗi thay đổi.

---

## Extensibility

Có thể bổ sung:

- Retry Engine
- Compensation Engine
- Human Approval
- Parallel Scheduler
- Distributed Execution

mà không làm thay đổi vai trò của các thành phần hiện tại.

---

# Related Documents

- README.md
- specification.md
- orchestration-model.md
- execution-command.md
- execution-model.md
- state-machine.md
- activity-executor.md
- execution-policy.md
- context-model.md
- business-rules.md
