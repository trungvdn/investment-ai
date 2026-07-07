---
title: Go Runtime Scheduler
volume: II - Go Runtime
chapter: 10
author: Nhat Trung Vo
reviewer: ChatGPT
version: 1.0
go_version: "1.25"

difficulty: ⭐⭐⭐⭐☆

estimated_reading_time: 90 minutes

prerequisites:
  - Goroutine
  - Stack & Heap
  - Operating System Basics
  - Context Switch

source_files:
  - runtime/proc.go
  - runtime/runtime2.go
  - runtime/preempt.go
  - runtime/lock_futex.go

last_updated: 2026-07-07
---

# Chapter 10 - Go Scheduler

> "Concurrency is the ability to deal with many things at once. Parallelism is the ability to do many things at once."
>
> **— Rob Pike**

---

# Mục tiêu

Sau khi hoàn thành chương này, bạn sẽ có thể:

- Hiểu tại sao Go cần một Runtime Scheduler.
- Phân biệt được Goroutine, OS Thread và Processor.
- Hiểu mô hình GMP (G-M-P).
- Hiểu cách Go lập lịch hàng triệu Goroutine.
- Hiểu cơ chế Work Stealing.
- Hiểu Scheduler hoạt động khi gặp Blocking System Call.
- Hiểu Asynchronous Preemption.
- Đọc và hiểu source code của `runtime/proc.go`.
- Trả lời hầu hết các câu hỏi phỏng vấn liên quan đến Go Scheduler.

---

# Mục lục

- 10.1 Giới thiệu
- 10.2 Tại sao Go cần Scheduler?
- 10.3 Sự tiến hóa của cơ chế lập lịch
- 10.4 Mục tiêu thiết kế của Go Scheduler
- 10.5 Tổng quan kiến trúc Scheduler
- 10.6 Mô hình GMP
- 10.7 Goroutine (G)
- 10.8 Machine (M)
- 10.9 Processor (P)
- 10.10 Tại sao cần Processor?
- 10.11 Local Run Queue
- 10.12 Global Run Queue
- 10.13 Tạo Goroutine
- 10.14 Scheduler Loop
- 10.15 Work Stealing
- 10.16 Blocking Syscall
- 10.17 Netpoller
- 10.18 Asynchronous Preemption
- 10.19 Phân tích Source Code
- 10.20 Performance
- 10.21 Best Practices
- 10.22 Câu hỏi phỏng vấn
- 10.23 Tổng kết

---

# 10.1 Giới thiệu

Nếu phải chọn một thành phần quan trọng nhất của Go Runtime, thì Scheduler gần như luôn là câu trả lời.

Không phải Garbage Collector.

Không phải Memory Allocator.

Mà chính là **Scheduler**.

Scheduler là trái tim của Go Runtime. Nó chịu trách nhiệm quyết định:

- Goroutine nào được chạy.
- Chạy trên OS Thread nào.
- Chạy khi nào.
- Chạy trong bao lâu.
- Khi nào phải tạm dừng.
- Khi nào tiếp tục thực thi.

Nói cách khác, mọi goroutine trong chương trình Go đều phải đi qua Scheduler.

Nếu Scheduler hoạt động không hiệu quả:

- CPU sẽ không được tận dụng hết.
- Goroutine sẽ phải chờ đợi.
- Throughput giảm.
- Latency tăng.
- Hệ thống khó mở rộng.

Ngược lại, một Scheduler tốt có thể giúp chương trình xử lý hàng triệu goroutine với chi phí rất thấp.

Đó chính là điều Go Runtime đã làm được.

---

# 10.2 Tại sao Go cần Scheduler?

Để hiểu lý do Go xây dựng Scheduler riêng, trước tiên hãy xem cách các ngôn ngữ truyền thống hoạt động.

Giả sử chúng ta xây dựng một HTTP Server.

Mỗi request sẽ được xử lý bởi một Thread.

```
HTTP Request

↓

OS Thread

↓

CPU
```

Nếu server nhận:

```
100 Requests
```

Hệ điều hành sẽ tạo:

```
100 Threads
```

Điều này vẫn hoàn toàn chấp nhận được.

Nhưng nếu hệ thống nhận:

```
1.000.000 Requests
```

Khi đó:

```
1.000.000 Threads
```

là điều gần như không thể.

Lý do nằm ở chi phí của Operating System Thread.

Mỗi Thread đều cần:

- Kernel Object
- Kernel Stack
- User Stack
- Scheduler Metadata
- Register Context
- Thread Local Storage

Ngoài ra, việc chuyển đổi giữa các Thread (Context Switching) cũng rất tốn kém.

```
Running Thread A

↓

Save Registers

↓

Kernel Mode

↓

Restore Registers

↓

Thread B

↓

User Mode
```

Việc chuyển đổi này diễn ra hàng nghìn hoặc hàng triệu lần mỗi giây.

CPU sẽ dành rất nhiều thời gian để chuyển đổi ngữ cảnh thay vì thực hiện công việc thực sự.

Đây chính là giới hạn của mô hình **One Thread Per Request**.

---

# 10.3 Giải pháp của Go

Go lựa chọn một hướng đi hoàn toàn khác.

Thay vì:

```
1 Request

↓

1 Thread
```

Go sử dụng:

```
1 Request

↓

1 Goroutine
```

Sau đó Runtime sẽ ánh xạ hàng triệu Goroutine xuống một số lượng rất nhỏ OS Thread.

```
          1,000,000 Goroutines

                    │
                    ▼

         Go Runtime Scheduler

                    │
                    ▼

             8 OS Threads

                    │
                    ▼

                8 CPU Cores
```

Điều này mang lại ba lợi ích cực kỳ lớn.

## 1. Tiết kiệm bộ nhớ

Một OS Thread thường có stack mặc định khá lớn (tùy nền tảng).

Trong khi đó một Goroutine chỉ khởi tạo với khoảng **2 KB stack** và có thể tự động mở rộng hoặc thu nhỏ khi cần.

Nhờ đó, chương trình có thể tạo hàng trăm nghìn hoặc hàng triệu goroutine mà không tiêu tốn lượng RAM khổng lồ.

---

## 2. Context Switching rẻ hơn

Context Switch giữa hai Thread yêu cầu hệ điều hành can thiệp.

Trong khi đó, chuyển đổi giữa hai Goroutine được thực hiện hoàn toàn trong User Space bởi Go Runtime.

Không cần:

- Trap vào Kernel.
- Chuyển User Mode ↔ Kernel Mode.
- Kernel Scheduler.

Điều này giúp Scheduler của Go có chi phí chuyển đổi thấp hơn đáng kể.

---

## 3. Runtime có toàn quyền quyết định

Thay vì giao việc lập lịch cho hệ điều hành, Go Runtime tự quyết định:

- Goroutine nào được ưu tiên.
- Khi nào phải nhường CPU.
- Khi nào phải đánh thức Goroutine.
- Khi nào cần tạo thêm OS Thread.
- Khi nào nên tái sử dụng Thread.

Nhờ đó Runtime có thể tối ưu theo đặc thù của ngôn ngữ Go thay vì phụ thuộc vào thuật toán lập lịch chung của hệ điều hành.

---

# 10.4 Sự tiến hóa của cơ chế lập lịch

Có thể chia lịch sử lập lịch thành ba giai đoạn.

## Giai đoạn 1 — One Thread Per Task

```
Task A

↓

Thread A

↓

CPU


Task B

↓

Thread B

↓

CPU
```

Đây là mô hình đơn giản nhất.

Ưu điểm:

- Dễ lập trình.
- Tận dụng Scheduler của hệ điều hành.

Nhược điểm:

- Thread rất nặng.
- Khó mở rộng.
- Context Switch đắt đỏ.

---

## Giai đoạn 2 — Thread Pool

Để giảm chi phí tạo Thread, nhiều hệ thống sử dụng Thread Pool.

```
Tasks

↓

Queue

↓

Thread Pool

↓

CPU
```

Thread được tái sử dụng.

Chi phí tạo Thread giảm đáng kể.

Tuy nhiên, Thread vẫn là tài nguyên của hệ điều hành nên số lượng vẫn bị giới hạn.

---

## Giai đoạn 3 — User-space Scheduler

Đây là mô hình hiện đại nhất.

```
Tasks

↓

Runtime Scheduler

↓

OS Threads

↓

CPU
```

Runtime chịu trách nhiệm lập lịch.

Operating System chỉ quản lý một số lượng nhỏ Thread.

Go, Erlang BEAM và Java Virtual Threads đều đi theo hướng này.

---

# 10.5 Mục tiêu thiết kế của Go Scheduler

Go Scheduler không chỉ nhằm mục tiêu chạy được nhiều Goroutine.

Nó còn phải đáp ứng đồng thời nhiều yêu cầu khác nhau.

## Lightweight Concurrency

Việc tạo một Goroutine phải cực kỳ rẻ.

Một chương trình Go có thể tạo hàng triệu Goroutine mà không làm cạn kiệt tài nguyên hệ thống.

---

## Khả năng mở rộng

Scheduler phải hoạt động hiệu quả trên:

- 2 CPU
- 8 CPU
- 64 CPU
- Hàng trăm CPU

mà không cần thay đổi mã nguồn.

---

## Tận dụng tối đa CPU

Nếu vẫn còn Goroutine sẵn sàng chạy thì CPU không nên ở trạng thái nhàn rỗi.

Scheduler luôn cố gắng giữ cho tất cả Processor đều có việc để thực hiện.

---

## Chi phí lập lịch thấp

Scheduler bản thân nó không được trở thành nút thắt hiệu năng.

Một Scheduler tốt là Scheduler "gần như vô hình", tiêu tốn rất ít CPU so với khối lượng công việc mà chương trình thực hiện.

---

## Cân bằng tải tự động

Trong thực tế, khối lượng công việc không bao giờ phân bố đều.

Có Processor rất bận.

Có Processor lại rảnh.

Go Runtime tự động cân bằng tải thông qua cơ chế **Work Stealing**, giúp các Processor nhàn rỗi lấy bớt công việc từ Processor đang quá tải.

Đây là một trong những kỹ thuật quan trọng nhất của Go Scheduler và sẽ được phân tích chi tiết ở phần sau của chương này.

---

---

# 10.6 Tổng quan kiến trúc Scheduler

Sau khi hiểu vì sao Go cần một Scheduler riêng, câu hỏi tiếp theo là:

> **Go Runtime Scheduler được xây dựng như thế nào?**

Trước Go 1.1, Scheduler của Go tương đối đơn giản và chưa tối ưu cho hệ thống đa nhân (multi-core).

Kể từ Go 1.1, nhóm phát triển Go đã giới thiệu mô hình **GMP (Goroutine - Machine - Processor)**.

Đây là một trong những quyết định thiết kế quan trọng nhất của Go Runtime.

Cho đến hiện nay (Go 1.25), GMP vẫn là nền tảng của toàn bộ Scheduler.

Kiến trúc tổng quát có thể hình dung như sau.

```text
                  +----------------------+
                  |     Goroutines       |
                  | G1 G2 G3 G4 G5 ...   |
                  +----------+-----------+
                             |
                             |
                     Runtime Scheduler
                             |
          +------------------+------------------+
          |                                     |
          ▼                                     ▼
      Processor P0                         Processor P1
      Local Queue                         Local Queue
          |                                     |
          ▼                                     ▼
      Machine M0                           Machine M1
     (OS Thread)                          (OS Thread)
          |                                     |
          +------------------+------------------+
                             |
                             ▼
                           CPU Cores
```

Có ba thành phần trung tâm:

- **G (Goroutine)**
- **M (Machine)**
- **P (Processor)**

Trong đó:

- Goroutine đại diện cho công việc cần thực hiện.
- Machine đại diện cho OS Thread.
- Processor đại diện cho tài nguyên Scheduler.

Ba thành phần này phối hợp với nhau để tạo nên khả năng xử lý đồng thời cực kỳ hiệu quả của Go.

---

# 10.7 Mô hình GMP

Mô hình GMP là trái tim của Go Runtime.

Rất nhiều kỹ sư Go có thể sử dụng Goroutine thành thạo nhưng lại chưa thực sự hiểu GMP hoạt động như thế nào.

Trong chương này chúng ta sẽ đi từ:

```
G

↓

M

↓

P

↓

Scheduler Loop

↓

Work Stealing

↓

Preemption
```

Sau khi hiểu GMP, bạn sẽ dễ dàng đọc source code trong `runtime/proc.go`.

---

# 10.8 Goroutine (G)

## Goroutine là gì?

Goroutine là đơn vị thực thi (execution unit) của Go Runtime.

Khi chúng ta viết:

```go
go processOrder(orderID)
```

Runtime sẽ tạo ra một đối tượng kiểu **G**.

Điều quan trọng cần nhớ là:

> **Scheduler không chạy Function. Scheduler chạy Goroutine (G).**

Function chỉ là đoạn mã.

G mới là đối tượng mà Runtime quản lý.

---

## Goroutine không chỉ là Stack

Nhiều người mới học Go nghĩ rằng Goroutine chỉ đơn giản là một Stack.

Điều này không đúng.

Một Goroutine còn chứa rất nhiều metadata phục vụ Scheduler.

Có thể hình dung như sau:

```text
                Goroutine (G)

        +-------------------------+
        | Stack                   |
        +-------------------------+
        | Program Counter (PC)    |
        +-------------------------+
        | Stack Pointer (SP)      |
        +-------------------------+
        | Status                  |
        +-------------------------+
        | Scheduler Metadata      |
        +-------------------------+
        | Panic / Defer Info      |
        +-------------------------+
        | GC Information          |
        +-------------------------+
```

Scheduler sử dụng các thông tin này để:

- Tạm dừng Goroutine.
- Tiếp tục Goroutine.
- Di chuyển Goroutine giữa các Processor.
- Thu hồi Goroutine khi hoàn thành.

---

## Cấu trúc G trong Runtime

Trong source code Go, Goroutine được định nghĩa bởi struct `g`.

```go
type g struct {
    stack stack
    sched gobuf
    m *m
    status uint32
    ...
}
```

Đây chỉ là một phần nhỏ.

Trong thực tế, struct `g` có hàng chục trường dữ liệu.

Một số trường quan trọng:

| Trường | Ý nghĩa |
|---------|----------|
| stack | Stack của Goroutine |
| sched | Thông tin context |
| m | Machine đang chạy Goroutine |
| atomicstatus | Trạng thái hiện tại |
| goid | ID của Goroutine |
| timer | Timer liên quan |
| labels | Metadata phục vụ tracing |

Trong các phần sau chúng ta sẽ gặp lại các trường này khi đọc source code.

---

# 10.9 Vòng đời của Goroutine

Một Goroutine không tồn tại mãi mãi.

Nó trải qua nhiều trạng thái khác nhau.

```
Create

↓

Runnable

↓

Running

↓

Waiting

↓

Runnable

↓

Running

↓

Dead
```

Chi tiết từng trạng thái sẽ được phân tích trong Chapter 11.

Trong chương Scheduler chỉ cần nhớ:

**Scheduler luôn tìm Goroutine ở trạng thái Runnable để thực thi.**

---

# 10.10 Machine (M)

Nếu G đại diện cho công việc thì M đại diện cho **Operating System Thread**.

Một Machine gần như tương ứng 1:1 với một Thread của hệ điều hành.

```
Machine (M)

↓

OS Thread
```

Ví dụ:

```
M0

↓

Thread 1

↓

Linux Scheduler
```

```
M1

↓

Thread 2
```

Go Runtime không trực tiếp chạy trên CPU.

Runtime vẫn phải sử dụng Thread do hệ điều hành cung cấp.

---

## M chứa những gì?

Một Machine thường chứa:

```text
Machine (M)

+-------------------------+
| OS Thread               |
+-------------------------+
| Current Goroutine       |
+-------------------------+
| Current Processor       |
+-------------------------+
| Registers               |
+-------------------------+
| Thread Local Storage    |
+-------------------------+
| Syscall Information     |
+-------------------------+
```

Một M chỉ có thể chạy **một Goroutine tại một thời điểm**.

Điều này rất quan trọng.

Nếu bạn có:

```
1000 Goroutines

2 Machines
```

Thì tại cùng một thời điểm chỉ có:

```
2 Goroutines
```

được thực thi.

Các Goroutine còn lại sẽ chờ trong Scheduler.

---

# 10.11 Processor (P)

Đây là thành phần khiến nhiều người nhầm lẫn nhất.

Tên gọi **Processor** dễ khiến chúng ta nghĩ đến CPU.

Thực tế:

> **P hoàn toàn không phải CPU.**

Đây chỉ là một đối tượng logic do Go Runtime tạo ra.

Có thể hiểu đơn giản:

```
CPU

≠

Processor (P)
```

Processor đại diện cho **quyền thực thi mã Go**.

Nếu không có Processor thì Machine không thể chạy Goroutine.

---

## Processor quản lý những gì?

Mỗi Processor sở hữu một tập tài nguyên riêng.

```text
Processor

+-------------------------+
| Local Run Queue         |
+-------------------------+
| mcache                  |
+-------------------------+
| Timer Cache             |
+-------------------------+
| Scheduler State         |
+-------------------------+
```

Điều đáng chú ý là:

**Local Run Queue thuộc về Processor, không phải Machine.**

Đây là quyết định thiết kế cực kỳ thông minh của Go Runtime.

Chúng ta sẽ hiểu lý do ở phần tiếp theo.

---

# 10.12 Quan hệ giữa G, M và P

Có thể tóm tắt mối quan hệ như sau.

## Quan hệ giữa G và M

```
N Goroutines

↓

1 Machine
```

Một Machine có thể chạy rất nhiều Goroutine theo thời gian.

Nhưng chỉ chạy được một Goroutine tại một thời điểm.

---

## Quan hệ giữa M và P

```
Machine

↓

owns

↓

Processor
```

Một Machine phải sở hữu Processor mới có thể thực thi Go code.

Nếu mất Processor:

```
Machine

↓

Cannot execute Go code
```

---

## Quan hệ giữa P và G

```
Processor

↓

Local Queue

↓

Runnable Goroutines
```

Processor quyết định Goroutine nào sẽ được giao cho Machine thực thi.

---

## Tóm tắt

Có thể hình dung toàn bộ GMP bằng sơ đồ sau.

```text
        Runnable Goroutines

      G1 G2 G3 G4 G5 G6 G7

                │

                ▼

          Local Run Queue

                │

                ▼

          Processor (P)

                │

                ▼

          Machine (M)

                │

                ▼

          Operating System Thread

                │

                ▼

                CPU
```

Từ góc nhìn của Scheduler:

- **G** là công việc cần chạy.
- **P** là nơi quản lý công việc.
- **M** là phương tiện để đưa công việc lên CPU.

Ba thành phần này tạo thành nền tảng của toàn bộ Go Runtime. Mọi cơ chế nâng cao như Work Stealing, Netpoller, Asynchronous Preemption hay Garbage Collector đều được xây dựng dựa trên mô hình GMP này.

---

---

# 10.13 Tại sao Go cần Processor (P)?

Đây là một trong những câu hỏi quan trọng nhất khi tìm hiểu Go Scheduler.

Nếu đã có:

- Goroutine (G)
- Machine (M)

thì tại sao Go Runtime lại phải tạo thêm một thành phần thứ ba là **Processor (P)**?

Thoạt nhìn, mô hình chỉ gồm G và M có vẻ đã đủ.

