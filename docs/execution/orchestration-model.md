# Orchestration Model

**Status:** Draft

**Version:** 1.0

---

# Purpose

Orchestration Model định nghĩa cách Execution Layer điều phối quá trình thực thi các Runtime Instance.

Orchestrator chịu trách nhiệm quản lý luồng thực thi từ Execution đến Activity, đảm bảo các Runtime Instance được khởi tạo, thực thi và hoàn thành theo đúng thứ tự được định nghĩa trong Business Process.

Orchestrator không chứa Business Logic.

---

# Objectives

Orchestration Model hướng đến các mục tiêu sau.

- Điều phối Runtime Execution.
- Quản lý vòng đời Runtime Instance.
- Thực thi Business Process theo đúng Definition.
- Quản lý thứ tự thực thi.
- Thu thập kết quả thực thi.

---

# Responsibilities

Orchestrator chịu trách nhiệm:

- Khởi tạo Runtime Instance.
- Điều phối Workflow.
- Điều phối Stage.
- Điều phối Task.
- Điều phối Activity.
- Thu thập kết quả.
- Kết thúc Runtime.

Orchestrator không chịu trách nhiệm:

- Business Logic.
- Business Rules.
- Activity Implementation.
- State Management.
- Context Management.

---

# Orchestration Hierarchy

Orchestrator điều phối theo cấu trúc sau.

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

Runtime cha chịu trách nhiệm điều phối Runtime con.

---

# Orchestration Flow

Quá trình điều phối diễn ra theo các bước sau.

```
Receive Execution Request

↓

Create Execution Instance

↓

Create Workflow Instance

↓

Execute Workflow

↓

Create Stage Instance

↓

Execute Stage

↓

Create Task Instance

↓

Execute Task

↓

Create Activity Instance

↓

Execute Activity

↓

Collect Result

↓

Complete Task

↓

Complete Stage

↓

Complete Workflow

↓

Complete Execution
```

---

# Execution Strategy

Phiên bản hiện tại sử dụng chiến lược thực thi tuần tự (Sequential Execution).

```
Stage 1

↓

Stage 2

↓

Stage 3
```

Trong mỗi Stage:

```
Task A

↓

Task B

↓

Task C
```

Trong mỗi Task:

```
Activity A

↓

Activity B

↓

Activity C
```

Các chiến lược song song sẽ được bổ sung trong các phiên bản sau.

---

# Runtime Creation

Execution Layer tạo Runtime Instance theo thứ tự:

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

Runtime cha chịu trách nhiệm tạo Runtime con.

---

# Result Propagation

Sau khi Activity hoàn thành:

```
Activity Result

↓

Task

↓

Stage

↓

Workflow

↓

Execution
```

Mỗi Runtime chịu trách nhiệm tổng hợp kết quả từ Runtime con.

---

# Failure Handling

Nếu một Runtime thất bại:

- Runtime hiện tại kết thúc.
- Runtime cha nhận thông tin lỗi.
- Việc xử lý tiếp theo được quyết định theo Business Rules.

Chi tiết được mô tả trong:

- business-rules.md

---

# Integration

Orchestrator làm việc cùng:

## Execution Model

Quản lý Runtime Instance.

---

## Context Model

Cung cấp dữ liệu Runtime.

---

## State Machine

Theo dõi trạng thái Runtime.

---

## Activity Model

Thực thi các Activity.

---

# Boundaries

Orchestrator chịu trách nhiệm:

- Điều phối Runtime.
- Khởi tạo Runtime Instance.
- Gọi Runtime con.
- Thu thập kết quả.

Orchestrator không chịu trách nhiệm:

- Business Logic.
- Capability Implementation.
- Context Storage.
- State Transition Rules.

---

# Design Principles

## Lightweight

Orchestrator chỉ điều phối.

---

## Definition Driven

Luồng thực thi được quyết định bởi Definition Models.

---

## Runtime Oriented

Orchestrator chỉ làm việc với Runtime Instance.

---

## Hierarchical

Runtime cha điều phối Runtime con.

---

## Extensible

Có thể bổ sung các chiến lược thực thi mới mà không thay đổi Definition Models.

---

# Example

```
Execution

│

└── Workflow

      │

      ├── Stage

      │      │

      │      ├── Task

      │      │      │

      │      │      ├── Activity

      │      │      ├── Activity

      │      │      └── Activity

      │      │

      │      └── Task

      │

      └── Stage
```

Trong ví dụ trên:

- Execution khởi tạo Workflow.
- Workflow điều phối các Stage.
- Stage điều phối các Task.
- Task điều phối các Activity.
- Activity thực hiện hành động kỹ thuật và trả kết quả ngược lên Runtime cha.

---

# Future Enhancements

Các tính năng sau không thuộc phạm vi phiên bản hiện tại:

- Parallel Execution
- Conditional Branching
- Retry Policy
- Timeout Policy
- Compensation
- Distributed Execution
- Event-driven Orchestration

Các khả năng này có thể được bổ sung mà không làm thay đổi Definition Models.

---

# Related Documents

- execution-model.md
- workflow-model.md
- stage-model.md
- task-model.md
- activity-model.md
- context-model.md
- state-machine.md
- business-rules.md
