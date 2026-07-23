# Workflow Model

**Status:** Draft

**Version:** 3.0

---

# Purpose

Workflow Model định nghĩa một Workflow Definition.

Workflow Definition mô tả một Business Process nhằm đạt được một Business Goal.

Workflow Definition xác định:

- Mục tiêu nghiệp vụ.
- Các giai đoạn nghiệp vụ (Stages).
- Quan hệ giữa các Stage.
- Điều kiện chuyển tiếp giữa các Stage.
- Input và Output của quy trình.

Workflow Definition là một bản thiết kế (Blueprint).

Workflow Definition không đại diện cho quá trình thực thi.

---

# Objectives

Workflow Model hướng đến các mục tiêu sau.

- Chuẩn hóa Business Process.
- Mô tả quy trình nghiệp vụ.
- Tách biệt Business Process và Runtime Execution.
- Cho phép tái sử dụng Workflow.
- Hỗ trợ Versioning.
- Hỗ trợ mở rộng quy trình.

---

# Workflow Definition

Workflow Definition mô tả:

- Business Goal
- Business Scope
- Input Definition
- Output Definition
- Stage Definitions
- Transition Rules
- Completion Criteria

Workflow Definition không lưu:

- Runtime State
- Runtime Context
- Runtime Result
- Runtime Scheduling

---

# Workflow Structure

```
Workflow Definition

├── Metadata
├── Business Goal
├── Business Scope
├── Input Definition
├── Output Definition
├── Stage Definitions
├── Transition Rules
└── Completion Criteria
```

---

# Workflow Components

## Metadata

Thông tin mô tả Workflow.

Ví dụ:

- Workflow ID
- Name
- Version
- Description
- Owner

---

## Business Goal

Mục tiêu cuối cùng mà Workflow hướng tới.

Ví dụ:

- Analyze Stock
- Evaluate Portfolio
- Discover Investment Opportunities
- Generate Investment Recommendation

Business Goal là lý do tồn tại của Workflow.

---

## Business Scope

Phạm vi nghiệp vụ mà Workflow chịu trách nhiệm.

Ví dụ:

Workflow chỉ chịu trách nhiệm phân tích cổ phiếu.

Workflow không chịu trách nhiệm quản lý danh mục.

---

## Input Definition

Định nghĩa dữ liệu đầu vào.

Ví dụ:

- Stock Symbol
- User Parameters
- Portfolio
- Watchlist

---

## Output Definition

Định nghĩa dữ liệu đầu ra.

Ví dụ:

- Recommendation
- Analysis Report
- Investment Opportunity

---

## Stage Definitions

Workflow được chia thành nhiều Stage.

Mỗi Stage đại diện cho một Business Milestone.

Workflow không quan tâm bên trong Stage có bao nhiêu Task.

---

## Transition Rules

Quy tắc chuyển giữa các Stage.

Ví dụ:

- Always
- On Success
- On Condition

Workflow chỉ mô tả quy tắc nghiệp vụ.

Execution Layer quyết định cách thực thi.

---

## Completion Criteria

Điều kiện kết thúc Workflow.

Ví dụ:

- All Stages Completed
- Business Goal Achieved

---

# Workflow Hierarchy

Workflow Definition được tổ chức như sau.

```
Workflow Definition
        │
        ▼
Stage Definition
        │
        ▼
Task Definition
        │
        ▼
Activity Definition
```

Workflow chỉ biết Stage.

Các mức còn lại thuộc chi tiết triển khai của Stage.

---

# Workflow Relationships

Workflow Definition có quan hệ với:

## Stage Model

Workflow bao gồm nhiều Stage Definition.

---

## Execution Model

Execution Layer tạo Workflow Instance từ Workflow Definition.

---

## Business Capability

Workflow không gọi Business Capability trực tiếp.

Business Capability được sử dụng trong Activity.

---

# Workflow Boundaries

Workflow chịu trách nhiệm:

- Mô tả Business Goal.
- Mô tả Business Process.
- Mô tả Business Milestones.
- Định nghĩa Input.
- Định nghĩa Output.
- Định nghĩa luồng chuyển giữa các Stage.

Workflow không chịu trách nhiệm:

- Runtime State.
- Runtime Context.
- Scheduling.
- Retry.
- Timeout.
- Activity Execution.
- Business Capability Invocation.

Các trách nhiệm trên thuộc Execution Layer.

---

# Design Principles

## Business Driven

Workflow phản ánh quy trình nghiệp vụ.

---

## Goal Oriented

Mỗi Workflow phục vụ một Business Goal.

---

## Stage Based

Workflow được chia thành các Business Milestone.

---

## Independent

Workflow không phụ thuộc Runtime.

---

## Reusable

Workflow có thể được sử dụng trong nhiều Execution.

---

## Versionable

Workflow có thể tồn tại nhiều phiên bản.

---

## Extensible

Có thể bổ sung Stage mới mà không ảnh hưởng đến Execution Engine.

---

# Example

```
Workflow Definition

Analyze Stock

Business Goal

Đánh giá một cổ phiếu và đưa ra khuyến nghị đầu tư.

│
├── Stage
│      Data Collection
│
├── Stage
│      Market Analysis
│
├── Stage
│      Fundamental Analysis
│
├── Stage
│      Valuation
│
└── Stage
       Investment Recommendation
```

Workflow chỉ mô tả các Business Milestone.

Execution Layer sẽ quyết định cách thực thi từng Stage.

---

# Related Documents

- execution-model.md
- stage-model.md
- task-model.md
- activity-model.md
- orchestration-model.md
- state-machine.md
