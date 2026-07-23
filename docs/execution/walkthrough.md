# Execution Walkthrough

**Status:** Draft

**Version:** 3.0

---

# 1. Overview

## Purpose

Tài liệu này mô tả toàn bộ vòng đời của một **Execution**, từ khi hệ thống nhận yêu cầu thực thi đến khi Workflow hoàn thành.

Khác với các tài liệu kiến trúc chỉ mô tả cấu trúc của hệ thống, tài liệu này tập trung vào **hành vi của Execution Engine trong quá trình thực thi**.

Thông qua một ví dụ cụ thể, người đọc sẽ quan sát được:

- Cách Workflow Definition được chuyển thành Runtime Hierarchy.
- Cách Execution Runtime được khởi tạo.
- Cách Orchestrator liên tục đánh giá Runtime.
- Cách Execution Command được sinh ra.
- Cách State Machine và Activity Executor thực thi các Command.
- Cách Runtime State và Runtime Context thay đổi theo thời gian.
- Cách Execution hoàn thành thông qua Event Loop.

Mục tiêu của tài liệu là giúp người đọc hiểu **Execution Engine hoạt động như thế nào**, thay vì chỉ biết **Execution Engine được thiết kế như thế nào**.

---

## Scope

Walkthrough tập trung vào luồng thực thi (Happy Path) của một Workflow hoàn chỉnh.

Các chủ đề bao gồm:

- Workflow Initialization
- Runtime Hierarchy Creation
- Execution Event Loop
- Runtime State Transition
- Context Propagation
- Activity Execution
- Workflow Completion

Các chủ đề nâng cao như Retry, Parallel Execution, Compensation hoặc Human Approval sẽ được trình bày trong các tài liệu riêng.

---

## Audience

Tài liệu này dành cho:

- Backend Engineers
- Workflow Engine Developers
- AI Agent Developers
- Contributors của Execution Layer
- Người muốn hiểu cách Execution Engine hoạt động

---

## Reading Prerequisites

Để dễ theo dõi tài liệu này, nên đọc trước:

- README.md
- architecture.md
- execution-model.md
- orchestration-model.md
- execution-command.md
- state-machine.md

---

## Walkthrough Strategy

Toàn bộ quá trình thực thi sẽ được mô tả theo đúng thứ tự mà Execution Engine xử lý Runtime.

Mỗi bước sẽ bao gồm các nội dung sau:

- Runtime Snapshot
- Context Snapshot (nếu có thay đổi)
- Orchestrator Evaluation
- Generated Execution Command
- Command Execution
- Runtime Update

Điều này cho phép người đọc theo dõi toàn bộ Execution giống như đang quan sát một phiên debug của Workflow Engine.

---

# 2. Business Scenario

Để minh họa cách Execution Engine hoạt động, tài liệu sử dụng một quy trình phân tích doanh nghiệp đơn giản trong hệ thống Investment AI.

Người dùng gửi yêu cầu phân tích một công ty niêm yết trên thị trường chứng khoán Việt Nam.

Ví dụ:

```http
POST /analysis
```

Request

```yaml
ticker: FPT
market: VN
```

Sau khi nhận yêu cầu, hệ thống sẽ thực hiện toàn bộ quy trình phân tích doanh nghiệp và tạo ra một báo cáo đầu tư.

Expected Output

```yaml
reportId: report-001

ticker: FPT

status: Completed

generatedAt: 2026-07-23T10:15:30Z
```

---

## Business Process

Business Process được định nghĩa dưới dạng một Workflow.

Workflow bao gồm ba Stage chính.

```text
Analyze Company

├── Load Company Data
├── Analyze Company
└── Generate Report
```

Mỗi Stage được chia thành nhiều Task.

```text
Analyze Company

├── Load Company Data
│
│   ├── Load Company Profile
│   └── Load Financial Statements
│
├── Analyze Company
│
│   ├── Analyze Financial Health
│   ├── Analyze Valuation
│   └── Analyze Business Risks
│
└── Generate Report
    │
    └── Build Investment Report
```

Mỗi Task tiếp tục được chia thành các Activity nhỏ hơn.

Ví dụ:

```text
Load Company Profile

├── Call Company API
├── Parse Response
└── Store Company Profile
```

Hoặc:

```text
Analyze Valuation

├── Calculate PE
├── Calculate PB
├── Compare Industry Average
└── Generate Valuation Summary
```