```
Goroutine

↓

Machine

↓

CPU
```

Nhưng trên thực tế, thiết kế này tồn tại nhiều vấn đề nghiêm trọng.

Để hiểu lý do Processor ra đời, chúng ta sẽ đi qua quá trình phát triển của Scheduler.

---

# 10.14 Scheduler chỉ gồm G và M

Giả sử Scheduler chỉ có hai thành phần.

```
G

↓

M

↓

CPU
```

Machine chịu trách nhiệm:

- Chạy Goroutine
- Quản lý Queue
- Quản lý Memory Cache
- Quản lý Timer
- Thực hiện Syscall

Thoạt nhìn mọi thứ đều hợp lý.

Tuy nhiên, hãy xem điều gì xảy ra khi Machine thực hiện một Blocking System Call.

```
read(socket)
```

hoặc

```
open(file)
```

Trong lúc này Thread sẽ bị hệ điều hành chặn.

```
             read()

               │

               ▼

      M0  -------------------- BLOCKED
```

Nếu Queue cũng nằm trong Machine thì toàn bộ Goroutine đang chờ trong Queue cũng bị "mắc kẹt".

Ví dụ:

```text
Machine

+----------------------+
| Local Queue          |
| G1 G2 G3 G4          |
+----------------------+
| OS Thread            |
+----------------------+
```

Khi Thread bị Block:

```text
Machine

↓

Blocked

↓

Queue cũng bị Block
```

Scheduler không còn cách nào để tiếp tục chạy:

```
G2

G3

G4
```

mặc dù CPU vẫn đang rảnh.

Đây là một sự lãng phí rất lớn.

---

# 10.15 Ý tưởng tách Scheduler khỏi Thread

Để giải quyết vấn đề trên, nhóm phát triển Go đưa ra một ý tưởng rất quan trọng.

> **Không gắn dữ liệu của Scheduler với OS Thread.**

Thay vào đó:

- Thread chỉ làm nhiệm vụ thực thi.
- Scheduler State sẽ được lưu ở một đối tượng khác.

Đối tượng đó chính là **Processor (P).**

```
Machine

↓

Execution

Processor

↓

Scheduling State
```

Đây là quyết định kiến trúc quan trọng nhất của Go Runtime.

---

# 10.16 Processor chính là "quyền chạy Go"

Có thể hiểu Processor như một **giấy phép thực thi Go Code**.

Một Machine chỉ có thể chạy Go khi đang sở hữu một Processor.

```
Machine

+

Processor

↓

Run Go Code
```

Nếu Machine không còn Processor:

```
Machine

↓

Cannot execute Go Code
```

Điều này nghe có vẻ kỳ lạ.

Tại sao Runtime lại làm như vậy?

Câu trả lời sẽ rõ hơn khi xem ví dụ dưới đây.

---

# 10.17 Ví dụ: Blocking System Call

Giả sử hệ thống có:

```
P0

M0

Queue:

G1
G2
G3
```

Ban đầu:

```text
           P0
      +-------------+
      | G1 G2 G3    |
      +-------------+
             │
             ▼
            M0
             │
             ▼
            CPU
```

Machine đang chạy:

```
G1
```

Trong lúc xử lý, G1 gọi:

```go
os.ReadFile(...)
```

Lệnh này thực hiện System Call.

Thread bị Kernel chặn.

```
Kernel

↓

Block Thread
```

Nếu không có Processor, toàn bộ Queue sẽ bị chặn theo.

```
G2

G3

↓

Cannot Run
```

CPU sẽ rơi vào trạng thái nhàn rỗi.

---

# 10.18 Điều gì xảy ra khi có Processor?

Bây giờ hãy xem cùng ví dụ trên nhưng có thêm Processor.

Ban đầu:

```text
P0

Queue

G1
G2
G3

↓

M0

↓

CPU
```

Khi M0 đi vào Blocking System Call:

```text
P0

↓

Detach

↓

M0 (Blocked)
```

Runtime sẽ thu hồi Processor.

```
P0

↓

Available
```

Sau đó Runtime tạo (hoặc lấy từ Idle List) một Machine khác.

```
M1
```

và gắn Processor cho Machine này.

```text
          P0

           │

           ▼

          M1

           │

           ▼

          CPU
```

Machine mới tiếp tục chạy:

```
G2

↓

G3
```

Trong khi:

```
M0
```

vẫn đang bị Block trong Kernel.

Kết quả là CPU không bị lãng phí.

---

# 10.19 Điều gì xảy ra khi System Call kết thúc?

Sau một khoảng thời gian:

```
read()

↓

Finished
```

Kernel đánh thức Thread.

```
M0

↓

Wake Up
```

Lúc này Runtime kiểm tra:

**Processor cũ còn không?**

Có hai khả năng.

## Trường hợp 1

Processor vẫn chưa được sử dụng.

```
P0

↓

Attach

↓

M0
```

Machine tiếp tục chạy.

---

## Trường hợp 2

Processor đã được Machine khác sử dụng.

```
P0

↓

M1
```

Khi đó M0 không thể chạy Go Code.

Nó sẽ:

```
Sleep

↓

Idle List
```

hoặc

```
Try Acquire Another Processor
```

Điều này giúp Runtime không tạo quá nhiều Thread.

---

# 10.20 Processor còn giúp tối ưu Memory

Processor không chỉ chứa Queue.

Nó còn chứa rất nhiều tài nguyên Runtime.

Ví dụ:

```text
Processor

+------------------------+
| Local Run Queue        |
+------------------------+
| mcache                 |
+------------------------+
| Tiny Allocator         |
+------------------------+
| Timer Cache            |
+------------------------+
| Scheduler State        |
+------------------------+
```

Đặc biệt là:

```
mcache
```

Trong Chapter Memory Allocator chúng ta sẽ học:

```
mheap

↓

mcentral

↓

mcache
```

Điều quan trọng là:

**mcache thuộc về Processor.**

Không phải Machine.

Nếu mcache gắn với Machine thì mỗi lần Thread bị Block:

```
Memory Cache

↓

Blocked
```

Runtime sẽ phải đồng bộ rất nhiều.

Hiệu năng giảm đáng kể.

---

# 10.21 Processor giúp giảm Lock Contention

Giả sử toàn bộ Goroutine dùng chung một Queue.

```
Global Queue

↓

100 Threads
```

Mỗi lần:

```
Push

Pop
```

đều phải khóa Queue.

```
Lock

↓

Unlock
```

Khi số CPU tăng lên:

```
32 CPU

↓

64 CPU

↓

128 CPU
```

Lock sẽ trở thành nút thắt.

Go giải quyết bằng cách:

```
P0

↓

Local Queue
```

```
P1

↓

Local Queue
```

```
P2

↓

Local Queue
```

Mỗi Processor có Queue riêng.

Nhờ vậy:

- Ít tranh chấp Lock hơn.
- Scheduler mở rộng tốt hơn trên hệ thống nhiều nhân.

Đây cũng là tiền đề cho cơ chế **Work Stealing**, sẽ được trình bày ở phần sau.

---

# 10.22 Tóm tắt vai trò của Processor

Processor được tạo ra để giải quyết đồng thời nhiều vấn đề:

| Vấn đề | Processor giải quyết như thế nào |
|---------|----------------------------------|
| Blocking System Call | Tách Scheduler khỏi Thread để có thể chuyển Processor sang Machine khác |
| Context Switching | Giảm chi phí chuyển đổi giữa các Goroutine |
| Memory Allocation | Mỗi Processor có `mcache` riêng, giảm tranh chấp bộ nhớ |
| Scheduling | Mỗi Processor có Local Run Queue riêng |
| Scalability | Giảm Lock Contention khi số CPU tăng |
| CPU Utilization | Giữ CPU luôn có Goroutine để thực thi |

Có thể xem **Processor** là tài nguyên trung tâm của Go Runtime.

Nếu Goroutine là **đơn vị công việc**, Machine là **phương tiện thực thi**, thì Processor chính là **bộ não điều phối**, nơi lưu giữ toàn bộ trạng thái cần thiết để Runtime lập lịch một cách hiệu quả.

---

> **Insight**
>
> Một trong những hiểu lầm phổ biến là cho rằng **Processor đại diện cho CPU**. Thực tế không phải vậy.
>
> CPU là phần cứng do hệ điều hành quản lý.
>
> Processor (`P`) là một đối tượng logic hoàn toàn do Go Runtime tạo ra.
>
> Số lượng Processor được quyết định bởi `GOMAXPROCS` và đại diện cho số lượng goroutine có thể thực thi mã Go song song tại một thời điểm, chứ không phải số lượng CPU vật lý của hệ thống.

---

---

# 10.23 Run Queue - Nơi Goroutine chờ được thực thi

Đến thời điểm này, chúng ta đã hiểu ba thành phần của Scheduler:

- G (Goroutine)
- M (Machine)
- P (Processor)

Tuy nhiên vẫn còn một câu hỏi quan trọng:

> **Sau khi một Goroutine được tạo ra, nó sẽ nằm ở đâu trước khi được CPU thực thi?**

Câu trả lời là:

> **Run Queue.**

Run Queue là hàng đợi chứa các Goroutine đang ở trạng thái **Runnable**, nghĩa là:

- Đã sẵn sàng chạy.
- Không bị block.
- Không đang chờ Channel.
- Không đang chờ Mutex.
- Không đang ngủ (`time.Sleep`).
- Chỉ còn chờ Scheduler cấp CPU.

Có thể hình dung:

```text
            go func()

                │

                ▼

         Runnable Goroutine

                │

                ▼

           Run Queue

                │

                ▼

           Scheduler

                │

                ▼

               CPU
```

Run Queue chính là "phòng chờ" của Scheduler.

---

# 10.24 Tại sao cần Run Queue?

Giả sử chúng ta tạo liên tiếp 1000 Goroutine.

```go
for i := 0; i < 1000; i++ {
    go worker(i)
}
```

CPU không thể chạy cả 1000 Goroutine ngay lập tức.

Ví dụ:

```
CPU Core = 8
```

Thì cùng một thời điểm chỉ tối đa:

```
8 Goroutine
```

được thực thi.

Vậy 992 Goroutine còn lại sẽ ở đâu?

Chúng sẽ nằm trong Run Queue.

```
Runnable

↓

Run Queue

↓

Waiting CPU
```

Run Queue đóng vai trò như một bộ đệm giữa việc tạo Goroutine và việc thực thi Goroutine.

---

# 10.25 Chỉ dùng một Global Queue có đủ không?

Một ý tưởng đơn giản là:

```
                Global Queue

G1 G2 G3 G4 G5 G6 G7 G8 ...

        │

        ▼

     Scheduler

        ▼

    CPU Threads
```

Mọi Goroutine đều được đưa vào cùng một Queue.

Mỗi Machine sẽ lấy Goroutine từ Queue này.

Thiết kế này rất dễ hiểu.

Nhưng nó không phù hợp với hệ thống nhiều CPU.

---

## Vấn đề đầu tiên: Lock Contention

Giả sử máy có:

```
64 CPU
```

và cũng có:

```
64 Machines
```

Khi đó cả 64 Machine đều muốn lấy Goroutine.

```
Global Queue

↓

Lock

↓

Pop

↓

Unlock
```

Điều này dẫn đến:

```
64 Threads

↓

1 Lock
```

Tất cả Thread phải tranh chấp cùng một khóa.

CPU sẽ mất rất nhiều thời gian để:

- Chờ Lock
- Giải phóng Lock
- Đồng bộ Cache giữa các CPU

Khi số CPU tăng lên, hiệu năng lại giảm.

Đây là hiện tượng **Lock Contention**.

---

## Vấn đề thứ hai: Cache Locality

CPU hiện đại đều có nhiều tầng Cache.

```
CPU

↓

L1 Cache

↓

L2 Cache

↓

L3 Cache

↓

RAM
```

Nếu một Thread liên tục lấy dữ liệu từ cùng một Queue chung:

```
Global Queue
```

thì dữ liệu phải di chuyển qua lại giữa nhiều CPU.

Điều này làm giảm hiệu quả của CPU Cache.

---

## Vấn đề thứ ba: Khả năng mở rộng

Một Queue duy nhất sẽ trở thành nút thắt cổ chai.

Khi hệ thống mở rộng lên:

- 32 CPU
- 64 CPU
- 128 CPU

thì Global Queue sẽ ngày càng trở thành điểm nóng (Hot Spot).

Điều này đi ngược lại mục tiêu thiết kế của Go Scheduler.

---

# 10.26 Thiết kế của Go Runtime

Để giải quyết các vấn đề trên, Go Runtime sử dụng mô hình kết hợp:

- Local Run Queue
- Global Run Queue

```text
                 Global Queue
               +-------------+
               | G20 G21 G22 |
               +-------------+

                    ▲
                    │

+-----------+   +-----------+   +-----------+

|   P0      |   |    P1     |   |    P2     |

| Local Q   |   | Local Q   |   | Local Q   |

| G1 G2 G3  |   | G8 G9     |   | G15 G16   |

+-----------+   +-----------+   +-----------+
```

Mỗi Processor có Queue riêng.

Ngoài ra Runtime vẫn duy trì một Global Queue để xử lý một số trường hợp đặc biệt.

Đây là một quyết định thiết kế rất quan trọng.

---

# 10.27 Local Run Queue

Mỗi Processor đều sở hữu một Local Queue.

```
P0

↓

Local Queue
```

Ví dụ:

```text
Processor P0

+------------------------+

G1

G2

G3

G4

+------------------------+
```

Machine gắn với Processor sẽ luôn ưu tiên lấy Goroutine từ Queue này.

```
P0

↓

Pop G1

↓

Run
```

Không cần truy cập Global Queue.

Không cần tranh chấp với Processor khác.

Điều này giúp Scheduler hoạt động rất nhanh.

---

## Đặc điểm của Local Queue

Local Queue có một số đặc điểm quan trọng.

### 1. Thuộc về Processor

Không phải Machine.

Điều này rất quan trọng.

Nếu Machine bị Block:

```
Machine

↓

Blocked
```

Processor vẫn còn.

Queue vẫn còn.

Runtime chỉ cần gắn Processor sang Machine khác.

Không cần di chuyển Queue.

---

### 2. Không chia sẻ với Processor khác

Ví dụ:

```
P0

↓

G1 G2 G3
```

Processor khác không được phép truy cập Queue này một cách tùy ý.

Điều này giảm đáng kể Lock Contention.

---

### 3. Scheduler luôn ưu tiên Local Queue

Thứ tự ưu tiên của Scheduler là:

```
Local Queue

↓

Global Queue

↓

Work Stealing

↓

Netpoller
```

Điều này giúp Goroutine được thực thi với chi phí thấp nhất.

---

# 10.28 Global Run Queue

Bên cạnh Local Queue, Runtime vẫn duy trì một Queue toàn cục.

```
Global Queue
```

Nhiều người thắc mắc:

> Đã có Local Queue rồi thì tại sao vẫn cần Global Queue?

Lý do là Local Queue không thể giải quyết mọi trường hợp.

Global Queue đóng vai trò như một "vùng đệm" dùng khi:

- Khởi tạo Goroutine trong một số ngữ cảnh đặc biệt.
- Phân phối lại công việc giữa các Processor.
- Local Queue bị đầy.
- Scheduler khởi động.

Nói cách khác:

```
Global Queue

↓

Fallback
```

Nó không phải nơi Scheduler truy cập thường xuyên.

---

# 10.29 Local Queue có kích thước giới hạn

Local Queue không thể lớn vô hạn.

Trong Runtime, mỗi Processor chỉ lưu được một số lượng Goroutine nhất định trong Local Queue (triển khai hiện tại sử dụng một hàng đợi vòng với kích thước cố định).

Điều này giúp:

- Truy cập nhanh.
- Không phải cấp phát bộ nhớ liên tục.
- Giảm áp lực cho Garbage Collector.

Nếu Local Queue đầy:

```
New Goroutine

↓

Local Queue Full
```

Runtime sẽ chuyển một phần Goroutine sang Global Queue để cân bằng tải.

Chi tiết thuật toán này sẽ được phân tích khi đọc `runtime/proc.go`.

---

# 10.30 Luồng hoạt động tổng quát

Có thể hình dung Scheduler hoạt động như sau.

```text
          go func()

              │

              ▼

      Runnable Goroutine

              │

              ▼

      Local Queue của P

              │

              ▼

         Machine (M)

              │

              ▼

             CPU
```

Nếu Local Queue rỗng:

```text
Local Queue

↓

Empty

↓

Global Queue

↓

Still Empty

↓

Work Stealing

↓

Still Empty

↓

Netpoller
```

Đây chính là thứ tự mà Scheduler sử dụng để tìm Goroutine tiếp theo.

Trong các phần tiếp theo, chúng ta sẽ phân tích chi tiết từng bước của thuật toán này.

---

# 10.31 Tóm tắt

Run Queue là nơi lưu trữ các Goroutine đã ở trạng thái **Runnable** nhưng chưa được cấp CPU.

Go Runtime không sử dụng một Queue duy nhất mà kết hợp:

- **Local Run Queue**: thuộc về từng Processor, là nơi Scheduler ưu tiên lấy Goroutine.
- **Global Run Queue**: dùng cho các trường hợp đặc biệt và hỗ trợ cân bằng tải.

Thiết kế này giúp:

- Giảm Lock Contention.
- Tăng khả năng mở rộng trên hệ thống nhiều CPU.
- Tận dụng tốt CPU Cache.
- Giảm chi phí lập lịch.
- Chuẩn bị nền tảng cho cơ chế **Work Stealing**, nội dung sẽ được phân tích trong phần tiếp theo.

---

---

# 10.32 Scheduler Loop - Trái tim của Go Runtime

Đến thời điểm này, chúng ta đã biết:

- Goroutine được lưu trong Run Queue.
- Mỗi Processor có Local Queue riêng.
- Runtime cũng duy trì một Global Queue.

Nhưng vẫn còn một câu hỏi quan trọng:

> **Scheduler quyết định Goroutine nào sẽ được chạy tiếp theo như thế nào?**

Câu trả lời nằm ở **Scheduler Loop**.

Có thể xem Scheduler Loop là "vòng lặp vô tận" của Go Runtime.

Miễn là chương trình vẫn còn chạy, Scheduler Loop vẫn tiếp tục hoạt động.

```
while (program is running) {

    tìm Goroutine

    chạy Goroutine

    Goroutine kết thúc?

        YES -> tìm Goroutine khác

        NO -> tiếp tục chạy
}
```

Tất nhiên, Go Runtime không thực sự viết như trên, nhưng về mặt ý tưởng thì Scheduler hoạt động gần như vậy.

---

# 10.33 Góc nhìn trực quan

Hãy tưởng tượng một nhà hàng.

Có:

- Khách hàng
- Nhân viên phục vụ
- Quản lý

Trong mô hình của Go:

```
Khách hàng

↓

Goroutine
```

```
Nhân viên

↓

Machine
```

```
Quản lý

↓

Processor
```

Mỗi khi một nhân viên rảnh, người quản lý sẽ giao cho họ một khách hàng tiếp theo.

Sau khi phục vụ xong:

```
Nhân viên

↓

Quay lại

↓

Nhận khách tiếp theo
```

Scheduler của Go cũng hoạt động tương tự.

---

# 10.34 Một vòng đời của Scheduler

Giả sử Processor có Local Queue như sau:

```text
+------------------+

G1

G2

G3

G4

+------------------+
```

Machine đang rảnh.

Scheduler sẽ:

```
Pop G1

↓

Run G1
```

Sau một khoảng thời gian, G1 có thể xảy ra một trong các trường hợp sau.

---

## Trường hợp 1 — Goroutine hoàn thành

Ví dụ:

```go
go func() {
    fmt.Println("Hello")
}()
```

Sau khi chạy xong:

```
G1

↓

Dead
```

Scheduler sẽ quay lại Queue.

```
Pop G2
```

Tiếp tục thực thi.

---

## Trường hợp 2 — Goroutine bị Block

Ví dụ:

```go
<-ch
```

hoặc

```go
mutex.Lock()
```

hoặc

```go
time.Sleep()
```

Lúc này:

```
G1

↓

Waiting
```

Scheduler sẽ:

```
Save Context

↓

Remove G1

↓

Run G2
```

Sau khi điều kiện chờ kết thúc:

```
Channel Ready

↓

G1

↓

Runnable

↓

Run Queue
```

Scheduler sẽ thực thi G1 sau.

---

## Trường hợp 3 — Hết Time Slice

Một Goroutine chạy quá lâu.

Ví dụ:

```go
for {

}
```

Nếu Runtime không can thiệp:

```
G1

↓

Forever
```

Các Goroutine khác sẽ không bao giờ được chạy.

Để tránh điều này, Go Runtime sử dụng **Preemption**.

Scheduler sẽ:

```
Pause G1

↓

Save Context

↓

Run G2
```

Chi tiết về Preemption sẽ được trình bày ở phần sau.

---

# 10.35 Scheduler không chạy Function

Đây là điểm rất nhiều lập trình viên mới học Go dễ hiểu nhầm.

Ví dụ:

```go
go processOrder()
```

Nhiều người nghĩ Runtime đang chạy:

```
Function processOrder()
```

Thực tế không phải vậy.

Runtime chạy:

```
Goroutine

↓

Stack

↓

Program Counter

↓

Registers

↓

Continue Execution
```

Function chỉ là đoạn mã.

Scheduler luôn thao tác trên đối tượng Goroutine.

Điều này cho phép Runtime:

- Tạm dừng.
- Tiếp tục.
- Di chuyển.
- Đánh thức.

một Goroutine bất cứ lúc nào.

---

# 10.36 Scheduler luôn ưu tiên Local Queue

Giả sử hệ thống có:

```text
P0

Queue:

G1

G2

G3
```

Scheduler sẽ luôn lấy:

```
G1

↓

Run
```

Sau đó:

```
G2

↓

Run
```

Rồi:

```
G3

↓

Run
```

Không cần truy cập Global Queue.

Không cần Work Stealing.

Không cần Netpoller.

Đây là con đường nhanh nhất.

---

# 10.37 Khi Local Queue rỗng

Sau khi chạy hết:

```
G1

G2

G3
```

Local Queue trở thành:

```text
Empty
```

Scheduler không dừng lại.

Nó tiếp tục tìm Goroutine.

Thứ tự tìm kiếm là:

```
1. Local Queue

↓

2. Global Queue

↓

3. Work Stealing

↓

4. Netpoller

↓

5. Idle
```

Đây là thứ tự rất quan trọng.

Hầu hết các câu hỏi phỏng vấn về Scheduler đều xoay quanh thứ tự này.

---

# 10.38 Bước 1 — Kiểm tra Local Queue

Scheduler luôn bắt đầu bằng Local Queue.

```text
Local Queue

↓

Có Goroutine?

YES

↓

Run
```

Đây là đường đi nhanh nhất.

Không cần đồng bộ.

Không cần khóa.

Không cần giao tiếp với Processor khác.

---

# 10.39 Bước 2 — Kiểm tra Global Queue

Nếu Local Queue rỗng:

```
Local Queue

↓

Empty
```

Scheduler sẽ kiểm tra:

```
Global Queue
```

Nếu Global Queue có Goroutine:

```text
Global Queue

↓

G100

↓

Move về Local Queue

↓

Run
```

Lưu ý rằng Scheduler thường không chạy trực tiếp từ Global Queue mà sẽ đưa Goroutine về Local Queue trước để tận dụng lợi thế của bộ nhớ cục bộ và giảm tranh chấp.

---

# 10.40 Bước 3 — Work Stealing

Nếu cả:

```
Local Queue

↓

Empty
```

và

```
Global Queue

↓

Empty
```

Scheduler sẽ nghĩ:

> "Có Processor nào khác còn nhiều việc không?"

Nếu có:

```text
P1

Queue

G20

G21

G22

G23

G24
```

Processor hiện tại sẽ lấy một phần công việc.

```
Steal

↓

G22

G23
```

Sau đó:

```
Move

↓

Local Queue
```

Tiếp tục thực thi.

Nhờ vậy, toàn bộ CPU đều được sử dụng hiệu quả.

---

# 10.41 Bước 4 — Netpoller

Giả sử:

- Local Queue rỗng.
- Global Queue rỗng.
- Không thể Work Stealing.

Scheduler vẫn chưa bỏ cuộc.

Nó sẽ hỏi:

> Có Goroutine nào đang chờ I/O không?

Ví dụ:

```go
conn.Read(...)
```

Khi dữ liệu mạng đến:

```
Socket Ready

↓

Wake Goroutine

↓

Runnable
```

Scheduler sẽ đưa Goroutine này trở lại Run Queue.

Điều này giúp Go xử lý hàng chục nghìn kết nối mạng mà không cần tạo hàng chục nghìn Thread.

Chúng ta sẽ tìm hiểu Netpoller chi tiết trong một chương riêng.

---

# 10.42 Bước cuối cùng — Idle

Nếu:

- Local Queue rỗng.
- Global Queue rỗng.
- Không Work Stealing được.
- Không có I/O hoàn thành.

Scheduler kết luận:

```
Không còn Goroutine Runnable.
```

Machine sẽ chuyển sang trạng thái Idle.

```text
Machine

↓

Sleep
```

Khi có Goroutine mới:

```go
go worker()
```

Runtime sẽ đánh thức một Machine để tiếp tục xử lý.

Nhờ vậy, Go không tiêu tốn CPU khi hệ thống không có công việc.

---

# 10.43 Thuật toán tổng quát

Có thể mô tả luồng hoạt động của Scheduler bằng sơ đồ sau:

```text
                Scheduler

                     │

                     ▼

          Local Queue có việc?

             │          │

           YES         NO

            │           │

            ▼           ▼

        Run G       Global Queue?

                        │

                YES            NO

                 │              │

                 ▼              ▼

           Move về Local     Work Stealing?

                                  │

                         YES               NO

                          │                 │

                          ▼                 ▼

                    Steal Goroutine     Netpoller?

                                              │

                                      YES            NO

                                       │              │

                                       ▼              ▼

                                   Run G         Machine Idle
```

Đây là vòng lặp diễn ra liên tục trong suốt vòng đời của chương trình Go.

---

# 10.44 Tóm tắt

Scheduler Loop là cơ chế trung tâm điều phối việc thực thi Goroutine.

Thay vì chọn ngẫu nhiên, Scheduler luôn tuân theo một thứ tự ưu tiên rõ ràng:

1. **Local Run Queue** – nơi nhanh nhất và ít tranh chấp nhất.
2. **Global Run Queue** – nguồn công việc dùng chung.
3. **Work Stealing** – lấy bớt công việc từ Processor khác khi cần cân bằng tải.
4. **Netpoller** – đánh thức các Goroutine đang chờ I/O.
5. **Idle** – đưa Machine vào trạng thái nghỉ nếu không còn công việc.

Nhờ chiến lược này, Go Runtime vừa tận dụng tốt CPU, vừa giảm chi phí đồng bộ, đồng thời vẫn đảm bảo khả năng mở rộng trên các hệ thống có nhiều lõi xử lý.

---

---

# 10.45 Goroutine được tạo ra như thế nào?

Trong phần trước, chúng ta đã biết Scheduler lấy Goroutine từ Run Queue để thực thi.

Bây giờ chúng ta sẽ trả lời câu hỏi tiếp theo:

> **Điều gì xảy ra khi chúng ta gọi từ khóa `go`?**

Ví dụ:

```go
func main() {

    go worker()

    fmt.Println("main")
}
```

Lệnh đơn giản này kích hoạt rất nhiều thành phần bên trong Go Runtime.

Quá trình này có thể được mô tả như sau:

```
go worker()

↓

Runtime tạo Goroutine (G)

↓

Khởi tạo Stack

↓

Khởi tạo Metadata

↓

Đưa vào Run Queue

↓

Scheduler tìm thấy

↓

Machine thực thi

↓

CPU
```

Điều quan trọng cần nhớ là:

> **Từ khóa `go` không có nghĩa là Goroutine sẽ chạy ngay lập tức.**

Đây là một hiểu lầm rất phổ biến.

---

# 10.46 "go" không đồng nghĩa với "Run"

Ví dụ:

```go
go worker()
```

Nhiều lập trình viên hình dung như sau:

```
go worker()

↓

CPU chạy worker()
```

Thực tế hoàn toàn khác.

Điều xảy ra là:

```
go worker()

↓

Create Goroutine

↓

Runnable

↓

Run Queue

↓

Scheduler

↓

CPU
```

Có thể phải mất:

- vài micro giây
- vài mili giây
- hoặc lâu hơn

Goroutine mới thực sự được CPU thực thi.

Điều này phụ thuộc vào:

- Scheduler
- CPU
- Load của hệ thống
- Các Goroutine khác

---

# 10.47 Một ví dụ đơn giản

```go
package main

import "fmt"

func main() {

    go fmt.Println("A")

    fmt.Println("B")
}
```

Kết quả có thể là:

```
B

A
```

Hoặc:

```
A

B
```

Hoặc thậm chí:

```
B
```

nếu chương trình kết thúc trước khi Goroutine được Scheduler chạy.

Điều này chứng minh rằng:

```
go

≠

Run Immediately
```

Nó chỉ có nghĩa là:

> **Đưa công việc này cho Scheduler xử lý sau.**

---

# 10.48 Các bước tạo Goroutine

Khi Runtime nhận:

```go
go worker()
```

Nó sẽ thực hiện một chuỗi các bước.

## Bước 1 — Cấp phát đối tượng G

Runtime tạo một struct `g`.

```
Allocate G
```

Bên trong chứa:

- Stack
- Program Counter
- Status
- Scheduler Metadata

Lúc này Goroutine vẫn chưa chạy.

---

## Bước 2 — Khởi tạo Stack

Runtime cấp phát Stack ban đầu.

```
Initial Stack

≈ 2 KB
```

Điểm đặc biệt của Go là Stack này có thể mở rộng khi cần.

Khác với Thread của hệ điều hành, Stack của Goroutine không cố định.

Đây là một trong những lý do giúp Goroutine rất nhẹ.

---

## Bước 3 — Thiết lập điểm bắt đầu

Runtime lưu địa chỉ của hàm:

```go
worker
```

vào Program Counter.

```
PC

↓

worker()
```

Khi Scheduler quyết định chạy Goroutine, CPU sẽ bắt đầu thực thi từ địa chỉ này.

---

## Bước 4 — Đánh dấu trạng thái

Lúc này:

```
Status

↓

Runnable
```

Goroutine đã sẵn sàng.

Chỉ còn chờ CPU.

---

## Bước 5 — Đưa vào Run Queue

Đây là bước quan trọng nhất.

```
Runnable

↓

Run Queue
```

Thông thường Runtime sẽ ưu tiên đưa Goroutine vào:

```
Local Queue
```

của Processor hiện tại.

Nếu Queue đầy thì mới sử dụng Global Queue.

---

# 10.49 Sau khi vào Run Queue

Lúc này Goroutine chỉ đang chờ.

```
Runnable

↓

Waiting CPU
```

Machine sẽ lấy Goroutine này khi:

- Machine đang rảnh.
- Scheduler chọn đến.
- Processor còn thời gian.

Nếu chưa đến lượt:

```
Runnable

↓

Continue Waiting
```

Không có gì đảm bảo Goroutine sẽ được chạy ngay sau khi được tạo.

---

# 10.50 Trạng thái của Goroutine sau khi tạo

Có thể hình dung như sau.

```text
go worker()

        │

        ▼

Allocate G

        │

        ▼

Create Stack

        │

        ▼

Initialize Metadata

        │

        ▼

Runnable

        │

        ▼

Run Queue

        │

        ▼

Scheduler

        │

        ▼

Running
```

Toàn bộ quá trình này thường chỉ mất một khoảng thời gian rất ngắn, giúp việc tạo Goroutine có chi phí thấp hơn rất nhiều so với việc tạo một OS Thread.

---

# 10.51 Tại sao tạo Goroutine lại nhanh?

Hãy so sánh với việc tạo Thread.

## Tạo OS Thread

```
Application

↓

Kernel

↓

Allocate Kernel Object

↓

Allocate Kernel Stack

↓

Allocate User Stack

↓

Initialize Registers

↓

Scheduler

↓

Ready
```

Quá trình này cần sự tham gia của Kernel.

Chi phí tương đối lớn.

---

## Tạo Goroutine

```
Application

↓

Go Runtime

↓

Allocate G

↓

Allocate Small Stack

↓

Runnable
```

Không cần Kernel.

Không cần tạo Thread mới.

Không cần chuyển sang Kernel Mode.

Đây chính là lý do Go có thể tạo hàng triệu Goroutine trong thời gian rất ngắn.

---

# 10.52 Goroutine có phải luôn tạo mới?

Không hẳn.

Trong nhiều trường hợp, Runtime sẽ tái sử dụng một số cấu trúc dữ liệu nội bộ để giảm chi phí cấp phát.

Ví dụ:

- Goroutine đã kết thúc có thể được Runtime giữ lại trong các danh sách nội bộ.
- Một số vùng nhớ được tái sử dụng thay vì luôn cấp phát mới.

Điều này giúp giảm:

- Chi phí cấp phát bộ nhớ.
- Áp lực lên Garbage Collector.
- Fragmentation của Heap.

Đây là một trong những kỹ thuật tối ưu hiệu năng của Go Runtime và hoàn toàn minh bạch với lập trình viên.

---

# 10.53 Một ví dụ trực quan

Giả sử chúng ta viết:

```go
go task1()

go task2()

go task3()
```

Runtime sẽ tạo:

```text
G1

Runnable

↓

Local Queue


G2

Runnable

↓

Local Queue


G3

Runnable

↓

Local Queue
```

Scheduler sau đó sẽ lấy lần lượt:

```
Run G1

↓

Run G2

↓

Run G3
```

Lưu ý rằng thứ tự thực thi **không nhất thiết trùng với thứ tự được tạo**.

Scheduler có thể thay đổi thứ tự dựa trên trạng thái của hệ thống.

---

# 10.54 Những hiểu lầm phổ biến

## Hiểu lầm 1

> `go` nghĩa là chạy song song ngay lập tức.

Sai.

`go` chỉ tạo một Goroutine và giao nó cho Scheduler.

---

## Hiểu lầm 2

> Goroutine luôn chạy theo đúng thứ tự được tạo.

Sai.

Scheduler không đảm bảo điều này.

---

## Hiểu lầm 3

> Sau khi gọi `go`, Goroutine chắc chắn sẽ chạy.

Không hoàn toàn đúng.

Nếu chương trình kết thúc trước:

```go
func main() {

    go worker()

}
```

thì Goroutine có thể chưa bao giờ được CPU thực thi.

Đó là lý do chúng ta thường sử dụng:

- `sync.WaitGroup`
- Channel
- Context

để đồng bộ việc kết thúc của các Goroutine.

---

# 10.55 Tóm tắt

Khi gặp từ khóa `go`, Go Runtime **không chạy ngay hàm được chỉ định**.

Thay vào đó, Runtime:

1. Tạo một đối tượng Goroutine (`G`).
2. Cấp phát Stack ban đầu.
3. Khởi tạo các metadata cần thiết.
4. Đánh dấu Goroutine ở trạng thái **Runnable**.
5. Đưa Goroutine vào **Run Queue**.
6. Chờ Scheduler phân phối CPU để bắt đầu thực thi.

Nhờ quy trình này, việc tạo Goroutine có chi phí rất thấp và hoàn toàn tách biệt với việc tạo OS Thread. Đây là nền tảng để Go có thể xử lý số lượng lớn tác vụ đồng thời một cách hiệu quả.

---

---

# 10.56 Work Stealing - Cơ chế cân bằng tải của Go Scheduler

Một trong những lý do giúp Go Scheduler có khả năng mở rộng rất tốt trên hệ thống nhiều CPU là **Work Stealing**.

Nếu chỉ có Local Queue, chúng ta sẽ gặp một vấn đề mới.

Giả sử hệ thống có 4 Processor.

```
P0

G1
G2
G3
G4
G5
G6
G7
G8
```

```
P1

Empty
```

```
P2

Empty
```

```
P3

Empty
```

Lúc này:

- CPU của P0 đang rất bận.
- CPU của P1, P2 và P3 lại gần như không có việc để làm.

Nếu không có cơ chế cân bằng tải:

```
P0

100% CPU
```

Trong khi:

```
P1

0%
```

```
P2

0%
```

```
P3

0%
```

Ba CPU đang bị lãng phí mặc dù hệ thống vẫn còn rất nhiều Goroutine chưa được xử lý.

Đây chính là vấn đề mà Work Stealing giải quyết.

---

# 10.57 Ý tưởng của Work Stealing

Ý tưởng rất đơn giản.

Nếu Processor của mình không còn việc:

```
Local Queue

↓

Empty
```

Thay vì ngồi chờ,

Processor sẽ hỏi:

> **"Có Processor nào khác còn nhiều việc không?"**

Nếu có:

```
Processor A

↓

Steal

↓

Processor B
```

Nó sẽ lấy một phần Goroutine từ Processor còn nhiều việc.

```
P0

G1
G2
G3
G4
G5
G6
```

↓

Steal

↓

```
P1

G4
G5
G6
```

Nhờ vậy cả hai CPU đều tiếp tục làm việc.

---

# 10.58 Tại sao không chuyển toàn bộ Queue?

Một câu hỏi rất hay là:

> Nếu đã đi lấy việc thì tại sao không lấy hết?

Ví dụ:

```
P0

G1

G2

G3

G4

G5

G6
```

Tại sao Runtime không chuyển toàn bộ sang:

```
P1
```

Có nhiều lý do.

---

## Lý do thứ nhất

Nếu chuyển toàn bộ.

```
P0

↓

Empty
```

Chỉ vài mili giây sau:

```
P0

↓

Idle
```

Processor ban đầu lại phải đi ăn cắp việc lần nữa.

Điều này làm tăng chi phí Scheduler.

---

## Lý do thứ hai

Việc di chuyển Goroutine cũng có chi phí.

Mỗi Goroutine đều có:

- Metadata
- Scheduler State
- Queue Operation

Nếu cứ liên tục di chuyển:

```
P0

↓

P1

↓

P2

↓

P3
```

Scheduler sẽ tốn nhiều thời gian cho việc cân bằng tải hơn là chạy chương trình.

---

## Lý do thứ ba

CPU Cache.

Giả sử:

```
P0
```

đã chạy:

```
G1

↓

G2

↓

G3
```

Dữ liệu của các Goroutine này đã nằm trong Cache của CPU.

Nếu chuyển hết sang Processor khác:

