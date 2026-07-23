# State Machine Model

**Status:** Draft

**Version:** 1.0

---

# Purpose

State Machine Model định nghĩa vòng đời và các quy tắc chuyển trạng thái của Runtime Instance trong Execution Layer.

State Machine là thành phần duy nhất chịu trách nhiệm xác định trạng thái hiện tại của Runtime và kiểm soát các chuyển đổi trạng thái hợp lệ.

State Machine không chứa Business Logic và không điều phối quá trình thực thi.

---

# Objectives

State Machine hướng đến các mục tiêu sau.

- Chuẩn hóa vòng đời Runtime.
- Quản lý Runtime State.
- Kiểm soát State Transition.
- Đảm bảo tính nhất quán của Runtime.
- Hỗ trợ mở rộng Execution Engine.

---

# Responsibilities

State Machine chịu trách nhiệm:

- Lưu trạng thái hiện tại của Runtime.
- Kiểm tra tính hợp lệ của State Transition.
- Thực hiện State Transition.
- Từ chối các Transition không hợp lệ.

State Machine không chịu trách nhiệm:

- Business Logic.
- Runtime Scheduling.
- Runtime Orchestration.
- Context Management.
- Activity Execution.

---

# Runtime Scope

Mỗi Runtime Instance đều có State riêng.

```
Execution
        │
        ▼
Workflow
        │
        ▼
Stage
        │
        ▼
Task
        │
        ▼
Activity
```

State Machine áp dụng cho tất cả Runtime Instance.

---

# Runtime States

Phiên bản hiện tại sử dụng các trạng thái sau.

```
Created

↓

Ready

↓

Running

↓

Completed
```

Ngoài ra còn có:

```
Failed

Cancelled
```

---

# State Definitions

## Created

Runtime Instance vừa được tạo.

Runtime chưa sẵn sàng để thực thi.

---

## Ready

Runtime đã được khởi tạo đầy đủ.

Có thể bắt đầu thực thi.

---

## Running

Runtime đang được thực thi.

---

## Completed

Runtime đã hoàn thành thành công.

---

## Failed

Runtime kết thúc do lỗi.

---

## Cancelled

Runtime bị hủy trước khi hoàn thành.

---

# State Transition

Các Transition hợp lệ.

```
Created

↓

Ready

↓

Running

├──────────────► Completed

├──────────────► Failed

└──────────────► Cancelled
```

---

# Transition Rules

Các Transition hợp lệ.

| From | To |
|------|----|
| Created | Ready |
| Ready | Running |
| Running | Completed |
| Running | Failed |
| Running | Cancelled |

Mọi Transition khác đều không hợp lệ.

---

# Invalid Transition

Ví dụ.

Không hợp lệ:

```
Created

↓

Completed
```

Không hợp lệ:

```
Completed

↓

Running
```

Không hợp lệ:

```
Failed

↓

Running
```

State Machine phải từ chối các Transition này.

---

# State Ownership

State thuộc về Runtime Instance.

Ví dụ.

```
Task Instance

└── Task State
```

Mỗi Runtime có State độc lập.

---

# State Transition Flow

```
Create Runtime

↓

Created

↓

Initialize Runtime

↓

Ready

↓

Start Execution

↓

Running

↓

Execution Success

↓

Completed
```

Hoặc:

```
Running

↓

Failed
```

Hoặc:

```
Running

↓

Cancelled
```

---

# Interaction

State Machine làm việc với các thành phần sau.

## Execution Runtime

Runtime giữ State Reference.

---

## Orchestrator

Orchestrator đọc Runtime State để quyết định Runtime tiếp theo.

Orchestrator không được thay đổi State.

---

## Context Model

Không phụ thuộc trực tiếp.

---

## Activity Executor

Executor báo kết quả thực thi.

State Machine quyết định Transition tương ứng.

---

# Boundaries

State Machine chịu trách nhiệm:

- Quản lý Runtime State.
- Kiểm tra Transition.
- Thực hiện Transition.

State Machine không chịu trách nhiệm:

- Business Logic.
- Runtime Scheduling.
- Activity Execution.
- Context Storage.
- Result Collection.

---

# Design Principles

## Single Source of Truth

State chỉ được quản lý bởi State Machine.

---

## Immutable Transition Rules

Các Transition được định nghĩa rõ ràng.

---

## Runtime Independent

Mọi Runtime đều sử dụng cùng một State Model.

---

## Predictable

Một Runtime luôn có đúng một State tại một thời điểm.

---

## Extensible

Có thể bổ sung State mới mà không ảnh hưởng các thành phần khác.

---

# Example

```
Task Instance

State = Ready

↓

Orchestrator quyết định thực thi Task

↓

State Machine

Ready → Running

↓

Activity Executor thực thi

↓

State Machine

Running → Completed
```

Trong ví dụ trên:

- Orchestrator chỉ đưa ra quyết định thực thi.
- State Machine thực hiện chuyển trạng thái.
- Activity Executor thực hiện công việc kỹ thuật.
- Execution Runtime lưu tham chiếu đến State hiện tại.

---

# Related Documents

- execution-model.md
- orchestration-model.md
- context-model.md
- workflow-model.md
- stage-model.md
- task-model.md
- activity-model.md
- business-rules.md