Đây là các đơn vị thực thi nhỏ nhất của Workflow.

---

## Workflow Definition

Điều quan trọng cần lưu ý là ở thời điểm này **chưa có bất kỳ Runtime nào được tạo ra**.

Workflow hiện chỉ tồn tại dưới dạng Definition.

```text
Workflow Definition
        │
        ├── Stage Definitions
        │
        ├── Task Definitions
        │
        └── Activity Definitions
```

Definition chỉ mô tả:

- Cấu trúc Workflow.
- Thứ tự thực hiện.
- Quan hệ cha - con giữa các thành phần.
- Metadata cần thiết cho việc thực thi.

Definition không chứa:

- Runtime State
- Runtime Context
- Execution Result
- Execution Progress

Những thông tin này chỉ xuất hiện sau khi hệ thống bắt đầu khởi tạo một Execution.

---

## Execution Goal

Mục tiêu của Execution lần này là:

1. Tạo một Execution Runtime từ Workflow Definition.
2. Khởi tạo toàn bộ Runtime Hierarchy.
3. Thực thi từng Activity theo đúng Workflow.
4. Thu thập dữ liệu trong quá trình thực thi.
5. Tổng hợp kết quả phân tích.
6. Sinh báo cáo đầu tư hoàn chỉnh.
7. Đánh dấu Execution ở trạng thái **Completed**.

Các phần tiếp theo của tài liệu sẽ mô tả chi tiết từng bước mà Execution Engine thực hiện để đạt được mục tiêu này.

## 3. Workflow Definition

Sau khi nhận được yêu cầu phân tích doanh nghiệp, Execution Engine **không tạo Runtime ngay lập tức**.

Bước đầu tiên là tải **Workflow Definition** đã được thiết kế trước.

Workflow Definition mô tả **Business Process** cần thực hiện, nhưng **không đại diện cho một lần thực thi cụ thể**.

Definition có thể được sử dụng lại bởi nhiều Execution khác nhau.

Ví dụ:

```text
Workflow Definition

Analyze Company
```

Workflow này bao gồm ba Stage.

```text
Analyze Company

├── Stage 1 : Load Company Data
├── Stage 2 : Analyze Company
└── Stage 3 : Generate Report
```

---

### Stage Definitions

Mỗi Stage mô tả một nhóm nghiệp vụ độc lập.

```text
Load Company Data

├── Load Company Profile
└── Load Financial Statements
```

```text
Analyze Company

├── Analyze Financial Health
├── Analyze Valuation
└── Analyze Business Risks
```

```text
Generate Report

└── Build Investment Report
```

Stage chỉ mô tả cấu trúc.

Stage không chứa trạng thái thực thi.

---

### Task Definitions

Mỗi Stage bao gồm một hoặc nhiều Task.

Ví dụ.

```text
Load Company Profile
```

Task này có thể được mô tả như sau.

```yaml
id: task-load-profile

name: Load Company Profile
```

Task Definition xác định:

- tên Task
- các Activity cần thực hiện
- quan hệ với Stage
- metadata phục vụ thực thi

Task vẫn chưa được thực thi.

---

### Activity Definitions

Activity là đơn vị thực thi nhỏ nhất.

Ví dụ.

```text
Load Company Profile

├── Call Company API
├── Parse Response
└── Store Company Profile
```

Hoặc.

```text
Analyze Valuation

├── Calculate PE
├── Calculate PB
├── Compare Industry Average
└── Generate Valuation Summary
```

Mỗi Activity sẽ được Activity Executor thực thi trong Runtime.

Definition chỉ mô tả Activity.

Không chứa:

- trạng thái
- context
- kết quả

---

### Definition Hierarchy

Toàn bộ Definition có thể được biểu diễn như sau.

```text
Workflow Definition

└── Stage Definition

      └── Task Definition

             └── Activity Definition
```

Đây chỉ là **template** của Business Process.

Definition hoàn toàn bất biến (Immutable) trong suốt quá trình thực thi.

Nhiều Execution có thể cùng tham chiếu đến một Workflow Definition mà không ảnh hưởng lẫn nhau.

---

### Definition Summary

Đến thời điểm này, hệ thống mới chỉ có:

- Workflow Definition
- Stage Definitions
- Task Definitions
- Activity Definitions

Chưa tồn tại:

- Execution Runtime
- Runtime State
- Runtime Context
- Execution Result

Những thành phần này sẽ được tạo ở bước tiếp theo.