```
CPU Cache Miss
```

Hiệu năng sẽ giảm.

---

# 10.59 Go chỉ Steal một phần Queue

Go Runtime lựa chọn một chiến lược thông minh hơn.

Thay vì lấy toàn bộ Queue.

Runtime chỉ lấy khoảng:

```
Một nửa
```

Ví dụ.

Ban đầu.

```
P0

G1

G2

G3

G4

G5

G6

G7

G8
```

Sau Work Stealing.

```
P0

G1

G2

G3

G4
```

```
P1

G5

G6

G7

G8
```

Hai Processor cùng làm việc.

Khả năng phải Work Stealing lần nữa cũng giảm đáng kể.

---

# 10.60 Ví dụ thực tế

Giả sử hệ thống có:

```
4 CPU
```

Tại một thời điểm.

```
P0

100 Goroutines
```

```
P1

0
```

```
P2

0
```

```
P3

0
```

Scheduler sẽ thực hiện.

```
P1

↓

Steal

↓

50 Goroutines
```

Kết quả.

```
P0

50
```

```
P1

50
```

Sau đó.

```
P2
```

tiếp tục Steal.

```
P0

25
```

```
P2

25
```

Cuối cùng.

```
P3

↓

Steal từ P1
```

Hệ thống trở thành.

```
P0

25
```

```
P1

25
```

```
P2

25
```

```
P3

25
```

Khối lượng công việc được phân bố khá đồng đều.

---

# 10.61 Work Stealing diễn ra khi nào?

Work Stealing **không phải** là hoạt động diễn ra liên tục.

Nếu Local Queue vẫn còn Goroutine.

```
P0

↓

Local Queue

↓

Run
```

Scheduler sẽ **không** đi lấy việc.

Chỉ khi:

```
Local Queue

↓

Empty
```

Runtime mới bắt đầu cân nhắc Work Stealing.

Điều này giúp giảm đáng kể chi phí lập lịch.

---

# 10.62 Chọn Processor nào để Steal?

Một câu hỏi khác là:

> Processor sẽ lấy việc từ Processor nào?

Nếu luôn lấy từ:

```
P0
```

thì:

```
P0

↓

Hot Spot
```

Mọi Processor đều tranh chấp với P0.

Điều này không tốt.

Go Runtime lựa chọn một chiến lược đơn giản nhưng hiệu quả.

```
Random
```

Runtime chọn ngẫu nhiên một Processor khác.

Ví dụ.

```
P2

↓

Random

↓

P7
```

Nếu P7 không có việc.

Runtime thử Processor khác.

```
Random

↓

P3
```

Cách làm này giúp:

- Phân tán truy cập.
- Giảm tranh chấp.
- Đơn giản hóa thuật toán.
- Đạt hiệu quả tốt trong thực tế.

---

# 10.63 Điều gì xảy ra nếu tất cả Queue đều rỗng?

Giả sử.

```
P0

Empty
```

```
P1

Empty
```

```
P2

Empty
```

```
P3

Empty
```

Work Stealing sẽ thất bại.

```
No Work
```

Scheduler sẽ chuyển sang bước tiếp theo.

```
Netpoller
```

Nếu Netpoller cũng không tìm thấy Goroutine.

Machine sẽ đi ngủ.

```
Idle
```

Đây là lý do CPU của chương trình Go gần như bằng 0% khi không có công việc.

---

# 10.64 Lợi ích của Work Stealing

Work Stealing mang lại rất nhiều lợi ích.

## 1. Cân bằng tải tự động

Không cần lập trình viên can thiệp.

Runtime tự điều chỉnh.

---

## 2. Tận dụng tối đa CPU

Nếu còn Goroutine.

CPU gần như luôn có việc.

---

## 3. Giảm Lock Contention

Mỗi Processor chủ yếu sử dụng Local Queue.

Chỉ khi cần mới truy cập Queue của Processor khác.

---

## 4. Khả năng mở rộng tốt

Khi số CPU tăng.

Hiệu năng Scheduler vẫn tăng gần như tuyến tính trong nhiều khối lượng công việc.

Đây là một trong những lý do Go hoạt động rất tốt trên:

- Server 8 Core
- Server 32 Core
- Server 64 Core

---

# 10.65 Ví dụ trực quan

Có thể hình dung Work Stealing như sau.

```
Ban đầu

P0

████████████████████

P1



P2



P3
```

Sau Work Stealing.

```
P0

████████
```

```
P1

████████
```

```
P2

████
```

```
P3

████
```

Khối lượng công việc được phân phối đều hơn.

CPU được sử dụng hiệu quả hơn.

---

# 10.66 Tóm tắt

Work Stealing là cơ chế cân bằng tải quan trọng nhất của Go Scheduler.

Khi một Processor không còn Goroutine để thực thi, nó sẽ:

1. Chọn ngẫu nhiên một Processor khác.
2. Kiểm tra Local Queue của Processor đó.
3. Nếu Queue đủ lớn, lấy khoảng **một nửa** số Goroutine.
4. Đưa các Goroutine vừa lấy về Local Queue của mình.
5. Tiếp tục thực thi.

Nhờ Work Stealing, Go Runtime có thể tận dụng gần như toàn bộ tài nguyên CPU mà không cần một Scheduler trung tâm phức tạp. Đây là một trong những yếu tố cốt lõi giúp Go đạt hiệu năng cao trên các hệ thống đa lõi.

---

---

# 10.15 Blocking System Call - Điều gì xảy ra khi Thread bị chặn?

Trong phần trước, chúng ta đã tìm hiểu cách Scheduler phân phối Goroutine và cân bằng tải thông qua Work Stealing.

Tuy nhiên vẫn còn một tình huống rất quan trọng.

> **Điều gì xảy ra nếu OS Thread đang chạy Goroutine bị Kernel chặn?**

Ví dụ:

```go
file, err := os.Open("large-file.txt")
```

hoặc

```go
n, err := syscall.Read(fd, buf)
```

hoặc

```go
syscall.Accept(fd)
```

Các lời gọi này đều là **System Call**.

Khi thực hiện System Call, quyền điều khiển sẽ được chuyển từ User Space sang Kernel Space.

```
Go Code

↓

System Call

↓

Kernel

↓

Operating System
```

Nếu System Call mất:

- 1 ms
- 10 ms
- 100 ms
- hoặc lâu hơn

thì OS Thread cũng sẽ bị chặn trong suốt khoảng thời gian đó.

Đây chính là một trong những thách thức lớn nhất của mọi Runtime Scheduler.

---

# 10.16 Blocking System Call là gì?

Thông thường CPU thực thi chương trình trong User Space.

```
Application

↓

User Mode
```

Khi chương trình cần:

- Đọc file
- Ghi file
- Đọc socket
- Tạo process
- Truy cập thiết bị

CPU phải chuyển sang Kernel.

```
Application

↓

Kernel

↓

Driver

↓

Hardware
```

Quá trình này gọi là **System Call**.

Một số System Call hoàn thành rất nhanh.

Một số khác có thể phải chờ:

- Đĩa cứng
- SSD
- Network
- Database
- Thiết bị ngoại vi

Trong khoảng thời gian này, Thread không thể tiếp tục thực thi.

```
Thread

↓

Waiting
```

---

# 10.17 Vấn đề nếu không có Processor

Hãy giả sử Scheduler chỉ có:

- Goroutine
- Machine

Giả sử Machine đang chạy:

```
G1
```

```
G1

↓

Read File

↓

Kernel

↓

Waiting...
```

Trong lúc này Thread bị Block.

```
Machine

↓

Blocked
```

Nếu Queue cũng thuộc Machine.

```
Machine

↓

Queue

↓

Blocked
```

Các Goroutine phía sau:

```
G2

G3

G4
```

không thể chạy.

Mặc dù:

```
CPU

↓

Idle
```

Đây chính là điều Go Runtime phải tránh.

---

# 10.18 Cách Go Runtime xử lý

Go Runtime lựa chọn một giải pháp rất thông minh.

Ngay trước khi Machine đi vào Blocking System Call.

Runtime sẽ:

```
Detach Processor
```

Có nghĩa là.

```
Before

P0

↓

M0

↓

Read()
```

Sau khi đi vào Kernel.

```
P0

Detached
```

```
M0

↓

Kernel

↓

Blocked
```

Processor lúc này không còn gắn với Machine.

Runtime có thể sử dụng Processor đó cho Thread khác.

---

# 10.19 Machine mới tiếp tục thực thi

Sau khi Processor được thu hồi.

Runtime sẽ tìm một Machine khác.

Ví dụ.

```
Idle Machine

↓

M1
```

Sau đó.

```
P0

↓

Attach

↓

M1
```

Machine mới tiếp tục chạy.

```
P0

↓

Local Queue

↓

G2

↓

CPU
```

Trong khi đó.

```
M0
```

vẫn đang bị Block trong Kernel.

Hai việc diễn ra hoàn toàn độc lập.

---

# 10.20 Minh họa toàn bộ quá trình

Ban đầu.

```text
        P0

         │

         ▼

        M0

         │

         ▼

        G1
```

G1 gọi:

```go
os.ReadFile(...)
```

Thread đi vào Kernel.

```text
        P0

      Detach

         │

         ▼

     Available
```

```text
        M0

         │

         ▼

     Kernel Waiting
```

Runtime lấy Machine khác.

```text
        P0

         │

         ▼

        M1

         │

         ▼

        G2
```

CPU tiếp tục hoạt động bình thường.

---

# 10.21 Điều gì xảy ra khi System Call hoàn thành?

Sau một khoảng thời gian.

```
Disk Read

↓

Completed
```

Kernel đánh thức Thread.

```
Wake Thread
```

Machine M0 quay trở lại Runtime.

```
Kernel

↓

M0

↓

Runtime
```

Lúc này Runtime sẽ kiểm tra.

```
Processor Available?
```

Có hai trường hợp.

---

## Trường hợp 1 - Có Processor rảnh

Nếu còn Processor.

```
P1

↓

Idle
```

Runtime sẽ:

```
Attach

↓

M0
```

Sau đó tiếp tục chạy Goroutine.

```
Continue G1
```

---

## Trường hợp 2 - Không còn Processor

Nếu tất cả Processor đều đang bận.

```
P0

Busy

P1

Busy

P2

Busy
```

Machine M0 không thể chạy Go Code.

Nó sẽ chuyển sang trạng thái:

```
Idle Thread
```

và chờ đến khi có Processor rảnh.

Điều này giúp Runtime không vượt quá số lượng Processor được phép thực thi đồng thời.

---

# 10.22 Tại sao không tạo thêm Processor?

Một câu hỏi rất thú vị là:

> **Nếu Thread bị Block thì tại sao không tạo luôn Processor mới?**

Câu trả lời là:

**Processor đại diện cho mức độ song song của chương trình.**

Số lượng Processor được quyết định bởi:

```
GOMAXPROCS
```

Ví dụ.

```
GOMAXPROCS = 8
```

Có nghĩa là.

```
Maximum

8 Goroutines

Running Simultaneously
```

Nếu Runtime tự ý tạo thêm Processor.

```
8

↓

9

↓

10

↓

20
```

thì chương trình sẽ vượt quá mức song song mà người dùng đã cấu hình.

Do đó Runtime chỉ:

- Tạo thêm Machine khi cần.
- Không tự ý tạo thêm Processor.

---

# 10.23 Machine có thể nhiều hơn Processor

Đây là một điểm rất nhiều lập trình viên chưa để ý.

Giả sử.

```
GOMAXPROCS = 8
```

Không có nghĩa Runtime chỉ có:

```
8 Threads
```

Trong thực tế có thể xảy ra.

```
8 Processor
```

```
12 Machines
```

hoặc.

```
8 Processor
```

```
20 Machines
```

Lý do là một số Machine đang:

- Block trong Kernel.
- Chờ System Call.
- Chờ CGO.
- Chờ Lock của hệ điều hành.

Runtime cần tạo thêm Machine để đảm bảo vẫn có đủ Thread phục vụ các Processor đang hoạt động.

Điều cần ghi nhớ là:

> **Số lượng Machine có thể lớn hơn số lượng Processor, nhưng tại một thời điểm chỉ có tối đa `GOMAXPROCS` Machine đang thực thi Go code.**

---

# 10.24 Blocking System Call và Work Stealing

Hai cơ chế này thường bị nhầm lẫn.

Thực tế chúng giải quyết hai vấn đề hoàn toàn khác nhau.

### Blocking System Call

Giải quyết vấn đề:

```
OS Thread

↓

Blocked
```

Runtime sẽ:

```
Detach Processor

↓

Attach sang Machine khác
```

---

### Work Stealing

Giải quyết vấn đề:

```
Processor

↓

Hết việc
```

Runtime sẽ:

```
Steal

↓

Processor khác
```

Có thể tóm tắt như sau:

| Tình huống | Cơ chế |
|------------|--------|
| Thread bị Block | Detach Processor và gắn sang Machine khác |
| Processor hết Goroutine | Work Stealing |
| Goroutine chờ I/O | Netpoller |
| Goroutine chạy quá lâu | Preemption |

Mỗi cơ chế giải quyết một loại vấn đề riêng, nhưng tất cả đều phối hợp để Scheduler hoạt động hiệu quả.

---

# 10.25 Tóm tắt

Blocking System Call là một trong những thách thức lớn nhất đối với bất kỳ Runtime Scheduler nào.

Go Runtime giải quyết vấn đề này bằng cách **tách Processor khỏi Machine** trước khi Thread đi vào Kernel. Processor sau đó được gắn cho một Machine khác để tiếp tục thực thi các Goroutine còn lại.

Thiết kế này mang lại nhiều lợi ích:

- CPU không bị bỏ trống khi một Thread chờ I/O.
- Goroutine khác vẫn tiếp tục được thực thi.
- Không cần tăng số lượng Processor.
- Vẫn đảm bảo giới hạn song song do `GOMAXPROCS` quy định.

Đây là một trong những lý do quan trọng giúp Go có thể xử lý đồng thời hàng chục nghìn kết nối hoặc tác vụ I/O mà không cần tạo tương ứng hàng chục nghìn OS Thread.

---

---

# 10.26 Netpoller - Bí quyết xử lý hàng trăm nghìn kết nối đồng thời

Trong phần trước chúng ta đã học:

> Điều gì xảy ra khi một Goroutine thực hiện Blocking System Call.

Go Runtime giải quyết bằng cách:

- Tách Processor khỏi Machine.
- Đưa Processor sang Machine khác.
- Tiếp tục chạy các Goroutine khác.

Điều này giải quyết được vấn đề **Thread bị Block**.

Tuy nhiên, lại xuất hiện một câu hỏi mới.

> **Làm thế nào để Runtime biết khi nào một Goroutine đang chờ I/O có thể tiếp tục chạy?**

Ví dụ.

```go
conn.Read(buf)
```

Lúc này Goroutine sẽ chờ:

```
Network

↓

Data
```

Có thể mất:

- vài mili giây
- vài giây
- thậm chí lâu hơn

Runtime không thể liên tục hỏi:

```
Có dữ liệu chưa?

Có dữ liệu chưa?

Có dữ liệu chưa?
```

Hành động này được gọi là **Polling**.

Nếu hàng chục nghìn Goroutine đều polling như vậy thì CPU sẽ bị lãng phí rất lớn.

Go Runtime cần một giải pháp hiệu quả hơn.

Đó chính là **Netpoller**.

---

# 10.27 Polling là gì?

Giả sử bạn đang chờ một người bạn gửi tin nhắn.

Có hai cách.

## Cách thứ nhất

Bạn cứ 1 giây mở điện thoại một lần.

```
Có tin nhắn chưa?

↓

Chưa.

↓

Có tin nhắn chưa?

↓

Chưa.

↓

Có tin nhắn chưa?
```

Đây chính là Polling.

CPU cũng hoạt động tương tự.

```
Socket Ready?

↓

No

↓

Socket Ready?

↓

No

↓

Socket Ready?
```

Nếu có:

```
100.000 Socket
```

thì CPU sẽ phải kiểm tra:

```
100.000 lần
```

rất nhiều lần mỗi giây.

Đây là một sự lãng phí.

---

## Cách thứ hai

Bạn không kiểm tra liên tục.

Bạn chỉ chờ.

Khi có tin nhắn.

```
Notification

↓

Wake Up
```

Đây là mô hình **Event Driven**.

Go Runtime lựa chọn cách này.

---

# 10.28 Netpoller hoạt động như thế nào?

Netpoller là cầu nối giữa:

```
Operating System

↓

Go Runtime
```

Nhiệm vụ của Netpoller là:

- Theo dõi các Socket.
- Theo dõi các File Descriptor.
- Phát hiện khi I/O hoàn thành.
- Đánh thức Goroutine tương ứng.

Có thể hình dung.

```
Socket

↓

Operating System

↓

Netpoller

↓

Scheduler

↓

Run Queue

↓

CPU
```

Scheduler không cần liên tục kiểm tra từng Socket.

Hệ điều hành sẽ chủ động thông báo.

---

# 10.29 Netpoller sử dụng cơ chế của hệ điều hành

Go Runtime không tự xây dựng một hệ thống theo dõi I/O từ đầu.

Nó tận dụng cơ chế sẵn có của từng hệ điều hành.

| Hệ điều hành | Cơ chế |
|--------------|--------|
| Linux | epoll |
| macOS / BSD | kqueue |
| Windows | IOCP |

Mặc dù tên gọi khác nhau, mục tiêu đều giống nhau:

> **Thông báo khi một thao tác I/O đã sẵn sàng.**

Nhờ đó Runtime không cần Polling liên tục.

---

# 10.30 Ví dụ với một Goroutine

Giả sử có Goroutine.

```go
go func() {

    conn.Read(buf)

}()
```

Quá trình diễn ra như sau.

Bước 1.

```
Read()
```

Không có dữ liệu.

---

Bước 2.

Goroutine chuyển sang trạng thái.

```
Waiting
```

---

Bước 3.

Machine tiếp tục chạy Goroutine khác.

```
G2

↓

Running
```

---

Bước 4.

Một lúc sau.

```
Network Packet

↓

Arrived
```

---

Bước 5.

Kernel thông báo.

```
Socket Ready
```

---

Bước 6.

Netpoller nhận sự kiện.

```
Socket Ready

↓

Find Goroutine
```

---

Bước 7.

Đánh dấu Goroutine.

```
Runnable
```

---

Bước 8.

Đưa trở lại Run Queue.

```
Runnable

↓

Local Queue
```

Scheduler sẽ tiếp tục chạy Goroutine này.

---

# 10.31 Điều gì xảy ra với Machine?

Một điểm rất quan trọng.

Trong khi Goroutine chờ dữ liệu.

```
G1

↓

Waiting
```

Machine **không** chờ theo.

Machine sẽ tiếp tục chạy.

```
G2

↓

Running
```

Sau đó.

```
G3

↓

Running
```

```
G4

↓

Running
```

Machine luôn cố gắng sử dụng CPU để chạy các Goroutine khác.

Đây chính là điểm mạnh của Goroutine.

Nếu sử dụng mô hình One Thread Per Connection.

```
Thread

↓

Waiting

↓

CPU Idle
```

Trong Go.

```
Goroutine Waiting

↓

Machine vẫn chạy Goroutine khác
```

---

# 10.32 Ví dụ với 100.000 kết nối

Giả sử một Server có.

```
100.000 TCP Connections
```

Nếu dùng:

