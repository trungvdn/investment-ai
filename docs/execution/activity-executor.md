# Activity Executor

**Status:** Draft

**Version:** 1.0

---

# Purpose

Activity Executor chịu trách nhiệm thực thi Activity Instance trong Execution Layer.

Activity Executor nhận một Activity Instance từ Orchestrator, thực hiện Activity tương ứng và trả kết quả cho Runtime.

Activity Executor không chứa Business Logic và không điều phối quá trình thực thi.

---

# Objectives

Activity Executor hướng đến các mục tiêu sau.

- Thực thi Activity.
- Chuẩn hóa quá trình thực thi.
- Trả kết quả thực thi.
- Báo cáo lỗi thực thi.
- Hỗ trợ mở rộng nhiều loại Activity.

---

# Responsibilities

Activity Executor chịu trách nhiệm:

- Nhận Activity Instance.
- Đọc Activity Definition.
- Chuẩn bị dữ liệu thực thi.
- Gọi Activity Handler.
- Thu thập kết quả.
- Báo cáo kết quả cho Runtime.

Activity Executor không chịu trách nhiệm:

- Business Logic.
- Workflow Orchestration.
- State Transition.
- Context Management.
- Runtime Scheduling.

---

# Execution Flow

Quá trình thực thi Activity.

```
Receive Activity Instance

↓

Load Activity Definition

↓

Prepare Execution Context

↓

Resolve Activity Handler

↓

Execute Activity

↓

Collect Result

↓

Return Result
```

---

# Activity Handler

Mỗi Activity Type được xử lý bởi một Activity Handler.

```
Activity

↓

Activity Handler

↓

Execution Result
```

Activity Executor không trực tiếp thực hiện Business Logic.

Executor chỉ định tuyến Activity đến Handler phù hợp.

---

# Supported Activity Types

Phiên bản hiện tại hỗ trợ các loại Activity sau.

## Capability Activity

Thực thi Capability nội bộ.

Ví dụ:

- Analyze Market Trend
- Calculate Indicator

---

## API Activity

Gọi dịch vụ bên ngoài.

Ví dụ:

- REST API
- GraphQL API

---

## Database Activity

Thao tác cơ sở dữ liệu.

Ví dụ:

- Query
- Insert
- Update

---

## Event Activity

Gửi hoặc nhận Event.

Ví dụ:

- Publish Event
- Consume Event

---

## LLM Activity

Gọi Large Language Model.

Ví dụ:

- Generate Summary
- Analyze News

---

## Human Activity

Yêu cầu người dùng hoặc chuyên gia thực hiện một bước xử lý.

Ví dụ:

- Review
- Approval

---

## Utility Activity

Thực hiện các tác vụ hỗ trợ.

Ví dụ:

- Validate Data
- Transform Data
- Format Output

---

# Activity Resolution

Activity Executor xác định Handler dựa trên Activity Definition.

```
Activity Definition

↓

Activity Type

↓

Activity Handler

↓

Execute
```

Executor không phụ thuộc vào từng Handler cụ thể.

---

# Execution Result

Mỗi Activity trả về một Execution Result.

Execution Result có thể bao gồm:

- Output
- Error
- Metadata

Execution Result được trả về cho Runtime.

---

# Error Handling

Nếu Activity xảy ra lỗi:

```
Execute Activity

↓

Execution Error

↓

Return Error
```

Activity Executor không quyết định Retry hoặc Compensation.

Các quyết định này thuộc về:

- Business Rules
- Orchestrator

---

# Integration

Activity Executor làm việc với các thành phần sau.

## Orchestrator

Cung cấp Activity cần thực thi.

---

## Execution Runtime

Cung cấp Activity Instance.

---

## Context Model

Cung cấp dữ liệu đầu vào cho Activity.

---

## State Machine

Quản lý trạng thái của Activity.

---

## Activity Model

Cung cấp Activity Definition.

---

# Boundaries

Activity Executor chịu trách nhiệm:

- Thực thi Activity.
- Gọi Activity Handler.
- Thu thập kết quả.
- Trả Execution Result.

Activity Executor không chịu trách nhiệm:

- Business Logic.
- Workflow Navigation.
- State Transition.
- Context Storage.
- Runtime Management.

---

# Design Principles

## Single Responsibility

Executor chỉ thực thi Activity.

---

## Handler Based

Mỗi Activity Type có Handler riêng.

---

## Definition Driven

Cách thực thi được xác định bởi Activity Definition.

---

## Pluggable

Có thể bổ sung Handler mới mà không thay đổi Executor.

---

## Reusable

Executor không phụ thuộc Workflow, Stage hoặc Task.

---

# Example

```
Activity Instance

↓

Activity Executor

↓

Resolve Handler

↓

API Handler

↓

REST API

↓

Execution Result

↓

Runtime
```

Ví dụ khác:

```
Activity Instance

↓

Activity Executor

↓

Resolve Handler

↓

LLM Handler

↓

LLM Provider

↓

Execution Result

↓

Runtime
```

Trong cả hai trường hợp:

- Orchestrator chỉ quyết định Activity nào cần chạy.
- Activity Executor thực hiện Activity.
- State Machine quản lý trạng thái.
- Context Model cung cấp dữ liệu đầu vào và nhận dữ liệu đầu ra.

---

# Future Enhancements

Các khả năng sau không thuộc phạm vi phiên bản hiện tại:

- Parallel Activity Execution
- Retry Executor
- Timeout Executor
- Circuit Breaker
- Activity Middleware
- Metrics & Tracing
- Plugin-based Handler Discovery

Các tính năng này có thể được bổ sung mà không thay đổi giao diện của Activity Executor.

---

# Related Documents

- execution-model.md
- orchestration-model.md
- state-machine.md
- context-model.md
- activity-model.md
- business-rules.md