---

## 4. Execution Initialization

Sau khi Workflow Definition được tải thành công, Execution Engine bắt đầu khởi tạo một phiên thực thi mới.

Mỗi yêu cầu thực thi sẽ tạo ra **một Execution Runtime độc lập**.

Execution Runtime đại diện cho toàn bộ vòng đời của Workflow trong lần thực thi này.

---

### Step 1. Create Execution Runtime

Execution Engine tạo Aggregate Root của Runtime Hierarchy.

```text
Execution Runtime
```

Ví dụ.

```yaml
id: exec-001

definitionId: workflow-analyze-company

status: Created

createdAt: 2026-07-23T09:00:00Z
```

Execution Runtime là điểm truy cập duy nhất tới toàn bộ Runtime Hierarchy.

---

### Step 2. Create Runtime Hierarchy

Từ Workflow Definition, Execution Engine tạo toàn bộ Runtime Objects.

```text
Execution Runtime

└── Workflow Runtime

      ├── Stage Runtime

      │      ├── Task Runtime

      │      │      ├── Activity Runtime

      │      │      └── Activity Runtime

      │      │
      │      └── Task Runtime

      │
      ├── Stage Runtime
      │
      └── Stage Runtime
```

Mỗi Runtime giữ tham chiếu đến Definition tương ứng.

Ví dụ.

```text
Task Runtime

↓

Task Definition
```

Runtime không sao chép Definition.

---

### Step 3. Initialize Runtime State

Sau khi Runtime Hierarchy được tạo, mọi Runtime đều bắt đầu với trạng thái **Created**.

```text
Execution Runtime    Created

Workflow Runtime     Created

Stage Runtime        Created

Task Runtime         Created

Activity Runtime     Created
```

Chưa có Runtime nào được phép thực thi.

---

### Step 4. Create Execution Context

Tiếp theo, Execution Engine khởi tạo Execution Context.

Context ban đầu chứa dữ liệu đầu vào của người dùng.

```yaml
ticker: FPT

market: VN
```

Execution Context sẽ được truyền xuống các Runtime con trong quá trình thực thi.

Tại thời điểm này chưa có dữ liệu phân tích.

---

### Step 5. Bind Runtime References

Mỗi Runtime được liên kết với Definition tương ứng.

Ví dụ.

```text
Workflow Runtime

↓

Workflow Definition
```

```text
Stage Runtime

↓

Stage Definition
```

```text
Task Runtime

↓

Task Definition
```

```text
Activity Runtime

↓

Activity Definition
```

Việc sử dụng Reference thay vì sao chép Definition giúp nhiều Execution có thể chia sẻ cùng một Definition.

---

### Step 6. Runtime Initialization Completed

Sau khi hoàn thành khởi tạo, hệ thống đã có đầy đủ các thành phần cần thiết để bắt đầu thực thi.

Runtime Snapshot.

```text
Execution Runtime

Status : Created

└── Workflow Runtime

      Status : Created

      ├── Stage Runtime

      │      Status : Created

      │      ├── Task Runtime

      │      │      Status : Created

      │      │      ├── Activity Runtime

      │      │      │      Status : Created

      │      │      └── Activity Runtime

      │      │             Status : Created

      │      └── Task Runtime

      │             Status : Created

      ├── Stage Runtime

      │      Status : Created

      └── Stage Runtime

             Status : Created
```

Context Snapshot.

```yaml
ticker: FPT

market: VN
```

---

### Initialization Result

Đến thời điểm này:

- Workflow Definition đã được tải.
- Execution Runtime đã được tạo.
- Runtime Hierarchy đã hoàn chỉnh.
- Runtime State đã được khởi tạo.
- Execution Context đã sẵn sàng.
- Runtime đã liên kết với Definition.

Tuy nhiên, **chưa có Activity nào được thực thi**.

Execution Engine chỉ mới chuẩn bị môi trường thực thi.

Ở phần tiếp theo, **Orchestrator sẽ bắt đầu Event Loop đầu tiên**, đánh giá Runtime Hierarchy và sinh ra Execution Command đầu tiên để khởi động Workflow.

## 5. Execution Event Loop

Sau khi Execution Runtime được khởi tạo hoàn tất, Execution Engine bắt đầu vòng lặp thực thi (Execution Event Loop).

Event Loop là cơ chế trung tâm của Execution Layer.

