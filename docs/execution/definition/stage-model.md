# Stage Model

**Status:** Draft

**Version:** 3.0

---

# Purpose

Stage Model định nghĩa một Stage Definition.

Stage Definition đại diện cho một Business Milestone trong một Workflow.

Mỗi Stage xác định một kết quả nghiệp vụ trung gian cần đạt được trước khi Workflow có thể chuyển sang giai đoạn tiếp theo.

Stage Definition là một phần của Workflow Definition.

Stage Definition không đại diện cho quá trình thực thi.

---

# Objectives

Stage Model hướng đến các mục tiêu sau.

- Chia Business Process thành các Business Milestones rõ ràng.
- Định nghĩa mục tiêu của từng giai đoạn.
- Tách biệt mục tiêu nghiệp vụ và cách thực hiện.
- Hỗ trợ tái sử dụng Stage.
- Hỗ trợ Versioning.
- Giúp Workflow dễ đọc, dễ mở rộng và dễ bảo trì.

---

# Stage Definition

Stage Definition mô tả:

- Business Milestone
- Business Scope
- Input Definition
- Output Definition
- Task Definitions
- Entry Criteria
- Exit Criteria
- Completion Criteria

Stage Definition không lưu:

- Runtime State
- Runtime Context
- Runtime Result
- Runtime Scheduling

---

# Stage Structure

```
Stage Definition

├── Metadata
├── Business Milestone
├── Business Scope
├── Input Definition
├── Output Definition
├── Task Definitions
├── Entry Criteria
├── Exit Criteria
└── Completion Criteria
```

---

# Stage Components

## Metadata

Thông tin mô tả Stage.

Ví dụ:

- Stage ID
- Name
- Description
- Version

---

## Business Milestone

Business Milestone là kết quả nghiệp vụ mà Stage cần đạt được.

Ví dụ:

- Required Data Collected
- Market Analysis Completed
- Financial Health Evaluated
- Intrinsic Value Estimated
- Recommendation Generated

Business Milestone mô tả trạng thái mong muốn sau khi Stage hoàn thành.

---

## Business Scope

Phạm vi nghiệp vụ của Stage.

Ví dụ:

Stage "Market Analysis" chỉ chịu trách nhiệm phân tích thị trường.

Không thực hiện định giá doanh nghiệp.

---

## Input Definition

Định nghĩa dữ liệu đầu vào của Stage.

Ví dụ:

- Market Data
- Company Profile
- Financial Statements

---

## Output Definition

Định nghĩa dữ liệu đầu ra.

Ví dụ:

- Market Analysis Result
- Risk Assessment
- Financial Metrics

---

## Task Definitions

Stage bao gồm một hoặc nhiều Task Definition.

Task đại diện cho các Business Work Unit cần thực hiện để đạt được Business Milestone.

Stage không quy định cách Task được thực thi.

---

## Entry Criteria

Điều kiện để Stage được phép bắt đầu.

Ví dụ:

- Previous Stage Completed
- Required Input Available

---

## Exit Criteria

Điều kiện để Stage được phép kết thúc.

Ví dụ:

- All Required Tasks Completed
- Required Outputs Produced

---

## Completion Criteria

Stage được xem là hoàn thành khi Business Milestone đã đạt được.

Completion Criteria mô tả kết quả nghiệp vụ, không mô tả trạng thái thực thi.

---

# Stage Hierarchy

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

Stage chỉ quan tâm đến Task Definition.

Chi tiết kỹ thuật được mô tả trong Activity Definition.

---

# Stage Relationships

## Workflow Model

Stage thuộc về một Workflow Definition.

Workflow chịu trách nhiệm điều phối các Stage.

---

## Task Model

Stage bao gồm nhiều Task Definition.

Task thực hiện các Business Work Unit để đạt được Business Milestone.

---

## Execution Model

Execution Layer tạo Stage Instance từ Stage Definition.

---

# Stage Boundaries

Stage chịu trách nhiệm:

- Định nghĩa một Business Milestone.
- Xác định phạm vi nghiệp vụ.
- Định nghĩa Input và Output.
- Tổ chức các Task.
- Xác định điều kiện bắt đầu và kết thúc.

Stage không chịu trách nhiệm:

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

## Milestone Driven

Mỗi Stage đại diện cho một Business Milestone.

---

## Business Oriented

Stage phản ánh một giai đoạn nghiệp vụ.

---

## Cohesive

Tất cả Task trong cùng một Stage phải cùng phục vụ một Business Milestone.

---

## Independent

Stage không phụ thuộc vào Runtime.

---

## Reusable

Stage có thể được tái sử dụng trong nhiều Workflow.

---

## Versionable

Stage hỗ trợ nhiều phiên bản.

---

## Extensible

Có thể bổ sung hoặc thay đổi Task mà không làm thay đổi Business Milestone.

---

# Example

```
Workflow Definition

Analyze Stock

│
├── Stage
│
│   Business Milestone
│
│   Required Data Collected
│
│   Tasks
│
│   ├── Acquire Market Data
│   ├── Acquire Financial Statements
│   └── Acquire Company Profile
│
├── Stage
│
│   Business Milestone
│
│   Market Analysis Completed
│
│   Tasks
│
│   ├── Analyze Trend
│   ├── Analyze Liquidity
│   └── Analyze Volume
│
├── Stage
│
│   Business Milestone
│
│   Financial Health Evaluated
│
├── Stage
│
│   Business Milestone
│
│   Intrinsic Value Estimated
│
└── Stage

    Business Milestone

    Investment Recommendation Generated
```

Trong ví dụ trên:

- Workflow định nghĩa toàn bộ quy trình phân tích cổ phiếu.
- Mỗi Stage đại diện cho một Business Milestone.
- Mỗi Task là một Business Work Unit để đạt được Milestone.
- Activity (được định nghĩa trong Activity Model) mô tả cách thực hiện từng Task.

---

# Related Documents

- execution-model.md
- workflow-model.md
- task-model.md
- activity-model.md
- orchestration-model.md
- state-machine.md
