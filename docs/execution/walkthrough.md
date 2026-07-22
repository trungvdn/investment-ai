# Execution Walkthrough

**Status:** Draft

**Version:** 1.0

---

# Overview

Tài liệu này mô tả toàn bộ vòng đời của một Execution trong Execution Layer, từ thời điểm hệ thống nhận một yêu cầu từ Client cho đến khi Workflow hoàn thành.

Mục tiêu của tài liệu là giúp người đọc hiểu cách các thành phần trong Execution Layer phối hợp với nhau trong quá trình thực thi.

Walkthrough không tập trung vào chi tiết cài đặt của từng thành phần mà mô tả cách chúng tương tác trong một phiên thực thi thực tế.

Trong suốt tài liệu, chúng ta sẽ theo dõi cùng một Execution và quan sát sự thay đổi của:

- Runtime Objects
- Runtime State
- Runtime Context
- Execution Flow
- Component Interaction

Sau khi hoàn thành walkthrough, người đọc sẽ hiểu:

- Một Execution được tạo như thế nào.
- Runtime Hierarchy được xây dựng ra sao.
- Orchestrator điều phối Workflow như thế nào.
- State Machine quản lý Runtime State ra sao.
- Activity Executor thực thi Activity như thế nào.
- Context được truyền và cập nhật trong Runtime Hierarchy.
- Một Workflow hoàn thành như thế nào.

---

# Business Scenario

Để minh họa cho toàn bộ quá trình thực thi, chúng ta sử dụng một Workflow có tên:

> **Analyze Company**

Workflow này phân tích một doanh nghiệp và tạo báo cáo đầu tư.

Client gửi yêu cầu phân tích một mã cổ phiếu.

Ví dụ:

```json
POST /analysis

{
    "ticker": "FPT",
    "market": "VN"
}
```

Execution Layer sẽ thực hiện các bước sau.

1. Tạo Execution Runtime.
2. Khởi tạo Runtime Context.
3. Khởi tạo Workflow Runtime.
4. Thực thi các Stage.
5. Thực thi các Task.
6. Thực thi từng Activity.
7. Tổng hợp kết quả.
8. Sinh báo cáo cuối cùng.

Đây là một ví dụ đơn giản nhưng đủ để minh họa gần như toàn bộ kiến trúc của Execution Layer.

---

# Workflow Definition

Workflow Definition mô tả toàn bộ Business Process cần thực hiện.

Workflow không chứa Runtime State và không được thực thi trực tiếp.

Nó chỉ đóng vai trò là blueprint cho Execution Runtime.

Workflow của ví dụ này được định nghĩa như sau.

```
Workflow
└── Analyze Company
```

Workflow bao gồm ba Stage.

```
Analyze Company

├── Load Company Data
├── Analyze Company
└── Generate Report
```

Mỗi Stage đại diện cho một Business Milestone.

Workflow chỉ mô tả cấu trúc nghiệp vụ.

Việc thực thi sẽ được thực hiện bởi Runtime Models trong Execution Layer.

---

# Workflow Hierarchy

Workflow được tổ chức theo cấu trúc phân cấp.

```
Workflow
    │
    ├── Stage
    │       │
    │       ├── Task
    │       │      │
    │       │      └── Activity
    │       │
    │       └── Task
    │
    └── Stage
```

Mỗi cấp có trách nhiệm khác nhau.

| Level | Responsibility |
|---------|----------------|
| Workflow | Business Process |
| Stage | Business Milestone |
| Task | Business Work |
| Activity | Technical Action |

Execution Layer sẽ tạo Runtime tương ứng cho từng Definition này.

---

# Stage Definition

Workflow của ví dụ bao gồm ba Stage.

## Stage 1

```
Load Company Data
```

Mục tiêu.

Thu thập toàn bộ dữ liệu đầu vào cần thiết cho việc phân tích.

Stage này không thực hiện phân tích.

