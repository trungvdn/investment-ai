# Execution Layer Specification

**Status:** Draft

**Version:** 3.0

---

# Purpose

Execution Layer chịu trách nhiệm chuyển đổi Business Process Definitions thành Runtime Execution.

Layer này cung cấp một cơ chế chuẩn để:

- Khởi tạo Runtime Hierarchy.
- Điều phối quá trình thực thi.
- Quản lý Runtime State.
- Quản lý Runtime Context.
- Thực thi Technical Actions.
- Áp dụng Business Rules và Execution Policies.

Execution Layer không định nghĩa Business Process và không chứa Business Logic.

---

# Objectives

Execution Layer được thiết kế nhằm đạt các mục tiêu sau.

- Chuẩn hóa quá trình thực thi Business Process.
- Tách biệt Business Logic khỏi Technical Execution.
- Chuẩn hóa Runtime Lifecycle.
- Chuẩn hóa Runtime State.
- Hỗ trợ mở rộng Workflow Engine.
- Đảm bảo khả năng quan sát (Observability).
- Đảm bảo khả năng bảo trì và mở rộng.

---

# Scope

Execution Layer chịu trách nhiệm:

- Chuyển Definition Models thành Runtime Models.
- Khởi tạo Runtime Hierarchy.
- Điều phối Runtime Execution.
- Đánh giá Runtime Hierarchy.
- Sinh Execution Commands.
- Quản lý Runtime State.
- Quản lý Runtime Context.
- Thực thi Activity.
- Áp dụng Execution Policies.
- Đánh giá Business Rules.
- Tổng hợp Execution Result.

Execution Layer không chịu trách nhiệm:

- Thiết kế Business Process.
- Cài đặt Business Logic.
- Triển khai Capability.
- Quản lý Infrastructure.
- Lưu trữ dữ liệu nghiệp vụ.
- Giao tiếp với người dùng.

---

# Execution Model

Execution Layer hoạt động theo mô hình Event Loop.

```
Execution Started
        │
        ▼
Evaluate Runtime Hierarchy
        │
        ▼
Generate Execution Command
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

Execution kết thúc khi không còn Runtime nào có thể thực thi.

---

# Core Responsibilities

Execution Layer bao gồm các nhóm trách nhiệm sau.

## Runtime Management

- Tạo Runtime Objects.
- Quản lý Runtime Hierarchy.
- Quản lý Runtime Lifecycle.
- Theo dõi Runtime Progress.

---

## Runtime Orchestration

- Đánh giá Runtime Hierarchy.
- Xác định Runtime tiếp theo.
- Sinh Execution Commands.

---

## Runtime State Management

- Quản lý Runtime State.
- Kiểm soát State Transition.

---

## Runtime Context Management

- Quản lý Runtime Context.
- Truyền Context theo Runtime Hierarchy.
- Tổng hợp Context sau khi Runtime hoàn thành.

---

## Activity Execution

- Thực thi Activity Runtime.
- Thu thập Execution Result.

---

## Execution Policy Enforcement

- Áp dụng Execution Strategy.
- Retry.
- Timeout.
- Concurrency.
- Cancellation.

---

## Business Validation

- Đánh giá Business Rules.
- Xác định Runtime có được phép tiếp tục hay không.

---

# Design Principles

Execution Layer tuân theo các nguyên tắc sau.

## Separation of Concerns

Business Logic và Technical Execution được tách biệt hoàn toàn.

---

## Single Responsibility

Mỗi thành phần chỉ chịu trách nhiệm cho một chức năng.

---

## Definition Driven

Business Process được định nghĩa bởi Definition Models.

Execution Layer chỉ thực thi các Definition này.

---

## Runtime Oriented

Execution Engine chỉ làm việc với Runtime Objects.

---

## Decision Before Execution

Mọi Runtime Action đều bắt đầu từ quyết định của Orchestrator.

---

## Command Driven Execution

Orchestrator không trực tiếp thay đổi Runtime.

Mọi hành động được biểu diễn thông qua Execution Commands.

---

## Centralized State Management

State chỉ được thay đổi bởi State Machine.

---

## Hierarchical Context

Context được tổ chức theo Runtime Hierarchy.

---

## Continuous Evaluation

Sau mỗi Runtime Update, Runtime Hierarchy được đánh giá lại để xác định bước tiếp theo.

---

## Extensibility

Execution Layer phải cho phép bổ sung Runtime Types, Execution Commands và Execution Policies mới mà không làm thay đổi kiến trúc cốt lõi.

---

# Functional Requirements

Execution Layer phải hỗ trợ các khả năng sau.

## Runtime

- Tạo Execution Runtime.
- Tạo Runtime Hierarchy.
- Quản lý Runtime Lifecycle.
- Theo dõi Runtime Progress.

---

## Orchestration

- Đánh giá Runtime Hierarchy.
- Sinh Execution Commands.
- Điều phối Execution thông qua Event Loop.

---

## Commands

- Biểu diễn Runtime Actions.
- Hỗ trợ nhiều loại Command.
- Đảm bảo Command bất biến (Immutable).

---

## State

- Quản lý Runtime State.
- Kiểm tra State Transition.
- Đảm bảo tính nhất quán của Runtime.

---

## Context

- Khởi tạo Runtime Context.
- Truyền Context.
- Hợp nhất Context.

---

## Activity

- Thực thi Activity Runtime.
- Thu thập Execution Result.

---

## Policy

- Áp dụng Execution Policy.

---

## Business Rules

- Đánh giá Business Rules.

---

# Non-Functional Requirements

Execution Layer phải đáp ứng các yêu cầu sau.

## Modularity

Các thành phần phải độc lập và có thể thay thế.

---

## Maintainability

Kiến trúc phải dễ bảo trì và mở rộng.

---

## Testability

Mỗi thành phần phải có thể kiểm thử độc lập.

---

## Scalability

Execution Engine phải hỗ trợ mở rộng theo chiều ngang và chiều dọc.

---

## Observability

Toàn bộ Runtime Lifecycle phải có thể theo dõi và kiểm tra.

---

## Reliability

Runtime phải luôn duy trì trạng thái nhất quán, ngay cả khi xảy ra lỗi.

---

## Determinism

Với cùng Definition, cùng Input và cùng Execution Policy, Execution Layer phải tạo ra cùng một chuỗi quyết định và cùng một kết quả, trừ khi có yếu tố bất định được khai báo rõ (ví dụ: LLM hoặc dữ liệu thời gian thực).

---

# Out of Scope

Các nội dung sau không thuộc phạm vi của Execution Layer.

- Business Process Design.
- Capability Implementation.
- Agent Collaboration.
- Infrastructure Management.
- Memory Management.
- External Service Implementation.
- User Interface.

---

# Success Criteria

Execution Layer được xem là đáp ứng Specification khi:

- Có thể thực thi một Workflow từ đầu đến cuối.
- Runtime Hierarchy luôn nhất quán.
- Mọi Runtime Transition đều thông qua State Machine.
- Mọi Runtime Action đều bắt nguồn từ Execution Command.
- Orchestrator không trực tiếp thay đổi Runtime.
- Business Logic không nằm trong Execution Layer.
- Các thành phần có thể được kiểm thử độc lập.

---

# Related Documents

## Overview

- README.md

## Architecture

- architecture.md

## Walkthrough

- walkthrough.md

## Definition Models

- workflow-model.md
- stage-model.md
- task-model.md
- activity-model.md

## Runtime Models

- execution-model.md
- context-model.md
- state-machine.md

## Execution Engine

- orchestration-model.md
- execution-command.md
- activity-executor.md
- execution-policy.md

## Business

- business-rules.md