Trong mỗi vòng lặp, Orchestrator sẽ:

1. Đánh giá Runtime Hierarchy.
2. Lựa chọn Runtime có thể thực thi.
3. Sinh Execution Command.
4. Chuyển Command cho thành phần thực thi phù hợp.
5. Cập nhật Runtime.
6. Bắt đầu vòng lặp tiếp theo.

Toàn bộ quá trình có thể được mô tả như sau.

```text
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

Execution chỉ kết thúc khi không còn Runtime nào có thể thực thi.

---

### Event Loop Iteration 1

#### Runtime Snapshot

```text
Execution Runtime    Created

Workflow Runtime     Created
```

#### Orchestrator Evaluation

Orchestrator bắt đầu từ Aggregate Root và duyệt Runtime Hierarchy.

Workflow Runtime đang ở trạng thái **Created** và đủ điều kiện để được kích hoạt.

#### Generated Execution Command

```yaml
type: TransitionState

runtime: workflow-runtime

from: Created

to: Ready
```

#### Command Execution

State Machine nhận Command và kiểm tra tính hợp lệ của State Transition.

Transition hợp lệ nên Workflow Runtime được chuyển sang trạng thái **Ready**.

#### Runtime Snapshot

```text
Workflow Runtime

Ready
```

---

### Event Loop Iteration 2

#### Runtime Snapshot

```text
Workflow Runtime

Ready
```

#### Orchestrator Evaluation

Workflow đã sẵn sàng để bắt đầu thực thi.

#### Generated Execution Command

```yaml
type: TransitionState

runtime: workflow-runtime

from: Ready

to: Running
```

#### Command Execution

State Machine chuyển Workflow sang trạng thái **Running**.

#### Runtime Snapshot

```text
Workflow Runtime

Running
```

---

### Event Loop Iteration 3

#### Runtime Snapshot

```text
Stage Runtime

Created
```

#### Orchestrator Evaluation

Workflow đã chạy.

Stage đầu tiên có thể được kích hoạt.

#### Generated Execution Command

```yaml
type: TransitionState

runtime: stage-runtime

from: Created

to: Ready
```

#### Runtime Snapshot

```text
Stage Runtime

Ready
```

---

### Event Loop Iteration 4

Stage Runtime chuyển từ **Ready** sang **Running**.

```yaml
type: TransitionState

runtime: stage-runtime

from: Ready

to: Running
```

---

### Event Loop Iteration 5

Task Runtime đầu tiên được kích hoạt.

```yaml
type: TransitionState

runtime: task-runtime

from: Created

to: Ready
```

---

### Event Loop Iteration 6

Task Runtime bắt đầu thực thi.

```yaml
type: TransitionState

runtime: task-runtime

from: Ready

to: Running
```

---

### Event Loop Iteration 7

Activity Runtime đầu tiên được kích hoạt.

```yaml
type: TransitionState

runtime: activity-runtime

from: Created

to: Ready
```

---

### Event Loop Iteration 8

Activity Runtime bắt đầu thực thi.

```yaml
type: TransitionState

runtime: activity-runtime

from: Ready

to: Running
```

---

### Event Loop Iteration 9

#### Runtime Snapshot

```text
Activity Runtime

Running
```

#### Orchestrator Evaluation

Activity đang ở trạng thái Running và sẵn sàng được thực thi.

#### Generated Execution Command

```yaml
type: ExecuteActivity

runtime: activity-runtime
```

#### Command Execution

Activity Executor nhận Command.

Activity Executor thực hiện các bước:

1. Resolve Activity Handler.
2. Chuẩn bị Input từ Context.
3. Gọi Handler.
4. Thu thập Output.
5. Cập nhật Context.

Ví dụ Context sau khi thực thi.

```yaml
companyProfile:
  ticker: FPT
  companyName: FPT Corporation
```

Sau khi Activity hoàn thành, State Machine cập nhật trạng thái.

```text
Activity Runtime

Completed
```

---

### Event Loop Iteration 10

Orchestrator đánh giá lại Runtime Hierarchy.

Activity tiếp theo được chọn và quy trình tương tự lặp lại.

```text
Transition Activity → Ready

↓

Transition Activity → Running

↓

Execute Activity

↓

Activity Completed
```

---

### Event Loop Continues

Execution Engine tiếp tục lặp lại quy trình trên cho đến khi toàn bộ Activity của Task hoàn thành.

Sau đó:

```text
Task Runtime