```
1 Connection

↓

1 Thread
```

thì cần.

```
100.000 Threads
```

Điều này gần như không khả thi.

Go làm khác.

```
100.000 Connections

↓

100.000 Goroutines

↓

Netpoller

↓

8 Machines

↓

8 CPU
```

Phần lớn thời gian.

100.000 Goroutine đều ở trạng thái:

```
Waiting
```

Chỉ khi có dữ liệu.

```
Socket Ready

↓

Runnable
```

Scheduler mới chạy Goroutine tương ứng.

Đây là lý do Go có thể xử lý hàng trăm nghìn kết nối đồng thời với số lượng Thread rất nhỏ.

---

# 10.33 Netpoller và Scheduler phối hợp như thế nào?

Netpoller **không trực tiếp chạy Goroutine**.

Đây là điểm rất quan trọng.

Netpoller chỉ làm nhiệm vụ.

```
Waiting

↓

Runnable
```

Sau đó.

```
Runnable

↓

Run Queue
```

Việc quyết định khi nào Goroutine được CPU thực thi vẫn thuộc về Scheduler.

Có thể mô tả bằng sơ đồ.

```
Network Event

↓

Kernel

↓

Netpoller

↓

Runnable

↓

Run Queue

↓

Scheduler

↓

CPU
```

Như vậy, Netpoller và Scheduler có trách nhiệm hoàn toàn khác nhau.

| Thành phần | Trách nhiệm |
|------------|-------------|
| Netpoller | Phát hiện I/O đã sẵn sàng |
| Scheduler | Chọn Goroutine để chạy |

---

# 10.34 Lợi ích của Netpoller

Netpoller mang lại rất nhiều lợi ích.

## Không cần Polling liên tục

CPU không phải kiểm tra hàng nghìn Socket mỗi giây.

---

## Giảm số lượng Thread

Một số ít Machine có thể phục vụ rất nhiều kết nối.

---

## Tận dụng CPU tốt hơn

Machine không phải chờ I/O.

Trong khi một Goroutine chờ dữ liệu.

Machine có thể chạy hàng nghìn Goroutine khác.

---

## Khả năng mở rộng cao

Đây là nền tảng giúp Go xây dựng hiệu quả:

- HTTP Server
- RPC Framework
- Reverse Proxy
- Gateway
- Message Broker
- Streaming System

---

# 10.35 Tóm tắt

Netpoller là thành phần chịu trách nhiệm theo dõi các sự kiện I/O và phối hợp với Scheduler để đánh thức các Goroutine đang chờ dữ liệu.

Khi một Goroutine thực hiện thao tác I/O:

1. Goroutine chuyển sang trạng thái **Waiting**.
2. Machine tiếp tục chạy các Goroutine khác.
3. Hệ điều hành theo dõi Socket hoặc File Descriptor.
4. Khi I/O hoàn thành, Kernel gửi thông báo.
5. Netpoller nhận sự kiện và chuyển Goroutine sang trạng thái **Runnable**.
6. Scheduler đưa Goroutine vào Run Queue và thực thi khi đến lượt.

Nhờ Netpoller, Go tránh được việc tạo một Thread cho mỗi kết nối và không cần Polling liên tục. Đây là một trong những thành phần quan trọng nhất giúp Go xử lý hiệu quả các ứng dụng mạng có mức độ đồng thời rất cao.

---

---

# 10.36 Asynchronous Preemption - Làm thế nào Go ngăn một Goroutine chiếm CPU mãi mãi?

Cho đến thời điểm này, chúng ta đã hiểu cách Scheduler:

- Phân phối Goroutine.
- Cân bằng tải bằng Work Stealing.
- Xử lý Blocking System Call.
- Đánh thức Goroutine thông qua Netpoller.

Tuy nhiên vẫn còn một tình huống rất nguy hiểm.

Hãy xem đoạn chương trình sau.

```go
func worker() {

    for {

    }

}
```

Sau đó:

```go
go worker()

go taskA()

go taskB()

go taskC()
```

Nếu `worker()` chiếm CPU mãi mãi thì điều gì sẽ xảy ra?

```
worker()

↓

Running

↓

Forever
```

Các Goroutine khác:

```
taskA()

taskB()

taskC()
```

sẽ không bao giờ được thực thi.

Scheduler gần như bị "đóng băng".

Go Runtime cần một cơ chế để giành lại CPU từ Goroutine đang chạy quá lâu.

Đó chính là **Preemption**.

---

# 10.37 Preemption là gì?

Preemption có nghĩa là:

> **Runtime chủ động tạm dừng một Goroutine đang chạy để nhường CPU cho Goroutine khác.**

Điều quan trọng là:

Việc tạm dừng này **không cần Goroutine tự nguyện nhường CPU**.

Ví dụ.

```
Running

↓

Runtime Interrupt

↓

Save Context

↓

Run Another Goroutine
```

Sau này.

```
Restore Context

↓

Continue
```

Goroutine tiếp tục chạy như chưa từng bị dừng.

---

# 10.38 Tại sao cần Preemption?

Giả sử có hai Goroutine.

```
G1
```

và

```
G2
```

G1 thực hiện.

```go
for {

}
```

Không có:

- Channel
- Mutex
- Sleep
- Function Call

Nếu Runtime không có Preemption.

```
CPU

↓

G1

↓

Forever
```

G2 sẽ không bao giờ có cơ hội chạy.

Điều này phá vỡ hoàn toàn tính công bằng của Scheduler.

---

# 10.39 Cooperative Scheduling

Ở những phiên bản Go đầu tiên, Scheduler sử dụng cơ chế gọi là:

```
Cooperative Scheduling
```

Ý tưởng rất đơn giản.

Runtime chỉ chuyển Goroutine khi Goroutine **tự nguyện nhường CPU**.

Ví dụ.

```go
channel <- value
```

hoặc.

```go
time.Sleep()
```

hoặc.

```go
mutex.Lock()
```

hoặc.

```go
function call
```

Khi gặp các điểm này.

Runtime mới có cơ hội chuyển sang Goroutine khác.

---

## Vấn đề của Cooperative Scheduling

Giả sử.

```go
func worker() {

    for {

    }

}
```

Không có:

- Function Call
- Channel
- Sleep
- Syscall

Runtime gần như không có cơ hội giành lại CPU.

```
worker()

↓

Forever
```

Đây là nhược điểm lớn nhất của Cooperative Scheduling.

---

# 10.40 Asynchronous Preemption

Để khắc phục nhược điểm trên, từ Go 1.14 Runtime giới thiệu:

```
Asynchronous Preemption
```

Ý tưởng là.

Runtime **không cần chờ Goroutine tự nhường CPU**.

Thay vào đó.

Runtime chủ động:

```
Interrupt

↓

Pause

↓

Save Context

↓

Run Another Goroutine
```

Điều này giúp Scheduler luôn duy trì được tính công bằng.

---

# 10.41 Minh họa

Giả sử.

```
CPU

↓

Running G1
```

Sau một khoảng thời gian.

Runtime quyết định.

```
Time Slice

↓

Expired
```

Runtime sẽ:

```text
G1

↓

Pause

↓

Save Registers

↓

Save Stack Pointer

↓

Save Program Counter
```

Sau đó.

```
Run G2
```

Một lúc sau.

```
Restore G1

↓

Continue
```

G1 tiếp tục từ đúng vị trí trước đó.

---

# 10.42 Goroutine không biết mình bị dừng

Đây là một điểm rất thú vị.

Ví dụ.

```go
for i := 0; i < 1000000; i++ {

    sum += i

}
```

Runtime có thể tạm dừng.

```
i = 527843
```

Sau đó chạy Goroutine khác.

Một lúc sau.

```
Continue

↓

i = 527844
```

Đối với lập trình viên.

Không có gì thay đổi.

Goroutine hoàn toàn không biết mình từng bị Scheduler tạm dừng.

---

# 10.43 Runtime lưu những gì khi Preemption?

Để có thể tiếp tục thực thi chính xác, Runtime phải lưu lại toàn bộ trạng thái của Goroutine.

Bao gồm.

```
Program Counter
```

```
Stack Pointer
```

```
Registers
```

```
Stack
```

```
Scheduler State
```

Có thể hình dung.

```text
Running

↓

Save Context

↓

Scheduler

↓

Restore Context

↓

Continue
```

Điều này tương tự như Context Switch giữa các Thread, nhưng được thực hiện ở mức Goroutine.

---

# 10.44 Khi nào Runtime thực hiện Preemption?

Runtime không dừng Goroutine một cách ngẫu nhiên.

Một số điều kiện phổ biến bao gồm:

- Goroutine đã chạy quá lâu.
- Scheduler cần CPU cho Goroutine khác.
- Garbage Collector cần tất cả Goroutine đạt đến trạng thái an toàn (Safe Point).

Mục tiêu là đảm bảo:

- Không có Goroutine nào chiếm CPU quá lâu.
- Scheduler luôn duy trì khả năng phản hồi.
- Garbage Collector có thể hoạt động hiệu quả.

---

# 10.45 Preemption và Safe Point

Không phải mọi vị trí trong chương trình đều có thể tạm dừng ngay lập tức.

Ví dụ.

```
Updating Pointer
```

Nếu Runtime dừng đúng lúc này.

Garbage Collector có thể nhìn thấy dữ liệu không nhất quán.

Do đó Runtime sẽ chỉ thực hiện Preemption tại các **Safe Point**.

Safe Point là vị trí mà:

- Stack ở trạng thái hợp lệ.
- Thanh ghi đã được đồng bộ.
- Garbage Collector có thể quét bộ nhớ một cách an toàn.

Nhờ vậy, việc tạm dừng Goroutine không làm hỏng trạng thái của chương trình.

---

# 10.46 Lợi ích của Asynchronous Preemption

Cơ chế này mang lại nhiều lợi ích quan trọng.

## Đảm bảo tính công bằng

Không Goroutine nào có thể chiếm CPU vô thời hạn.

---

## Tăng khả năng phản hồi

Các Goroutine mới không phải chờ quá lâu để được thực thi.

---

## Hỗ trợ Garbage Collector

Garbage Collector có thể yêu cầu Runtime tạm dừng hoặc đồng bộ các Goroutine tại Safe Point để thực hiện việc quét bộ nhớ.

---

## Không yêu cầu lập trình viên can thiệp

Lập trình viên không cần tự gọi:

```go
runtime.Gosched()
```

để nhường CPU trong hầu hết các trường hợp.

Scheduler sẽ tự động xử lý.

---

# 10.47 Ví dụ thực tế

Giả sử có hai Goroutine.

```go
go infiniteLoop()

go processOrders()
```

Nếu không có Preemption.

```
infiniteLoop()

↓

Running Forever
```

```
processOrders()

↓

Never Run
```

Khi có Asynchronous Preemption.

```text
infiniteLoop()

↓

Running

↓

Pause

↓

processOrders()

↓

Pause

↓

infiniteLoop()

↓

Continue
```

Hai Goroutine sẽ được luân phiên thực thi.

---

# 10.48 Những hiểu lầm phổ biến

## Hiểu lầm 1

> Runtime dừng Goroutine ở bất kỳ dòng lệnh nào.

Sai.

Runtime chỉ Preempt tại những thời điểm an toàn.

---

## Hiểu lầm 2

> Preemption làm mất dữ liệu.

Sai.

Runtime luôn lưu và khôi phục đầy đủ ngữ cảnh trước khi tiếp tục thực thi.

---

## Hiểu lầm 3

> Cần tự gọi `runtime.Gosched()` để Scheduler hoạt động.

Trong phần lớn trường hợp, điều này không còn cần thiết.

Scheduler hiện đại đã có Asynchronous Preemption để đảm bảo tính công bằng giữa các Goroutine.

---

# 10.49 Tóm tắt

Asynchronous Preemption là cơ chế cho phép Go Runtime chủ động giành lại CPU từ một Goroutine đang chạy quá lâu mà không cần Goroutine tự nguyện nhường quyền thực thi.

Quy trình tổng quát:

1. Runtime phát hiện Goroutine đã sử dụng CPU đủ lâu hoặc cần nhường CPU.
2. Chờ đến một Safe Point thích hợp.
3. Lưu toàn bộ ngữ cảnh thực thi của Goroutine.
4. Chuyển CPU cho Goroutine khác.
5. Khi đến lượt, khôi phục ngữ cảnh và tiếp tục thực thi từ đúng vị trí trước đó.

Nhờ cơ chế này, Go Scheduler duy trì được tính công bằng, tăng khả năng phản hồi của hệ thống và tạo điều kiện cho Garbage Collector hoạt động hiệu quả mà không yêu cầu lập trình viên phải chủ động nhường CPU.

---

---

# 10.50 GOMAXPROCS - Có bao nhiêu Goroutine thực sự chạy cùng lúc?

Sau khi tìm hiểu toàn bộ Scheduler, nhiều người thường đặt ra câu hỏi:

> **Nếu một chương trình tạo 1.000.000 Goroutine thì có phải 1.000.000 Goroutine sẽ chạy đồng thời?**

Câu trả lời là:

**Không.**

Trong Go, số lượng Goroutine **đang thực sự được CPU thực thi tại cùng một thời điểm** luôn bị giới hạn bởi số lượng **Processor (P)**.

Và số lượng Processor lại được quyết định bởi:

```
GOMAXPROCS
```

---

# 10.51 GOMAXPROCS là gì?

`GOMAXPROCS` xác định:

> **Số lượng Processor (P) mà Go Runtime được phép tạo.**

Hay nói cách khác:

> **Đây là số lượng Goroutine tối đa có thể thực thi song song trên CPU tại cùng một thời điểm.**

Ví dụ:

```
GOMAXPROCS = 8
```

Runtime sẽ tạo:

```
P0

P1

P2

P3

P4

P5

P6

P7
```

Có tổng cộng:

```
8 Processor
```

Điều đó đồng nghĩa:

```
Maximum Running Goroutines = 8
```

Dù chương trình có:

```
10

100

1000

1.000.000 Goroutines
```

thì cũng chỉ có tối đa:

```
8 Goroutines
```

được CPU thực thi cùng lúc.

Các Goroutine còn lại sẽ ở trạng thái:

```
Runnable

↓

Run Queue
```

để chờ đến lượt.

---

# 10.52 GOMAXPROCS có phải số CPU không?

Đây là hiểu lầm rất phổ biến.

Nhiều người nghĩ:

```
GOMAXPROCS

=

Number of CPUs
```

Điều này **không hoàn toàn đúng**.

Mặc định, Go Runtime sẽ đặt:

```
GOMAXPROCS

=

Logical CPUs
```

của hệ thống.

Ví dụ.

Máy tính có:

```
8 Physical Cores

16 Logical CPUs
```

Go sẽ mặc định:

```
GOMAXPROCS = 16
```

Nhưng lập trình viên hoàn toàn có thể thay đổi.

Ví dụ:

```go
runtime.GOMAXPROCS(4)
```

Khi đó Runtime chỉ tạo:

```
4 Processor
```

mặc dù máy vẫn có:

```
16 Logical CPUs
```

---

# 10.53 Ví dụ trực quan

Giả sử máy có:

```
8 CPU
```

Ta tạo:

```go
for i := 0; i < 1000; i++ {

    go worker()

}
```

Nếu:

```
GOMAXPROCS = 8
```

thì:

```
Running

↓

8 Goroutines
```

```
Runnable

↓

992 Goroutines
```

Scheduler sẽ liên tục luân phiên các Goroutine trong Run Queue.

---

Nếu:

```
GOMAXPROCS = 1
```

thì:

```
Running

↓

1 Goroutine
```

```
Runnable

↓

999 Goroutines
```

Toàn bộ chương trình sẽ chỉ sử dụng một CPU để thực thi Go code.

---

# 10.54 Thay đổi GOMAXPROCS

Go cho phép thay đổi giá trị này trong lúc chương trình đang chạy.

Ví dụ:

```go
package main

import (
    "runtime"
)

func main() {

    runtime.GOMAXPROCS(4)

}
```

Sau khi gọi.

```
Processor

↓

Resize
```

Runtime sẽ điều chỉnh số lượng Processor tương ứng.

Tuy nhiên, việc thay đổi thường xuyên không được khuyến khích vì Scheduler cần thời gian để tái cân bằng.

---

# 10.55 Khi nào nên thay đổi GOMAXPROCS?

Trong phần lớn trường hợp.

**Không cần thay đổi.**

Go Runtime đã tự động lựa chọn giá trị phù hợp với số CPU của hệ thống.

Chỉ nên thay đổi trong một số tình huống đặc biệt.

---

## Trường hợp 1 - Giới hạn tài nguyên

Ví dụ.

Máy có:

```
32 CPUs
```

Nhưng chương trình chỉ muốn sử dụng:

```
8 CPUs
```

Có thể đặt.

```
GOMAXPROCS = 8
```

---

## Trường hợp 2 - Chạy nhiều ứng dụng

Ví dụ.

Một Server chạy:

- Database
- Kafka
- Redis
- Go Service

Ta không muốn Go Service sử dụng toàn bộ CPU.

Có thể giới hạn.

```
GOMAXPROCS
```

để nhường tài nguyên cho các ứng dụng khác.

---

## Trường hợp 3 - Benchmark

Khi Benchmark.

Chúng ta thường thử.

```
1

2

4

8

16
```

Processor.

để xem hiệu năng thay đổi như thế nào.

---

# 10.56 GOMAXPROCS có làm chương trình nhanh hơn không?

Không phải lúc nào.

Đây cũng là một hiểu lầm rất phổ biến.

Ví dụ.

```
1 CPU
```

↓

```
2 CPU
```

↓

```
4 CPU
```

↓

```
8 CPU
```

Không có nghĩa chương trình sẽ nhanh gấp:

```
2x

4x

8x
```

Hiệu năng còn phụ thuộc vào:

- Loại công việc.
- Đồng bộ giữa Goroutine.
- Lock Contention.
- Memory Bandwidth.
- Cache Miss.
- Garbage Collector.

Một số chương trình gần như không tăng tốc khi tăng `GOMAXPROCS`.

---

# 10.57 CPU-bound và IO-bound

Hiệu quả của `GOMAXPROCS` phụ thuộc rất nhiều vào đặc điểm của chương trình.

## CPU-bound

Ví dụ.

- Tính toán.
- Mã hóa.
- Machine Learning.
- Image Processing.

Những chương trình này thường hưởng lợi khi tăng `GOMAXPROCS`.

---

## IO-bound

Ví dụ.

- HTTP Server.
- Database Client.
- RPC Service.

Phần lớn thời gian Goroutine ở trạng thái:

```
Waiting
```

Tăng `GOMAXPROCS` quá nhiều đôi khi không mang lại khác biệt đáng kể.

---

# 10.58 Những hiểu lầm phổ biến

## Hiểu lầm 1

> Tăng `GOMAXPROCS` càng cao càng tốt.

Sai.

Quá nhiều Processor có thể làm tăng:

- Context Switching.
- Lock Contention.
- Cache Miss.

---

## Hiểu lầm 2

> `GOMAXPROCS` là số lượng Thread.

Sai.

`GOMAXPROCS` quyết định số lượng **Processor (P)**.

Số lượng Machine (OS Thread) có thể nhiều hơn.

---

## Hiểu lầm 3

> Một Goroutine tương ứng với một Processor.

Sai.

Một Processor có thể lần lượt thực thi hàng triệu Goroutine.

