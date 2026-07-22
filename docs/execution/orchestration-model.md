# Orchestration Model

**Status:** Draft

**Version:** 2.0

---

# Purpose

Orchestrator chịu trách nhiệm điều phối quá trình thực thi của Execution Layer.

Thay vì trực tiếp thay đổi Runtime hoặc thực thi Activity, Orchestrator liên tục đánh giá Runtime Hierarchy và sinh ra các Execution Command để điều khiển Execution Engine.

Orchestrator là trung tâm ra quyết định (Decision Engine) của Execution Layer.

---

# Responsibilities

Orchestrator chịu trách nhiệm:

- Đánh giá Runtime Hierarchy.
- Xác định Runtime có thể thực thi tiếp theo.
- Xác định hành động cần thực hiện.
- Sinh Execution Command.
- Lặp lại quá trình này cho đến khi Execution hoàn thành.

Orchestrator không chịu trách nhiệm:

- Thực thi Activity.
- Thay đổi Runtime State.
- Cập nhật Runtime Context.
- Thực hiện Business Logic.
- Thực thi Execution Command.

---

# Architecture

```
                Runtime Hierarchy
                        │
                        ▼
                 Evaluate Runtime
                        │
                        ▼
                 Decide Next Action
                        │
                        ▼
              Generate Execution Command
                        │
                        ▼
               Execution Engine Components
```

---

# Runtime Evaluation

Trong mỗi vòng lặp, Orchestrator đánh giá toàn bộ Runtime Hierarchy.

Thông tin được sử dụng bao gồm:

- Runtime Structure
- Runtime State
- Runtime Context
- Business Rules
- Execution Policy

Mục tiêu là xác định Runtime tiếp theo có thể được xử lý.

---

# Event Loop

Execution được điều phối thông qua một Event Loop.

```
Execution Started
        │
        ▼
Evaluate Runtime Hierarchy
        │
        ▼
Executable Runtime ?
      /       \
    No         Yes
    │           │
    ▼           ▼
Execution   Generate
Complete    Execution Command
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

Execution chỉ kết thúc khi không còn Runtime nào có thể thực thi.

---

# Runtime Selection

Orchestrator luôn chọn Runtime theo Runtime Hierarchy.

```
Execution Runtime
        │
        ▼
Workflow Runtime
        │
        ▼
Stage Runtime
        │
        ▼
Task Runtime
        │
        ▼
Activity Runtime
```

Runtime cha luôn được kích hoạt trước Runtime con.

Ví dụ.

```
Workflow

↓

Stage

↓

Task

↓

Activity
```

Điều này đảm bảo Runtime Hierarchy luôn nhất quán.

---

# Execution Commands

Orchestrator không thao tác trực tiếp với Runtime.

Thay vào đó, nó sinh ra Execution Command.

Ví dụ.

```
Evaluate Workflow

↓

Transition Workflow → Ready
```

Hoặc.

```
Evaluate Activity

↓

Execute Activity
```

Execution Command sẽ được chuyển cho thành phần phù hợp của Execution Engine để thực thi.

---

# Decision Flow

Ví dụ.

```
Workflow Runtime

Created
```

↓

Orchestrator đánh giá Runtime.

↓

Sinh Command.

```yaml
type: TransitionState

runtime: workflow-001

from: Created

to: Ready
```

↓

State Machine xử lý.

↓

Workflow Runtime

Ready

↓

Orchestrator tiếp tục vòng lặp.

---

# Decision Criteria

Để đưa ra quyết định, Orchestrator xem xét:

- Runtime State.
- Runtime Hierarchy.
- Parent Runtime.
- Child Runtime.
- Business Rules.
- Execution Policy.

Orchestrator không quan tâm Activity sẽ được thực thi như thế nào.

---

# Design Principles

## Stateless

Orchestrator không lưu Runtime Data.

Mọi thông tin đều được lấy từ Runtime hiện tại.

---

## Decision Only

Orchestrator chỉ chịu trách nhiệm đưa ra quyết định.

Mọi thay đổi Runtime đều được thực hiện bởi các thành phần khác.

---

## Hierarchical Evaluation

Runtime luôn được đánh giá theo cấu trúc phân cấp.

---

## Command Driven

Orchestration được thực hiện thông qua Execution Command.

---

## Continuous Evaluation

Orchestrator liên tục đánh giá Runtime sau mỗi thay đổi.

Mỗi Runtime Update đều có thể dẫn đến một quyết định mới.

---

# Related Components

Orchestrator phối hợp với các thành phần sau.

| Component | Purpose |
|----------|---------|
| Execution Runtime | Cung cấp Runtime Hierarchy |
| Context Model | Cung cấp Runtime Data |
| Business Rules | Xác thực điều kiện nghiệp vụ |
| Execution Policy | Xác định chiến lược thực thi |
| Execution Command | Biểu diễn quyết định |
| State Machine | Thực thi Transition Command |
| Activity Executor | Thực thi Activity Command |

---

# Future Evolution

Kiến trúc hiện tại cho phép mở rộng dễ dàng.

Ví dụ:

- Parallel Runtime Scheduling
- Priority Scheduling
- Retry Scheduling
- Human Approval
- Compensation
- Distributed Execution

Những khả năng này có thể được bổ sung bằng cách mở rộng quá trình sinh Execution Command mà không cần thay đổi vai trò của Orchestrator.
