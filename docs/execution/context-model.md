# Context Model

**Status:** Draft

**Version:** 1.0

---

# Purpose

Context Model định nghĩa cách dữ liệu runtime được tạo, quản lý, chia sẻ và truyền trong suốt vòng đời của một Execution.

Context là cơ chế trao đổi dữ liệu giữa các Runtime Instance, đảm bảo mọi thành phần trong Execution có thể truy cập đúng dữ liệu tại đúng thời điểm mà vẫn giữ được ranh giới và tính độc lập.

Context Model không định nghĩa Business Logic.

---

# Objectives

Context Model hướng đến các mục tiêu sau.

- Chuẩn hóa dữ liệu runtime.
- Chia sẻ dữ liệu giữa các Runtime Instance.
- Giảm phụ thuộc giữa Workflow, Stage, Task và Activity.
- Hỗ trợ mở rộng Execution Engine.
- Hỗ trợ khả năng theo dõi và gỡ lỗi.

---

# Context Definition

Context là tập hợp dữ liệu runtime tồn tại trong một Execution.

Context được tạo khi Execution bắt đầu và tồn tại đến khi Execution kết thúc.

Context có thể được:

- Đọc.
- Bổ sung.
- Cập nhật.
- Truyền giữa các Runtime Instance.

Context không đại diện cho:

- Runtime State.
- Runtime Result.
- Business Definition.

---

# Context Hierarchy

Execution Layer sử dụng mô hình Context phân cấp.

```
Execution Context
        │
        ▼
Workflow Context
        │
        ▼
Stage Context
        │
        ▼
Task Context
        │
        ▼
Activity Context
```

Mỗi cấp Runtime có Context riêng.

Context cấp dưới kế thừa dữ liệu từ Context cấp trên theo quy tắc được định nghĩa.

---

# Context Components

## Execution Context

Execution Context chứa dữ liệu dùng chung cho toàn bộ Execution.

Ví dụ:

- Request
- Correlation ID
- User
- Tenant
- Global Parameters
- Shared Runtime Data

Execution Context được tạo khi Execution bắt đầu.

---

## Workflow Context

Workflow Context chứa dữ liệu phục vụ một Workflow Instance.

Ví dụ:

- Workflow Input
- Workflow Output
- Workflow Variables

Workflow Context kế thừa Execution Context.

---

## Stage Context

Stage Context chứa dữ liệu của một Stage Instance.

Ví dụ:

- Stage Input
- Stage Output
- Intermediate Data

Stage Context kế thừa Workflow Context.

---

## Task Context

Task Context chứa dữ liệu của một Task Instance.

Ví dụ:

- Task Input
- Task Output
- Working Data

Task Context kế thừa Stage Context.

---

## Activity Context

Activity Context chứa dữ liệu cần thiết để thực hiện một Activity.

Ví dụ:

- Activity Parameters
- Temporary Variables
- Execution Metadata

Activity Context kế thừa Task Context.

---

# Context Propagation

Context được truyền theo chiều từ trên xuống.

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

Mỗi Runtime Instance nhận Context từ Runtime cha khi được tạo.

---

# Context Visibility

Mỗi Runtime Instance chỉ có thể truy cập các Context nằm trong phạm vi của mình.

| Runtime | Có thể truy cập |
|----------|-----------------|
| Execution | Execution Context |
| Workflow | Workflow Context + Execution Context |
| Stage | Stage Context + Workflow Context + Execution Context |
| Task | Task Context + Stage Context + Workflow Context + Execution Context |
| Activity | Activity Context + Task Context + Stage Context + Workflow Context + Execution Context |

---

# Context Lifecycle

Context được quản lý theo vòng đời của Runtime.

```
Create

↓

Populate

↓

Read / Update

↓

Propagate

↓

Archive

↓

Destroy
```

---

# Context Ownership

Mỗi Context có một Owner.

| Context | Owner |
|----------|-------|
| Execution Context | Execution Instance |
| Workflow Context | Workflow Instance |
| Stage Context | Stage Instance |
| Task Context | Task Instance |
| Activity Context | Activity Instance |

Owner chịu trách nhiệm quản lý dữ liệu trong phạm vi Context của mình.

---

# Context Isolation

Mỗi Runtime chỉ được phép ghi vào Context của chính mình.

Ví dụ:

- Task chỉ được cập nhật Task Context.
- Stage chỉ được cập nhật Stage Context.
- Workflow chỉ được cập nhật Workflow Context.

Các Context của Runtime cha chỉ được đọc, trừ khi có chính sách cho phép.

Nguyên tắc này giúp giảm phụ thuộc và tránh các thay đổi ngoài ý muốn.

---

# Context Merge

Khi một Runtime hoàn thành, dữ liệu đầu ra có thể được hợp nhất vào Context của Runtime cha.

Ví dụ:

```
Task Output

↓

Stage Context
```

```
Stage Output

↓

Workflow Context
```

```
Workflow Output

↓

Execution Context
```

Quy tắc hợp nhất được điều phối bởi Execution Engine.

---

# Context Boundaries

Context chịu trách nhiệm:

- Lưu dữ liệu runtime.
- Truyền dữ liệu giữa các Runtime Instance.
- Cung cấp dữ liệu đầu vào cho Runtime con.
- Thu thập dữ liệu đầu ra từ Runtime con.

Context không chịu trách nhiệm:

- Business Logic.
- State Management.
- Scheduling.
- Retry.
- Error Handling.

Các trách nhiệm này thuộc các thành phần khác của Execution Layer.

---

# Design Principles

## Scoped

Mỗi Runtime có Context riêng.

---

## Hierarchical

Context được tổ chức theo cấu trúc phân cấp.

---

## Inherited

Runtime con kế thừa Context của Runtime cha.

---

## Isolated

Mỗi Runtime chỉ được phép ghi vào Context của mình.

---

## Composable

Context có thể được hợp nhất theo các quy tắc xác định.

---

## Extensible

Có thể mở rộng Context mà không ảnh hưởng đến Runtime khác.

---

# Example

```
Execution Context

├── request
├── user
├── tenant
└── correlationId

        │
        ▼

Workflow Context

├── stockSymbol
├── analysisType
└── market

        │
        ▼

Stage Context

├── marketData
└── news

        │
        ▼

Task Context

├── indicators
└── trend

        │
        ▼

Activity Context

├── apiEndpoint
├── timeout
└── requestPayload
```

Trong ví dụ trên:

- Execution Context chứa dữ liệu toàn cục.
- Workflow Context chứa dữ liệu của quy trình phân tích.
- Stage Context chứa dữ liệu của từng giai đoạn.
- Task Context chứa dữ liệu của từng công việc.
- Activity Context chỉ chứa dữ liệu cần thiết để thực hiện một hành động kỹ thuật.

---

# Related Documents

- execution-model.md
- workflow-model.md
- stage-model.md
- task-model.md
- activity-model.md
- orchestration-model.md
- state-machine.md
