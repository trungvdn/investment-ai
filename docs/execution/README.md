# Execution Layer

> Transform Business Process Definitions into Runtime Execution.

---

# Overview

Execution Layer là thành phần chịu trách nhiệm thực thi Business Process được định nghĩa trong Definition Layer.

Layer này chuyển đổi các Definition Models thành Runtime Objects, sau đó điều phối toàn bộ quá trình thực thi thông qua một tập hợp các thành phần độc lập.

Execution Layer không chứa Business Logic.

Business Logic được định nghĩa bởi Definition Models và Business Rules.

---

# Goals

Execution Layer được thiết kế nhằm:

- Chuyển Business Process thành Runtime Execution.
- Chuẩn hóa Runtime Lifecycle.
- Chuẩn hóa Runtime State.
- Tách biệt Decision và Execution.
- Tách biệt Business Logic và Technical Execution.
- Hỗ trợ mở rộng Workflow Engine.

---

# High-Level Architecture

Execution Layer bao gồm năm nhóm thành phần.

```
Definition Models
        │
        ▼
Runtime Models
        │
        ▼
Orchestration Engine
        │
        ▼
Execution Commands
        │
        ▼
Execution Engine
        │
        ▼
Supporting Components
```

Mỗi nhóm thành phần chỉ chịu trách nhiệm cho một vai trò riêng biệt.

---

# Architecture Summary

## Definition Models

Mô tả Business Process.

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

Đại diện cho một lần thực thi của Definition Models.

```
Execution
    ↓
Workflow
    ↓
Stage
    ↓
Task
    ↓
Activity
```

---

## Orchestration Engine

Đánh giá Runtime Hierarchy và quyết định hành động tiếp theo.

Orchestrator không trực tiếp thay đổi Runtime.

---

## Execution Commands

Biểu diễn quyết định của Orchestrator dưới dạng các Command.

Execution Commands đóng vai trò là contract giữa Decision và Execution.

---

## Execution Engine

Thực thi các Execution Commands.

Bao gồm:

- State Machine
- Activity Executor
- Execution Policy

---

## Supporting Components

Cung cấp dữ liệu và quy tắc cho quá trình thực thi.

Bao gồm:

- Context Model
- Business Rules

---

# Execution Flow

Execution Layer hoạt động theo mô hình Event Loop.

```
Execution Started
        │
        ▼
Evaluate Runtime
        │
        ▼
Generate Command
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
               Evaluate Runtime
```

Execution chỉ kết thúc khi không còn Runtime nào có thể thực thi.

---

# Design Principles

Execution Layer được xây dựng dựa trên các nguyên tắc sau.

- Separation of Concerns
- Single Responsibility
- Runtime Oriented
- Command Driven Execution
- Decision Before Execution
- Hierarchical Runtime
- Continuous Evaluation
- Extensibility

---

# Documentation

## Architecture

Mô tả kiến trúc tổng thể.

- architecture.md

---

## Specifications

Mô tả mục tiêu, phạm vi và yêu cầu của Execution Layer.

- specification.md

---

## Walkthrough

Mô tả chi tiết một Execution từ đầu đến cuối.

- walkthrough.md

---

## Definition Models

- workflow-model.md
- stage-model.md
- task-model.md
- activity-model.md

---

## Runtime Models

- execution-model.md
- context-model.md
- state-machine.md

---

## Execution Engine

- orchestration-model.md
- execution-command.md
- activity-executor.md
- execution-policy.md

---

## Business

- business-rules.md

---

# Directory Structure

```
execution/

README.md
architecture.md
specification.md
walkthrough.md

workflow-model.md
stage-model.md
task-model.md
activity-model.md

execution-model.md
context-model.md
state-machine.md

orchestration-model.md
execution-command.md
activity-executor.md
execution-policy.md

business-rules.md
```

---

# Recommended Reading Order

1. README.md
2. architecture.md
3. specification.md
4. walkthrough.md
5. workflow-model.md
6. stage-model.md
7. task-model.md
8. activity-model.md
9. execution-model.md
10. context-model.md
11. state-machine.md
12. orchestration-model.md
13. execution-command.md
14. activity-executor.md
15. execution-policy.md
16. business-rules.md