Nó chỉ chuẩn bị dữ liệu.

---

## Stage 2

```
Analyze Company
```

Mục tiêu.

Thực hiện toàn bộ quá trình phân tích doanh nghiệp.

Bao gồm.

- Financial Analysis
- Valuation
- Risk Analysis

Đây là Stage quan trọng nhất của Workflow.

---

## Stage 3

```
Generate Report
```

Mục tiêu.

Tổng hợp toàn bộ kết quả phân tích thành báo cáo cuối cùng.

---

# Task Definition

Mỗi Stage bao gồm nhiều Task.

Task đại diện cho một đơn vị công việc có ý nghĩa nghiệp vụ.

## Stage 1

```
Load Company Data

├── Load Company Profile
└── Load Financial Statements
```

---

## Stage 2

```
Analyze Company

├── Analyze Financials
├── Analyze Valuation
└── Analyze Risks
```

---

## Stage 3

```
Generate Report

├── Build Investment Report
└── Publish Report
```

Task không trực tiếp gọi API hoặc Database.

Task chỉ mô tả Business Work.

Các Technical Actions sẽ được chia nhỏ thành Activity.

---

# Activity Definition

Activity là đơn vị thực thi nhỏ nhất trong Execution Layer.

Mỗi Activity thực hiện đúng một Technical Action.

Ví dụ.

Task:

```
Load Company Profile
```

Bao gồm ba Activity.

```
Load Company Profile

├── Call Company API
├── Parse Response
└── Store Company Profile
```

Một ví dụ khác.

Task:

```
Analyze Financials
```

Bao gồm.

```
Analyze Financials

├── Calculate Revenue Growth
├── Calculate Profit Margin
├── Calculate ROE
└── Generate Financial Summary
```

Mỗi Activity có thể được thực hiện bởi một Activity Handler khác nhau.

Ví dụ.

| Activity | Handler |
|-----------|----------|
| Call API | API Handler |
| Execute SQL | Database Handler |
| Invoke LLM | LLM Handler |
| Publish Event | Event Handler |
| Save File | File Handler |

Activity không biết ai gọi nó.

Activity chỉ mô tả hành động kỹ thuật cần thực hiện.

---

# Complete Workflow Definition

Toàn bộ Workflow được biểu diễn như sau.

```
Workflow
└── Analyze Company
    │
    ├── Stage
    │     Load Company Data
    │     │
    │     ├── Task
    │     │     Load Company Profile
    │     │
    │     │     ├── Call Company API
    │     │     ├── Parse Response
    │     │     └── Store Company Profile
    │     │
    │     └── Task
    │           Load Financial Statements
    │
    ├── Stage
    │     Analyze Company
    │     │
    │     ├── Analyze Financials
    │     ├── Analyze Valuation
    │     └── Analyze Risks
    │
    └── Stage
          Generate Report
          │
          ├── Build Investment Report
          └── Publish Report
```

Đây chỉ là Definition.

Tại thời điểm này:

- Chưa có Runtime.
- Chưa có Context.
- Chưa có State.
- Chưa có Execution.
- Chưa có Activity được thực thi.

Execution Layer sẽ sử dụng Definition này để tạo Runtime Hierarchy trong phần tiếp theo của walkthrough.

---

# Next Step

Ở phần tiếp theo, hệ thống sẽ nhận yêu cầu từ Client và bắt đầu khởi tạo Execution Runtime.

Chúng ta sẽ quan sát quá trình tạo:

- Execution Runtime
- Workflow Runtime
- Stage Runtime
- Task Runtime
- Activity Runtime
- Runtime Context

Đây là thời điểm Execution chính thức bắt đầu.

# Execution Initialization

Execution Initialization là giai đoạn đầu tiên của Runtime Execution.

Mục tiêu của giai đoạn này là chuyển một Client Request thành một Runtime Hierarchy có thể được Execution Engine thực thi.

Tại thời điểm này:

- Workflow Definition đã tồn tại.
- Chưa có Runtime Objects.
- Chưa có Activity được thực thi.
- Chưa có Business Logic được xử lý.

Execution Layer sẽ khởi tạo toàn bộ Runtime trước khi bắt đầu điều phối.

---

# Execution Initialization Flow

Quá trình khởi tạo được thực hiện theo thứ tự sau.

```
Client Request
        │
        ▼
Load Workflow Definition
        │
        ▼
Create Execution Runtime
        │
        ▼
Create Runtime Context
        │
        ▼
Create Workflow Runtime
        │
        ▼
Create Stage Runtime
        │
        ▼
Create Task Runtime
        │
        ▼
Create Activity Runtime
        │
        ▼
Runtime Ready
```

Sau bước này, Execution Engine đã có đầy đủ Runtime Objects để bắt đầu thực thi.

---

# Step 1 — Client Request

Execution bắt đầu khi Client gửi một yêu cầu.

Ví dụ.

```http
POST /analysis
```

Body.

```json
{
    "ticker": "FPT",
    "market": "VN"
}
```

Execution Layer nhận Request và xác định Workflow cần thực thi.

Ví dụ.

```
Workflow

Analyze Company
```

Tại thời điểm này, hệ thống mới chỉ biết:

- Workflow cần chạy.
- Input của Workflow.

Chưa có Runtime nào được tạo.

---

# Runtime Snapshot

```
Execution Runtime

Not Created
```

```
Workflow Runtime

Not Created
```

```
Stage Runtime

Not Created
```

```
Task Runtime

Not Created
```

```
Activity Runtime

Not Created
```

---

# Step 2 — Load Workflow Definition

Execution Layer tải Workflow Definition từ Definition Repository.

```
Workflow

Analyze Company
```

Definition được nạp vào bộ nhớ để phục vụ việc tạo Runtime.

Lưu ý.

Definition chỉ là blueprint.

Definition không thay đổi trong suốt vòng đời của Execution.

Một Workflow Definition có thể được nhiều Execution Runtime sử dụng đồng thời.

Ví dụ.

```
Execution A

↓

Analyze Company
```

```
Execution B

↓

Analyze Company
```

Hai Execution chia sẻ cùng một Definition nhưng có Runtime độc lập.

---

# Runtime Snapshot

```
Definition

Loaded
```

```
Runtime

Not Created
```

---

# Step 3 — Create Execution Runtime

Execution Runtime là Aggregate Root của toàn bộ Runtime Hierarchy.

Execution Layer tạo một Execution Runtime mới.

Ví dụ.

```yaml
Execution Runtime

id: exec-001

workflow: Analyze Company

state: Created
```

Execution Runtime chịu trách nhiệm quản lý:

- Workflow Runtime
- Stage Runtime
- Task Runtime
- Activity Runtime

Execution Runtime chưa bắt đầu thực thi.

---

# Runtime Snapshot

```yaml
Execution Runtime

id: exec-001

state: Created
```

---

# Step 4 — Create Execution Context

Execution Context được khởi tạo từ Client Request.

Ví dụ.

```yaml
Execution Context

ticker: FPT

market: VN
```

Execution Context là Context cấp cao nhất.

Toàn bộ Runtime Hierarchy có thể đọc dữ liệu từ Execution Context.

Các Runtime phía dưới sẽ kế thừa Context này trong quá trình thực thi.

---

# Context Snapshot

```yaml
Execution Context

ticker: FPT

market: VN
```

---

# Step 5 — Create Workflow Runtime

Execution Runtime tạo Workflow Runtime.

```yaml
Workflow Runtime

id: workflow-001

definition: Analyze Company

state: Created
```

Workflow Runtime đại diện cho một lần thực thi của Workflow Definition.

Workflow Runtime tham chiếu tới:

- Workflow Definition
- Workflow Context
- Workflow State

Workflow Runtime không sao chép Definition.

---

# Runtime Snapshot

