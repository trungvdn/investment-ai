# Execution Layer

> Transform Business Process Definitions into Runtime Execution.

---

# Overview

Execution Layer chịu trách nhiệm thực thi các Business Process được định nghĩa trong Definition Layer.

Layer này chuyển các Definition Models thành Runtime Instances và điều phối quá trình thực thi thông qua một tập hợp các thành phần độc lập.

Execution Layer không chứa Business Logic.

Business Logic được định nghĩa trong Business Rules và các Definition Models.

---

# Design Goals

Execution Layer được thiết kế với các mục tiêu sau.

- Thực thi Business Process.
- Quản lý Runtime Objects.
- Điều phối quá trình thực thi.
- Chuẩn hóa Runtime Lifecycle.
- Chuẩn hóa Runtime State.
- Tách biệt Business Logic khỏi Technical Execution.
- Hỗ trợ mở rộng Execution Engine.

---

# High-Level Architecture

Execution Layer được chia thành bốn nhóm thành phần.

```
                    Execution Layer

        +--------------------------------------+
        |          Definition Models           |
        |--------------------------------------|
        | Workflow → Stage → Task → Activity   |
        +-------------------+------------------+
                            |
                            v
        +--------------------------------------+
        |          Runtime Models              |
        |--------------------------------------|
        | Execution → Workflow → Stage         |
        |            → Task → Activity         |
        +-------------------+------------------+
                            |
                            v
        +--------------------------------------+
        |          Execution Engine            |
        |--------------------------------------|
        | Orchestrator                         |
        | State Machine                        |
        | Activity Executor                    |
        | Execution Policy                     |
        +-------------------+------------------+
                            |
                            v
        +--------------------------------------+
        |        Supporting Components         |
        |--------------------------------------|
        | Context Model                        |
        | Business Rules                       |
        +--------------------------------------+
```

---

# Core Components

## Definition Models

Definition Models mô tả Business Process.

Chúng đóng vai trò là blueprint và không được thực thi trực tiếp.

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

## Runtime Models

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

Execution Runtime là Aggregate Root của toàn bộ Runtime Hierarchy.

---

## Execution Engine

Execution Engine chịu trách nhiệm điều khiển quá trình thực thi Runtime.

### Orchestrator

Quyết định Runtime Instance nào sẽ được thực thi tiếp theo.

Không quản lý State hoặc Context.

---

### State Machine

Quản lý Runtime State và kiểm soát các State Transition hợp lệ.

State Machine là thành phần duy nhất được phép thay đổi Runtime State.

---

### Activity Executor

Thực thi Activity Instance bằng cách gọi Activity Handler phù hợp.

Không chứa Business Logic.

---

### Execution Policy

Định nghĩa các chính sách kỹ thuật của quá trình thực thi.

Ví dụ:

- Execution Strategy
- Retry Strategy
- Timeout
- Error Handling
- Concurrency

---

## Context Model

Context Model quản lý dữ liệu Runtime.

Context được tổ chức theo cấu trúc phân cấp.

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

## Business Rules

Business Rules định nghĩa các quy tắc nghiệp vụ.

Business Rules trả lời câu hỏi:

> Runtime có được phép tiếp tục hay không?

Business Rules hoàn toàn độc lập với Execution Engine.

---

# Execution Flow

Luồng thực thi tổng quát.

```
Client Request
        │
        ▼
Create Execution Runtime
        │
        ▼
Load Workflow Definition
        │
        ▼
Orchestrator
        │
        ▼
Select Runtime Instance
        │
        ▼
State Machine
        │
        ▼
Transition Runtime State
        │
        ▼
Activity Executor
        │
        ▼
Activity Handler
        │
        ▼
Execution Result
        │
        ▼
Update Runtime
        │
        ▼
Next Runtime
        │
        ▼
Execution Completed
```

---

# Responsibility Matrix

| Component | Responsibility |
|-----------|----------------|
| Workflow | Define Business Process |
| Stage | Define Business Milestone |
| Task | Define Business Work |
| Activity | Define Technical Action |
| Execution Runtime | Manage Runtime Objects |
| Orchestrator | Decide Execution Flow |
| State Machine | Manage Runtime State |
| Activity Executor | Execute Activity |
| Execution Policy | Define Execution Strategy |
| Context Model | Manage Runtime Data |
| Business Rules | Validate Business Constraints |

---

# Design Principles

## Separation of Concerns

Business Logic và Technical Execution được tách biệt hoàn toàn.

---

## Single Responsibility

Mỗi thành phần chỉ có một trách nhiệm.

---

## Definition Driven

Business Process được định nghĩa bởi Definition Models.

---

## Runtime Oriented

Execution Engine chỉ làm việc với Runtime Instances.

---

## Stateless Orchestrator

Orchestrator không lưu Runtime Data.

---

## Centralized State Management

State chỉ được quản lý bởi State Machine.

---

## Hierarchical Context

Runtime Context được kế thừa theo Runtime Hierarchy.

---

## Policy Driven Execution

Các hành vi kỹ thuật được điều khiển bởi Execution Policy.

---

# Documentation Structure

```
execution/

README.md

specification.md

workflow-model.md
stage-model.md
task-model.md
activity-model.md

execution-model.md
context-model.md
state-machine.md

orchestration-model.md
activity-executor.md
execution-policy.md

business-rules.md
```

---

# Recommended Reading Order

1. README.md
2. specification.md
3. workflow-model.md
4. stage-model.md
5. task-model.md
6. activity-model.md
7. execution-model.md
8. context-model.md
9. state-machine.md
10. orchestration-model.md
11. activity-executor.md
12. execution-policy.md
13. business-rules.md

---

# Related Layers

Execution Layer phối hợp với các tầng khác trong hệ thống.

- Capability Layer
- Agent Layer
- Memory Layer
- Infrastructure Layer
