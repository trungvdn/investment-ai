# Execution Command

**Status:** Draft

**Version:** 1.0

---

# Purpose

Execution Command đại diện cho một hành động mà Execution Engine cần thực hiện đối với Runtime.

Execution Command được tạo bởi Orchestrator sau khi đánh giá Runtime Hierarchy.

Execution Command không chứa Business Logic.

Execution Command chỉ mô tả **điều cần thực hiện**, không mô tả **cách thực hiện**.

---

# Motivation

Execution Layer tách biệt quá trình:

- Ra quyết định
- Thực thi quyết định

Điều này giúp Orchestrator chỉ tập trung vào việc phân tích Runtime và xác định bước tiếp theo.

Việc thực thi sẽ được giao cho các thành phần chuyên biệt như State Machine hoặc Activity Executor.

---

# Architecture

```
          Runtime Hierarchy
                  │
                  ▼
           Orchestrator
                  │
                  ▼
         Execution Command
                  │
      +-----------+------------+
      │                        │
      ▼                        ▼
State Machine         Activity Executor
```

---

# Responsibilities

Execution Command chịu trách nhiệm:

- Biểu diễn một hành động thực thi.
- Chuyển quyết định của Orchestrator thành một lệnh rõ ràng.
- Mang theo dữ liệu cần thiết để thực thi.

Execution Command không chịu trách nhiệm:

- Đánh giá Runtime.
- Thay đổi Runtime.
- Thực thi Activity.
- Chuyển State.

---

# Command Lifecycle

```
Runtime Changed
        │
        ▼
Orchestrator Evaluate
        │
        ▼
Generate Command
        │
        ▼
Execute Command
        │
        ▼
Runtime Updated
```

---

# Command Types

Execution Layer hiện hỗ trợ hai nhóm Command.

## Transition State Command

Yêu cầu State Machine chuyển Runtime sang trạng thái mới.

Ví dụ:

```yaml
type: TransitionState

runtime: workflow-001

from: Created

to: Ready
```

---

## Execute Activity Command

Yêu cầu Activity Executor thực thi một Activity Runtime.

Ví dụ:

```yaml
type: ExecuteActivity

activity: activity-001
```

---

# Command Flow

Ví dụ.

```
Orchestrator

↓

TransitionStateCommand

↓

State Machine

↓

Workflow Ready
```

Ví dụ khác.

```
Orchestrator

↓

ExecuteActivityCommand

↓

Activity Executor

↓

Activity Handler

↓

Activity Completed
```

---

# Design Principles

## Immutable

Execution Command là immutable.

Sau khi được tạo, Command không được thay đổi.

---

## Single Purpose

Một Command chỉ biểu diễn một hành động.

Không được kết hợp nhiều hành động trong cùng một Command.

---

## Explicit

Execution Command phải mô tả rõ:

- Runtime nào.
- Hành động gì.
- Dữ liệu cần thiết.

---

# Future Evolution

Trong tương lai có thể bổ sung:

- RetryActivityCommand
- CancelRuntimeCommand
- ResumeRuntimeCommand
- CompensationCommand
- TimeoutCommand
- SuspendRuntimeCommand

Mỗi loại Command sẽ được xử lý bởi thành phần chuyên biệt tương ứng.
