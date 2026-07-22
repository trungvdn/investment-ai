# Execution Runtime Model

**Status:** Draft

**Version:** 3.0

---

# Purpose

Execution Runtime đại diện cho một lần thực thi của một Workflow Definition.

Mỗi Execution Runtime là một phiên thực thi (execution session) độc lập, chứa toàn bộ Runtime Objects, Runtime State và Runtime Context cần thiết để hoàn thành một Business Process.

Execution Runtime là **Aggregate Root** của toàn bộ Runtime Hierarchy.

---

# Responsibilities

Execution Runtime chịu trách nhiệm:

- Quản lý Runtime Hierarchy.
- Quản lý Runtime Lifecycle.
- Liên kết Runtime với Definition.
- Quản lý Runtime References.
- Tổng hợp Execution Result.
- Cung cấp Runtime Information cho Execution Engine.

Execution Runtime không chịu trách nhiệm:

- Điều phối Execution.
- Thực thi Activity.
- Chuyển Runtime State.
- Đánh giá Business Rules.
- Sinh Execution Commands.

---

# Runtime Hierarchy

Execution Runtime quản lý toàn bộ Runtime Objects.

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

Mỗi Runtime chỉ có một Parent.

Runtime Hierarchy phản ánh trực tiếp cấu trúc của Definition Models.

---

# Runtime Composition

Execution Runtime bao gồm các thành phần sau.

```
Execution Runtime

├── Metadata
├── Runtime Hierarchy
├── Runtime References
├── Execution Context
├── Execution Result
└── Lifecycle Information
```

Execution Runtime không chứa Business Logic.

---

# Metadata

Metadata mô tả Execution Runtime.

Ví dụ.

```yaml
id: exec-001

definition: Analyze Company

createdAt: ...

startedAt: ...

completedAt: ...

status: Running
```

Metadata giúp định danh và theo dõi một Execution.

---

# Runtime Hierarchy

Execution Runtime quản lý toàn bộ Runtime Objects.

Ví dụ.

```text
Execution

└── Workflow

    ├── Stage

    │     ├── Task

    │     │     ├── Activity

    │     │     └── Activity

    │     │
    │     └── Task

    └── Stage
```

Execution Runtime không quyết định Runtime nào sẽ được thực thi.

Đó là trách nhiệm của Orchestrator.

---

# Runtime References

Execution Runtime không sao chép Definition.

Mỗi Runtime chỉ giữ Reference tới Definition tương ứng.

```
Execution Runtime

↓

Workflow Definition
```

Tương tự.

```
Task Runtime

↓

Task Definition
```

Điều này cho phép nhiều Execution cùng chia sẻ một Definition.

---

# Context

Execution Runtime giữ Execution Context.

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

Execution Runtime không xử lý Context Propagation.

Đó là trách nhiệm của Context Model.

---

# Runtime State

Mỗi Runtime có Runtime State riêng.

Ví dụ.

```
Execution

Running
```

```
Workflow

Running
```

```
Task

Completed
```

Execution Runtime lưu Runtime State nhưng không thay đổi Runtime State.

Runtime State chỉ được thay đổi bởi State Machine.

---

# Execution Result

Sau khi Runtime hoàn thành.

Execution Runtime tổng hợp kết quả.

Ví dụ.

```yaml
status: Success

reportId: report-001

duration: 18s
```

Execution Result đại diện cho kết quả cuối cùng của toàn bộ Execution.

---

# Runtime Lifecycle

Execution Runtime trải qua các giai đoạn.

```
Created

↓

Initialized

↓

Running

↓

Completed
```

Ngoài ra.

```
Failed

Cancelled
```

Lifecycle phản ánh vòng đời của toàn bộ Execution.

Không phản ánh trạng thái của từng Runtime Object.

---

# Aggregate Root

Execution Runtime là Aggregate Root.

Mọi Runtime Objects đều thuộc về Execution Runtime.

```
Execution Runtime
        │
        ├── Workflow Runtime
        ├── Stage Runtime
        ├── Task Runtime
        └── Activity Runtime
```

Execution Engine luôn bắt đầu từ Execution Runtime khi đánh giá Runtime Hierarchy.

---

# Interaction

Execution Runtime được sử dụng bởi nhiều thành phần.

```
Execution Runtime

        │

        ├── Orchestrator

        ├── State Machine

        ├── Activity Executor

        ├── Context Model

        └── Business Rules
```

Execution Runtime chỉ cung cấp dữ liệu.

Không thực hiện xử lý nghiệp vụ.

---

# Design Principles

## Aggregate Root

Execution Runtime là điểm truy cập duy nhất tới Runtime Hierarchy.

---

## Definition Reference

Runtime luôn tham chiếu Definition.

Không sao chép Definition.

---

## Runtime Only

Execution Runtime chỉ lưu Runtime Information.

Không lưu Business Logic.

---

## Immutable Structure

Runtime Hierarchy được tạo trong giai đoạn Initialization.

Cấu trúc Runtime không thay đổi trong quá trình thực thi.

Chỉ Runtime State, Context và Result thay đổi.

---

## Single Responsibility

Execution Runtime chỉ quản lý Runtime Objects và dữ liệu Runtime.

---

# Related Components

| Component | Relationship |
|-----------|--------------|
| Workflow Definition | Runtime Reference |
| Orchestrator | Runtime Evaluation |
| Execution Command | Runtime Action Target |
| State Machine | Runtime State Management |
| Activity Executor | Activity Execution |
| Context Model | Context Propagation |
| Business Rules | Runtime Validation |

---

# Future Evolution

Execution Runtime có thể được mở rộng để hỗ trợ:

- Execution History
- Execution Metrics
- Checkpoint / Resume
- Snapshot Persistence
- Distributed Execution
- Event Sourcing
