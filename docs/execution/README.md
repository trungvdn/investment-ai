# Execution Layer

**Status:** Draft

**Version:** 2.0

---

# Overview

Execution Layer chịu trách nhiệm thực thi các Business Process được định nghĩa trong Definition Layer.

Execution Layer chuyển các Definition Models thành các Runtime Instance, điều phối quá trình thực thi và đảm bảo Business Process được thực hiện một cách nhất quán.

Execution Layer không chứa Business Logic.

Business Logic được định nghĩa trong Definition Models và Business Rules.

---

# Goals

Execution Layer được thiết kế với các mục tiêu sau.

- Thực thi Business Process.
- Quản lý Runtime.
- Điều phối quá trình thực thi.
- Chuẩn hóa Runtime Lifecycle.
- Hỗ trợ mở rộng Execution Engine.
- Tách biệt Business Logic khỏi Technical Execution.

---

# Architecture

Execution Layer bao gồm bốn nhóm thành phần chính.

```
Execution Layer

├── Definitions
│
├── Runtime
│
├── Execution Engine
│
└── Business Rules
```

Mỗi nhóm chịu trách nhiệm cho một khía cạnh riêng của quá trình thực thi.

---

# Definitions

Definitions mô tả Business Process.

Definitions không được thực thi trực tiếp.

```
Workflow

↓

Stage

↓

Task

↓

Activity
```

Responsibilities

- Định nghĩa Business Process.
- Định nghĩa Business Milestones.
- Định nghĩa Business Work.
- Định nghĩa Technical Actions.

---

# Runtime

Runtime đại diện cho một lần thực thi của Definition Models.

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

Responsibilities

- Quản lý Runtime Objects.
- Quản lý Runtime Hierarchy.
- Liên kết State.
- Liên kết Context.
- Tổng hợp Runtime Results.

---

# Execution Engine

Execution Engine chịu trách nhiệm điều khiển quá trình thực thi Runtime.

```
Execution Engine

├── Orchestrator
├── State Machine
├── Activity Executor
└── Execution Policy
```

Responsibilities

## Orchestrator

- Quyết định Runtime tiếp theo cần được thực thi.
- Điều hướng luồng thực thi.

---

## State Machine

- Quản lý Runtime State.
- Kiểm soát State Transition.

---

## Activity Executor

- Thực thi Activity.
- Gọi Activity Handler.
- Trả Execution Result.

---

## Execution Policy

- Cung cấp các chính sách thực thi.
- Quy định cách Runtime được thực hiện.

---

# Business Rules

Business Rules định nghĩa các quy tắc nghiệp vụ.

Business Rules trả lời câu hỏi:

> Runtime có được phép tiếp tục hay không?

Business Rules hoàn toàn độc lập với Execution Engine.

---

# Context Model

Context Model cung cấp dữ liệu Runtime cho toàn bộ Execution.

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

# Execution Flow

Quá trình thực thi tổng quát.

```
Client Request

↓

Create Execution Runtime

↓

Orchestrator

↓

Select Runtime

↓

State Machine

↓

Transition State

↓

Activity Executor

↓

Activity Handler

↓

Execution Result

↓

Runtime Updated

↓

Next Runtime

↓

Execution Completed
```

---

# Responsibility Matrix

| Component | Responsibility |
|-----------|----------------|
| Workflow | Business Process Definition |
| Stage | Business Milestone Definition |
| Task | Business Work Definition |
| Activity | Technical Action Definition |
| Execution Runtime | Runtime Object Management |
| Orchestrator | Execution Flow Decision |
| State Machine | Runtime State Management |
| Activity Executor | Activity Execution |
| Execution Policy | Technical Execution Strategy |
| Context Model | Runtime Data Management |
| Business Rules | Business Validation |

---

# Design Principles

## Separation of Concerns

Business Logic và Technical Execution được tách biệt.

---

## Single Responsibility

Mỗi thành phần chỉ có một trách nhiệm duy nhất.

---

## Definition Driven

Business Process được xác định bởi Definition Models.

---

## Runtime Oriented

Execution Engine chỉ làm việc với Runtime Instance.

---

## Stateless Orchestrator

Orchestrator không lưu trữ Runtime Data.

---

## Centralized State Management

State chỉ được quản lý bởi State Machine.

---

## Context Isolation

Mỗi Runtime có Context riêng.

---

## Policy Driven Execution

Hành vi thực thi được xác định bởi Execution Policy.

---

# Directory Structure

```
execution/

├── README.md
├── specification.md
│
├── definitions/
│   ├── workflow-model.md
│   ├── stage-model.md
│   ├── task-model.md
│   └── activity-model.md
│
├── runtime/
│   ├── execution-model.md
│   ├── context-model.md
│   └── state-machine.md
│
├── engine/
│   ├── orchestration-model.md
│   ├── activity-executor.md
│   └── execution-policy.md
│
└── business-rules.md
```

---

# Documentation Guide

Khuyến nghị đọc tài liệu theo thứ tự sau.

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

Execution Layer làm việc với các tầng sau.

- Capability Layer
- Agent Layer
- Memory Layer
- Infrastructure Layer