---

# 10.59 Tóm tắt

`GOMAXPROCS` là tham số quan trọng quyết định mức độ song song của Go Runtime.

Nó xác định số lượng **Processor (P)** được tạo, từ đó giới hạn số lượng Goroutine có thể thực sự chạy trên CPU tại cùng một thời điểm.

Những điểm cần ghi nhớ:

- `GOMAXPROCS` **không phải** là số lượng Goroutine.
- `GOMAXPROCS` **không phải** là số lượng OS Thread.
- `GOMAXPROCS` xác định số lượng **Processor**.
- Mặc định, Go Runtime đặt giá trị này bằng số **Logical CPU** của hệ thống.
- Trong đa số trường hợp, không cần thay đổi giá trị mặc định.

Hiểu đúng vai trò của `GOMAXPROCS` sẽ giúp bạn lý giải vì sao một chương trình có thể tạo hàng triệu Goroutine nhưng chỉ có một số rất nhỏ trong đó thực sự được CPU thực thi cùng một thời điểm.

---

---

# 10.60 Trạng thái của Goroutine

Trong các phần trước, chúng ta thường nhắc đến các trạng thái như:

- Runnable
- Running
- Waiting

Nhưng chúng ta chưa tìm hiểu đầy đủ:

> **Một Goroutine có thể trải qua những trạng thái nào trong suốt vòng đời của nó?**

Việc hiểu các trạng thái này rất quan trọng vì Scheduler hoạt động gần như hoàn toàn dựa trên việc chuyển đổi giữa các trạng thái.

---

# 10.61 Vòng đời của một Goroutine

Một Goroutine không chỉ đơn giản là:

```
Create

↓

Run

↓

Finish
```

Thực tế vòng đời của nó phức tạp hơn nhiều.

Có thể hình dung như sau.

```text
            New

             │

             ▼

         Runnable

             │

             ▼

         Running

        ╱     │     ╲

       ▼      ▼      ▼

 Waiting   Runnable   Dead

       ▲

       │

       └───────────────
```

Trong suốt thời gian tồn tại, một Goroutine có thể nhiều lần chuyển đổi giữa:

```
Runnable

↓

Running

↓

Waiting

↓

Runnable

↓

Running
```

cho đến khi kết thúc.

---

# 10.62 Trạng thái New

Ngay sau khi lập trình viên gọi.

```go
go worker()
```

Runtime bắt đầu:

- Cấp phát đối tượng `G`.
- Khởi tạo Stack.
- Khởi tạo Scheduler Metadata.

Trong khoảng thời gian rất ngắn này.

Goroutine đang ở trạng thái:

```
New
```

Đây chỉ là trạng thái nội bộ trong quá trình khởi tạo.

Sau khi hoàn tất.

Runtime chuyển ngay sang:

```
Runnable
```

---

# 10.63 Trạng thái Runnable

Đây là trạng thái xuất hiện nhiều nhất.

Runnable có nghĩa là.

> **Goroutine đã sẵn sàng chạy nhưng chưa được CPU thực thi.**

Có thể hình dung.

```
Runnable

↓

Waiting CPU
```

Scheduler sẽ đưa Goroutine vào:

```
Local Queue
```

hoặc.

```
Global Queue
```

để chờ đến lượt.

Điều quan trọng là:

Runnable **không có nghĩa là đang chạy**.

Nó chỉ có nghĩa:

```
Ready to Run
```

---

# 10.64 Trạng thái Running

Khi Scheduler chọn Goroutine.

```
Runnable

↓

Running
```

Lúc này.

Machine bắt đầu thực thi các lệnh của Goroutine trên CPU.

```
CPU

↓

Execute Instructions
```

Một Goroutine chỉ ở trạng thái Running khi:

- Đã được Scheduler chọn.
- Có Processor.
- Có Machine.

Đây là trạng thái duy nhất thực sự sử dụng CPU.

---

# 10.65 Trạng thái Waiting

Trong quá trình thực thi.

Goroutine có thể phải chờ.

Ví dụ.

```go
<-ch
```

hoặc.

```go
mutex.Lock()
```

hoặc.

```go
time.Sleep()
```

hoặc.

```go
conn.Read()
```

Lúc này.

```
Running

↓

Waiting
```

Scheduler sẽ:

- Lưu Context.
- Gỡ Goroutine khỏi CPU.
- Chạy Goroutine khác.

Điểm rất quan trọng là.

```
Waiting

≠

Running
```

Waiting Goroutine **không tiêu tốn CPU**.

---

# 10.66 Trạng thái Dead

Sau khi hàm kết thúc.

Ví dụ.

```go
func worker() {

    fmt.Println("Done")

}
```

Runtime chuyển.

```
Running

↓

Dead
```

Dead có nghĩa.

```
Finished
```

Scheduler sẽ không bao giờ chạy Goroutine này nữa.

Sau đó.

Runtime sẽ thu hồi các tài nguyên liên quan.

---

# 10.67 Ví dụ trực quan

Giả sử có chương trình.

```go
go processOrder()
```

Vòng đời của Goroutine.

```text
Create

↓

Runnable

↓

Running

↓

Waiting (Read DB)

↓

Runnable

↓

Running

↓

Waiting (Channel)

↓

Runnable

↓

Running

↓

Dead
```

Trong thực tế.

Một Goroutine có thể chuyển giữa:

```
Running

↓

Waiting

↓

Runnable
```

hàng trăm hoặc hàng nghìn lần.

---

# 10.68 Scheduler chỉ quan tâm Goroutine nào?

Một câu hỏi rất hay.

> Scheduler có quan tâm mọi Goroutine không?

Câu trả lời là.

**Không.**

Scheduler chủ yếu quan tâm tới:

```
Runnable
```

Đây là các Goroutine đang chờ CPU.

```
Runnable

↓

Scheduler

↓

CPU
```

Các Goroutine ở trạng thái Waiting sẽ được:

- Channel
- Mutex
- Timer
- Netpoller

đánh thức khi điều kiện chờ kết thúc.

Sau đó chúng mới trở lại Runnable.

---

# 10.69 Điều gì làm Goroutine chuyển trạng thái?

Có thể tóm tắt như sau.

| Từ | Sang | Nguyên nhân |
|-----|------|------------|
| New | Runnable | Khởi tạo hoàn tất |
| Runnable | Running | Scheduler chọn |
| Running | Waiting | Chờ Channel, Mutex, I/O, Sleep... |
| Waiting | Runnable | Điều kiện chờ kết thúc |
| Running | Runnable | Scheduler Preemption |
| Running | Dead | Hàm kết thúc |

Đây là sơ đồ quan trọng cần ghi nhớ.

---

# 10.70 Ví dụ với Channel

```go
func worker(ch chan int) {

    value := <-ch

    fmt.Println(value)

}
```

Quá trình.

```
Runnable

↓

Running

↓

Receive Channel

↓

Waiting
```

Một Goroutine khác.

```go
ch <- 100
```

Sau khi gửi.

```
Waiting

↓

Runnable
```

Scheduler sẽ chạy tiếp.

```
Running
```

---

# 10.71 Ví dụ với time.Sleep

```go
time.Sleep(time.Second)
```

Quá trình.

```
Running

↓

Waiting
```

Một giây sau.

```
Timer

↓

Runnable
```

Scheduler tiếp tục.

```
Running
```

---

# 10.72 Ví dụ với Network

```go
conn.Read(buf)
```

Quá trình.

```
Running

↓

Waiting
```

Sau khi dữ liệu đến.

```
Netpoller

↓

Runnable
```

Scheduler.

```
Running
```

---

# 10.73 Ví dụ với Mutex

```go
mutex.Lock()
```

Nếu Mutex đang bị giữ.

```
Running

↓

Waiting
```

Sau khi Goroutine khác.

```go
mutex.Unlock()
```

Runtime đánh thức.

```
Runnable
```

Scheduler tiếp tục.

```
Running
```

---

# 10.74 Tại sao phải hiểu các trạng thái này?

Trong thực tế.

Khi Debug một chương trình Go.

Bạn sẽ thường gặp.

```
goroutine 128 [running]
```

hoặc.

```
goroutine 201 [chan receive]
```

hoặc.

```
goroutine 77 [IO wait]
```

hoặc.

```
goroutine 65 [sleep]
```

Những thông tin này phản ánh trạng thái hiện tại của Goroutine.

Nếu hiểu vòng đời của Goroutine.

Bạn sẽ dễ dàng:

- Phân tích Goroutine Dump.
- Debug Deadlock.
- Debug Goroutine Leak.
- Phân tích Performance.

---

# 10.75 Tóm tắt

Trong suốt vòng đời của mình, một Goroutine liên tục chuyển đổi giữa nhiều trạng thái khác nhau.

Các trạng thái quan trọng nhất gồm:

- **Runnable**: đã sẵn sàng, đang chờ CPU.
- **Running**: đang được CPU thực thi.
- **Waiting**: đang chờ một sự kiện như Channel, Mutex, Timer hoặc I/O.
- **Dead**: đã hoàn thành và sẽ không được Scheduler chạy lại.

Scheduler chủ yếu làm việc với các Goroutine ở trạng thái **Runnable**, trong khi các thành phần như Channel, Mutex, Timer và Netpoller chịu trách nhiệm đánh thức Goroutine từ trạng thái **Waiting** và đưa chúng trở lại hàng đợi để tiếp tục thực thi.

Việc hiểu rõ vòng đời của Goroutine là nền tảng để phân tích hiệu năng, phát hiện Goroutine Leak và đọc các báo cáo Goroutine Dump trong các hệ thống Go thực tế.

---

---

# 10.76 Scheduler và Garbage Collector

Đến thời điểm này, chúng ta đã tìm hiểu gần như toàn bộ cơ chế lập lịch của Go Runtime.

Tuy nhiên vẫn còn một thành phần rất quan trọng.

Đó là **Garbage Collector (GC).**

Nhiều lập trình viên thường xem Scheduler và Garbage Collector là hai thành phần độc lập.

Thực tế hoàn toàn ngược lại.

> **Scheduler và Garbage Collector phối hợp với nhau rất chặt chẽ.**

Nếu Scheduler chịu trách nhiệm:

```
Ai được chạy?
```

thì Garbage Collector chịu trách nhiệm:

```
Bộ nhớ nào còn được sử dụng?
```

Hai thành phần này liên tục trao đổi thông tin trong suốt vòng đời của chương trình.

---

# 10.77 Tại sao Garbage Collector cần Scheduler?

Hãy tưởng tượng một chương trình đang có:

```
100.000 Goroutines
```

Mỗi Goroutine đều có:

- Stack
- Local Variables
- Pointer
- Heap Objects

Garbage Collector cần biết:

```
Object nào còn được tham chiếu?
```

Nếu một Goroutine đang chạy.

```
Running

↓

Updating Pointer
```

Trong đúng thời điểm Garbage Collector quét bộ nhớ.

Có thể xảy ra:

```
Object

↓

Incorrect State
```

Điều này có thể khiến Garbage Collector:

- Thu hồi nhầm Object.
- Hoặc giữ lại Object không còn sử dụng.

Cả hai đều rất nguy hiểm.

---

# 10.78 Safe Point

Để giải quyết vấn đề trên.

Go Runtime giới thiệu khái niệm:

```
Safe Point
```

Safe Point là vị trí mà tại đó:

- Stack đã nhất quán.
- Registers đã được đồng bộ.
- Pointer đều ở trạng thái hợp lệ.

Chỉ tại Safe Point.

Garbage Collector mới được phép:

```
Scan Stack
```

Có thể hình dung.

```text
Running

↓

Safe Point

↓

GC Scan

↓

Continue
```

Điều này đảm bảo GC luôn nhìn thấy trạng thái chính xác của Goroutine.

---

# 10.79 Scheduler giúp GC như thế nào?

Trong quá trình Garbage Collection.

Runtime cần tất cả Goroutine đều đi qua Safe Point.

Scheduler sẽ phối hợp theo trình tự.

```
Running Goroutine

↓

Reach Safe Point

↓

GC Scan

↓

Continue Running
```

Nếu Goroutine chưa đến Safe Point.

Scheduler sẽ chờ.

```
Waiting Safe Point
```

Sau khi tất cả Goroutine đã an toàn.

Garbage Collector mới bắt đầu quét Heap.

---

# 10.80 Stop The World (STW)

Có lẽ bạn đã từng nghe đến khái niệm:

```
Stop The World
```

Nhiều người nghĩ.

> Khi Garbage Collector chạy thì toàn bộ chương trình sẽ dừng rất lâu.

Điều này đúng với các Runtime cũ.

Nhưng không còn đúng với Go Runtime hiện đại.

Go chỉ dừng chương trình trong một khoảng thời gian rất ngắn.

Mục đích của Stop The World là:

- Đồng bộ trạng thái.
- Đảm bảo mọi Goroutine ở Safe Point.
- Chuẩn bị cho quá trình Garbage Collection.

Sau đó.

Phần lớn công việc của Garbage Collector diễn ra đồng thời với chương trình.

```
Application Running

+

Garbage Collector Running
```

Đây gọi là:

```
Concurrent Garbage Collection
```

---

# 10.81 Scheduler trong thời gian GC

Trong khi Garbage Collector hoạt động.

Scheduler **không dừng hoàn toàn**.

Thay vào đó.

Scheduler vẫn tiếp tục lập lịch.

```text
Runnable

↓

Running

↓

Runnable

↓

Running
```

Tuy nhiên.

Một phần CPU sẽ được dành cho Garbage Collector.

Có thể hình dung.

```
8 Processor
```

↓

```
6 Processor

↓

Application
```

```
2 Processor

↓

Garbage Collector
```

Tỷ lệ này được Runtime tự động điều chỉnh dựa trên tải của hệ thống.

---

# 10.82 Mark Worker cũng là Goroutine

Một điều khá thú vị.

Garbage Collector của Go **không phải là một Thread đặc biệt**.

Trong nhiều giai đoạn.

GC cũng sử dụng:

```
Goroutine
```

để thực hiện việc đánh dấu (Mark).

Có thể hình dung.

```
Application Goroutine

↓

Running
```

```
GC Worker

↓

Running
```

Scheduler xem cả hai đều là Goroutine.

Điểm khác biệt chỉ là:

```
GC Worker

↓

Special Priority
```

---

# 10.83 Scheduler phải cân bằng CPU

Giả sử.

Garbage Collector sử dụng toàn bộ CPU.

```
GC

↓

100%
```

Khi đó.

Application gần như không thể chạy.

Ngược lại.

Nếu Application sử dụng:

```
100%
```

Garbage Collector sẽ không có cơ hội thu hồi bộ nhớ.

Runtime phải cân bằng.

```
Application

↓

CPU
```

```
GC

↓

CPU
```

Điều này được thực hiện thông qua Scheduler.

---

# 10.84 Goroutine Assist GC

Một đặc điểm thú vị của Go Runtime là:

Không phải lúc nào Garbage Collector cũng tự làm toàn bộ công việc.

Trong một số trường hợp.

Goroutine đang cấp phát bộ nhớ cũng sẽ phải hỗ trợ Garbage Collector.

Khái niệm này gọi là:

```
GC Assist
```

Ví dụ.

```
Allocate Memory

↓

Heap quá nhanh

↓

Application Goroutine

↓

Assist GC

↓

Continue
```

Điều này giúp Garbage Collector không bị tụt quá xa so với tốc độ cấp phát bộ nhớ của chương trình.

---

# 10.85 Scheduler và GC phối hợp để giảm Pause Time

Mục tiêu của Go Runtime là:

```
GC Pause

↓

As Small As Possible
```

Scheduler đóng vai trò rất quan trọng trong việc đạt mục tiêu này.

Thông qua:

- Safe Point.
- Concurrent Execution.
- GC Worker Scheduling.
- GC Assist.

Go Runtime đã giảm thời gian dừng chương trình từ hàng trăm mili giây (ở các phiên bản rất cũ) xuống chỉ còn mức rất nhỏ trong hầu hết các ứng dụng hiện đại.

Đây là một trong những cải tiến lớn nhất của Go Runtime qua nhiều phiên bản.

---

# 10.86 Tóm tắt

Scheduler và Garbage Collector không hoạt động độc lập mà phối hợp chặt chẽ với nhau trong suốt quá trình thực thi chương trình.

Vai trò của Scheduler bao gồm:

- Đưa Goroutine đến Safe Point.
- Phân phối CPU giữa Application và Garbage Collector.
- Lập lịch cho các GC Worker.
- Hỗ trợ cơ chế GC Assist khi cần.

Trong khi đó, Garbage Collector dựa vào Scheduler để đảm bảo việc quét bộ nhớ diễn ra trên trạng thái nhất quán của chương trình.

Sự phối hợp này giúp Go vừa duy trì hiệu năng thực thi cao, vừa giữ thời gian tạm dừng (Pause Time) ở mức rất thấp, ngay cả đối với các hệ thống có hàng trăm nghìn Goroutine đang hoạt động.

---

---

# 10.87 Scheduler Performance

Đến thời điểm này, chúng ta đã hiểu gần như toàn bộ cơ chế hoạt động của Go Scheduler.

Một câu hỏi tự nhiên được đặt ra là:

> **Tại sao Go Scheduler lại có hiệu năng cao như vậy?**

Nếu nhìn từ bên ngoài, Scheduler dường như chỉ làm một việc đơn giản:

- Chọn Goroutine.
- Chạy Goroutine.
- Chuyển sang Goroutine khác.

Nhưng trên thực tế, để đạt được hiệu năng cao trên các hệ thống nhiều CPU, Go Runtime đã áp dụng rất nhiều kỹ thuật tối ưu.

Trong phần này, chúng ta sẽ phân tích những yếu tố quan trọng nhất ảnh hưởng đến hiệu năng của Scheduler.

---

# 10.87.1 Mục tiêu của Scheduler

Một Scheduler tốt không phải là Scheduler chạy nhanh nhất.

Mà là Scheduler đáp ứng đồng thời nhiều mục tiêu.

- Chi phí lập lịch thấp.
- Tận dụng tối đa CPU.
- Đảm bảo tính công bằng.
- Có khả năng mở rộng.
- Không trở thành nút thắt của hệ thống.

Có thể mô tả bằng công thức đơn giản.

```
Scheduler Overhead

↓

As Small As Possible

CPU Utilization

↓

As High As Possible
```

---

# 10.87.2 Chi phí của một lần Scheduling

Mỗi lần Scheduler chuyển từ Goroutine này sang Goroutine khác đều có chi phí.

Ví dụ.

```
G1

↓

Pause

↓

Save Context

↓

Select G2

↓

Restore Context

↓

Run G2
```

Các bước này bao gồm:

- Lưu Program Counter.
- Lưu Stack Pointer.
- Lưu Registers.
- Cập nhật trạng thái Goroutine.
- Cập nhật Queue.

Nếu việc chuyển đổi diễn ra quá thường xuyên.

```
Scheduling

>>

Business Logic
```

thì phần lớn CPU sẽ dành cho Scheduler thay vì xử lý nghiệp vụ.

Đây là điều Runtime luôn cố gắng tránh.

---

# 10.87.3 Local Queue giúp Scheduler nhanh hơn

Một trong những quyết định thiết kế quan trọng nhất là:

> **Mỗi Processor có Local Queue riêng.**

Giả sử chỉ có Global Queue.

```
64 Processor

↓

1 Queue
```

Mọi Processor đều phải:

```
Lock

↓

Pop

↓

Unlock
```

Khi số CPU tăng.

Lock Contention cũng tăng theo.

Ngược lại.

```
P0

↓

Local Queue
```

```
P1

↓

Local Queue
```

```
P2

↓

Local Queue
```

Phần lớn thao tác đều diễn ra trên Queue cục bộ.

Không cần đồng bộ với Processor khác.

Đây là yếu tố giúp Scheduler mở rộng rất tốt trên hệ thống nhiều lõi.

---

# 10.87.4 Cache Locality

CPU hiện đại có nhiều tầng bộ nhớ.

```
Register

↓

L1 Cache

↓

L2 Cache

↓

L3 Cache

↓

RAM
```

Truy cập:

```
Register

↓

Rất nhanh
```

Trong khi truy cập:

```
RAM

↓

Chậm hơn rất nhiều
```

Nếu một Processor liên tục chạy các Goroutine trong Local Queue của chính mình.

Dữ liệu sẽ có xu hướng nằm sẵn trong Cache của CPU.

Điều này gọi là:

```
Cache Locality
```

Scheduler cố gắng giữ Goroutine trên cùng một Processor càng lâu càng tốt để tận dụng lợi thế này.

---

# 10.87.5 Giảm Lock Contention

Một nguyên tắc quan trọng trong thiết kế Scheduler là:

> **Giảm tối đa việc sử dụng Lock.**

Các thành phần như:

- Local Queue
- mcache
- Timer Cache

đều được đặt trong Processor thay vì dùng chung.

Nhờ đó.

Phần lớn thao tác không cần khóa.

Chỉ trong một số trường hợp đặc biệt như:

- Global Queue.
- Work Stealing.
- Điều chỉnh số lượng Processor.

Runtime mới cần đồng bộ giữa nhiều Processor.

---

# 10.87.6 Work Stealing chỉ diễn ra khi cần

Nếu mỗi vòng lặp Scheduler đều thực hiện Work Stealing.

```
Scheduler

↓

Random Processor

↓

Steal

↓

Repeat
```

Chi phí sẽ rất lớn.

Thay vào đó.

Scheduler luôn ưu tiên:

```
Local Queue
```

Chỉ khi Local Queue rỗng.

Runtime mới thực hiện Work Stealing.

Điều này giúp giảm đáng kể chi phí lập lịch.

---

# 10.87.7 Goroutine rất nhẹ

Một yếu tố khác góp phần vào hiệu năng của Scheduler là:

```
Goroutine

↓

Small Stack
```

Ban đầu.

Một Goroutine chỉ có Stack khoảng:

```
≈ 2 KB
```

Trong khi một OS Thread thường cần Stack lớn hơn rất nhiều.

Nhờ đó.

Runtime có thể quản lý hàng triệu Goroutine mà không tiêu tốn quá nhiều bộ nhớ.

---

# 10.87.8 Scheduler tránh tạo Thread mới

Việc tạo một OS Thread tương đối tốn kém.

Nó yêu cầu:

- Tạo Kernel Object.
- Cấp phát Stack.
- Khởi tạo Thread Context.

Do đó Runtime luôn ưu tiên:

```
Reuse Machine
```

Chỉ khi thật sự cần thiết.

Runtime mới tạo thêm Machine mới.

Điều này giúp giảm chi phí của hệ thống.

---

# 10.87.9 Khi nào Scheduler trở thành Bottleneck?

Mặc dù được tối ưu rất tốt.

Scheduler vẫn có giới hạn.

Một số tình huống có thể làm Scheduler hoạt động kém hiệu quả.

### Tạo quá nhiều Goroutine

Ví dụ.

```go
for i := 0; i < 10000000; i++ {

    go worker()

}
```

Việc tạo hàng chục triệu Goroutine trong thời gian rất ngắn sẽ làm tăng:

- Chi phí cấp phát.
- Chi phí lập lịch.
- Áp lực lên Garbage Collector.

---

### Goroutine sống quá ngắn

Ví dụ.

```go
go func() {

    x++

}()
```

Nếu Goroutine chỉ thực hiện một vài phép tính rồi kết thúc.

Chi phí:

- Tạo Goroutine.
- Đưa vào Queue.
- Scheduler.
- Thu hồi.

có thể lớn hơn chính công việc mà Goroutine thực hiện.

Trong trường hợp này.

Việc sử dụng Goroutine không còn hiệu quả.

---

### Quá nhiều đồng bộ

Ví dụ.

```go
mutex.Lock()

...

mutex.Unlock()
```

Nếu hàng nghìn Goroutine liên tục tranh chấp cùng một Mutex.

Scheduler sẽ phải chuyển đổi trạng thái giữa:

```
Running

↓

Waiting

↓

Runnable
```

rất thường xuyên.

Hiệu năng sẽ giảm đáng kể.

---

### Blocking kéo dài

Nếu Goroutine thường xuyên thực hiện các Blocking System Call.

Runtime phải:

- Tách Processor.
- Gắn Processor sang Machine khác.
- Quản lý nhiều Machine hơn.

Mặc dù Runtime xử lý tốt tình huống này, nhưng đây vẫn là một chi phí cần thiết.

---

# 10.87.10 Scheduler không thể làm mọi thứ nhanh hơn

Một điều rất quan trọng cần nhớ là:

> Scheduler chỉ quyết định **ai được chạy**.

Nó không làm cho mã nguồn của bạn chạy nhanh hơn.

Ví dụ.

```go
func calculate() {

    for i := 0; i < 1000000000; i++ {

    }

}
```

Nếu thuật toán chậm.

Scheduler không thể tối ưu thay cho lập trình viên.

Hiệu năng của hệ thống luôn là kết quả của nhiều yếu tố:

- Thuật toán.
- Cấu trúc dữ liệu.
- Scheduler.
- Garbage Collector.
- Memory Allocator.
- Cache.
- Hệ điều hành.

---

# 10.87.11 Những nguyên tắc giúp Scheduler đạt hiệu năng cao

Có thể tóm tắt các nguyên tắc thiết kế của Go Scheduler như sau.

| Nguyên tắc | Mục tiêu |
|------------|----------|
| Local Queue | Giảm Lock Contention |
| Work Stealing | Cân bằng tải |
| Processor | Tách Scheduler khỏi Thread |
| Netpoller | Không Block Thread khi chờ I/O |
| Asynchronous Preemption | Đảm bảo tính công bằng |
| Small Stack | Giảm chi phí bộ nhớ |
| Reuse Machine | Giảm chi phí tạo Thread |
| Concurrent GC | Giảm Pause Time |

Đây là những yếu tố phối hợp với nhau để tạo nên hiệu năng tổng thể của Go Runtime.

---

# 10.87.12 Tóm tắt

Hiệu năng của Go Scheduler không đến từ một thuật toán duy nhất mà là kết quả của nhiều quyết định thiết kế được kết hợp chặt chẽ.

Những yếu tố quan trọng nhất bao gồm:

- Mỗi Processor có Local Run Queue riêng để giảm tranh chấp.
- Work Stealing giúp cân bằng tải giữa các Processor.
- Processor tách trạng thái Scheduler khỏi OS Thread.
- Netpoller cho phép xử lý số lượng lớn tác vụ I/O mà không cần nhiều Thread.
- Asynchronous Preemption đảm bảo mọi Goroutine đều có cơ hội được thực thi.
- Goroutine nhẹ và có Stack tự mở rộng giúp giảm chi phí bộ nhớ.
- Runtime ưu tiên tái sử dụng Machine thay vì liên tục tạo Thread mới.

Nhờ những cơ chế này, Go Scheduler có thể mở rộng hiệu quả trên các hệ thống đa lõi và duy trì hiệu năng ổn định ngay cả khi chương trình có hàng trăm nghìn hoặc hàng triệu Goroutine đang hoạt động.

---

---

# 10.89 Những hiểu lầm phổ biến về Go Scheduler

Go Scheduler là một trong những thành phần phức tạp nhất của Go Runtime.

Rất nhiều lập trình viên sử dụng Goroutine hằng ngày nhưng vẫn có những hiểu lầm về cách Scheduler hoạt động.

Trong phần này chúng ta sẽ tổng hợp những quan niệm sai phổ biến nhất và giải thích nguyên nhân.

---

# Hiểu lầm 1 - Goroutine chính là Thread

Đây có lẽ là hiểu lầm phổ biến nhất.

Nhiều người nghĩ:

```
1 Goroutine

=

1 OS Thread
```

Điều này hoàn toàn sai.

Thực tế.

```
100.000 Goroutines

↓

8 Processor

↓

12 Machines

↓

8 CPU
```

Một Machine có thể lần lượt thực thi hàng triệu Goroutine.

Một Goroutine cũng có thể được thực thi trên nhiều Machine khác nhau trong suốt vòng đời của nó.

Không tồn tại mối quan hệ 1:1 giữa Goroutine và Thread.

---

# Hiểu lầm 2 - Từ khóa go nghĩa là chạy ngay lập tức

Ví dụ.

```go
go worker()
```

Nhiều người nghĩ.

```
go

↓

Run Immediately
```

Thực tế.

```
go

↓

Create Goroutine

↓

Runnable

↓

Run Queue

↓

Scheduler

↓

Running
```

Từ khóa `go` chỉ có nghĩa:

> **Yêu cầu Runtime tạo một Goroutine mới.**

Việc Goroutine đó khi nào được chạy hoàn toàn do Scheduler quyết định.

---

# Hiểu lầm 3 - Goroutine luôn chạy song song

Giả sử.

```go
go A()

go B()
```

Nhiều người nghĩ.

```
A

||

B
```

Luôn chạy song song.

Điều này không đúng.

Nếu.

```
GOMAXPROCS = 1
```

Thì.

```
A

↓

Scheduler

↓

B

↓

Scheduler

↓

A
```

Hai Goroutine chỉ chạy **đồng thời về mặt logic (Concurrent)** chứ không chạy **song song về mặt vật lý (Parallel)**.

Điều này cũng là lý do Rob Pike luôn nhấn mạnh:

> **Concurrency is not Parallelism.**

---

# Hiểu lầm 4 - GOMAXPROCS là số lượng Thread

Nhiều lập trình viên nghĩ.

```
GOMAXPROCS = 8

↓

8 Threads
```

Thực tế.

```
GOMAXPROCS

↓

Processor
```

Số lượng Machine có thể là.

```
8

10

15

20
```

đặc biệt khi chương trình có nhiều Blocking System Call hoặc CGO.

Điều bị giới hạn là số lượng **Processor**, không phải số lượng **OS Thread**.

---

# Hiểu lầm 5 - Scheduler sử dụng Round Robin đơn giản

Một số người cho rằng Scheduler chỉ đơn giản là:

```
G1

↓

G2

↓

G3

↓

G4
```

Điều này không chính xác.

Go Scheduler còn phải xem xét rất nhiều yếu tố.

- Local Queue
- Global Queue
- Work Stealing
- Netpoller
- Blocking Syscall
- Preemption
- Processor Availability

Thứ tự thực thi của Goroutine không hề cố định.

---

# Hiểu lầm 6 - Work Stealing diễn ra liên tục

Nhiều người tưởng tượng.

```
Processor

↓

Steal

↓

Steal

↓

Steal
```

Thực tế.

Work Stealing chỉ xảy ra khi.

```
Local Queue

↓

Empty
```

Nếu Local Queue vẫn còn Goroutine.

Scheduler sẽ không thực hiện Work Stealing.

Đây là một tối ưu rất quan trọng để giảm chi phí lập lịch.

---

# Hiểu lầm 7 - Processor là CPU

Tên gọi Processor rất dễ gây nhầm lẫn.

Nhiều người nghĩ.

```
P

=

CPU
```

Thực tế.

Processor chỉ là một đối tượng logic trong Go Runtime.

Nó chứa.

- Local Queue
- mcache
- Timer Cache
- Scheduler State

CPU là phần cứng.

Processor là tài nguyên của Runtime.

Hai khái niệm này hoàn toàn khác nhau.

---

# Hiểu lầm 8 - Scheduler chỉ có một Queue

Nếu mới học Go.

Nhiều người thường nghĩ.

```
Scheduler

↓

Queue

↓

CPU
```

Thực tế Runtime sử dụng nhiều Queue.

```
Processor

↓

Local Queue
```

```
Global Queue
```

Ngoài ra còn có các danh sách nội bộ khác phục vụ việc quản lý Goroutine.

Việc chia nhỏ Queue giúp Runtime giảm Lock Contention và cải thiện khả năng mở rộng.

---

# Hiểu lầm 9 - Goroutine Waiting vẫn tiêu tốn CPU

Ví dụ.

```go
<-ch
```

Lúc này.

```
Running

↓

Waiting
```

Một Goroutine ở trạng thái Waiting **không sử dụng CPU**.

Machine sẽ tiếp tục chạy Goroutine khác.

Điều này khác hoàn toàn với mô hình One Thread Per Connection, nơi một Thread có thể bị chặn trong thời gian dài.

---

# Hiểu lầm 10 - Blocking System Call sẽ làm toàn bộ Scheduler dừng lại

Điều này đúng với mô hình chỉ có Thread.

Nhưng không đúng với Go.

Khi một Machine đi vào Blocking System Call.

```
Machine

↓

Blocked
```

Runtime sẽ.

```
Detach Processor

↓

Attach sang Machine khác
```

Các Goroutine khác vẫn tiếp tục chạy bình thường.

---

# Hiểu lầm 11 - Scheduler quyết định hiệu năng của toàn bộ chương trình

Một số người cho rằng.

> Chỉ cần Scheduler tốt thì chương trình sẽ nhanh.

Điều này không đúng.

Scheduler chỉ quyết định:

```
Ai được chạy?
```

Hiệu năng của chương trình còn phụ thuộc vào rất nhiều yếu tố.

- Thuật toán.
- Bộ nhớ.
- Cache.
- Garbage Collector.
- Network.
- Database.
- Đồng bộ giữa các Goroutine.

Scheduler không thể biến một thuật toán có độ phức tạp kém thành một chương trình hiệu năng cao.

---

# Hiểu lầm 12 - Goroutine rất rẻ nên có thể tạo vô hạn

Đúng là Goroutine nhẹ hơn Thread rất nhiều.

Nhưng điều đó không có nghĩa.

```
Unlimited
```

Mỗi Goroutine vẫn tiêu tốn.

- Stack.
- Metadata.
- Scheduler State.
- Bộ nhớ Heap.

Việc tạo hàng chục triệu Goroutine có thể:

- Tăng áp lực lên Garbage Collector.
- Tăng chi phí Scheduler.
- Tiêu tốn nhiều RAM.

Nguyên tắc là:

> **Tạo đủ Goroutine để giải quyết công việc, không phải càng nhiều càng tốt.**

---

# Hiểu lầm 13 - Goroutine luôn thực thi theo đúng thứ tự được tạo

Ví dụ.

```go
go A()

go B()

go C()
```

Không có gì đảm bảo.

```
A

↓

B

↓

C
```

Scheduler có thể chạy.

```
B

↓

A

↓

C
```

hoặc bất kỳ thứ tự nào khác.

Lập trình viên không nên dựa vào thứ tự thực thi của Goroutine nếu không có cơ chế đồng bộ rõ ràng.

---

# Hiểu lầm 14 - Scheduler là thành phần duy nhất quản lý Goroutine

Thực tế có nhiều thành phần cùng tham gia.

| Thành phần | Vai trò |
|------------|----------|
| Scheduler | Chọn Goroutine để chạy |
| Netpoller | Đánh thức Goroutine chờ I/O |
| Channel | Đánh thức Goroutine chờ gửi/nhận |
| Mutex | Đánh thức Goroutine chờ khóa |
| Timer | Đánh thức Goroutine sau Sleep |
| Garbage Collector | Đồng bộ Goroutine tại Safe Point |

Scheduler chỉ là một mắt xích trong toàn bộ Go Runtime.

---

# Hiểu lầm 15 - Scheduler của Go là hoàn hảo

Go Scheduler được đánh giá là một trong những Scheduler tốt nhất hiện nay.

Tuy nhiên.

Nó vẫn có những giới hạn.

Ví dụ.

- Lock Contention.
- False Sharing.
- NUMA.
- Cache Miss.
- Workload không đồng đều.
- Chi phí Garbage Collector.

Không có Scheduler nào có thể tối ưu cho mọi loại ứng dụng.

Go Runtime liên tục được cải tiến qua từng phiên bản để giải quyết những hạn chế này.

---

# Tổng kết

Việc hiểu đúng về Scheduler quan trọng không kém việc học cách sử dụng Goroutine.

Phần lớn các lỗi về hiệu năng trong chương trình Go không đến từ việc Scheduler hoạt động sai, mà đến từ việc lập trình viên có những giả định không chính xác về cách Scheduler vận hành.

Những điều cần ghi nhớ nhất là:

- Goroutine **không phải** là Thread.
- `go` **không có nghĩa** là chạy ngay lập tức.
- Concurrency **không đồng nghĩa** với Parallelism.
- `GOMAXPROCS` quyết định số lượng **Processor**, không phải số lượng Thread.
- Scheduler ưu tiên **Local Queue**, chỉ Work Stealing khi cần.
- Goroutine ở trạng thái Waiting **không tiêu tốn CPU**.
- Scheduler chỉ là một phần của Go Runtime và luôn phối hợp với Netpoller, Garbage Collector, Timer, Channel và Mutex để quản lý vòng đời của Goroutine.

Hiểu rõ những điểm này sẽ giúp bạn tránh được nhiều lỗi thiết kế, đồng thời có nền tảng vững chắc để phân tích hiệu năng và trả lời các câu hỏi phỏng vấn chuyên sâu về Go Runtime.

---

---

# 10.90 Interview Questions

Sau khi hoàn thành chương này, bạn nên có khả năng trả lời các câu hỏi dưới đây.

Các câu hỏi được sắp xếp theo nhiều mức độ khác nhau, từ Junior đến Senior.

---

# Cấp độ Junior

## Câu 1

### Goroutine là gì?

**Ý cần trả lời**

- Là đơn vị thực thi (Execution Unit) của Go Runtime.
- Được Scheduler quản lý.
- Nhẹ hơn OS Thread.
- Có Stack tự mở rộng.
- Không ánh xạ 1:1 với Thread.

---

## Câu 2

### Goroutine khác Thread ở điểm nào?

Các điểm cần đề cập.

| Goroutine | Thread |
|------------|---------|
| Do Go Runtime quản lý | Do OS quản lý |
| Stack nhỏ (~2KB ban đầu) | Stack lớn hơn nhiều |
| User-space scheduling | Kernel scheduling |
| Tạo rất nhanh | Tạo chậm hơn |
| Có thể tạo hàng triệu | Số lượng hạn chế |

---

## Câu 3

### Scheduler là gì?

Một câu trả lời tốt nên bao gồm.

- Thành phần của Go Runtime.
- Chịu trách nhiệm lập lịch Goroutine.
- Ánh xạ Goroutine xuống OS Thread.
- Tận dụng CPU hiệu quả.

---

## Câu 4

### GMP là gì?

Ứng viên nên giải thích.

```
G

↓

Execution Context
```

```
M

↓

OS Thread
```

```
P

↓

Scheduling Resource
```

Không nên trả lời:

```
P = Processor của CPU
```

Đây là câu trả lời sai.

---

## Câu 5

### GOMAXPROCS dùng để làm gì?