Completed
```

Tiếp theo:

```text
Stage Runtime

Completed
```

Cuối cùng:

```text
Workflow Runtime

Completed
```

và sau khi tất cả Workflow Runtime hoàn thành:

```text
Execution Runtime

Completed
```

---

### Event Loop Termination

Sau mỗi Runtime Update, Orchestrator luôn bắt đầu một vòng đánh giá mới.

Khi không còn Runtime nào có thể thực thi:

```text
Evaluate Runtime Hierarchy

↓

No Executable Runtime

↓

Execution Completed
```

Đây là điều kiện dừng duy nhất của Execution Event Loop.

## 6. Execution Completion

Sau nhiều vòng lặp của Execution Event Loop, toàn bộ Activity trong Workflow đã được thực thi thành công.

Khi Activity cuối cùng hoàn thành, Execution Engine tiếp tục chạy một vòng đánh giá Runtime mới.

Lần đánh giá này, Orchestrator nhận thấy:

- Không còn Activity nào ở trạng thái `Created` hoặc `Ready`.
- Không còn Task nào đang thực thi.
- Tất cả Task trong Stage hiện tại đều đã hoàn thành.
- Tất cả Stage trong Workflow đều đã hoàn thành.
- Workflow đã đạt điều kiện kết thúc.

Từ đó, Orchestrator lần lượt sinh các Execution Command để hoàn tất các Runtime còn lại.

Quá trình kết thúc diễn ra theo đúng Runtime Hierarchy.

```text
Activity Runtime

Completed

↓

Task Runtime

Completed

↓

Stage Runtime

Completed

↓

Workflow Runtime

Completed

↓

Execution Runtime

Completed
```

Điều quan trọng là **Execution Engine không "nhảy" trực tiếp từ Activity lên Execution**.

Mỗi Runtime chỉ được đánh dấu hoàn thành khi tất cả Runtime con trực tiếp của nó đã hoàn thành.

Điều này đảm bảo Runtime Hierarchy luôn nhất quán trong suốt vòng đời của Execution.

---

### Final Event Loop

Sau khi Execution Runtime chuyển sang trạng thái `Completed`, Orchestrator tiếp tục thực hiện thêm một vòng đánh giá cuối cùng.

```text
Evaluate Runtime Hierarchy

↓

No Executable Runtime

↓

Execution Completed

↓

Exit Event Loop
```

Đây là điều kiện dừng duy nhất của Execution Engine.

Execution không kết thúc vì một Runtime được đánh dấu `Completed`.

Execution chỉ kết thúc khi **không còn Runtime nào có thể thực thi hoặc chuyển trạng thái**.

---

### Execution Result

Sau khi Execution hoàn thành, hệ thống tổng hợp kết quả cuối cùng.

Ví dụ.

```yaml
executionId: exec-001

status: Completed

startedAt: 2026-07-23T09:00:00Z

completedAt: 2026-07-23T09:00:18Z

duration: 18s

result:

  reportId: report-001
```

Execution Result đại diện cho kết quả của toàn bộ Business Process.

Các Runtime trung gian vẫn được giữ lại để phục vụ:

- Audit
- Monitoring
- Debugging
- Retry (nếu được hỗ trợ)
- Execution History

---

## 7. Final Runtime Snapshot

Sau khi Execution hoàn thành, toàn bộ Runtime Hierarchy có trạng thái như sau.

```text
Execution Runtime
Status : Completed

└── Workflow Runtime
    Status : Completed

    ├── Stage Runtime
    │   Status : Completed
    │
    │   ├── Task Runtime
    │   │   Status : Completed
    │   │
    │   │   ├── Activity Runtime
    │   │   │   Status : Completed
    │   │   │
    │   │   ├── Activity Runtime
    │   │   │   Status : Completed
    │   │   │
    │   │   └── Activity Runtime
    │   │       Status : Completed
    │   │
    │   └── Task Runtime
    │       Status : Completed
    │
    ├── Stage Runtime
    │   Status : Completed
    │
    └── Stage Runtime
        Status : Completed
```

Toàn bộ Runtime đều đã hoàn thành và không còn Runtime nào ở trạng thái:

- Created
- Ready
- Running

---

### Final Context Snapshot

Execution Context lúc này đã được bổ sung đầy đủ dữ liệu được tạo ra trong suốt quá trình thực thi.

Ví dụ.

```yaml
input:

  ticker: FPT

  market: VN