```yaml
Execution Runtime
  state: Created

Workflow Runtime
  state: Created
```

---

# Step 6 — Create Stage Runtime

Execution Layer đọc Workflow Definition.

Workflow gồm ba Stage.

```
Load Company Data

Analyze Company

Generate Report
```

Từ Definition này, hệ thống tạo ba Stage Runtime.

```yaml
Stage Runtime

- stage-001
- stage-002
- stage-003
```

Ban đầu.

```yaml
state: Created
```

Tất cả Stage Runtime đã tồn tại nhưng chưa được thực thi.

---

# Runtime Snapshot

```yaml
Execution
  Created

Workflow
  Created

Stages

- Created
- Created
- Created
```

---

# Step 7 — Create Task Runtime

Execution Layer tiếp tục tạo Task Runtime.

Ví dụ.

```
Stage

Load Company Data
```

Bao gồm.

```
Task

Load Company Profile

Load Financial Statements
```

Execution Layer tạo.

```yaml
Task Runtime

task-001

task-002
```

Mỗi Task Runtime tham chiếu tới Task Definition tương ứng.

---

# Runtime Snapshot

```yaml
Execution

Created

Workflow

Created

Stages

3 Created

Tasks

7 Created
```

---

# Step 8 — Create Activity Runtime

Execution Layer tiếp tục tạo Activity Runtime.

Ví dụ.

Task.

```
Load Company Profile
```

Bao gồm.

```
Call Company API

Parse Response

Store Company Profile
```

Execution Runtime tạo ba Activity Runtime.

```yaml
Activity Runtime

activity-001

activity-002

activity-003
```

Tất cả Activity Runtime đều ở trạng thái.

```yaml
Created
```

Lưu ý.

Activity Runtime được tạo trước khi thực thi.

Việc tạo trước toàn bộ Runtime giúp:

- Quan sát toàn bộ Execution.
- Theo dõi tiến độ.
- Hỗ trợ Resume.
- Hỗ trợ Retry.
- Hỗ trợ Monitoring.

---

# Runtime Snapshot

```yaml
Execution
  Created

Workflow
  Created

Stages
  3 Created

Tasks
  7 Created

Activities
  18 Created
```

---

# Step 9 — Runtime Hierarchy Completed

Sau khi hoàn tất quá trình khởi tạo, Runtime Hierarchy có cấu trúc như sau.

```
Execution Runtime
│
└── Workflow Runtime
      │
      ├── Stage Runtime
      │      │
      │      ├── Task Runtime
      │      │      │
      │      │      ├── Activity Runtime
      │      │      ├── Activity Runtime
      │      │      └── Activity Runtime
      │      │
      │      └── Task Runtime
      │
      ├── Stage Runtime
      │
      └── Stage Runtime
```

Toàn bộ Runtime đã sẵn sàng.

Tuy nhiên.

- Chưa có Activity nào chạy.
- Chưa có State Transition.
- Chưa có Context Update.
- Chưa có Business Rule được đánh giá.

Execution Engine sẽ bắt đầu điều phối ở phần tiếp theo.

---

# Final Runtime Snapshot

```yaml
Execution Runtime

id: exec-001

state: Created

Workflow Runtime

state: Created

Stages

3 Created

Tasks

7 Created

Activities

18 Created

Execution Context

ticker: FPT

market: VN
```

---

# Summary

Sau giai đoạn Execution Initialization, hệ thống đã:

- Nhận yêu cầu từ Client.
- Tải Workflow Definition.
- Tạo Execution Runtime.
- Khởi tạo Execution Context.
- Tạo toàn bộ Runtime Hierarchy.
- Liên kết Runtime với Definition tương ứng.

Execution Layer đã hoàn tất việc chuẩn bị.

Ở phần tiếp theo, Execution Engine sẽ bắt đầu điều phối Runtime thông qua Orchestrator, State Machine và Activity Executor để thực hiện Workflow.


