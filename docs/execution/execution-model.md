# Execution Model

**Status:** Draft

**Version:** 3.0

---

# Purpose

Execution Model định nghĩa cấu trúc và vòng đời của các Runtime Models trong Execution Layer.

Execution Runtime là Aggregate Root của toàn bộ Runtime, chịu trách nhiệm quản lý các Runtime Instance và duy trì tính nhất quán của quá trình thực thi.

Execution Runtime không chứa Business Logic và không điều phối luồng thực thi.

---

# Objectives

Execution Model hướng đến các mục tiêu sau.

- Quản lý Runtime Objects.
- Quản lý vòng đời Runtime.
- Tổ chức quan hệ giữa các Runtime Instance.
- Cung cấp Runtime Container cho Execution Engine.
- Hỗ trợ mở rộng Runtime.

---

# Runtime Models

Execution Layer bao gồm các Runtime Models sau.

```
Execution
        │
        ▼
Workflow Instance
        │
        ▼
Stage Instance
        │
        ▼
Task Instance
        │
        ▼
Activity Instance
```

Mỗi Runtime Instance đại diện cho một Definition Model tương ứng.

---

# Execution Runtime

Execution Runtime là Runtime Root của một lần thực thi.

Execution Runtime được tạo khi hệ thống nhận một Execution Request.

Execution Runtime kết thúc khi toàn bộ Workflow hoàn thành hoặc Execution bị hủy.

---

# Responsibilities

Execution Runtime chịu trách nhiệm:

- Quản lý Workflow Instance.
- Quản lý Runtime Hierarchy.
- Lưu trữ Runtime Reference.
- Liên kết Context.
- Liên kết Runtime State.
- Tổng hợp Execution Result.

Execution Runtime không chịu trách nhiệm:

- Business Logic.
- Runtime Scheduling.
- Runtime Orchestration.
- Activity Execution.
- State Transition.
- Context Storage.

---

# Runtime Hierarchy

Runtime được tổ chức theo mô hình cây.

```
Execution

└── Workflow Instance
      │
      ├── Stage Instance
      │      │
      │      ├── Task Instance
      │      │      │
      │      │      ├── Activity Instance
      │      │      └── Activity Instance
      │      │
      │      └── Task Instance
      │
      └── Stage Instance
```

Execution Runtime là Root của toàn bộ cây Runtime.

---

# Runtime Components

## Execution

Runtime Root.

Chứa:

- Execution ID
- Workflow Instance
- Context Reference
- State Reference
- Result Reference

---

## Workflow Instance

Runtime của Workflow Definition.

Chứa:

- Workflow Definition Reference
- Stage Instances
- Context Reference
- State Reference
- Result Reference

---

## Stage Instance

Runtime của Stage Definition.

Chứa:

- Stage Definition Reference
- Task Instances
- Context Reference
- State Reference
- Result Reference

---

## Task Instance

Runtime của Task Definition.

Chứa:

- Task Definition Reference
- Activity Instances
- Context Reference
- State Reference
- Result Reference

---

## Activity Instance

Runtime của Activity Definition.

Chứa:

- Activity Definition Reference
- Context Reference
- State Reference
- Result Reference

---

# Runtime References

Runtime Instance không sở hữu trực tiếp:

- State
- Context
- Definition

Thay vào đó Runtime chỉ giữ Reference.

```
Task Instance

├── Task Definition Ref
├── Context Ref
├── State Ref
└── Result Ref
```

Việc quản lý các thành phần này thuộc các Model tương ứng.

---

# Runtime Relationships

Execution Runtime làm việc với các thành phần sau.

## Workflow Definition

Cung cấp Blueprint.

---

## Context Model

Cung cấp Runtime Context.

---

## State Machine

Quản lý Runtime State.

---

## Orchestrator

Quyết định Runtime nào được thực thi tiếp theo.

---

## Activity Executor

Thực hiện Activity.

---

# Runtime Lifecycle

Runtime Object có vòng đời chung.

```
Create

↓

Initialize

↓

Running

↓

Completed

↓

Archived

↓

Destroyed
```

Chi tiết chuyển trạng thái được định nghĩa trong:

- state-machine.md

---

# Runtime Ownership

Runtime cha sở hữu Runtime con.

```
Execution

owns

Workflow

↓

Workflow

owns

Stage

↓

Stage

owns

Task

↓

Task

owns

Activity
```

Việc sở hữu chỉ áp dụng với Runtime Instance.

Definition Models không có quan hệ sở hữu.

---

# Runtime Result

Mỗi Runtime Instance tạo Result riêng.

```
Activity Result

↓

Task Result

↓

Stage Result

↓

Workflow Result

↓

Execution Result
```

Execution Runtime chịu trách nhiệm tổng hợp Execution Result.

---

# Runtime Boundaries

Execution Runtime chịu trách nhiệm:

- Quản lý Runtime Objects.
- Quản lý Runtime Hierarchy.
- Duy trì Runtime References.
- Tổng hợp Runtime Results.

Execution Runtime không chịu trách nhiệm:

- Business Logic.
- Runtime Scheduling.
- Runtime Orchestration.
- State Transition.
- Context Storage.
- Activity Execution.

---

# Design Principles

## Aggregate Root

Execution là Aggregate Root của Runtime.

---

## Reference Based

Runtime chỉ giữ Reference tới Definition, Context, State và Result.

---

## Hierarchical

Runtime được tổ chức theo mô hình cây.

---

## Lightweight

Runtime không chứa Business Logic.

---

## Single Responsibility

Execution Runtime chỉ quản lý Runtime Objects.

---

## Extensible

Có thể mở rộng Runtime Models mà không ảnh hưởng Definition Models.

---

# Example

```
Execution

Context Ref
State Ref

│

└── Workflow Instance

      Context Ref
      State Ref

      │

      ├── Stage Instance

      │      Context Ref
      │      State Ref

      │      │

      │      ├── Task Instance

      │      │      Context Ref
      │      │      State Ref

      │      │      │

      │      │      ├── Activity Instance
      │      │      └── Activity Instance

      │      │

      │      └── Task Instance

      │

      └── Stage Instance
```

Execution Runtime chỉ quản lý các Runtime Instance.

Việc điều phối được thực hiện bởi Orchestrator.

Việc chuyển trạng thái được thực hiện bởi State Machine.

Việc thực thi Activity được thực hiện bởi Activity Executor.

---

# Related Documents

- README.md
- specification.md
- workflow-model.md
- stage-model.md
- task-model.md
- activity-model.md
- context-model.md
- orchestration-model.md
- state-machine.md
- business-rules.md
