# Orchestration Model

**Status:** Draft

**Version:** 2.0

---

# Purpose

Orchestration Model định nghĩa cách Execution Layer điều phối quá trình thực thi của các Runtime Instance.

Orchestrator chịu trách nhiệm xác định Runtime Instance nào sẽ được thực thi tiếp theo dựa trên Definition Models, Runtime State và Execution Context.

Orchestrator không chứa Business Logic và không quản lý Runtime Data.

---

# Objectives

Orchestration Model hướng đến các mục tiêu sau.

- Điều phối Runtime Execution.
- Điều hướng luồng thực thi.
- Kích hoạt Runtime Instance.
- Điều phối Activity Execution.
- Đảm bảo Business Process được thực thi đúng thứ tự.

---

# Responsibilities

Orchestrator chịu trách nhiệm:

- Khởi động Execution.
- Kích hoạt Workflow Instance.
- Kích hoạt Stage Instance.
- Kích hoạt Task Instance.
- Kích hoạt Activity Instance.
- Điều hướng luồng thực thi.
- Quyết định Runtime tiếp theo cần được thực thi.

Orchestrator không chịu trách nhiệm:

- Business Logic.
- Runtime State.
- Runtime Context.
- Runtime Result.
- Activity Implementation.
- State Transition.
- Context Storage.

---

# Orchestration Scope

Orchestrator chỉ làm việc với Runtime Models.

```
Execution

↓

Workflow Instance

↓

Stage Instance

↓

Task Instance

↓

Activity Instance
```

Definition Models chỉ được sử dụng làm Blueprint.

---

# Orchestration Flow

Execution được điều phối theo trình tự sau.

```
Receive Request

↓

Create Execution

↓

Activate Workflow

↓

Activate Stage

↓

Activate Task

↓

Activate Activity

↓

Activity Finished

↓

Activate Next Activity

↓

Task Completed

↓

Activate Next Task

↓

Stage Completed

↓

Activate Next Stage

↓

Workflow Completed

↓

Execution Completed
```

Orchestrator luôn điều phối từ Runtime cha xuống Runtime con.

---

# Execution Strategy

Phiên bản hiện tại sử dụng Sequential Execution.

```
Workflow

↓

Stage 1

↓

Stage 2

↓

Stage 3
```

Task trong Stage:

```
Task A

↓

Task B

↓

Task C
```

Activity trong Task:

```
Activity A

↓

Activity B

↓

Activity C
```

Các chiến lược khác sẽ được mở rộng trong các phiên bản tiếp theo.

---

# Decision Sources

Để điều phối Runtime, Orchestrator sử dụng ba nguồn thông tin.

## Workflow Definition

Xác định cấu trúc Business Process.

---

## Runtime State

Kiểm tra Runtime đã sẵn sàng để thực thi hay chưa.

Runtime State được quản lý bởi State Machine.

---

## Runtime Context

Cung cấp dữ liệu cần thiết cho Runtime.

Runtime Context được quản lý bởi Context Model.

---

# Runtime Activation

Orchestrator chỉ kích hoạt Runtime.

Ví dụ:

```
Task Ready

↓

Activate Task

↓

Task Running
```

Việc thay đổi trạng thái được thực hiện bởi State Machine.

---

# Runtime Completion

Sau khi Runtime hoàn thành:

```
Activity

↓

Task

↓

Stage

↓

Workflow

↓

Execution
```

Orchestrator xác định Runtime tiếp theo cần được kích hoạt.

Việc tổng hợp dữ liệu được thực hiện thông qua Context Model.

---

# Runtime Relationships

Orchestrator làm việc với các thành phần sau.

## Execution Model

Quản lý Runtime Instance.

---

## Workflow Model

Cung cấp Workflow Definition.

---

## Stage Model

Cung cấp Stage Definition.

---

## Task Model

Cung cấp Task Definition.

---

## Activity Model

Cung cấp Activity Definition.

---

## State Machine

Cung cấp Runtime State.

---

## Context Model

Cung cấp Runtime Context.

---

# Boundaries

Orchestrator chịu trách nhiệm:

- Điều hướng Runtime.
- Kích hoạt Runtime.
- Xác định Runtime tiếp theo.

Orchestrator không chịu trách nhiệm:

- Quản lý State.
- Quản lý Context.
- Thực hiện Activity.
- Lưu trữ dữ liệu.
- Business Logic.

---

# Design Principles

## Lightweight

Orchestrator chỉ điều phối.

---

## Stateless

Orchestrator không lưu Runtime Data.

---

## Definition Driven

Luồng điều phối được xác định bởi Definition Models.

---

## State Aware

Orchestrator sử dụng Runtime State để đưa ra quyết định.

---

## Context Aware

Orchestrator sử dụng Runtime Context để kích hoạt Runtime.

---

## Extensible

Có thể bổ sung các chiến lược điều phối mới mà không thay đổi Definition Models.

---

# Example

```
Execution

↓

Workflow Instance

↓

Stage Instance

↓

Task Instance

↓

Activity Instance

↓

Activity Completed

↓

Next Activity

↓

Task Completed

↓

Next Task

↓

Stage Completed

↓

Next Stage

↓

Workflow Completed

↓

Execution Completed
```

Trong ví dụ trên:

- Orchestrator chỉ quyết định Runtime nào sẽ được kích hoạt tiếp theo.
- State Machine quản lý trạng thái của từng Runtime Instance.
- Context Model cung cấp dữ liệu cho Runtime.
- Activity thực hiện hành động kỹ thuật.
- Definition Models xác định cấu trúc Business Process.

---

# Future Enhancements

Các khả năng sau không thuộc phạm vi phiên bản hiện tại:

- Parallel Execution
- Conditional Branching
- Retry Strategy
- Timeout Strategy
- Compensation
- Distributed Orchestration
- Event-driven Orchestration

Những khả năng này có thể được bổ sung mà không thay đổi Definition Models.

---

# Related Documents

- README.md
- specification.md
- execution-model.md
- workflow-model.md
- stage-model.md
- task-model.md
- activity-model.md
- context-model.md
- state-machine.md
- business-rules.md