Điểm cần nhấn mạnh.

- Quyết định số lượng Processor.
- Giới hạn số Goroutine có thể Running đồng thời.
- Không phải số Thread.

---

# Cấp độ Intermediate

## Câu 6

### Tại sao Go phải tạo thêm Processor (P)?

Đây là câu hỏi rất phổ biến.

Một câu trả lời tốt nên đề cập.

- Tách Scheduler khỏi OS Thread.
- Hỗ trợ Blocking System Call.
- Local Queue.
- mcache.
- Timer Cache.
- Giảm Lock Contention.

Đây là một trong những quyết định thiết kế quan trọng nhất của Go Runtime.

---

## Câu 7

### Tại sao Scheduler không chỉ dùng Global Queue?

Ý chính.

Nếu chỉ có Global Queue.

```
64 CPU

↓

1 Lock
```

Sẽ xảy ra.

- Lock Contention.
- Cache Miss.
- Khả năng mở rộng kém.

Local Queue giúp giảm đáng kể các vấn đề này.

---

## Câu 8

### Local Queue khác Global Queue như thế nào?

Ứng viên nên trình bày.

| Local Queue | Global Queue |
|--------------|--------------|
| Thuộc từng Processor | Dùng chung |
| Scheduler ưu tiên | Chỉ dùng khi cần |
| Ít Lock | Có đồng bộ |
| Truy cập nhanh | Chậm hơn |

---

## Câu 9

### Work Stealing hoạt động như thế nào?

Điểm cần nhấn mạnh.

- Chỉ xảy ra khi Local Queue rỗng.
- Chọn ngẫu nhiên Processor khác.
- Lấy khoảng một nửa Queue.
- Không diễn ra liên tục.

---

## Câu 10

### Scheduler xử lý Blocking System Call như thế nào?

Ứng viên nên mô tả.

```
Thread

↓

Blocking Syscall

↓

Detach Processor

↓

Attach sang Machine khác

↓

Continue Running
```

Đây là câu hỏi rất hay xuất hiện trong phỏng vấn Senior.

---

## Câu 11

### Netpoller là gì?

Điểm chính.

- Theo dõi I/O.
- Không Polling liên tục.
- Đánh thức Goroutine.
- Chuyển Waiting → Runnable.

---

## Câu 12

### Khi nào Goroutine chuyển sang trạng thái Waiting?

Ví dụ.

- Channel.
- Mutex.
- I/O.
- time.Sleep().
- Cond.Wait().

---

## Câu 13

### Preemption là gì?

Một câu trả lời tốt.

- Runtime chủ động giành CPU.
- Không cần Goroutine tự nhường.
- Đảm bảo Fairness.
- Hỗ trợ Garbage Collector.

---

# Cấp độ Senior

## Câu 14

### Tại sao mcache nằm trong Processor thay vì Machine?

Ứng viên giỏi sẽ nói.

- Machine có thể Block.
- Processor đại diện cho Scheduling Resource.
- mcache đi theo Processor.
- Giảm Lock.
- Tăng Cache Locality.

---

## Câu 15

### Tại sao Machine có thể nhiều hơn Processor?

Điểm cần nhấn mạnh.

- Blocking Syscall.
- CGO.
- Runtime tạo thêm Machine.
- Nhưng số Processor vẫn không đổi.

---

## Câu 16

### Tại sao Runtime không tạo thêm Processor khi Thread bị Block?

Đây là câu hỏi đánh giá khả năng hiểu bản chất.

Cần trả lời.

- Processor đại diện mức độ song song.
- GOMAXPROCS quyết định số Processor.
- Runtime chỉ tạo thêm Machine.
- Không thay đổi Processor.

---

## Câu 17

### Scheduler có phải Round Robin không?

Đây là câu hỏi đánh lừa.

Câu trả lời.

Không.

Scheduler còn phải xem xét.

- Local Queue.
- Global Queue.
- Work Stealing.
- Netpoller.
- Blocking Syscall.
- Preemption.

---

## Câu 18

### Scheduler và Garbage Collector phối hợp như thế nào?

Ý chính.

- Safe Point.
- GC Worker.
- Concurrent GC.
- Scheduler phân phối CPU.

---

## Câu 19

### Vì sao Go có thể xử lý hàng trăm nghìn kết nối?

Ứng viên nên nói đến.

- Goroutine.
- Netpoller.
- epoll/kqueue/IOCP.
- Không tạo Thread cho mỗi Connection.
- Scheduler.

---

## Câu 20

### Vì sao Goroutine có Stack nhỏ nhưng vẫn chạy được chương trình lớn?

Điểm chính.

- Stack động.
- Stack Growth.
- Stack Copy.
- Runtime tự mở rộng.

(Chi tiết sẽ học ở Chapter Goroutine.)

---

# Câu hỏi mở

Những câu hỏi dưới đây thường được sử dụng trong các buổi phỏng vấn Senior hoặc Staff Engineer để đánh giá khả năng phân tích.

---

## Nếu bỏ Processor thì điều gì xảy ra?

Một câu trả lời tốt nên đề cập.

- Queue sẽ gắn với Thread.
- Blocking Syscall làm Queue bị Block.
- Không còn Local Queue.
- Lock Contention tăng.
- Khả năng mở rộng giảm.

---

## Nếu chỉ có một Global Queue thì điều gì xảy ra?

Điểm cần phân tích.

- Hot Spot.
- Lock Contention.
- Cache Miss.
- Throughput giảm.

---

## Vì sao Work Stealing chỉ lấy một nửa Queue?

Ý cần trả lời.

- Giảm chi phí di chuyển Goroutine.
- Giữ Cache Locality.
- Tránh Queue bị rỗng hoàn toàn.
- Giảm số lần Work Stealing.

---

## Nếu một Goroutine chạy vòng lặp vô hạn thì điều gì xảy ra?

Ứng viên nên đề cập.

- Asynchronous Preemption.
- Safe Point.
- Fairness.
- Scheduler sẽ giành lại CPU.

---

## Tại sao Waiting Goroutine không tiêu tốn CPU?

Điểm cần trả lời.

- Không ở trạng thái Running.
- Machine chạy Goroutine khác.
- Netpoller/Channel/Mutex sẽ đánh thức khi cần.

---

## Tại sao Blocking I/O không làm toàn bộ chương trình bị Block?

Ứng viên nên mô tả.

```
Blocking Thread

↓

Detach Processor

↓

Machine khác tiếp tục chạy
```

---

# Câu hỏi tình huống

## Tình huống 1

Một chương trình tạo:

```
5.000.000 Goroutines
```

Hỏi:

Điều gì sẽ xảy ra?

Ứng viên nên phân tích.

- Scheduler vẫn hoạt động.
- Memory tăng.
- GC tăng.
- Scheduling Cost tăng.
- Không có nghĩa 5 triệu Goroutine cùng chạy.

---

## Tình huống 2

Server có:

```
100.000 HTTP Connections
```

Tại sao Go không tạo:

```
100.000 Threads?
```

Đáp án.

- Goroutine.
- Netpoller.
- Waiting Goroutine.
- Scheduler.

---

## Tình huống 3

Một chương trình chỉ có:

```
1 CPU
```

nhưng tạo:

```
10.000 Goroutines
```

Hỏi.

Có bao nhiêu Goroutine thực sự Running?

Đáp án.

```
1
```

Nếu.

```
GOMAXPROCS = 1
```

---

## Tình huống 4

Một Processor hết việc.

Processor khác còn:

```
1000 Goroutines
```

Hỏi.

Runtime xử lý như thế nào?

Đáp án.

Work Stealing.

---

## Tình huống 5

Một Goroutine đang:

```go
time.Sleep(time.Minute)
```

Hỏi.

Trong thời gian đó CPU có bị chiếm không?

Đáp án.

Không.

Goroutine ở trạng thái Waiting.

Machine sẽ chạy Goroutine khác.

---

# Lời khuyên khi trả lời phỏng vấn

Đối với các câu hỏi về Scheduler, nhà tuyển dụng thường không chỉ muốn nghe định nghĩa.

Họ muốn biết bạn có hiểu **tại sao Go Runtime lại được thiết kế như vậy** hay không.

Ví dụ.

Nếu được hỏi.

> Tại sao Go cần Processor?

Đừng chỉ trả lời.

> "Để chạy Goroutine."

Hãy giải thích.

- Processor tách Scheduler khỏi Thread.
- Hỗ trợ Blocking System Call.
- Giảm Lock Contention.
- Chứa Local Queue.
- Chứa mcache.
- Giúp Scheduler mở rộng trên nhiều CPU.

Một câu trả lời giải thích **nguyên nhân thiết kế (Design Rationale)** luôn tạo ấn tượng tốt hơn việc chỉ mô tả **cơ chế hoạt động (Mechanism)**.

---

# Tổng kết

Nếu bạn có thể trả lời trôi chảy hầu hết các câu hỏi trong phần này, đồng thời giải thích được **vì sao** Go Runtime được thiết kế như vậy thay vì chỉ mô tả **Go làm gì**, thì bạn đã có nền tảng rất vững về Go Scheduler.

Đây cũng là mức kiến thức thường được kỳ vọng ở các vị trí **Middle**, **Senior Go Engineer** và là tiền đề để tiếp tục nghiên cứu sâu hơn về các chương tiếp theo như **Goroutine**, **Memory Allocator** và **Garbage Collector**.

---

---

# 10.92 Tổng kết chương

Scheduler là trái tim của Go Runtime.

Mọi Goroutine được tạo ra đều phải đi qua Scheduler trước khi có cơ hội được CPU thực thi.

Nếu không có Scheduler, Goroutine chỉ là những đối tượng được cấp phát trong bộ nhớ và sẽ không bao giờ được chạy.

Trong chương này, chúng ta đã đi từ những khái niệm nền tảng nhất cho đến những cơ chế cốt lõi giúp Go Runtime có thể quản lý hàng triệu Goroutine với chi phí rất thấp.

---

# Chúng ta đã học được những gì?

## 1. Tại sao Go cần Scheduler?

Chúng ta bắt đầu bằng câu hỏi:

> **Tại sao Go không sử dụng trực tiếp Scheduler của hệ điều hành?**

Câu trả lời nằm ở giới hạn của mô hình:

```
One Thread Per Request
```

OS Thread là tài nguyên tương đối nặng.

Việc tạo hàng trăm nghìn hoặc hàng triệu Thread sẽ dẫn đến:

- Tiêu tốn nhiều bộ nhớ.
- Context Switching tốn kém.
- Khả năng mở rộng kém.

Go giải quyết vấn đề này bằng cách giới thiệu **Goroutine** và một **User-space Scheduler**, cho phép Runtime tự quản lý việc lập lịch thay vì phụ thuộc hoàn toàn vào Kernel.

---

## 2. Mô hình GMP

Tiếp theo, chúng ta tìm hiểu mô hình **GMP**, nền tảng của toàn bộ Scheduler.

Ba thành phần cốt lõi gồm:

- **G (Goroutine)**: đơn vị thực thi.
- **M (Machine)**: đại diện cho OS Thread.
- **P (Processor)**: tài nguyên lập lịch của Go Runtime.

Điều quan trọng nhất cần ghi nhớ là:

> **Processor không phải là CPU.**

Processor là một đối tượng logic do Runtime tạo ra, chứa toàn bộ trạng thái cần thiết để lập lịch như:

- Local Run Queue.
- mcache.
- Timer Cache.
- Scheduler State.

Chính việc tách Processor khỏi Machine đã giúp Go giải quyết bài toán Blocking System Call và mở rộng tốt trên hệ thống nhiều CPU.

---

## 3. Run Queue

Chúng ta đã tìm hiểu cách Scheduler quản lý các Goroutine ở trạng thái Runnable thông qua:

- Local Run Queue.
- Global Run Queue.

Scheduler luôn ưu tiên lấy Goroutine từ Local Queue để:

- Giảm Lock Contention.
- Tận dụng Cache Locality.
- Giảm chi phí đồng bộ.

Global Queue chỉ được sử dụng trong một số tình huống đặc biệt như:

- Khởi tạo.
- Cân bằng tải.
- Local Queue đầy.

---

## 4. Scheduler Loop

Chúng ta đã phân tích vòng lặp trung tâm của Scheduler.

Mỗi khi một Machine cần công việc mới, Scheduler sẽ tìm Goroutine theo thứ tự ưu tiên:

1. Local Run Queue.
2. Global Run Queue.
3. Work Stealing.
4. Netpoller.
5. Idle.

Đây là chu trình diễn ra liên tục trong suốt vòng đời của chương trình.

---

## 5. Work Stealing

Work Stealing là cơ chế cân bằng tải giữa các Processor.

Khi Local Queue của một Processor rỗng, Runtime sẽ:

- Chọn ngẫu nhiên một Processor khác.
- Lấy khoảng một nửa số Goroutine trong Local Queue của Processor đó.
- Đưa về Local Queue của mình để tiếp tục thực thi.

Nhờ vậy, Go Runtime tận dụng tốt tài nguyên CPU mà không cần một Scheduler trung tâm phức tạp.

---

## 6. Blocking System Call

Một trong những điểm nổi bật nhất của Go Scheduler là cách xử lý Blocking System Call.

Khi một Machine đi vào Kernel và bị chặn:

- Processor được tách khỏi Machine.
- Processor được gắn sang một Machine khác.
- Các Goroutine còn lại vẫn tiếp tục chạy.

Thiết kế này giúp CPU không bị lãng phí khi một Thread đang chờ I/O.

---

## 7. Netpoller

Netpoller là cầu nối giữa hệ điều hành và Scheduler.

Thay vì liên tục kiểm tra xem I/O đã hoàn thành hay chưa, Go Runtime sử dụng các cơ chế như:

- `epoll` trên Linux.
- `kqueue` trên macOS/BSD.
- `IOCP` trên Windows.

Khi có sự kiện I/O, Netpoller sẽ:

- Đánh thức Goroutine.
- Chuyển trạng thái từ **Waiting** sang **Runnable**.
- Đưa Goroutine trở lại Run Queue.

Nhờ đó, Go có thể xử lý hàng trăm nghìn kết nối mạng mà không cần tạo tương ứng hàng trăm nghìn Thread.

---

## 8. Asynchronous Preemption

Để đảm bảo không Goroutine nào chiếm CPU quá lâu, Go Runtime sử dụng **Asynchronous Preemption**.

Scheduler có thể:

- Tạm dừng Goroutine đang chạy.
- Lưu ngữ cảnh thực thi.
- Chạy Goroutine khác.
- Khôi phục và tiếp tục Goroutine sau đó.

Cơ chế này giúp:

- Đảm bảo tính công bằng.
- Tăng khả năng phản hồi.
- Hỗ trợ Garbage Collector.

---

## 9. GOMAXPROCS

Chúng ta cũng đã tìm hiểu vai trò của `GOMAXPROCS`.

Điểm quan trọng nhất là:

> `GOMAXPROCS` quyết định số lượng **Processor**, không phải số lượng Thread.

Điều này cũng đồng nghĩa với việc:

- Có thể tạo hàng triệu Goroutine.
- Nhưng chỉ có tối đa `GOMAXPROCS` Goroutine thực sự chạy trên CPU tại cùng một thời điểm.

---

## 10. Scheduler và Garbage Collector

Scheduler không hoạt động độc lập.

Nó phối hợp chặt chẽ với Garbage Collector thông qua:

- Safe Point.
- Concurrent GC.
- GC Worker.
- GC Assist.

Sự phối hợp này giúp Go duy trì thời gian tạm dừng rất thấp trong khi vẫn đảm bảo bộ nhớ được quản lý hiệu quả.

---

# Bức tranh tổng thể của Scheduler

Sau khi hoàn thành chương này, chúng ta có thể mô tả toàn bộ luồng hoạt động của Scheduler như sau:

```text
              go func()

                  │

                  ▼

          Tạo Goroutine (G)

                  │

                  ▼

             Runnable

                  │

                  ▼

         Local Run Queue

                  │

                  ▼

        Scheduler lựa chọn

                  │

                  ▼

          Running trên M + P

                  │

     ┌────────────┼─────────────┐
     │            │             │
     ▼            ▼             ▼

 Waiting     Preemption      Dead

     │
     │
     ▼

 Runnable

     │
     └─────────────────────────────► Scheduler
```

Đây là vòng đời cơ bản của hầu hết các Goroutine trong một chương trình Go.

---

# Những điều quan trọng nhất cần ghi nhớ

Nếu chỉ được phép nhớ một số ít kiến thức từ chương này, hãy ghi nhớ những điểm sau:

- Scheduler của Go là **User-space Scheduler**, không phải Kernel Scheduler.
- Goroutine là đơn vị thực thi do Runtime quản lý.
- Machine đại diện cho OS Thread.
- Processor là tài nguyên lập lịch, không phải CPU.
- Scheduler ưu tiên Local Run Queue trước Global Run Queue.
- Work Stealing giúp cân bằng tải giữa các Processor.
- Blocking System Call không làm toàn bộ chương trình bị dừng.
- Netpoller giúp xử lý I/O với số lượng Thread rất nhỏ.
- Asynchronous Preemption đảm bảo mọi Goroutine đều có cơ hội được thực thi.
- `GOMAXPROCS` quyết định số lượng Processor.
- Scheduler và Garbage Collector luôn phối hợp chặt chẽ với nhau.

Nếu hiểu rõ các điểm trên, bạn đã nắm được phần lớn nguyên lý hoạt động của Go Scheduler.

---

# Chương tiếp theo

Trong chương này, chúng ta chủ yếu nhìn Goroutine từ góc nhìn của Scheduler.

Scheduler chỉ quan tâm:

- Goroutine đang ở trạng thái nào?
- Có Runnable hay không?
- Được chạy trên Machine nào?
- Thuộc Processor nào?

Nhưng Scheduler không trả lời những câu hỏi như:

- Một Goroutine thực chất được biểu diễn như thế nào trong bộ nhớ?
- Runtime lưu Program Counter và Stack Pointer ở đâu?
- Tại sao Stack của Goroutine chỉ bắt đầu với khoảng 2 KB nhưng vẫn có thể chạy các hàm đệ quy sâu?
- Cơ chế mở rộng và thu nhỏ Stack hoạt động ra sao?
- Runtime lưu thông tin `defer`, `panic` và context của Goroutine như thế nào?

Đó sẽ là nội dung của chương tiếp theo.

---

# Chapter 11 - Goroutine

Trong Chapter 11, chúng ta sẽ đi sâu vào cấu trúc nội tại của Goroutine, bao gồm:

- Kiến trúc của struct `g`.
- `gobuf` và cơ chế lưu ngữ cảnh thực thi.
- Stack của Goroutine.
- Stack Growth và Stack Shrink.
- Stack Copy.
- Context Switching.
- Defer.
- Panic.
- Goroutine Leak.
- Goroutine Dump.
- Phân tích các cấu trúc dữ liệu trong `runtime/runtime2.go`.

Nếu Chapter 10 giúp bạn hiểu **Scheduler điều phối Goroutine như thế nào**, thì Chapter 11 sẽ giúp bạn hiểu **một Goroutine thực sự được xây dựng và hoạt động ra sao bên trong Go Runtime**.

Đây cũng là nền tảng để tiếp tục nghiên cứu sâu hơn về **Memory Allocator**, **Garbage Collector** và các thành phần cốt lõi khác của Go Runtime trong các chương tiếp theo.

---