companyProfile:

  ...

financialStatements:

  ...

financialAnalysis:

  ...

valuation:

  ...

riskAnalysis:

  ...

investmentReport:

  reportId: report-001
```

Execution Context đại diện cho toàn bộ dữ liệu được tích lũy trong suốt vòng đời của Execution.

---

### Runtime Timeline

Có thể hình dung toàn bộ vòng đời của Execution như sau.

```text
Created

↓

Initialized

↓

Running

↓

Completed
```

Trong giai đoạn `Running`, Execution Engine liên tục thực hiện:

```text
Evaluate Runtime

↓

Generate Execution Command

↓

Execute Command

↓

Update Runtime

↓

Repeat
```

Đây chính là cơ chế Event Loop của Execution Layer.

---

## 8. Key Takeaways

Qua walkthrough này, chúng ta có thể rút ra một số nguyên tắc quan trọng của Execution Layer.

### Definition và Runtime được tách biệt

Workflow Definition chỉ mô tả Business Process.

Execution Runtime đại diện cho một lần thực thi cụ thể của Workflow đó.

Một Workflow Definition có thể được sử dụng bởi nhiều Execution khác nhau.

---

### Execution Runtime là Aggregate Root

Execution Runtime là điểm truy cập duy nhất tới Runtime Hierarchy.

Mọi Runtime đều thuộc về một Execution Runtime.

---

### Orchestrator chỉ đưa ra quyết định

Orchestrator không trực tiếp:

- thay đổi Runtime State;
- thực thi Activity;
- cập nhật Context.

Nhiệm vụ duy nhất của Orchestrator là đánh giá Runtime Hierarchy và sinh Execution Command phù hợp.

---

### Execution Command là cầu nối giữa Decision và Execution

Execution Command biểu diễn một quyết định của Orchestrator.

Các Command được chuyển cho các thành phần chuyên trách như:

- State Machine
- Activity Executor

để thực hiện hành động tương ứng.

---

### State chỉ được thay đổi bởi State Machine

Mọi thay đổi Runtime State đều phải đi qua State Machine.

Điều này đảm bảo:

- State Transition nhất quán.
- Runtime Lifecycle được kiểm soát.
- Không có thành phần nào tự ý thay đổi trạng thái Runtime.

---

### Activity chỉ được thực thi bởi Activity Executor

Activity Executor chịu trách nhiệm:

- Resolve Handler.
- Chuẩn bị Input.
- Thực thi Activity.
- Thu thập Output.
- Cập nhật Context.

Business Logic không nằm trong Execution Engine mà nằm trong các Capability hoặc Handler được Activity gọi tới.

---

### Context được tích lũy trong suốt Execution

Execution Context bắt đầu từ dữ liệu đầu vào của người dùng.

Qua mỗi Activity, Context được mở rộng bằng các kết quả trung gian và cuối cùng trở thành dữ liệu đầu ra của toàn bộ Workflow.

---

### Execution Engine hoạt động theo Event Loop

Execution Engine không thực thi Workflow theo cách đệ quy (recursive) hay theo một luồng tuyến tính cố định.

Thay vào đó, hệ thống liên tục:

1. Đánh giá Runtime Hierarchy.
2. Đưa ra quyết định.
3. Sinh Execution Command.
4. Thực thi Command.
5. Cập nhật Runtime.
6. Lặp lại cho đến khi không còn Runtime nào có thể thực thi.

Mô hình này giúp Execution Layer dễ dàng mở rộng để hỗ trợ các khả năng như:

- Parallel Execution
- Retry
- Timeout
- Compensation
- Human Approval
- Distributed Execution

mà không cần thay đổi kiến trúc cốt lõi của hệ thống.

---

### Tổng kết

Thông qua walkthrough này, chúng ta đã theo dõi toàn bộ vòng đời của một Execution:

1. Tải Workflow Definition.
2. Khởi tạo Execution Runtime.
3. Xây dựng Runtime Hierarchy.
4. Khởi tạo Context.
5. Thực thi Workflow thông qua Execution Event Loop.
6. Hoàn thành toàn bộ Runtime.
7. Sinh Execution Result.

Đây là luồng thực thi chuẩn (Happy Path) của Execution Layer và là nền tảng cho các kịch bản nâng cao như xử lý lỗi, Retry, Parallel Execution và Compensation sẽ được trình bày trong các tài liệu chuyên biệt.