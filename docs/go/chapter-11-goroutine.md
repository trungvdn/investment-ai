---
title: Goroutine
volume: II - Go Runtime
chapter: 11
section: 11.1
---

# Chapter 11 - Goroutine

> "Don't communicate by sharing memory; share memory by communicating."
>
> **— Andrew Gerrand**

---

# Mục tiêu chương

Trong Chapter 10, chúng ta đã nhìn Goroutine từ góc nhìn của **Scheduler**.

Scheduler chỉ quan tâm đến những câu hỏi như:

- Goroutine nào đang ở trạng thái Runnable?
- Goroutine nào sẽ được chạy tiếp theo?
- Goroutine đang thuộc Processor nào?
- Goroutine đang chạy trên Machine nào?

Tuy nhiên, Scheduler không quan tâm:

- Một Goroutine được biểu diễn như thế nào trong bộ nhớ?
- Runtime lưu Program Counter ở đâu?
- Runtime lưu Stack Pointer ở đâu?
- Vì sao Goroutine có thể bị tạm dừng rồi tiếp tục chạy?
- Stack của Goroutine được mở rộng bằng cách nào?

Đó chính là nội dung của chương này.

Sau khi hoàn thành Chapter 11, bạn sẽ hiểu được cách Go Runtime xây dựng và quản lý Goroutine từ bên trong, bao gồm:

- Kiến trúc của `struct g`.
- Cấu trúc của Stack.
- Context Switching.
- Dynamic Stack.
- Stack Growth.
- Stack Shrink.
- Stack Copy.
- Defer.
- Panic.
- Goroutine Leak.
- Goroutine Dump.

Đây là một trong những chương quan trọng nhất của toàn bộ Volume II vì hầu như mọi thành phần của Go Runtime đều liên quan đến Goroutine.

---

# Mục lục

## Phần I - Tổng quan

- 11.1 Giới thiệu
- 11.2 Goroutine và OS Thread
- 11.3 Vòng đời của Goroutine

---

## Phần II - Cấu trúc dữ liệu nội bộ

- 11.4 `struct g`
- 11.5 `gobuf`
- 11.6 Stack
- 11.7 Stack Frame

---

## Phần III - Dynamic Stack

- 11.8 Initial Stack
- 11.9 Stack Growth
- 11.10 Stack Copy
- 11.11 Stack Shrink
- 11.12 Stack Overflow Detection

---

## Phần IV - Context Switching

- 11.13 Context Switching
- 11.14 Save Context
- 11.15 Restore Context
- 11.16 Function Call

---

## Phần V - Runtime Features

- 11.17 Defer
- 11.18 Panic
- 11.19 Recover
- 11.20 Goroutine Exit

---

## Phần VI - Scheduler Interaction

- 11.21 Goroutine States
- 11.22 Waiting Queue
- 11.23 Park / Unpark
- 11.24 Preemption

---

## Phần VII - Debugging

- 11.25 Goroutine Leak
- 11.26 Goroutine Dump
- 11.27 Trace
- 11.28 Common Mistakes

---

## Phần VIII - Engineering

- 11.29 Best Practices
- 11.30 Interview Questions
- 11.31 Tổng kết

---

# 11.1 Giới thiệu

Trong ngôn ngữ Go, khi nhắc đến Concurrency, khái niệm đầu tiên mà mọi lập trình viên nghĩ tới chính là **Goroutine**.

Chỉ với một từ khóa rất đơn giản:

```go
go processOrder(order)
```

Go Runtime có thể tạo ra một đơn vị thực thi mới gần như ngay lập tức.

Đối với lập trình viên, Goroutine chỉ đơn giản là một hàm được chạy đồng thời với phần còn lại của chương trình.

Nhưng bên trong Go Runtime, Goroutine là một cấu trúc dữ liệu rất phức tạp.

Nó chứa:

- Stack.
- Program Counter.
- Stack Pointer.
- Registers.
- Scheduler Metadata.
- Panic Stack.
- Defer Stack.
- Garbage Collector Metadata.

Nói cách khác:

> **Một Goroutine không phải là một hàm.**

Cũng không phải là một Thread.

Mà là một **Execution Context** hoàn chỉnh.

---

# Execution Context là gì?

Execution Context có thể hiểu là:

> **Toàn bộ trạng thái cần thiết để một chương trình có thể tạm dừng và tiếp tục thực thi tại đúng vị trí trước đó.**

Ví dụ.

Một Goroutine đang chạy đoạn mã sau.

```go
func calculate() {

    a := 10

    b := 20

    c := a + b

    fmt.Println(c)

}
```

Giả sử Scheduler quyết định tạm dừng Goroutine ngay sau dòng:

```go
c := a + b
```

Runtime cần lưu lại rất nhiều thông tin.

Ví dụ.

```
Program Counter

↓

Đang ở dòng nào
```

```
Stack Pointer

↓

Stack hiện tại
```

```
Registers

↓

Giá trị CPU Registers
```

```
Stack

↓

Biến cục bộ
```

Một lúc sau.

Scheduler khôi phục Goroutine.

```
Restore Context

↓

Continue Execution
```

Đối với lập trình viên.

Không có gì khác biệt.

Chương trình tiếp tục chạy như chưa từng bị dừng.

Đó chính là sức mạnh của Execution Context.

---

# Goroutine không phải là Function

Một hiểu lầm rất phổ biến là:

```
go foo()
```

đồng nghĩa với:

```
Function foo()
```

Thực tế không phải vậy.

Function chỉ là:

```
Instructions
```

Tức là mã máy sẽ được CPU thực thi.

Trong khi Goroutine bao gồm:

```text
Function

+

Stack

+

Registers

+

Program Counter

+

Scheduler State

+

GC Metadata
```

Function chỉ là một phần rất nhỏ trong Goroutine.

Điều mà Scheduler thực sự quản lý không phải là Function mà là toàn bộ Execution Context.

---

# Goroutine không phải là Thread

Một hiểu lầm khác là:

```
Goroutine

=

OS Thread
```

Điều này cũng không đúng.

Một Thread là đối tượng của hệ điều hành.

Một Goroutine là đối tượng của Go Runtime.

Có thể hình dung.

```text
          Goroutine

        +-----------------+
        | Execution State |
        +-----------------+
        | Stack           |
        +-----------------+
        | Registers       |
        +-----------------+
        | Scheduler Info  |
        +-----------------+

                │

                ▼

          Go Scheduler

                │

                ▼

            OS Thread

                │

                ▼

               CPU
```

Scheduler có thể:

- Di chuyển Goroutine.
- Tạm dừng Goroutine.
- Tiếp tục Goroutine.
- Chuyển Goroutine sang Machine khác.

Trong khi đó.

OS Thread hoàn toàn không biết Goroutine là gì.

---

# Tại sao Goroutine nhẹ?

Một trong những đặc điểm nổi bật nhất của Go là:

> Có thể tạo hàng triệu Goroutine.

Điều này nghe có vẻ khó tin nếu so sánh với Thread.

Lý do nằm ở thiết kế của Goroutine.

Ngay khi được tạo.

Runtime chỉ cấp phát:

```
≈ 2 KB Stack
```

Thay vì cấp phát một Stack rất lớn như Thread của hệ điều hành.

Không những vậy.

Stack này còn có thể:

- Mở rộng.
- Thu nhỏ.
- Sao chép.
- Di chuyển.

hoàn toàn tự động.

Nhờ vậy.

Một Goroutine có chi phí bộ nhớ rất nhỏ.

Đây là nền tảng giúp Go xây dựng các hệ thống có mức độ đồng thời cực kỳ cao.

---

# Goroutine là nền tảng của Go Runtime

Nếu quan sát toàn bộ Go Runtime.

Chúng ta sẽ thấy hầu như mọi thành phần đều xoay quanh Goroutine.

```text
              Scheduler

                  │

                  ▼

             Goroutine

        ┌────────┼────────┐
        │        │        │

        ▼        ▼        ▼

     Channel   Timer     GC

        │        │        │

        └────────┼────────┘

                 ▼

             Memory
```

Ví dụ.

Scheduler lập lịch:

```
Goroutine
```

Channel đánh thức:

```
Goroutine
```

Netpoller đánh thức:

```
Goroutine
```

Timer đánh thức:

```
Goroutine
```

Garbage Collector quét:

```
Goroutine Stack
```

Có thể nói.

**Goroutine chính là trung tâm của toàn bộ Go Runtime.**

---

# Những câu hỏi mà chương này sẽ trả lời

Trong các phần tiếp theo, chúng ta sẽ lần lượt trả lời những câu hỏi mà hầu hết lập trình viên Go đều từng thắc mắc.

Ví dụ.

### Runtime lưu Goroutine ở đâu?

---

### Vì sao Goroutine có thể bị dừng rồi chạy tiếp?

---

### Stack của Goroutine tăng trưởng như thế nào?

---

### Runtime phát hiện Stack Overflow bằng cách nào?

---

### Vì sao Stack của Goroutine có thể được sao chép?

---

### Scheduler lưu Program Counter ở đâu?

---

### Runtime thực hiện Context Switching như thế nào?

---

### Defer và Panic được lưu ở đâu?

---

### Làm thế nào Runtime phát hiện Goroutine Leak?

---

### Vì sao Goroutine chỉ cần khoảng 2 KB Stack ban đầu?

---

Đây đều là những câu hỏi sẽ được giải đáp trong Chapter 11.

---

# Tóm tắt

Goroutine là đơn vị thực thi cốt lõi của Go Runtime.

Khác với cách nhìn đơn giản của lập trình viên, một Goroutine không chỉ là một hàm được chạy đồng thời mà là một **Execution Context** hoàn chỉnh, bao gồm Stack, Program Counter, Stack Pointer, Registers và nhiều thông tin nội bộ khác phục vụ Scheduler và Garbage Collector.

Hiểu đúng Goroutine là bước đầu tiên để hiểu cách Go Runtime hoạt động.

Trong các phần tiếp theo, chúng ta sẽ đi sâu vào cấu trúc nội tại của Goroutine, bắt đầu với câu hỏi:

> **Goroutine khác gì so với một OS Thread?**

Đó sẽ là nội dung của **Section 11.2 – Goroutine và OS Thread**.

---

---

# 11.2 Goroutine và OS Thread

Trong phần trước, chúng ta đã biết:

> **Goroutine không phải là OS Thread.**

Đây là một trong những khái niệm quan trọng nhất của Go Runtime.

Tuy nhiên, rất nhiều lập trình viên mới học Go vẫn vô thức xem Goroutine như một phiên bản "nhẹ hơn" của Thread.

Cách hiểu này chỉ đúng một phần.

Để thực sự hiểu Goroutine, chúng ta cần so sánh nó với OS Thread từ góc nhìn của hệ điều hành, Runtime và phần cứng.

---

# Thread là gì?

Thread là đơn vị thực thi do **hệ điều hành (Operating System)** quản lý.

Một tiến trình (Process) có thể chứa nhiều Thread.

Ví dụ.

```text
                Process

        +----------------------+
        |      Thread 1        |
        +----------------------+
        |      Thread 2        |
        +----------------------+
        |      Thread 3        |
        +----------------------+
```

Mỗi Thread có:

- Program Counter (PC)
- Stack Pointer (SP)
- Registers
- Stack
- Thread Local Storage (TLS)
- Scheduling Information

Hệ điều hành chịu trách nhiệm:

- Tạo Thread.
- Hủy Thread.
- Chuyển đổi giữa các Thread.
- Phân phối Thread lên CPU.

Lập trình viên gần như không can thiệp vào quá trình lập lịch này.

---

# Goroutine là gì?

Khác với Thread.

Goroutine không phải là đối tượng của hệ điều hành.

Nó là một đối tượng do **Go Runtime** tạo ra và quản lý.

Có thể hình dung.

```text
            Go Runtime

                │

      +---------+---------+

      │                   │

      ▼                   ▼

 Goroutine 1         Goroutine 2

      │                   │

      +---------+---------+

                │

                ▼

           Scheduler

                │

                ▼

           OS Thread
```

Hệ điều hành **không biết** Goroutine tồn tại.

Đối với Kernel.

Chỉ có:

```
OS Thread
```

đang thực thi.

Mọi việc liên quan đến Goroutine đều do Go Runtime xử lý.

---

# So sánh tổng quan

| Tiêu chí | Goroutine | OS Thread |
|----------|-----------|-----------|
| Được quản lý bởi | Go Runtime | Operating System |
| Scheduler | User-space Scheduler | Kernel Scheduler |
| Stack ban đầu | Khoảng 2 KB | Lớn hơn nhiều, tùy hệ điều hành |
| Stack | Tự động mở rộng/thu nhỏ | Thường có kích thước cố định khi tạo |
| Chi phí tạo | Thấp | Cao hơn |
| Chi phí Context Switch | Thấp | Cao hơn |
| Số lượng có thể tạo | Hàng trăm nghìn đến hàng triệu | Thường chỉ vài nghìn |
| Nhận biết bởi Kernel | Không | Có |

Bảng trên chỉ là phần bề mặt.

Điểm khác biệt thực sự nằm ở kiến trúc bên dưới.

---

# Khác biệt đầu tiên - Ai quản lý?

Đây là khác biệt quan trọng nhất.

## OS Thread

```
Application

↓

Kernel API

↓

Operating System

↓

Create Thread
```

Khi tạo một Thread.

Ứng dụng phải yêu cầu Kernel.

Kernel sẽ:

- Cấp phát Kernel Object.
- Cấp phát Stack.
- Khởi tạo Scheduler State.
- Đăng ký Thread.

Việc này tương đối tốn kém.

---

## Goroutine

```
Application

↓

Go Runtime

↓

Create Goroutine
```

Không cần Kernel.

Không cần System Call.

Không cần tạo Thread mới.

Runtime chỉ:

- Tạo `struct g`.
- Cấp phát Stack ban đầu.
- Đưa Goroutine vào Run Queue.

Đó là lý do việc tạo Goroutine nhanh hơn rất nhiều.

---

# Khác biệt thứ hai - Scheduler

OS Thread được lập lịch bởi Kernel.

```text
Thread

↓

Kernel Scheduler

↓

CPU
```

Kernel phải lập lịch cho:

- Trình duyệt.
- Database.
- IDE.
- Docker.
- Go Program.
- Mọi ứng dụng khác.

Do đó.

Kernel Scheduler phải là một thuật toán rất tổng quát.

---

Ngược lại.

Go Runtime chỉ cần lập lịch:

```
Goroutine
```

Bên trong một chương trình Go.

Điều này cho phép Runtime tối ưu rất nhiều.

Ví dụ.

- Local Queue.
- Work Stealing.
- Netpoller.
- Asynchronous Preemption.

Đây là những tối ưu mà Kernel không thể thực hiện vì Kernel không biết Goroutine là gì.

---

# Khác biệt thứ ba - Bộ nhớ

Mỗi Thread cần một Stack riêng.

```text
Thread

↓

Stack
```

Stack của Thread thường được cấp phát với kích thước tương đối lớn ngay từ đầu.

Nếu tạo:

```
100.000 Threads
```

Bộ nhớ dành riêng cho Stack sẽ rất lớn, ngay cả khi phần lớn Stack chưa được sử dụng.

---

Trong khi đó.

Goroutine chỉ bắt đầu với khoảng:

```
≈ 2 KB
```

Stack.

Nếu Stack chưa đủ.

Runtime sẽ:

```
Grow Stack
```

Nếu Stack dư thừa.

Runtime có thể:

```
Shrink Stack
```

Điều này giúp sử dụng bộ nhớ hiệu quả hơn rất nhiều.

---

# Khác biệt thứ tư - Context Switching

Context Switching là quá trình chuyển CPU từ đơn vị thực thi này sang đơn vị thực thi khác.

## Với Thread

Quá trình chuyển đổi cần Kernel tham gia.

```text
Thread A

↓

Kernel Mode

↓

Save Registers

↓

Restore Registers

↓

Thread B
```

Việc chuyển sang Kernel Mode và quay lại User Mode có chi phí không nhỏ.

---

## Với Goroutine

Runtime tự thực hiện việc chuyển đổi.

```text
G1

↓

Save Context

↓

Restore Context

↓

G2
```

Phần lớn quá trình diễn ra trong User Space.

Không cần chuyển sang Kernel.

Điều này giúp giảm đáng kể chi phí Context Switching.

---

# Khác biệt thứ năm - Khả năng mở rộng

Giả sử chúng ta cần xử lý:

```
1.000.000 Requests
```

Nếu mỗi Request sử dụng:

```
1 Thread
```

thì cần:

```
1.000.000 Threads
```

Điều này gần như không khả thi.

---

Với Go.

```
1.000.000 Requests

↓

1.000.000 Goroutines

↓

8 Processor

↓

8 Machines

↓

8 CPU
```

Runtime chỉ cần một số lượng nhỏ Thread để phục vụ rất nhiều Goroutine.

Đây là nền tảng giúp Go xây dựng các hệ thống có khả năng mở rộng rất cao.

---

# Khác biệt thứ sáu - Blocking I/O

Nếu một Thread thực hiện:

```go
read()
```

Thread sẽ bị Kernel chặn.

Trong nhiều mô hình truyền thống.

```
Thread

↓

Waiting
```

đồng nghĩa với việc CPU không thể sử dụng Thread đó cho công việc khác.

---

Trong Go.

Nếu Goroutine thực hiện I/O.

```
Goroutine

↓

Waiting
```

Machine sẽ tiếp tục chạy:

```
Goroutine khác
```

Thông qua:

- Netpoller.
- Scheduler.
- Processor.

Đây là một trong những khác biệt quan trọng nhất giữa Goroutine và Thread.

---

# Khác biệt thứ bảy - Di chuyển giữa các Thread

Một Thread luôn thực thi trên chính nó.

Không thể có chuyện:

```
Thread A

↓

Continue

↓

Thread B
```

---

Trong khi đó.

Một Goroutine có thể được thực thi trên nhiều Machine khác nhau.

Ví dụ.

```text
Time 1

G1

↓

M0
```

Sau khi bị Preempt.

```text
Time 2

G1

↓

M2
```

Sau Work Stealing.

```text
Time 3

G1

↓

M1
```

Đối với Goroutine.

Điều này hoàn toàn bình thường.

Runtime chỉ cần khôi phục đúng Execution Context.

---

# Góc nhìn của Go Runtime

Có thể hình dung mối quan hệ giữa Goroutine và Thread như sau.

```text
           Go Runtime

                │

        +-------+-------+

        │               │

        ▼               ▼

     G1  G2  G3  G4  G5  G6

        │               │

        +-------+-------+

                │

           Scheduler

                │

      +---------+---------+

      ▼                   ▼

     M0                  M1

      │                   │

      ▼                   ▼

 Thread 1            Thread 2

      │                   │

      +---------+---------+

                │

               CPU
```

Từ góc nhìn của Runtime:

- Goroutine là đơn vị công việc.
- Machine là phương tiện thực thi.
- Thread chỉ là tài nguyên mà Runtime sử dụng để tiếp cận CPU.

---

# Khi nào nên nghĩ theo Goroutine, khi nào nên nghĩ theo Thread?

Trong phần lớn trường hợp.

Khi lập trình Go.

Bạn nên suy nghĩ theo:

```
Goroutine
```

thay vì:

```
Thread
```

Ví dụ.

Đừng tự hỏi:

> Có bao nhiêu Thread?

Hãy hỏi:

> Có bao nhiêu Goroutine đang hoạt động?

Chỉ khi:

- Debug Runtime.
- Phân tích Performance.
- Làm việc với CGO.
- Phân tích Thread Dump.

thì Thread mới trở thành mối quan tâm trực tiếp.

---

# Những điểm cần ghi nhớ

Có thể tóm tắt sự khác biệt giữa Goroutine và Thread bằng những ý sau.

- Goroutine là đối tượng của Go Runtime.
- Thread là đối tượng của hệ điều hành.
- Goroutine được Scheduler của Go lập lịch.
- Thread được Kernel lập lịch.
- Goroutine có Stack động.
- Thread thường có Stack cố định.
- Goroutine có thể di chuyển giữa nhiều Machine.
- Thread không thể di chuyển sang Thread khác.
- Một Machine có thể thực thi rất nhiều Goroutine theo thời gian.
- Một Goroutine không gắn cố định với một Thread.

---

# Tóm tắt

Mặc dù Goroutine và OS Thread đều là đơn vị thực thi, nhưng chúng hoạt động ở hai tầng hoàn toàn khác nhau.

Thread là đơn vị thực thi của hệ điều hành, còn Goroutine là đơn vị thực thi của Go Runtime.

Go Runtime sử dụng Scheduler để ánh xạ hàng triệu Goroutine xuống một số lượng rất nhỏ OS Thread, từ đó giảm đáng kể chi phí tạo mới, chi phí Context Switching và mức sử dụng bộ nhớ.

Chính kiến trúc này đã tạo nên khả năng xử lý đồng thời nổi bật của Go và là nền tảng cho toàn bộ các thành phần khác của Go Runtime.

Trong phần tiếp theo, chúng ta sẽ tiếp tục tìm hiểu **vòng đời của một Goroutine**, từ khi được tạo ra cho đến khi kết thúc, đồng thời phân tích chi tiết các trạng thái và sự chuyển đổi giữa chúng trong Go Runtime.

---

---

# 11.3 Vòng đời của Goroutine

Trong Chapter 10, chúng ta đã nhiều lần nhắc đến các trạng thái của Goroutine như:

- Runnable
- Running
- Waiting
- Dead

Tuy nhiên, chúng ta mới chỉ nhìn các trạng thái này từ góc nhìn của **Scheduler**.

Trong phần này, chúng ta sẽ nghiên cứu vòng đời của Goroutine từ góc nhìn của **Go Runtime**.

Mục tiêu không chỉ là biết Goroutine có những trạng thái nào, mà còn hiểu:

- Runtime chuyển trạng thái khi nào?
- Thành phần nào thực hiện việc chuyển đổi?
- Vì sao Goroutine phải liên tục chuyển đổi giữa các trạng thái?
- Scheduler, Netpoller, Channel, Mutex và Timer ảnh hưởng đến vòng đời Goroutine như thế nào?

Đây là nền tảng để hiểu sâu hơn các phần sau như `gopark()`, `goready()`, Context Switching và Garbage Collector.

---

# Góc nhìn tổng quát

Nếu nhìn từ bên ngoài, vòng đời của Goroutine có vẻ rất đơn giản.

```go
go worker()
```

↓

```
Run
```

↓

```
Finish
```

Nhưng bên trong Runtime, Goroutine trải qua nhiều trạng thái khác nhau.

```text
                 New
                  │
                  ▼
              Runnable
                  │
                  ▼
              Running
            ╱    │    ╲
           ▼     ▼     ▼
      Waiting Runnable Dead
           ▲
           │
           └───────────────
```

Trong một chương trình thực tế, Goroutine có thể chuyển đổi giữa:

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

hàng nghìn hoặc hàng triệu lần trước khi kết thúc.

---

# 11.3.1 Trạng thái New

Khi lập trình viên viết:

```go
go worker()
```

Runtime chưa chạy ngay hàm `worker()`.

Thay vào đó, Runtime bắt đầu quá trình khởi tạo Goroutine.

Bao gồm:

- Cấp phát `struct g`.
- Cấp phát Stack ban đầu.
- Khởi tạo Program Counter.
- Khởi tạo Scheduler Metadata.

Trong khoảng thời gian rất ngắn này.

Goroutine ở trạng thái:

```
New
```

Có thể mô tả.

```text
go worker()

↓

Allocate struct g

↓

Allocate Stack

↓

Initialize Context
```

Sau khi hoàn tất.

Runtime chuyển Goroutine sang trạng thái:

```
Runnable
```

Lưu ý rằng trạng thái **New** tồn tại trong thời gian rất ngắn và hầu như lập trình viên sẽ không bao giờ quan sát được.

---

# 11.3.2 Trạng thái Runnable

Runnable là trạng thái quan trọng nhất của Scheduler.

Nó có nghĩa là:

> **Goroutine đã sẵn sàng chạy nhưng chưa được CPU thực thi.**

Có thể hình dung.

```text
Runnable

↓

Waiting for CPU
```

Runtime sẽ đưa Goroutine vào:

- Local Run Queue.
- Hoặc Global Run Queue.

để chờ Scheduler.

Điều cần nhớ là.

```
Runnable

≠

Running
```

Runnable chỉ có nghĩa:

```
Ready to Run
```

---

# 11.3.3 Trạng thái Running

Khi Scheduler chọn Goroutine.

```
Runnable

↓

Running
```

Machine bắt đầu thực thi các lệnh của Goroutine trên CPU.

```text
CPU

↓

Instruction 1

↓

Instruction 2

↓

Instruction 3
```

Đây là trạng thái duy nhất thực sự sử dụng CPU.

Một Goroutine chỉ có thể ở trạng thái Running nếu đồng thời có:

- Một Processor.
- Một Machine.
- Một CPU Core để thực thi.

Nếu thiếu một trong ba thành phần trên.

Goroutine sẽ không thể chạy.

---

# 11.3.4 Trạng thái Waiting

Trong quá trình chạy.

Goroutine có thể phải chờ một sự kiện nào đó.

Ví dụ.

```go
<-ch
```

```go
mutex.Lock()
```

```go
time.Sleep(time.Second)
```

```go
conn.Read(buf)
```

Lúc này Runtime sẽ:

```
Save Context

↓

Remove khỏi CPU

↓

Waiting
```

Điều quan trọng là.

Một Goroutine ở trạng thái Waiting **không sử dụng CPU**.

CPU sẽ ngay lập tức được sử dụng để chạy Goroutine khác.

Đây là điểm tạo nên khả năng đồng thời rất cao của Go.

---

# 11.3.5 Trạng thái Dead

Sau khi hàm kết thúc.

Ví dụ.

```go
func worker() {

    fmt.Println("Done")

}
```

Runtime sẽ chuyển.

```
Running

↓

Dead
```

Dead có nghĩa.

```
Execution Finished
```

Scheduler sẽ không bao giờ chạy Goroutine này nữa.

Sau đó Runtime sẽ:

- Giải phóng Stack (khi thích hợp).
- Thu hồi các cấu trúc dữ liệu nội bộ.
- Đưa một số tài nguyên vào các pool để tái sử dụng.

---

# 11.3.6 Chuyển đổi giữa các trạng thái

Điều thú vị là Goroutine không chỉ đi theo một chiều.

Ví dụ.

```text
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

Waiting

↓

Runnable

↓

Running

↓

Dead
```

Một Goroutine có thể trải qua rất nhiều chu kỳ:

```
Running

↓

Waiting

↓

Runnable
```

trong suốt vòng đời của mình.

Đây là điều hoàn toàn bình thường.

---

# 11.3.7 Điều gì làm Goroutine chuyển trạng thái?

Có nhiều thành phần khác nhau của Go Runtime tham gia vào quá trình này.

| Thành phần | Chuyển đổi |
|------------|------------|
| Scheduler | Runnable → Running |
| Scheduler | Running → Runnable (Preemption) |
| Channel | Waiting → Runnable |
| Mutex | Waiting → Runnable |
| Timer | Waiting → Runnable |
| Netpoller | Waiting → Runnable |
| Runtime | Running → Dead |

Điều này cho thấy Scheduler **không phải là thành phần duy nhất quản lý vòng đời Goroutine**.

---

# 11.3.8 Ví dụ với Channel

```go
func worker(ch chan int) {

    value := <-ch

    fmt.Println(value)

}
```

Ban đầu.

```text
Runnable

↓

Running
```

Khi gặp.

```go
<-ch
```

Runtime phát hiện Channel chưa có dữ liệu.

```
Running

↓

Waiting
```

Sau đó.

Một Goroutine khác.

```go
ch <- 100
```

Runtime đánh thức Goroutine.

```
Waiting

↓

Runnable
```

Scheduler sẽ chọn lại Goroutine.

```
Runnable

↓

Running
```

---

# 11.3.9 Ví dụ với Mutex

```go
mutex.Lock()
```

Nếu Mutex đang được Goroutine khác giữ.

```
Running

↓

Waiting
```

Sau khi.

```go
mutex.Unlock()
```

Runtime sẽ:

```
Waiting

↓

Runnable
```

Scheduler tiếp tục thực thi Goroutine.

---

# 11.3.10 Ví dụ với Network

```go
conn.Read(buf)
```

Khi chưa có dữ liệu.

```
Running

↓

Waiting
```

Netpoller theo dõi Socket.

Sau khi dữ liệu đến.

```
Socket Ready

↓

Netpoller

↓

Runnable
```

Scheduler tiếp tục.

```
Running
```

---

# 11.3.11 Ví dụ với Preemption

Giả sử Goroutine chạy quá lâu.

```go
for {

}
```

Runtime quyết định.

```
Running

↓

Preemption

↓

Runnable
```

Sau đó Scheduler chuyển CPU cho Goroutine khác.

Một lúc sau.

```
Runnable

↓

Running
```

Goroutine tiếp tục từ đúng vị trí trước đó.

---

# 11.3.12 Góc nhìn từ Runtime

Có thể mô tả toàn bộ vòng đời như sau.

```text
                 go func()

                     │

                     ▼

                  New

                     │

                     ▼

                Runnable

                     │

         Scheduler chọn chạy

                     │

                     ▼

                 Running

        ┌────────┼─────────┐
        │        │         │

        ▼        ▼         ▼

    Waiting  Preemption   Dead

        │

        ▼

    Runnable

        │

        └──────────────► Scheduler
```

Đây là sơ đồ quan trọng cần ghi nhớ vì hầu hết các cơ chế của Go Runtime đều dựa trên các chuyển đổi trạng thái này.

---

# 11.3.13 Góc nhìn từ Source Code

Trong Runtime.

Trạng thái của Goroutine được lưu trong trường:

```go
g.atomicstatus
```

Giá trị của trường này sẽ thay đổi liên tục trong suốt vòng đời Goroutine.

Ví dụ.

```
_Grunnable
```

```
_Grunning
```

```
_Gwaiting
```

```
_Gdead
```

Trong các chương tiếp theo, khi phân tích `runtime/runtime2.go`, chúng ta sẽ tìm hiểu chi tiết từng giá trị và cách Runtime thực hiện việc chuyển đổi trạng thái một cách an toàn trong môi trường đa luồng.

---

# 11.3.14 Những hiểu lầm phổ biến

### Hiểu lầm 1

> Goroutine chỉ chạy một lần rồi kết thúc.

Sai.

Một Goroutine có thể nhiều lần chuyển giữa:

```
Running

↓

Waiting

↓

Runnable
```

trước khi kết thúc.

---

### Hiểu lầm 2

> Waiting Goroutine vẫn chiếm CPU.

Sai.

Waiting Goroutine hoàn toàn không sử dụng CPU.

Machine sẽ chuyển sang chạy Goroutine khác.

---

### Hiểu lầm 3

> Scheduler quản lý toàn bộ vòng đời Goroutine.

Không hoàn toàn đúng.

Scheduler chỉ chịu trách nhiệm lập lịch.

Các thành phần như:

- Channel.
- Mutex.
- Timer.
- Netpoller.
- Garbage Collector.

cũng tham gia vào quá trình chuyển đổi trạng thái của Goroutine.

---

# Tóm tắt

Một Goroutine không chỉ đơn giản được tạo ra, chạy rồi kết thúc.

Trong suốt vòng đời của mình, nó liên tục chuyển đổi giữa các trạng thái như **Runnable**, **Running** và **Waiting** dưới sự phối hợp của nhiều thành phần trong Go Runtime.

Hiểu rõ vòng đời của Goroutine giúp chúng ta lý giải được:

- Vì sao Goroutine có thể chờ I/O mà không chiếm CPU.
- Vì sao Scheduler có thể tạm dừng và tiếp tục Goroutine bất kỳ lúc nào.
- Vì sao Channel, Mutex, Timer và Netpoller có thể "đánh thức" Goroutine.
- Vì sao Garbage Collector cần Safe Point để quét Stack.

Trong phần tiếp theo, chúng ta sẽ bắt đầu đi vào cấu trúc dữ liệu quan trọng nhất của Goroutine:

> **`struct g` - trái tim của Goroutine trong Go Runtime.**

Đây là nơi Runtime lưu giữ toàn bộ thông tin cần thiết để Scheduler có thể quản lý, tạm dừng và tiếp tục một Goroutine.

---

---

# Phần II - Cấu trúc dữ liệu nội bộ

Trong phần trước, chúng ta đã tìm hiểu Goroutine từ góc nhìn khái niệm:

- Goroutine là gì?
- Goroutine khác Thread như thế nào?
- Vòng đời của Goroutine.

Tuy nhiên, đó mới chỉ là góc nhìn bên ngoài.

Để thực sự hiểu Go Runtime, chúng ta cần trả lời câu hỏi:

> **Runtime lưu Goroutine ở đâu?**

Câu trả lời là:

```
struct g
```

Đây là một trong những cấu trúc dữ liệu quan trọng nhất trong toàn bộ Go Runtime.

Có thể nói:

- Scheduler quản lý `g`.
- Garbage Collector quét `g`.
- Channel đánh thức `g`.
- Netpoller đánh thức `g`.
- Timer đánh thức `g`.

Hầu như mọi thành phần của Go Runtime đều tương tác với `struct g`.

---

# 11.4 struct g

Trong source code của Go Runtime, Goroutine được biểu diễn bởi một cấu trúc dữ liệu có tên:

```go
type g struct {
    ...
}
```

Cấu trúc này được định nghĩa trong:

```
runtime/runtime2.go
```

Đây là một struct rất lớn.

Tùy theo phiên bản Go, `struct g` có thể chứa hơn một trăm trường dữ liệu.

Điều này cho thấy một Goroutine không chỉ đơn thuần là một Stack.

Nó là một đối tượng Runtime rất phức tạp.

---

# Vai trò của struct g

Có thể hình dung `struct g` như "hồ sơ" của một Goroutine.

Trong hồ sơ này Runtime lưu mọi thông tin cần thiết để:

- Chạy Goroutine.
- Tạm dừng Goroutine.
- Tiếp tục Goroutine.
- Di chuyển Goroutine sang Machine khác.
- Garbage Collector quét Stack.
- Panic.
- Defer.
- Tracing.
- Debugging.

Có thể minh họa như sau.

```text
                 struct g

        +---------------------------+
        | Stack                     |
        +---------------------------+
        | Program Counter           |
        +---------------------------+
        | Stack Pointer             |
        +---------------------------+
        | Goroutine Status          |
        +---------------------------+
        | Current Machine           |
        +---------------------------+
        | Scheduler Context         |
        +---------------------------+
        | Panic                     |
        +---------------------------+
        | Defer                     |
        +---------------------------+
        | GC Metadata               |
        +---------------------------+
```

Khi Scheduler muốn chuyển sang Goroutine khác.

Nó chỉ cần:

```
Save struct g

↓

Load struct g khác
```

CPU có thể tiếp tục chạy từ đúng vị trí trước đó.

---

# Cấu trúc tổng quát

Trong phiên bản Go hiện đại.

`struct g` có thể được chia thành các nhóm thông tin sau.

```text
struct g

├── Stack
├── Scheduler Context
├── Runtime State
├── Machine Information
├── Panic / Defer
├── GC Metadata
├── Synchronization
├── Timer
├── Tracing
└── Debug Information
```

Trong các phần tiếp theo chúng ta sẽ phân tích từng nhóm.

---

# Nhóm 1 - Stack Information

Đây là nhóm quan trọng nhất.

Ví dụ.

```go
stack
```

```
stackguard0
```

```
stackguard1
```

Nhóm này mô tả:

- Stack nằm ở đâu.
- Stack bắt đầu từ đâu.
- Stack kết thúc ở đâu.
- Khi nào cần mở rộng Stack.

Ví dụ.

```text
Stack

High Address

+----------------+

|

|

|

+----------------+

Low Address
```

Runtime sử dụng thông tin này để phát hiện:

- Stack Overflow.
- Stack Growth.
- Stack Shrink.

Chúng ta sẽ phân tích chi tiết trong Chapter Stack.

---

# Nhóm 2 - Scheduler Context

Đây là nhóm dữ liệu phục vụ Context Switching.

Ví dụ.

```go
sched gobuf
```

Đây là nơi Runtime lưu:

- Program Counter.
- Stack Pointer.
- Registers.

Có thể hình dung.

```text
Running

↓

Save Context

↓

sched

↓

Restore Context

↓

Continue
```

Không có `sched`.

Scheduler sẽ không thể tạm dừng và tiếp tục Goroutine.

---

# Nhóm 3 - Machine Information

Một Goroutine đang Running sẽ biết mình đang chạy trên Machine nào.

Ví dụ.

```go
m *m
```

Đây là con trỏ tới Machine hiện tại.

Ví dụ.

```text
G1

↓

M0
```

Sau khi bị Preempt.

```text
G1

↓

M2
```

Giá trị này sẽ thay đổi.

Điều này cho phép Runtime di chuyển Goroutine giữa nhiều Thread.

---

# Nhóm 4 - Runtime State

Một trong những trường quan trọng nhất là.

```go
atomicstatus
```

Đây là trạng thái hiện tại của Goroutine.

Ví dụ.

```
_Grunnable
```

```
_Grunning
```

```
_Gwaiting
```

```
_Gdead
```

Scheduler liên tục đọc và cập nhật trường này.

Đây là "nguồn sự thật" về trạng thái của Goroutine.

---

# Nhóm 5 - Panic và Defer

Khi chương trình sử dụng.

```go
defer
```

hoặc.

```go
panic()
```

Runtime cần lưu các thông tin liên quan.

Ví dụ.

```go
_defer
```

```go
_panic
```

Có thể hình dung.

```text
panic()

↓

struct g

↓

panic stack
```

Các trường này giúp Runtime:

- Thực thi `defer`.
- Lan truyền `panic`.
- Thực hiện `recover`.

Chúng ta sẽ phân tích chi tiết ở phần sau.

---

# Nhóm 6 - Garbage Collector

Garbage Collector cũng cần rất nhiều thông tin từ Goroutine.

Ví dụ.

- Stack.
- Pointer.
- Scan State.

GC sẽ đọc thông tin trong `struct g` để:

```
Scan Stack

↓

Find Live Objects
```

Nếu không có các thông tin này.

Garbage Collector không thể xác định Object nào còn được sử dụng.

---

# Nhóm 7 - Synchronization

Trong quá trình chờ.

Ví dụ.

```go
mutex.Lock()
```

hoặc.

```go
<-channel
```

Runtime cần biết.

```
Waiting Reason
```

Ví dụ.

```
Channel Receive
```

```
Mutex Wait
```

```
Sleep
```

```
IO Wait
```

Thông tin này được lưu trong `struct g`.

Nhờ đó.

Khi Debug.

Chúng ta có thể thấy.

```
goroutine 18 [chan receive]
```

hoặc.

```
goroutine 25 [sleep]
```

---

# Nhóm 8 - Tracing

Go Runtime hỗ trợ.

```
runtime/trace
```

và.

```
pprof
```

Để thực hiện việc Trace.

Runtime cần lưu nhiều Metadata.

Ví dụ.

- Goroutine ID.
- Labels.
- Trace State.

Các thông tin này đều được lưu trong `struct g`.

---

# struct g không nằm trên Stack

Đây là một điểm rất nhiều người hiểu nhầm.

Nhiều lập trình viên nghĩ.

```
Goroutine

↓

Stack
```

Thực tế.

`struct g` và Stack là **hai đối tượng khác nhau**.

Có thể hình dung.

```text
                struct g

        +---------------------+

        | stack pointer        |

        | scheduler state      |

        | panic               |

        | defer               |

        +---------------------+

                  │

                  ▼

              Stack Memory
```

`struct g` chỉ chứa:

```
Pointer

↓

Stack
```

Stack là một vùng nhớ riêng.

Điều này rất quan trọng.

Bởi vì.

Sau này Runtime có thể:

```
Move Stack
```

mà không cần di chuyển `struct g`.

---

# struct g có được tạo thường xuyên không?

Có.

Mỗi lần.

```go
go worker()
```

Runtime sẽ cần một `struct g`.

Tuy nhiên.

Để giảm chi phí cấp phát.

Runtime thường tái sử dụng các `struct g` đã kết thúc thông qua các pool nội bộ.

Điều này giúp:

- Giảm Allocation.
- Giảm GC Pressure.
- Tăng hiệu năng.

---

# Góc nhìn tổng thể

Có thể hình dung mối quan hệ giữa Goroutine và `struct g`.

```text
                Goroutine

                     │

                     ▼

                 struct g

     ┌────────────┼────────────┐

     ▼            ▼            ▼

 Stack       Scheduler      Runtime

               Context        State

     ▼            ▼            ▼

 Panic       Defer        GC Metadata
```

`struct g` chính là "bản sắc" của Goroutine.

Mọi thành phần trong Go Runtime đều làm việc thông qua cấu trúc dữ liệu này.

---

# Những điều cần ghi nhớ

- Một Goroutine được biểu diễn bởi `struct g`.
- `struct g` được định nghĩa trong `runtime/runtime2.go`.
- `struct g` không phải là Stack.
- `struct g` chỉ giữ con trỏ đến Stack.
- Scheduler quản lý `struct g`, không quản lý Function.
- Garbage Collector quét Stack thông qua thông tin trong `struct g`.
- Mọi lần Context Switching đều liên quan đến `struct g`.

---

# Tóm tắt

`struct g` là cấu trúc dữ liệu trung tâm đại diện cho một Goroutine trong Go Runtime.

Nó không chỉ lưu Stack mà còn chứa toàn bộ trạng thái thực thi, thông tin Scheduler, Panic, Defer, Garbage Collector và nhiều Metadata khác.

Hiểu `struct g` là bước đầu tiên để đọc và hiểu source code của Go Runtime, bởi gần như mọi thành phần quan trọng như Scheduler, Garbage Collector, Netpoller hay Channel đều tương tác trực tiếp hoặc gián tiếp với cấu trúc dữ liệu này.

Trong phần tiếp theo, chúng ta sẽ đi sâu vào trường quan trọng nhất bên trong `struct g`:

> **`gobuf` - nơi Runtime lưu Program Counter, Stack Pointer và toàn bộ Execution Context để thực hiện Context Switching.**

---

---

# 11.5 `gobuf` - Trái tim của Context Switching

Trong phần trước, chúng ta đã biết rằng `struct g` là "hồ sơ" của một Goroutine.

Bên trong `struct g` có rất nhiều trường dữ liệu, nhưng có một trường đặc biệt quan trọng.

```go
sched gobuf
```

Đây là nơi Go Runtime lưu **Execution Context** của Goroutine.

Nếu `struct g` đại diện cho toàn bộ Goroutine thì `gobuf` chính là phần giúp Scheduler có thể:

- Tạm dừng Goroutine.
- Khôi phục Goroutine.
- Chuyển CPU sang Goroutine khác.
- Tiếp tục Goroutine từ đúng vị trí trước đó.

Có thể nói:

> **Không có `gobuf` thì sẽ không có Goroutine Scheduling.**

---

# Context Switching là gì?

Hãy bắt đầu bằng một câu hỏi.

Giả sử một Goroutine đang chạy đoạn mã sau.

```go
func worker() {

    step1()

    step2()

    step3()

}
```

Scheduler quyết định chuyển CPU sang Goroutine khác ngay sau:

```go
step2()
```

Một lúc sau.

Scheduler cần tiếp tục Goroutine này.

Câu hỏi là.

> **Làm thế nào Runtime biết phải tiếp tục từ đâu?**

Nếu Runtime chạy lại từ đầu.

```
step1()

↓

step2()

↓

step3()
```

thì chương trình sẽ sai.

Runtime phải tiếp tục đúng tại vị trí:

```
step3()
```

Muốn làm được điều đó.

Runtime phải lưu toàn bộ trạng thái của CPU.

Đó chính là nhiệm vụ của `gobuf`.

---

# `gobuf` là gì?

Có thể hiểu đơn giản.

`gobuf` là một cấu trúc dữ liệu dùng để lưu:

> **Execution Context của Goroutine.**

Nó chứa đủ thông tin để Runtime có thể:

```
Pause

↓

Save Context

↓

Restore Context

↓

Continue
```

Việc chuyển đổi này diễn ra trong thời gian rất ngắn và hoàn toàn trong User Space.

---

# Góc nhìn trực quan

Hãy tưởng tượng CPU giống như một người đang đọc sách.

```
Trang 100
```

Nếu muốn tạm dừng.

Bạn cần một chiếc **bookmark**.

```
Trang 100

↓

Bookmark
```

Ngày hôm sau.

Bạn chỉ cần mở đúng dấu trang.

```
Bookmark

↓

Trang 100

↓

Continue Reading
```

`gobuf` chính là chiếc bookmark của Goroutine.

Nó ghi nhớ:

- Đang thực thi ở đâu.
- Stack hiện tại ở đâu.
- CPU đang giữ những giá trị gì.

---

# `gobuf` nằm ở đâu?

Trong `struct g`.

```go
type g struct {

    ...

    sched gobuf

    ...

}
```

Có thể hình dung.

```text
                struct g

        +----------------------+

        | Stack                |

        +----------------------+

        | sched (gobuf)        |

        +----------------------+

        | Status               |

        +----------------------+
```

Mỗi Goroutine đều có một `gobuf` riêng.

Điều này cho phép Runtime lưu Execution Context độc lập cho từng Goroutine.

---

# Bên trong `gobuf`

Trong source code của Go Runtime.

`gobuf` có dạng gần giống như sau.

```go
type gobuf struct {

    sp   uintptr

    pc   uintptr

    g    guintptr

    ctxt unsafe.Pointer

    bp   uintptr

    ...

}
```

(Lưu ý: cấu trúc có thể thay đổi đôi chút giữa các phiên bản Go.)

Nhìn qua có vẻ đơn giản.

Nhưng mỗi trường đều có vai trò cực kỳ quan trọng.

---

# Program Counter (PC)

Trường đầu tiên.

```
pc
```

Program Counter lưu:

> **Địa chỉ của lệnh CPU sẽ thực thi tiếp theo.**

Ví dụ.

```go
foo()

bar()

baz()
```

Nếu Runtime tạm dừng sau:

```go
bar()
```

Program Counter sẽ trỏ tới.

```go
baz()
```

Khi Scheduler khôi phục.

CPU sẽ tiếp tục từ:

```
baz()
```

chứ không phải quay lại đầu hàm.

Đây là nền tảng của Context Switching.

---

# Stack Pointer (SP)

Trường tiếp theo.

```
sp
```

Stack Pointer trỏ tới:

```
Đỉnh của Stack
```

Ví dụ.

```text
High Address

+----------------+

|

|

SP

|

+----------------+

Low Address
```

Khi Runtime khôi phục Goroutine.

Giá trị SP cũng được khôi phục.

CPU sẽ tiếp tục sử dụng đúng Stack trước đó.

---

# Base Pointer (BP)

Trên một số kiến trúc CPU.

Runtime cũng lưu.

```
bp
```

Base Pointer giúp CPU xác định:

- Stack Frame hiện tại.
- Biến cục bộ.
- Tham số hàm.

Không phải mọi kiến trúc đều sử dụng BP giống nhau, nhưng Runtime vẫn hỗ trợ khi cần.

---

# Con trỏ đến Goroutine

Một trường khác.

```
g
```

Trường này trỏ ngược trở lại.

```
struct g
```

Có thể hình dung.

```text
gobuf

↓

struct g
```

Nhờ vậy.

Khi Runtime khôi phục Context.

Nó biết Context này thuộc Goroutine nào.

---

# Context (`ctxt`)

Một số trường hợp.

Runtime cần truyền thêm thông tin.

Ví dụ.

- Closure.
- Deferred Call.
- Trampoline.

Thông tin này được lưu trong.

```
ctxt
```

Đây là trường phục vụ nhiều mục đích nội bộ của Runtime.

---

# `gobuf` không lưu toàn bộ Stack

Đây là một hiểu lầm rất phổ biến.

Nhiều người nghĩ.

```
gobuf

↓

Stack
```

Điều này không đúng.

`gobuf` chỉ lưu.

```
SP

↓

Pointer tới Stack
```

Stack thực sự nằm ở một vùng nhớ khác.

Có thể hình dung.

```text
            gobuf

      +---------------+

      | PC            |

      | SP ----------+---------+

      | BP           |         |

      +---------------+         |

                                ▼

                          Stack Memory
```

Điều này giúp `gobuf` rất nhỏ và việc Context Switching diễn ra nhanh.

---

# Runtime sử dụng `gobuf` khi nào?

`gobuf` được sử dụng trong rất nhiều tình huống.

Ví dụ.

## Scheduler

```
Save Context

↓

Run Goroutine khác
```

---

## Preemption

```
Pause

↓

Save PC/SP

↓

Continue sau
```

---

## Blocking System Call

```
Save Context

↓

Detach Processor
```

---

## Garbage Collector

GC cần biết Stack bắt đầu từ đâu để quét.

Thông tin này được lấy gián tiếp thông qua `gobuf` và `struct g`.

---

# Minh họa Context Switching

Giả sử.

```
G1

↓

Running
```

Runtime quyết định chuyển sang.

```
G2
```

Quá trình diễn ra.

```text
G1

↓

Save PC

↓

Save SP

↓

Save Registers

↓

gobuf
```

Sau đó.

```text
Load G2

↓

Restore PC

↓

Restore SP

↓

Restore Registers

↓

Continue
```

Đối với CPU.

Việc chuyển đổi gần như diễn ra tức thì.

---

# Tại sao `gobuf` phải nhỏ?

Scheduler thực hiện Context Switching rất thường xuyên.

Có thể hàng triệu lần mỗi giây.

Nếu `gobuf` quá lớn.

Ví dụ.

```
100 KB
```

Mỗi lần Context Switch sẽ phải sao chép rất nhiều dữ liệu.

Điều này làm Scheduler chậm đi đáng kể.

Do đó.

Runtime chỉ lưu những thông tin tối thiểu cần thiết.

- Program Counter.
- Stack Pointer.
- Một số Registers.
- Metadata quan trọng.

Mọi thứ khác đều nằm trong Stack hoặc `struct g`.

Đây là một quyết định thiết kế rất quan trọng.

---

# Mối quan hệ giữa `struct g`, `gobuf` và Stack

Có thể mô tả như sau.

```text
                struct g

        +----------------------+

        | Status               |

        | Panic                |

        | Defer                |

        | Stack Pointer ------+

        | sched (gobuf)        |----+

        +----------------------+    |

                                    |

                                    ▼

                               Stack Memory

                                    ▲

                                    |

                             Program Counter

                             Stack Pointer

                             Base Pointer
```

`struct g` quản lý toàn bộ Goroutine.

`gobuf` quản lý Execution Context.

Stack lưu dữ liệu thực thi.

Ba thành phần này luôn đi cùng nhau trong suốt vòng đời Goroutine.

---

# Những điều cần ghi nhớ

- `gobuf` là nơi Runtime lưu Execution Context.
- `gobuf` nằm trong `struct g`.
- `gobuf` không chứa Stack.
- `gobuf` chỉ chứa các con trỏ và thanh ghi quan trọng.
- Scheduler dựa vào `gobuf` để thực hiện Context Switching.
- Program Counter quyết định Goroutine sẽ tiếp tục từ đâu.
- Stack Pointer quyết định Goroutine sử dụng Stack nào.

---

# Tóm tắt

`gobuf` là một trong những cấu trúc dữ liệu quan trọng nhất của Go Runtime.

Nó lưu toàn bộ ngữ cảnh thực thi cần thiết để Scheduler có thể tạm dừng một Goroutine và tiếp tục nó sau đó mà không làm thay đổi hành vi của chương trình.

Mặc dù có kích thước nhỏ, `gobuf` đóng vai trò trung tâm trong Context Switching và là cầu nối giữa Scheduler với CPU.

Trong phần tiếp theo, chúng ta sẽ rời khỏi Execution Context để nghiên cứu thành phần quan trọng không kém:

> **Stack của Goroutine** - nơi lưu trữ Stack Frame, biến cục bộ, tham số hàm và toàn bộ quá trình thực thi của Goroutine.

Đây cũng là nền tảng để hiểu cơ chế **Stack Growth**, **Stack Copy** và **Stack Shrink** trong các chương tiếp theo.

---

---

# 11.6 Stack - Nơi Goroutine thực thi

Trong phần trước, chúng ta đã tìm hiểu về `gobuf` - nơi Runtime lưu Execution Context của Goroutine.

Tuy nhiên, `gobuf` chỉ lưu một lượng thông tin rất nhỏ:

- Program Counter.
- Stack Pointer.
- Một số Registers.
- Metadata của Scheduler.

Điều đó dẫn đến một câu hỏi quan trọng.

> **Vậy toàn bộ biến cục bộ, tham số hàm, giá trị trả về và lời gọi hàm được lưu ở đâu?**

Câu trả lời là:

```
Stack
```

Stack là một trong những thành phần quan trọng nhất của Goroutine.

Nếu Scheduler là **bộ não** điều phối Goroutine.

Thì Stack chính là **không gian làm việc** của Goroutine.

Mọi lời gọi hàm, biến cục bộ và quá trình thực thi đều diễn ra trên Stack.

---

# Stack là gì?

Stack là một vùng nhớ đặc biệt được sử dụng để lưu trữ:

- Tham số của hàm.
- Giá trị trả về.
- Biến cục bộ.
- Địa chỉ quay về (Return Address).
- Stack Frame.

Có thể hình dung.

```text
                Stack

        +----------------------+

        | Frame main()         |

        +----------------------+

        | Frame process()      |

        +----------------------+

        | Frame validate()     |

        +----------------------+

        | Frame saveDB()       |

        +----------------------+
```

Mỗi khi gọi một hàm mới.

Runtime sẽ tạo thêm một **Stack Frame**.

---

# Tại sao gọi là Stack?

Tên gọi Stack xuất phát từ cách dữ liệu được quản lý.

Stack hoạt động theo nguyên tắc.

```
LIFO

Last In

First Out
```

Ví dụ.

```go
main()

↓

foo()

↓

bar()

↓

baz()
```

Thứ tự Push.

```
main

↓

foo

↓

bar

↓

baz
```

Thứ tự Pop.

```
baz

↓

bar

↓

foo

↓

main
```

Đây chính là nguyên lý hoạt động của Call Stack trong hầu hết các ngôn ngữ lập trình.

---

# Stack của Goroutine

Mỗi Goroutine có:

```
Một Stack riêng
```

Ví dụ.

```text
             G1

        +-------------+

        | Stack G1    |

        +-------------+

             G2

        +-------------+

        | Stack G2    |

        +-------------+

             G3

        +-------------+

        | Stack G3    |

        +-------------+
```

Điều này rất quan trọng.

Bởi vì các Goroutine có thể:

- Chạy độc lập.
- Tạm dừng độc lập.
- Tiếp tục độc lập.

Nếu dùng chung một Stack.

Điều đó gần như không thể thực hiện được.

---

# Stack và Heap

Đây là hai vùng nhớ thường bị nhầm lẫn.

| Stack | Heap |
|---------|------|
| Mỗi Goroutine có một Stack | Toàn bộ Goroutine dùng chung Heap |
| Quản lý theo Stack Frame | Quản lý bởi Memory Allocator |
| Truy cập rất nhanh | Chậm hơn |
| Không cần Garbage Collector | Được Garbage Collector quản lý |

Ví dụ.

```go
func calculate() {

    x := 10

}
```

Thông thường.

```
x

↓

Stack
```

Nếu biến:

- Escape khỏi hàm.
- Được tham chiếu bên ngoài.

Runtime có thể đưa nó lên:

```
Heap
```

Thông qua Escape Analysis.

---

# Stack nằm ở đâu?

Một hiểu lầm phổ biến là.

```
Stack

↓

CPU
```

Điều này không đúng.

Stack là một vùng nhớ trong:

```
RAM
```

Có thể hình dung.

```text
Memory

+--------------------------+

| Heap                     |

+--------------------------+

|                          |

| Goroutine Stack          |

|                          |

+--------------------------+

| Runtime                  |

+--------------------------+
```

CPU chỉ sử dụng:

```
Stack Pointer (SP)
```

để biết đỉnh hiện tại của Stack.

---

# Stack tăng theo hướng nào?

Trên hầu hết các kiến trúc CPU hiện đại.

Stack phát triển từ:

```
High Address

↓

Low Address
```

Có thể hình dung.

```text
High Address

+--------------------+

|                    |

|        main()      |

+--------------------+

|      process()     |

+--------------------+

|      validate()    |

+--------------------+

|      saveDB()      |

+--------------------+

Low Address
```

Mỗi lời gọi hàm sẽ làm Stack Pointer giảm xuống.

Sau khi hàm kết thúc.

Stack Pointer tăng trở lại.

---

# Điều gì xảy ra khi gọi một hàm?

Giả sử.

```go
func main() {

    foo()

}
```

```go
func foo() {

    bar()

}
```

```go
func bar() {

}
```

Quá trình.

```text
main()

↓

Push Frame

↓

foo()

↓

Push Frame

↓

bar()
```

Sau khi.

```
bar()

↓

Return
```

Frame của `bar()` được Pop khỏi Stack.

Tiếp theo.

```
foo()

↓

Return
```

Stack lại tiếp tục được thu gọn.

---

# Stack Frame là gì?

Mỗi lời gọi hàm tương ứng với một:

```
Stack Frame
```

Có thể hình dung.

```text
Stack

+-------------------------+

| Return Address          |

| Parameters              |

| Local Variables         |

| Temporary Values        |

+-------------------------+
```

Mỗi Stack Frame là một "không gian làm việc" của một lời gọi hàm.

Các Stack Frame xếp chồng lên nhau tạo thành Call Stack.

---

# Stack Pointer (SP)

CPU luôn giữ một thanh ghi đặc biệt.

```
SP

↓

Stack Pointer
```

Thanh ghi này luôn trỏ tới:

```
Current Stack Frame
```

Ví dụ.

```text
Stack

+----------------+

|

|

SP

|

+----------------+
```

Khi Runtime thực hiện Context Switching.

Giá trị SP sẽ được lưu vào:

```
gobuf.sp
```

Sau khi khôi phục.

CPU tiếp tục sử dụng đúng Stack trước đó.

---

# Program Counter và Stack

Program Counter và Stack luôn đi cùng nhau.

Ví dụ.

```
PC

↓

Instruction
```

```
SP

↓

Current Frame
```

Nếu chỉ lưu Program Counter mà không lưu Stack Pointer.

Runtime sẽ không biết các biến cục bộ đang nằm ở đâu.

Ngược lại.

Nếu chỉ lưu Stack Pointer.

Runtime sẽ không biết phải tiếp tục thực thi từ dòng lệnh nào.

Do đó.

Context Switching luôn cần cả:

```
PC

+

SP
```

---

# Stack của Goroutine rất nhỏ

Một trong những điểm đặc biệt nhất của Go là:

```
Initial Stack

≈ 2 KB
```

Đây là kích thước rất nhỏ nếu so với Thread truyền thống.

Tại sao Runtime dám làm điều này?

Bởi vì Stack của Goroutine:

```
Dynamic
```

Nó có thể:

- Grow.
- Copy.
- Shrink.

Điều này sẽ được phân tích rất chi tiết trong các phần tiếp theo.

---

# Vì sao Stack nhỏ lại quan trọng?

Giả sử.

Một Goroutine cần:

```
8 MB
```

Stack ngay khi được tạo.

Nếu chương trình tạo:

```
1.000.000 Goroutines
```

Bộ nhớ chỉ dành cho Stack sẽ lên tới:

```
8 PB
```

Điều này hoàn toàn không khả thi.

Thay vào đó.

Go Runtime chỉ cấp phát khoảng:

```
2 KB
```

cho mỗi Goroutine.

Nếu Goroutine thực sự cần thêm bộ nhớ.

Runtime mới mở rộng Stack.

Đây là một trong những lý do giúp Go có thể tạo hàng triệu Goroutine.

---

# Stack không phải là cố định

Một khác biệt lớn giữa Goroutine và Thread là.

```
Thread Stack

↓

Thường cố định
```

Trong khi.

```
Goroutine Stack

↓

Dynamic
```

Runtime có thể:

```
Grow

↓

Copy

↓

Shrink
```

mà lập trình viên hoàn toàn không cần can thiệp.

Đây là một trong những thành tựu quan trọng nhất của Go Runtime.

---

# Mối quan hệ giữa Stack và struct g

Có thể hình dung.

```text
             struct g

      +----------------------+

      | Stack Pointer -------+

      | gobuf               |

      | Status              |

      +----------------------+


                 │

                 ▼

             Stack Memory

      +----------------------+

      | Frame 1             |

      +----------------------+

      | Frame 2             |

      +----------------------+

      | Frame 3             |

      +----------------------+
```

`struct g` không chứa Stack.

Nó chỉ lưu:

```
Pointer

↓

Stack
```

Điều này cho phép Runtime di chuyển Stack khi cần mà không phải di chuyển toàn bộ `struct g`.

---

# Những điều cần ghi nhớ

- Mỗi Goroutine có một Stack riêng.
- Stack lưu Stack Frame, biến cục bộ và tham số hàm.
- Stack hoạt động theo nguyên tắc LIFO.
- Stack Pointer luôn trỏ tới Stack Frame hiện tại.
- Program Counter và Stack Pointer luôn được lưu trong `gobuf`.
- Stack của Goroutine ban đầu rất nhỏ và có thể tự mở rộng.
- Stack và Heap là hai vùng nhớ hoàn toàn khác nhau.

---

# Tóm tắt

Stack là không gian thực thi của Goroutine.

Mọi lời gọi hàm, biến cục bộ, tham số và giá trị tạm thời đều được lưu trong các Stack Frame trên Stack.

Khác với Stack của Thread truyền thống, Stack của Goroutine có kích thước ban đầu rất nhỏ và có thể tự động mở rộng hoặc thu nhỏ theo nhu cầu của chương trình.

Thiết kế này giúp Go Runtime vừa tiết kiệm bộ nhớ vừa hỗ trợ tạo ra hàng triệu Goroutine mà vẫn duy trì hiệu năng cao.

Trong phần tiếp theo, chúng ta sẽ đi sâu hơn vào thành phần quan trọng nhất của Stack:

> **Stack Frame** - đơn vị cơ bản tạo nên Call Stack và là nơi Runtime lưu toàn bộ thông tin của mỗi lời gọi hàm.

Hiểu rõ Stack Frame sẽ giúp chúng ta lý giải cơ chế gọi hàm, truyền tham số và chuẩn bị nền tảng cho việc nghiên cứu **Stack Growth**, **Stack Copy** và **Context Switching** trong các phần tiếp theo.

---

---

# 11.7 Stack Frame - Đơn vị cơ bản của Stack

Trong phần trước, chúng ta đã biết rằng Stack là nơi Goroutine thực thi.

Tuy nhiên, Stack không phải là một vùng nhớ "phẳng".

Nó được chia thành nhiều phần nhỏ.

Mỗi phần tương ứng với **một lần gọi hàm**.

Những phần nhỏ này được gọi là:

```
Stack Frame
```

Có thể xem Stack Frame là đơn vị cơ bản cấu thành nên Stack.

Nếu Stack giống như một tòa nhà nhiều tầng.

Thì mỗi Stack Frame chính là một căn phòng trong tòa nhà đó.

Mỗi lần gọi một hàm mới.

Runtime sẽ tạo thêm một Stack Frame mới.

Khi hàm kết thúc.

Stack Frame đó sẽ bị loại bỏ.

---

# Tại sao cần Stack Frame?

Hãy xem ví dụ.

```go
func main() {

    process()

}
```

```go
func process() {

    validate()

}
```

```go
func validate() {

}
```

Nếu tất cả các hàm cùng sử dụng một vùng nhớ.

```text
main

↓

process

↓

validate
```

Các biến cục bộ sẽ ghi đè lên nhau.

Ví dụ.

```go
func foo() {

    x := 10

}
```

```go
func bar() {

    x := 20

}
```

Nếu không có Stack Frame.

Hai biến:

```
x
```

sẽ cùng sử dụng một địa chỉ.

Điều này rõ ràng là không thể.

Do đó.

Mỗi lời gọi hàm cần một không gian riêng.

Đó chính là Stack Frame.

---

# Một lời gọi hàm tương ứng một Stack Frame

Giả sử.

```go
func main() {

    foo()

}
```

```go
func foo() {

    bar()

}
```

```go
func bar() {

}
```

Quá trình thực thi.

```text
main()

↓

Push Frame(main)

↓

foo()

↓

Push Frame(foo)

↓

bar()

↓

Push Frame(bar)
```

Lúc này.

Stack sẽ có dạng.

```text
High Address

+---------------------------+

| Frame main()              |

+---------------------------+

| Frame foo()               |

+---------------------------+

| Frame bar()               |

+---------------------------+

Low Address
```

Sau khi.

```
bar()

↓

Return
```

Frame của `bar()` bị Pop khỏi Stack.

---

# Bên trong một Stack Frame có gì?

Một Stack Frame thường chứa nhiều loại dữ liệu.

Có thể hình dung.

```text
+-----------------------------+

| Return Address              |

+-----------------------------+

| Saved Registers             |

+-----------------------------+

| Parameters                  |

+-----------------------------+

| Local Variables             |

+-----------------------------+

| Temporary Values            |

+-----------------------------+
```

Không phải mọi hàm đều có đầy đủ các thành phần này.

Tuy nhiên đây là cấu trúc tổng quát mà Runtime và CPU cùng phối hợp sử dụng.

---

# Return Address

Đây là thành phần quan trọng nhất.

Giả sử.

```go
func main() {

    foo()

    fmt.Println("Done")

}
```

Khi CPU gọi.

```go
foo()
```

Một câu hỏi xuất hiện.

> Sau khi `foo()` kết thúc thì CPU quay về đâu?

Câu trả lời là.

```
Return Address
```

CPU sẽ lưu địa chỉ của lệnh tiếp theo.

```go
fmt.Println("Done")
```

vào Stack Frame.

Sau khi `foo()` kết thúc.

CPU đọc lại địa chỉ này và tiếp tục thực thi.

Có thể mô tả.

```text
main()

↓

CALL foo()

↓

Save Return Address

↓

foo()

↓

RET

↓

Continue
```

---

# Parameters

Khi gọi hàm.

```go
sum(10, 20)
```

Các tham số cần được truyền cho hàm.

Runtime sẽ đặt chúng vào Stack Frame.

```text
Frame sum()

+-------------------+

| a = 10            |

| b = 20            |

+-------------------+
```

(Tùy theo kiến trúc CPU và tối ưu của Compiler, một số tham số có thể được truyền qua Register. Tuy nhiên về mặt khái niệm, chúng vẫn thuộc ngữ cảnh của Stack Frame.)

---

# Local Variables

Ví dụ.

```go
func calculate() {

    x := 10

    y := 20

}
```

Thông thường.

```
x
```

và.

```
y
```

được lưu trong Stack Frame.

```text
Frame calculate()

+-------------------+

| x = 10            |

| y = 20            |

+-------------------+
```

Khi hàm kết thúc.

Toàn bộ vùng nhớ này sẽ không còn được sử dụng.

---

# Temporary Values

Trong quá trình tính toán.

CPU cũng cần lưu các giá trị trung gian.

Ví dụ.

```go
result := (a+b)*(c+d)
```

Trong lúc thực hiện.

Các giá trị.

```
a+b
```

và.

```
c+d
```

có thể được lưu tạm.

Đây chính là Temporary Values.

Compiler sẽ quyết định:

- Lưu vào Register.
- Hoặc lưu vào Stack Frame.

---

# Saved Registers

Một số Register của CPU cần được bảo toàn khi gọi hàm.

Ví dụ.

```
RBP
```

```
R12
```

```
R13
```

Runtime hoặc Compiler sẽ lưu các Register này vào Stack Frame.

Sau khi hàm kết thúc.

Chúng sẽ được khôi phục.

Đây là một phần quan trọng của Calling Convention.

---

# Minh họa Call Stack

Giả sử.

```go
main()

↓

login()

↓

validate()

↓

checkPassword()
```

Trong lúc.

```
checkPassword()
```

đang chạy.

Call Stack sẽ có dạng.

```text
High Address

+-----------------------------+

| Frame main()                |

+-----------------------------+

| Frame login()               |

+-----------------------------+

| Frame validate()            |

+-----------------------------+

| Frame checkPassword()       |

+-----------------------------+

Low Address
```

Frame trên cùng luôn là:

```
Current Function
```

---

# Khi hàm trả về

Sau khi.

```go
checkPassword()
```

kết thúc.

Stack trở thành.

```text
High Address

+-----------------------------+

| Frame main()                |

+-----------------------------+

| Frame login()               |

+-----------------------------+

| Frame validate()            |

+-----------------------------+

Low Address
```

Sau đó.

```
validate()

↓

Return
```

Frame tiếp tục bị loại bỏ.

Stack luôn phản ánh chính xác chuỗi lời gọi hàm hiện tại.

---

# Stack Pointer thay đổi như thế nào?

Giả sử.

```
SP
```

đang trỏ tới.

```
Frame login()
```

Khi gọi.

```go
validate()
```

CPU thực hiện.

```
Push Frame
```

SP giảm xuống.

```text
Before

SP

↓

login()
```

```text
After

SP

↓

validate()
```

Sau khi.

```
RET
```

SP tăng trở lại.

---

# Program Counter và Stack Frame

Program Counter và Stack Frame luôn phối hợp với nhau.

```
Program Counter

↓

Instruction tiếp theo
```

```
Stack Pointer

↓

Current Frame
```

Nếu chỉ có Program Counter.

CPU sẽ không biết.

- Biến cục bộ ở đâu.
- Tham số nằm ở đâu.

Nếu chỉ có Stack Pointer.

CPU sẽ không biết.

- Tiếp tục thực thi từ lệnh nào.

Do đó.

Mỗi Context Switch đều phải lưu cả:

```
PC

+

SP
```

---

# Stack Frame và Goroutine

Một Goroutine có thể có:

```
1

10

100

1000
```

Stack Frame.

Tùy thuộc vào độ sâu của lời gọi hàm.

Ví dụ.

```go
main()

↓

A()

↓

B()

↓

C()

↓

D()

↓

E()
```

Lúc này Stack sẽ có.

```
6 Stack Frames
```

Mỗi Stack Frame đại diện cho một lời gọi hàm đang hoạt động.

---

# Stack Frame và Garbage Collector

Garbage Collector cần quét:

- Local Variables.
- Pointer.
- Reference.

Các thông tin này đều nằm trong Stack Frame.

Có thể hình dung.

```text
Stack

↓

Frame

↓

Pointer

↓

Heap Object
```

GC sẽ lần lượt quét từng Frame để tìm các đối tượng còn sống (Live Objects).

Đây là lý do Stack Frame đóng vai trò rất quan trọng trong Garbage Collection.

---

# Stack Frame và Stack Trace

Khi xảy ra.

```go
panic()
```

Go Runtime sẽ in.

```
goroutine 1

↓

main()

↓

login()

↓

validate()

↓

checkPassword()
```

Đây chính là:

```
Stack Trace
```

Runtime tạo Stack Trace bằng cách duyệt lần lượt các Stack Frame từ trên xuống dưới.

Nhờ vậy.

Lập trình viên biết chính xác đường đi của chương trình trước khi xảy ra lỗi.

---

# Góc nhìn từ Source Code

Mặc dù Go Runtime không biểu diễn mỗi Stack Frame bằng một `struct` riêng.

Nhưng Compiler và Runtime cùng tuân theo một quy ước (Calling Convention) về cách bố trí dữ liệu trên Stack.

Có thể hình dung.

```text
Stack Memory

+----------------------------+

| Frame A                    |

+----------------------------+

| Frame B                    |

+----------------------------+

| Frame C                    |

+----------------------------+
```

Mỗi Frame được xác định thông qua:

- Stack Pointer.
- Program Counter.
- Metadata do Compiler sinh ra.

Điều này cho phép Runtime:

- Context Switching.
- Stack Walking.
- Garbage Collection.
- Panic Recovery.

mà không cần tạo thêm một đối tượng quản lý cho từng Frame.

---

# Những điều cần ghi nhớ

- Mỗi lời gọi hàm tạo ra một Stack Frame.
- Stack là tập hợp của nhiều Stack Frame.
- Stack Frame chứa Return Address, Local Variables, Parameters và dữ liệu tạm thời.
- Stack Pointer luôn trỏ tới Stack Frame hiện tại.
- Khi hàm kết thúc, Stack Frame bị loại bỏ theo nguyên tắc LIFO.
- Garbage Collector quét Stack Frame để tìm Pointer còn sống.
- Stack Trace được tạo bằng cách duyệt các Stack Frame.

---

# Tóm tắt

Stack Frame là đơn vị cơ bản tạo nên Stack của Goroutine.

Mỗi lần gọi hàm, Runtime và CPU sẽ tạo một Stack Frame mới để lưu toàn bộ ngữ cảnh của lời gọi đó, bao gồm tham số, biến cục bộ, địa chỉ quay về và các giá trị cần thiết cho việc thực thi.

Hiểu rõ Stack Frame là chìa khóa để lý giải cách Go Runtime thực hiện Context Switching, tạo Stack Trace, hỗ trợ Garbage Collection và quản lý lời gọi hàm một cách hiệu quả.

Trong phần tiếp theo, chúng ta sẽ khám phá một trong những thiết kế độc đáo nhất của Go Runtime:

> **Initial Stack** - vì sao mỗi Goroutine chỉ bắt đầu với khoảng **2 KB Stack**, và làm thế nào Runtime vẫn có thể hỗ trợ các chương trình với hàng nghìn lời gọi hàm mà không gây Stack Overflow.

---

---

# 11.8 Initial Stack - Vì sao Goroutine chỉ bắt đầu với khoảng 2 KB Stack?

Trong phần trước, chúng ta đã tìm hiểu:

- Stack là gì.
- Stack Frame hoạt động như thế nào.
- Mỗi Goroutine có một Stack riêng.

Một câu hỏi rất thú vị xuất hiện.

> **Nếu mỗi Goroutine đều có Stack, vậy Runtime sẽ cấp phát bao nhiêu Stack khi tạo Goroutine?**

Nếu bạn từng lập trình C/C++ hoặc Java, có thể bạn sẽ nghĩ:

```
1 MB

hoặc

8 MB
```

Đây là kích thước Stack khá phổ biến của một OS Thread.

Nhưng Go Runtime làm điều hoàn toàn khác.

Khi tạo một Goroutine.

Runtime chỉ cấp phát khoảng:

```
≈ 2 KB
```

Điều này nghe có vẻ không hợp lý.

Liệu:

```
2 KB
```

có đủ để chạy một chương trình không?

Câu trả lời là:

> **Đủ để bắt đầu, nhưng không phải để kết thúc.**

Đây chính là một trong những thiết kế thông minh nhất của Go Runtime.

---

# Stack của OS Thread

Trong hầu hết các hệ điều hành.

Một Thread khi được tạo sẽ được cấp phát:

```
Một vùng Stack cố định
```

Ví dụ.

```
Linux

↓

8 MB
```

Hoặc.

```
Windows

↓

1 MB
```

(Con số cụ thể phụ thuộc vào hệ điều hành và cấu hình.)

Điều quan trọng là.

Stack này được cấp phát **ngay khi Thread được tạo**.

Cho dù Thread chỉ thực hiện.

```c
printf("Hello");
```

Nó vẫn sở hữu toàn bộ vùng Stack đó.

---

# Vấn đề của Stack cố định

Giả sử.

Một Server tạo.

```
100.000 Threads
```

Nếu mỗi Thread có:

```
8 MB Stack
```

Tổng bộ nhớ chỉ dành cho Stack sẽ là.

```
100.000

×

8 MB

=

800 GB
```

Đây là một con số rất lớn.

Trong khi thực tế.

Phần lớn các Thread chỉ sử dụng một phần rất nhỏ của Stack.

Điều này dẫn đến lãng phí bộ nhớ.

---

# Cách tiếp cận của Go

Go Runtime đặt ra một câu hỏi rất đơn giản.

> **Tại sao phải cấp phát một Stack rất lớn ngay từ đầu khi chưa biết Goroutine có thực sự cần đến nó hay không?**

Thay vì cấp phát:

```
8 MB
```

Runtime chỉ cấp phát.

```
≈ 2 KB
```

Có thể hình dung.

```text
OS Thread

+----------------------------+

|                            |

|          8 MB              |

|                            |

+----------------------------+
```

Trong khi.

```text
Goroutine

+----------------------------+

|           2 KB             |

+----------------------------+
```

Nếu Goroutine cần thêm bộ nhớ.

Runtime sẽ mở rộng Stack sau.

---

# Tại sao 2 KB là đủ?

Hầu hết Goroutine trong thực tế đều rất đơn giản.

Ví dụ.

```go
go func() {

    fmt.Println("Hello")

}()
```

Hoặc.

```go
go handleRequest(conn)
```

Những Goroutine này thường chỉ có:

- Một vài lời gọi hàm.
- Một số biến cục bộ nhỏ.

Toàn bộ Stack thực tế có thể chỉ chiếm vài trăm byte.

Việc cấp phát ngay:

```
8 MB
```

sẽ rất lãng phí.

---

# Ví dụ thực tế

Giả sử.

Một HTTP Server xử lý.

```
200.000 Connections
```

Nếu mỗi Connection tương ứng:

```
1 Goroutine
```

Tổng Stack ban đầu chỉ khoảng.

```
200.000

×

2 KB

≈ 400 MB
```

Nếu sử dụng Thread truyền thống.

```
200.000

×

8 MB

≈ 1.6 PB
```

Điều này gần như không thể.

Đây chính là lý do Go có thể tạo hàng trăm nghìn Goroutine.

---

# Runtime có biết trước Stack sẽ lớn bao nhiêu không?

Không.

Runtime hoàn toàn không biết.

Ví dụ.

```go
go worker()
```

Có thể.

```go
func worker() {

    fmt.Println("Hello")

}
```

Stack chỉ cần.

```
500 Bytes
```

Hoặc.

```go
func worker() {

    recursive(100000)

}
```

Stack có thể cần.

```
Hàng MB
```

Runtime không thể đoán trước.

Do đó.

Chiến lược tốt nhất là.

```
Start Small

↓

Grow Later
```

---

# Initial Stack chỉ là điểm khởi đầu

Điều rất quan trọng cần nhớ.

```
Initial Stack

≠

Maximum Stack
```

Initial Stack chỉ là:

```
Starting Size
```

Nếu Goroutine tiếp tục gọi hàm.

```go
A()

↓

B()

↓

C()

↓

D()
```

Stack sẽ dần đầy.

Khi đó Runtime sẽ:

```
Grow Stack
```

Điều này hoàn toàn tự động.

---

# Ai chịu trách nhiệm cấp phát Initial Stack?

Khi Runtime thực hiện.

```go
go worker()
```

Quá trình diễn ra như sau.

```text
Create struct g

↓

Allocate Initial Stack

↓

Initialize gobuf

↓

Put vào Run Queue
```

Sau khi Scheduler chọn.

Goroutine bắt đầu chạy.

---

# Initial Stack nằm ở đâu?

Giống như mọi Stack khác.

Initial Stack cũng nằm trên:

```
Heap Memory
```

Đây là một điểm rất khác so với nhiều ngôn ngữ.

Có thể hình dung.

```text
Heap

+----------------------------+

| Stack G1                   |

+----------------------------+

| Stack G2                   |

+----------------------------+

| Stack G3                   |

+----------------------------+
```

Điều này cho phép Runtime:

- Di chuyển Stack.
- Mở rộng Stack.
- Thu nhỏ Stack.

Nếu Stack nằm cố định như Thread.

Các thao tác này sẽ rất khó thực hiện.

---

# Runtime quản lý Stack như thế nào?

Trong `struct g`.

Có trường.

```go
stack
```

Bên trong.

Runtime lưu.

```text
stack

├── lo
└── hi
```

Trong đó.

```
lo
```

là địa chỉ thấp nhất của Stack.

```
hi
```

là địa chỉ cao nhất.

Có thể minh họa.

```text
High Address

stack.hi

+----------------------+

|                      |

|      Stack           |

|                      |

+----------------------+

stack.lo

Low Address
```

Scheduler và Garbage Collector đều sử dụng thông tin này.

---

# Có phải mọi Goroutine đều bắt đầu với đúng 2 KB?

Không hẳn.

Con số:

```
≈ 2 KB
```

là giá trị phổ biến trong các phiên bản Go hiện đại.

Kích thước này có thể thay đổi giữa các phiên bản hoặc kiến trúc CPU.

Điều quan trọng không phải là:

```
2 KB
```

mà là nguyên tắc.

> **Bắt đầu với Stack rất nhỏ và mở rộng khi cần.**

---

# Tại sao không bắt đầu với 512 Byte?

Nếu Stack quá nhỏ.

Runtime sẽ phải:

```
Grow Stack
```

rất thường xuyên.

Điều này làm tăng chi phí.

Ngược lại.

Nếu Stack quá lớn.

Bộ nhớ bị lãng phí.

Do đó Runtime phải lựa chọn một giá trị cân bằng giữa:

- Chi phí bộ nhớ.
- Chi phí mở rộng.

Đây là kết quả của nhiều năm tối ưu Go Runtime.

---

# Initial Stack và Garbage Collector

Initial Stack nhỏ còn mang lại một lợi ích khác.

Garbage Collector phải quét Stack của mọi Goroutine.

Nếu mỗi Goroutine đều có Stack rất lớn.

Thời gian GC sẽ tăng đáng kể.

Stack nhỏ giúp:

- Giảm dữ liệu cần quét.
- Giảm Pause Time.
- Tăng Throughput.

---

# Góc nhìn tổng thể

Có thể mô tả toàn bộ quá trình như sau.

```text
go worker()

↓

Create struct g

↓

Allocate Initial Stack (~2 KB)

↓

Initialize gobuf

↓

Runnable

↓

Scheduler

↓

Running
```

Trong suốt quá trình này.

Stack vẫn chỉ khoảng:

```
2 KB
```

Chỉ khi chương trình thực sự cần nhiều hơn.

Runtime mới mở rộng.

---

# Những điều cần ghi nhớ

- Mỗi Goroutine bắt đầu với một Stack rất nhỏ.
- Initial Stack chỉ là kích thước ban đầu.
- Runtime không biết trước Goroutine sẽ cần bao nhiêu Stack.
- Stack nằm trên Heap để có thể di chuyển.
- Initial Stack nhỏ giúp tiết kiệm bộ nhớ.
- Initial Stack nhỏ giúp Garbage Collector hoạt động hiệu quả hơn.
- Khi Stack đầy, Runtime sẽ tự động mở rộng.

---

# Tóm tắt

Thay vì cấp phát một vùng Stack lớn như OS Thread, Go Runtime chỉ cấp phát một Initial Stack rất nhỏ cho mỗi Goroutine.

Thiết kế này giúp giảm đáng kể lượng bộ nhớ sử dụng, cho phép chương trình tạo hàng trăm nghìn hoặc hàng triệu Goroutine mà vẫn duy trì hiệu năng và khả năng mở rộng cao.

Tuy nhiên, Initial Stack chỉ là điểm khởi đầu.

Khi chương trình tiếp tục gọi nhiều hàm hoặc sử dụng nhiều biến cục bộ hơn, Runtime sẽ phải giải quyết một bài toán khó hơn:

> **Làm thế nào để mở rộng Stack của một Goroutine đang chạy mà không làm gián đoạn chương trình?**

Đó chính là nội dung của phần tiếp theo:

**11.9 - Stack Growth**.

---

---

# 11.9 Stack Growth - Làm thế nào Go Runtime mở rộng Stack?

Trong phần trước, chúng ta đã biết rằng mỗi Goroutine chỉ bắt đầu với khoảng:

```
≈ 2 KB
```

Stack.

Điều này giúp Go có thể tạo hàng triệu Goroutine mà không tiêu tốn quá nhiều bộ nhớ.

Nhưng ngay lập tức xuất hiện một câu hỏi rất quan trọng.

> **Điều gì xảy ra nếu Goroutine cần nhiều hơn 2 KB Stack?**

Ví dụ.

```go
func main() {

    recursive(100000)

}
```

Hoặc.

```go
func process() {

    var buffer [64 * 1024]byte

}
```

Rõ ràng.

```
2 KB
```

không đủ.

Runtime phải tìm cách mở rộng Stack.

Điều thú vị là.

Go Runtime thực hiện việc này **hoàn toàn tự động**, lập trình viên không cần can thiệp.

---

# Vấn đề cần giải quyết

Giả sử Stack hiện tại.

```text
High Address

+----------------------+

| Frame main()         |

+----------------------+

| Frame A()            |

+----------------------+

| Frame B()            |

+----------------------+

Low Address
```

Goroutine tiếp tục gọi.

```go
C()

↓

D()

↓

E()
```

Stack dần đầy.

```text
High Address

+----------------------+

| main                 |

+----------------------+

| A                    |

+----------------------+

| B                    |

+----------------------+

| C                    |

+----------------------+

| D                    |

+----------------------+

| ???                  |

Low Address
```

Nếu tiếp tục ghi xuống dưới.

Stack sẽ ghi đè lên vùng nhớ khác.

Điều này sẽ phá hỏng chương trình.

Runtime phải phát hiện điều này trước khi nó xảy ra.

---

# Stack Overflow

Nếu không mở rộng Stack.

CPU sẽ tiếp tục:

```
Push Frame
```

cho đến khi.

```
Out of Stack
```

Điều này gọi là:

```
Stack Overflow
```

Trong nhiều ngôn ngữ.

Stack Overflow thường dẫn đến:

- Crash.
- Segmentation Fault.
- Process Termination.

Go Runtime không muốn điều đó xảy ra.

---

# Ý tưởng của Go Runtime

Thay vì cấp phát Stack lớn ngay từ đầu.

Runtime chọn chiến lược.

```text
Start Small

↓

Grow On Demand
```

Khi Stack gần đầy.

Runtime sẽ:

```
Allocate Stack lớn hơn

↓

Copy dữ liệu

↓

Continue Running
```

Đối với lập trình viên.

Quá trình này hoàn toàn vô hình.

---

# Runtime phát hiện Stack sắp đầy như thế nào?

Đây là phần rất thú vị.

Trong `struct g`.

Có hai trường.

```go
stackguard0
```

và.

```go
stackguard1
```

Có thể hình dung.

```text
High Address

+----------------------+

|                      |

|      Stack           |

|                      |

+----------------------+

| stackguard0          |

+----------------------+

Low Address
```

Trước mỗi lời gọi hàm.

Compiler sinh thêm một đoạn mã kiểm tra.

Có thể mô tả đơn giản.

```go
if SP < stackguard0 {

    morestack()

}
```

Nếu Stack Pointer đã tiến sát vùng bảo vệ.

Runtime sẽ không cho phép gọi hàm tiếp.

Thay vào đó.

Nó chuyển sang mở rộng Stack.

---

# Hàm `morestack()`

Khi phát hiện Stack không đủ.

Runtime sẽ gọi.

```
morestack()
```

Đây là một hàm rất quan trọng trong:

```
runtime/asm_*.s
```

Tùy theo kiến trúc CPU.

`morestack()` không trực tiếp mở rộng Stack.

Nó chỉ làm nhiệm vụ:

```
Pause Current Execution

↓

Prepare Runtime

↓

Call newstack()
```

---

# Hàm `newstack()`

Sau khi vào Runtime.

`morestack()` sẽ gọi.

```
newstack()
```

Đây là nơi diễn ra phần lớn công việc.

Có thể mô tả.

```text
Current Stack

↓

Allocate New Stack

↓

Copy Old Stack

↓

Adjust Pointers

↓

Continue
```

Đây chính là trái tim của cơ chế Stack Growth.

---

# Runtime cấp phát Stack mới

Giả sử Stack hiện tại.

```
2 KB
```

đã đầy.

Runtime sẽ cấp phát.

```
4 KB
```

Stack mới.

```text
Old Stack

2 KB
```

↓

```text
New Stack

4 KB
```

Nếu sau này.

```
4 KB
```

lại đầy.

Runtime tiếp tục.

```
8 KB
```

↓

```
16 KB
```

↓

```
32 KB
```

Việc mở rộng diễn ra theo cấp số nhân để giảm số lần phải sao chép.

---

# Vì sao không mở rộng ngay trên Stack cũ?

Một câu hỏi rất hay.

Tại sao Runtime không làm.

```
Realloc
```

ngay tại chỗ?

Lý do là.

Không có gì đảm bảo vùng nhớ ngay sau Stack hiện tại còn trống.

Ví dụ.

```text
Stack G1

+

Heap Object

+

Stack G2
```

Nếu Runtime cố mở rộng tại chỗ.

Có thể ghi đè lên dữ liệu khác.

Do đó.

Giải pháp an toàn nhất là.

```
Allocate New Stack

↓

Copy
```

---

# Ví dụ minh họa

Ban đầu.

```text
Old Stack

+----------------------+

| main                 |

+----------------------+

| A                    |

+----------------------+

| B                    |

+----------------------+
```

Runtime tạo.

```text
New Stack

+----------------------+

|                      |

|                      |

|                      |

|                      |

+----------------------+
```

Sau đó.

```
Copy
```

Kết quả.

```text
New Stack

+----------------------+

| main                 |

+----------------------+

| A                    |

+----------------------+

| B                    |

+----------------------+

| Free Space           |

+----------------------+
```

Goroutine tiếp tục chạy.

---

# Runtime phải cập nhật những gì?

Sau khi tạo Stack mới.

Runtime không thể đơn giản bỏ Stack cũ.

Nó còn phải:

- Cập nhật Stack Pointer.
- Cập nhật các Pointer trong Stack.
- Cập nhật `struct g`.
- Cập nhật Metadata của GC.

Nếu bỏ sót một Pointer.

Chương trình sẽ Crash.

Đây là lý do Stack Growth phức tạp hơn nhiều so với việc chỉ sao chép bộ nhớ.

---

# Stack Growth có ảnh hưởng tới chương trình không?

Đối với lập trình viên.

Câu trả lời là:

```
Không.
```

Ví dụ.

```go
recursive(100000)
```

Chương trình vẫn chạy bình thường.

Không cần:

- malloc.
- realloc.
- free.

Runtime tự xử lý toàn bộ.

---

# Stack Growth có xảy ra thường xuyên không?

Thông thường.

Không.

Đa số Goroutine chỉ dùng vài KB Stack.

Ví dụ.

- HTTP Handler.
- RPC Handler.
- Worker.

Thường không bao giờ cần mở rộng.

Chỉ những Goroutine có:

- Đệ quy sâu.
- Nhiều lời gọi hàm.
- Biến cục bộ lớn.

mới kích hoạt Stack Growth.

---

# Stack Growth có tốn chi phí không?

Có.

Mỗi lần mở rộng.

Runtime phải:

- Cấp phát vùng nhớ mới.
- Sao chép toàn bộ Stack.
- Điều chỉnh Pointer.
- Cập nhật Metadata.

Do đó.

Stack Growth là một thao tác tương đối tốn kém.

Tuy nhiên.

Nó diễn ra rất ít lần trong vòng đời của một Goroutine.

Nhờ chiến lược tăng theo cấp số nhân.

Chi phí trung bình trên mỗi lời gọi hàm vẫn rất thấp.

---

# Runtime phối hợp với Garbage Collector

Khi Stack thay đổi.

Garbage Collector cũng phải biết.

Ví dụ.

```text
Old Stack

↓

Heap Pointer
```

Sau khi Copy.

```text
New Stack

↓

Heap Pointer
```

GC phải quét đúng Stack mới.

Do đó Runtime sẽ cập nhật Metadata liên quan trước khi Goroutine tiếp tục chạy.

---

# Góc nhìn tổng thể

Có thể mô tả toàn bộ quá trình.

```text
Function Call

↓

Check stackguard

↓

Enough Stack?

     │

 ┌───┴────┐

 │        │

Yes       No

 │        │

 ▼        ▼

Run   morestack()

           │

           ▼

      newstack()

           │

           ▼

 Allocate New Stack

           │

           ▼

   Copy Old Stack

           │

           ▼

 Adjust Pointers

           │

           ▼

 Continue Execution
```

Đây là một trong những luồng quan trọng nhất của Go Runtime.

---

# Những điều cần ghi nhớ

- Goroutine bắt đầu với Stack nhỏ.
- Compiler kiểm tra `stackguard` trước mỗi lời gọi hàm.
- Khi Stack gần đầy, Runtime gọi `morestack()`.
- `morestack()` chuyển sang `newstack()`.
- Runtime cấp phát Stack mới lớn hơn.
- Toàn bộ Stack cũ được sao chép sang Stack mới.
- Sau khi cập nhật Pointer, Goroutine tiếp tục chạy như bình thường.

---

# Tóm tắt

Stack Growth là cơ chế cho phép Go Runtime tự động mở rộng Stack của Goroutine khi không còn đủ chỗ cho các lời gọi hàm mới.

Thay vì cấp phát một Stack rất lớn ngay từ đầu, Runtime chỉ bắt đầu với một Stack nhỏ và chỉ mở rộng khi thật sự cần thiết.

Cơ chế này giúp Go vừa tiết kiệm bộ nhớ, vừa hỗ trợ các chương trình có lời gọi hàm sâu hoặc sử dụng nhiều biến cục bộ mà không yêu cầu lập trình viên phải quản lý Stack bằng tay.

Tuy nhiên, mở rộng Stack mới chỉ là một nửa câu chuyện.

Sau khi cấp phát vùng Stack mới, Runtime còn phải giải quyết một bài toán khó hơn nhiều:

> **Làm thế nào để sao chép toàn bộ Stack sang vùng nhớ mới và cập nhật mọi Pointer mà không làm hỏng chương trình?**

Đó sẽ là nội dung của **11.10 - Stack Copy**.

---

---

# 11.10 Stack Copy - Làm thế nào Go Runtime di chuyển một Stack đang hoạt động?

Trong phần trước, chúng ta đã tìm hiểu quá trình **Stack Growth**.

Khi Stack không còn đủ chỗ.

Runtime sẽ:

```text
Allocate New Stack

↓

Copy Old Stack

↓

Continue Execution
```

Thoạt nhìn, việc này có vẻ rất đơn giản.

Chỉ cần gọi:

```c
memcpy(...)
```

là xong.

Nhưng thực tế hoàn toàn khác.

> **Sao chép Stack là một trong những thao tác phức tạp nhất của Go Runtime.**

Bởi vì Stack không chỉ chứa dữ liệu.

Nó còn chứa:

- Pointer.
- Return Address.
- Stack Frame.
- Deferred Calls.
- Panic Context.
- Metadata của Garbage Collector.

Nếu chỉ Copy bộ nhớ.

Chương trình gần như chắc chắn sẽ bị lỗi.

---

# Tại sao phải Copy Stack?

Hãy nhớ lại.

Stack cũ.

```text
Old Stack

2 KB
```

đã đầy.

Runtime tạo.

```text
New Stack

4 KB
```

Bây giờ.

Runtime phải chuyển toàn bộ dữ liệu từ Stack cũ sang Stack mới.

Có thể hình dung.

```text
Old Stack

+----------------------+

| main                 |

+----------------------+

| foo                  |

+----------------------+

| bar                  |

+----------------------+
```

↓

```text
New Stack

+----------------------+

| main                 |

+----------------------+

| foo                  |

+----------------------+

| bar                  |

+----------------------+

| Free                 |

+----------------------+
```

Sau đó.

Runtime sẽ tiếp tục chạy trên Stack mới.

---

# Vì sao không thể dùng memcpy?

Đây là câu hỏi quan trọng nhất.

Giả sử Stack chứa.

```go
func foo() {

    var x int

    bar(&x)

}
```

Stack có thể trông như sau.

```text
Stack

+----------------------+

| x = 100              |

+----------------------+

| Pointer → x          |

+----------------------+
```

Nếu Runtime Copy nguyên Stack sang nơi khác.

```text
New Stack

+----------------------+

| x = 100              |

+----------------------+

| Pointer → ???        |

+----------------------+
```

Pointer vẫn trỏ tới:

```
Old Stack
```

Sau khi Stack cũ bị giải phóng.

Pointer trở thành:

```
Dangling Pointer
```

Đây là lỗi rất nghiêm trọng.

---

# Ví dụ trực quan

Giả sử.

Old Stack.

```text
Address

0x1000

+----------------------+

| x                    |

+----------------------+

0x1010

| Pointer -> 0x1000    |

+----------------------+
```

Sau khi Copy.

```text
Address

0x5000

+----------------------+

| x                    |

+----------------------+

0x5010

| Pointer -> 0x1000    |

+----------------------+
```

Pointer vẫn trỏ về.

```
0x1000
```

Trong khi dữ liệu mới nằm ở.

```
0x5000
```

Nếu Runtime không sửa Pointer.

Chương trình sẽ truy cập vùng nhớ đã bị giải phóng.

---

# Bài toán của Runtime

Sau khi Copy.

Runtime phải tìm.

```
Tất cả Pointer

↓

Update
```

Có thể mô tả.

```text
Old Address

↓

New Address
```

Đây chính là lý do việc Copy Stack khó hơn rất nhiều so với Copy một mảng thông thường.

---

# Runtime biết Pointer ở đâu bằng cách nào?

Một câu hỏi rất thú vị.

> Runtime làm sao biết ô nhớ nào là Pointer và ô nhớ nào chỉ là số nguyên?

Ví dụ.

```go
type User struct {

    Age int

    Name string

}
```

Trong Stack.

```text
100

0x5000

42

0x8000
```

Runtime không thể đoán.

Giá trị:

```
0x5000
```

là:

- Pointer.

Hay.

- Chỉ là một số nguyên.

Để giải quyết.

Compiler tạo thêm:

```
Stack Map
```

---

# Stack Map

Stack Map là Metadata do Compiler sinh ra.

Nó mô tả.

```
Slot nào

↓

Pointer
```

```
Slot nào

↓

Scalar
```

Ví dụ.

```text
Stack Frame

+----------------------+

| int                  |

+----------------------+

| pointer              |

+----------------------+

| pointer              |

+----------------------+

| float                |

+----------------------+
```

Nhờ Stack Map.

Runtime biết chính xác.

```
Update

↓

Slot 2

↓

Slot 3
```

Không cần sửa các ô khác.

---

# Hàm `copystack()`

Trong source code.

Việc sao chép Stack được thực hiện bởi.

```
copystack()
```

Có thể mô tả.

```text
Allocate New Stack

↓

Copy Memory

↓

Adjust Pointers

↓

Update struct g

↓

Continue
```

Đây là một trong những hàm quan trọng nhất của `runtime/stack.go`.

---

# Điều chỉnh Pointer

Sau khi Copy.

Runtime duyệt toàn bộ Stack.

Ví dụ.

```text
Pointer

↓

Old Stack
```

↓

```text
Pointer

↓

New Stack
```

Mỗi Pointer sẽ được tính lại dựa trên khoảng chênh lệch địa chỉ giữa hai Stack.

Ví dụ.

```
Old Base

0x1000
```

↓

```
New Base

0x5000
```

Khoảng dịch.

```
+0x4000
```

Pointer.

```
0x1020
```

sẽ trở thành.

```
0x5020
```

Toàn bộ quá trình được Runtime thực hiện tự động.

---

# Return Address có cần sửa không?

Thông thường.

Không.

Return Address trỏ tới:

```
Instruction Memory
```

không phải Stack.

Ví dụ.

```text
Code Segment

0x400000
```

Địa chỉ này không thay đổi khi Stack được Copy.

Do đó.

Runtime không cần cập nhật Return Address.

---

# Pointer tới Heap có cần sửa không?

Ví dụ.

```go
obj := &User{}
```

Stack chứa.

```
Pointer

↓

Heap
```

Sau khi Copy Stack.

Heap Object vẫn nằm ở cùng vị trí.

Do đó.

Pointer tới Heap:

```
Không thay đổi.
```

Runtime chỉ cập nhật.

```
Pointer

↓

Old Stack
```

---

# Pointer tới Stack mới cần sửa

Ví dụ.

```go
func foo() {

    var x int

    bar(&x)

}
```

Đây là Pointer nội bộ.

```
Stack

↓

Stack
```

Các Pointer kiểu này bắt buộc phải được điều chỉnh.

Nếu không.

Chương trình sẽ bị lỗi ngay sau khi Stack cũ được giải phóng.

---

# Runtime dừng Goroutine trong lúc Copy

Trong suốt quá trình.

```
Copy Stack
```

Runtime sẽ:

```
Pause Goroutine
```

Không có lệnh nào của Goroutine được thực thi.

Sau khi.

```
Copy Finished
```

Runtime mới:

```
Restore Context

↓

Continue
```

Điều này đảm bảo Stack luôn nhất quán.

---

# Stack Copy và Garbage Collector

Garbage Collector cũng cần biết.

```
Stack Changed
```

Sau khi Copy.

Runtime sẽ cập nhật.

- Stack Boundaries.
- Stack Pointer.
- GC Metadata.

Nhờ đó.

Garbage Collector sẽ quét đúng Stack mới.

---

# Stack Copy có tốn chi phí không?

Có.

Chi phí bao gồm.

- Allocate Memory.
- Copy Memory.
- Walk Stack.
- Adjust Pointers.
- Update Metadata.

Tuy nhiên.

Runtime cố gắng:

- Tăng Stack theo cấp số nhân.
- Hạn chế số lần Copy.

Thông thường.

Một Goroutine chỉ Copy Stack vài lần trong suốt vòng đời.

---

# Vì sao Go không dùng Segmented Stack nữa?

Ở các phiên bản Go rất cũ.

Runtime sử dụng.

```
Segmented Stack
```

Mỗi khi Stack đầy.

Runtime tạo thêm.

```
Segment
```

Ví dụ.

```text
Segment A

↓

Segment B

↓

Segment C
```

Thiết kế này tránh việc Copy.

Nhưng phát sinh nhiều vấn đề.

- Chi phí chuyển Segment.
- Cache Locality kém.
- Khó tối ưu Compiler.
- Hot Split.

Từ Go 1.3.

Runtime chuyển sang.

```
Contiguous Stack

+

Stack Copy
```

Thiết kế hiện nay đơn giản hơn và có hiệu năng tốt hơn trong phần lớn trường hợp.

---

# Minh họa toàn bộ quá trình

```text
Old Stack Full

↓

Allocate New Stack

↓

Copy Memory

↓

Walk Stack

↓

Adjust Stack Pointers

↓

Update struct g

↓

Update GC Metadata

↓

Free Old Stack

↓

Continue Execution
```

Đây là toàn bộ luồng hoạt động của `copystack()`.

---

# Những điều cần ghi nhớ

- Stack Growth luôn đi kèm Stack Copy.
- Runtime không thể chỉ dùng `memcpy()`.
- Chỉ các Pointer trỏ vào Stack mới cần được điều chỉnh.
- Pointer tới Heap không thay đổi.
- Compiler sinh Stack Map để Runtime biết đâu là Pointer.
- `copystack()` là thành phần trung tâm của cơ chế Dynamic Stack.
- Sau khi Copy xong, Goroutine tiếp tục chạy như chưa từng bị di chuyển.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Stack Copy chỉ là sao chép bộ nhớ.

Sai.

Sau khi sao chép, Runtime còn phải cập nhật tất cả các Pointer nội bộ và Metadata liên quan.

---

### Hiểu lầm 2

> Mọi Pointer đều phải cập nhật.

Sai.

Chỉ các Pointer trỏ vào vùng Stack cũ mới cần điều chỉnh.

Pointer tới Heap hoặc Code Segment vẫn giữ nguyên.

---

### Hiểu lầm 3

> Stack Copy xảy ra rất thường xuyên.

Sai.

Nhờ chiến lược tăng Stack theo cấp số nhân, một Goroutine thông thường chỉ phải Copy Stack rất ít lần trong suốt vòng đời của nó.

---

# Tóm tắt

Stack Copy là bước quan trọng nhất trong cơ chế Dynamic Stack của Go Runtime.

Sau khi phát hiện Stack không còn đủ chỗ, Runtime sẽ cấp phát một Stack lớn hơn, sao chép toàn bộ dữ liệu sang vùng nhớ mới và điều chỉnh các Pointer nội bộ để Goroutine có thể tiếp tục thực thi một cách chính xác.

Mặc dù đây là một thao tác phức tạp, nó diễn ra hoàn toàn tự động và gần như vô hình đối với lập trình viên.

Thiết kế này giúp Go vừa tiết kiệm bộ nhớ, vừa hỗ trợ các chương trình có Stack rất sâu mà không cần cấp phát một vùng Stack lớn ngay từ khi Goroutine được tạo.

Trong phần tiếp theo, chúng ta sẽ tìm hiểu chiều ngược lại của quá trình này:

> **Stack Shrink** - khi Goroutine không còn sử dụng nhiều Stack nữa, Go Runtime có thu nhỏ Stack để trả lại bộ nhớ hay không, và nếu có thì thực hiện điều đó như thế nào.

---

---

# 11.11 Stack Shrink - Khi nào Go Runtime thu nhỏ Stack?

Trong phần trước, chúng ta đã tìm hiểu cách Go Runtime mở rộng Stack.

Quá trình đó diễn ra như sau.

```text
Initial Stack

↓

Stack Growth

↓

Stack Copy

↓

Continue Running
```

Nhờ cơ chế này.

Một Goroutine có thể mở rộng Stack từ:

```
2 KB

↓

4 KB

↓

8 KB

↓

16 KB

↓

...
```

mà lập trình viên hoàn toàn không cần can thiệp.

Nhưng ngay lập tức xuất hiện một câu hỏi mới.

> **Nếu Goroutine đã sử dụng Stack rất lớn, sau đó không còn cần nữa thì sao?**

Ví dụ.

Một Goroutine từng sử dụng:

```
4 MB
```

Stack để xử lý một bài toán đệ quy.

Sau khi hoàn thành.

Nó chỉ còn thực hiện.

```go
for {

    time.Sleep(time.Second)

}
```

Lúc này.

```
4 MB
```

Stack gần như không còn được sử dụng.

Liệu Go Runtime có:

```
Thu nhỏ Stack
```

để trả lại bộ nhớ hay không?

Câu trả lời là:

> **Có.**

---

# Vì sao cần Stack Shrink?

Giả sử.

Một HTTP Server xử lý.

```
500.000 Requests
```

Một số Request đặc biệt khiến Goroutine phải mở rộng Stack lên.

```
2 MB
```

Sau khi xử lý xong.

Nếu Runtime giữ nguyên.

```
2 MB
```

cho mọi Goroutine.

Bộ nhớ sẽ tăng rất nhanh.

Ví dụ.

```
100.000 Goroutines

×

2 MB

=

200 GB
```

Trong khi thực tế.

Phần lớn Goroutine chỉ còn cần vài KB.

Do đó.

Runtime cần cơ chế:

```
Shrink Stack
```

để giải phóng bộ nhớ không còn sử dụng.

---

# Stack Shrink là gì?

Stack Shrink là quá trình.

```text
Large Stack

↓

Allocate Smaller Stack

↓

Copy Live Frames

↓

Continue
```

Có thể thấy.

Quá trình này gần giống với Stack Growth.

Chỉ khác.

```
Grow

↓

Lớn hơn
```

Trong khi.

```
Shrink

↓

Nhỏ hơn
```

---

# Runtime có thu nhỏ Stack ngay lập tức không?

Không.

Đây là một quyết định thiết kế rất quan trọng.

Giả sử.

Một Goroutine.

```
Grow

↓

Shrink

↓

Grow

↓

Shrink
```

liên tục.

Nếu Runtime thực hiện mỗi lần.

Hiệu năng sẽ rất kém.

Bởi vì.

Mỗi lần Shrink đều cần.

- Allocate.
- Copy.
- Adjust Pointer.

Chi phí không hề nhỏ.

Do đó.

Runtime chỉ Shrink khi thực sự cần.

---

# Khi nào Runtime xem xét Shrink?

Khác với Stack Growth.

Stack Shrink **không diễn ra trong lúc Goroutine đang thực thi bình thường**.

Thay vào đó.

Runtime thường xem xét việc thu nhỏ Stack trong quá trình:

```
Garbage Collection
```

Đây là thời điểm Runtime đã dừng hoặc đồng bộ trạng thái của Goroutine và có đầy đủ thông tin để đánh giá mức sử dụng Stack.

Có thể mô tả.

```text
GC

↓

Scan Stack

↓

Check Usage

↓

Need Shrink?

↓

Shrink
```

---

# Tại sao chọn Garbage Collector?

Garbage Collector vốn đã phải:

- Quét Stack.
- Quét Pointer.
- Cập nhật Metadata.

Do đó.

Việc kiểm tra.

```
Stack quá lớn?
```

gần như không làm phát sinh thêm nhiều chi phí.

Đây là một ví dụ rất hay về việc các thành phần của Go Runtime phối hợp với nhau.

---

# Runtime đánh giá như thế nào?

Runtime không chỉ nhìn vào:

```
Stack Size
```

Mà còn quan tâm.

```
Stack Usage
```

Ví dụ.

```text
Allocated

2 MB
```

Nhưng.

```text
Actually Used

32 KB
```

Lúc này.

Runtime có thể quyết định.

```
Shrink
```

Ngược lại.

Nếu phần lớn Stack vẫn đang được sử dụng.

Runtime sẽ giữ nguyên.

---

# Ví dụ trực quan

Trước khi Shrink.

```text
High Address

+----------------------------+

| main                       |

+----------------------------+

| worker                     |

+----------------------------+

| parse                      |

+----------------------------+

|                            |

|                            |

|      Free Space            |

|                            |

|                            |

+----------------------------+

Low Address
```

Có thể thấy.

Phần lớn Stack đang trống.

Runtime sẽ tạo Stack nhỏ hơn.

---

# Sau khi Shrink

```text
High Address

+----------------------------+

| main                       |

+----------------------------+

| worker                     |

+----------------------------+

| parse                      |

+----------------------------+

Low Address
```

Toàn bộ vùng nhớ dư thừa đã được loại bỏ.

---

# Stack Shrink có phải xóa dữ liệu không?

Không.

Runtime chỉ Copy:

```
Live Stack Frames
```

Những Stack Frame không còn tồn tại sẽ không được sao chép.

Có thể mô tả.

```text
Old Stack

↓

Copy Active Frames

↓

New Stack
```

Đây cũng chính là cách Runtime thực hiện Stack Growth.

---

# Runtime sử dụng hàm nào?

Trong `runtime/stack.go`.

Việc thu nhỏ Stack được thực hiện thông qua.

```
shrinkstack()
```

Có thể mô tả.

```text
Check Usage

↓

Allocate Smaller Stack

↓

Copy Active Frames

↓

Adjust Pointers

↓

Update struct g

↓

Free Old Stack
```

Nhìn rất giống.

```
copystack()
```

Chỉ khác ở kích thước Stack mới.

---

# Runtime có thể Shrink về 2 KB không?

Không nhất thiết.

Giả sử Goroutine vẫn đang sử dụng.

```
64 KB
```

Stack.

Runtime sẽ không Shrink xuống.

```
2 KB
```

Mục tiêu của Runtime là.

```
Đủ dùng

+

Tiết kiệm bộ nhớ
```

không phải nhỏ nhất có thể.

---

# Stack Shrink có ảnh hưởng tới Goroutine không?

Đối với lập trình viên.

Hoàn toàn không.

Ví dụ.

```go
func worker() {

    processHugeData()

    time.Sleep(time.Hour)

}
```

Sau khi.

```
processHugeData()
```

kết thúc.

Runtime có thể thu nhỏ Stack.

Khi Goroutine thức dậy.

Nó vẫn tiếp tục chạy bình thường.

---

# Stack Shrink và Pointer

Giống Stack Growth.

Sau khi tạo Stack nhỏ hơn.

Runtime phải cập nhật.

- Stack Pointer.
- Pointer nội bộ.
- Metadata của Garbage Collector.

Nếu không.

Pointer vẫn sẽ trỏ về Stack cũ.

---

# Stack Shrink và Garbage Collector

Garbage Collector đóng vai trò rất quan trọng.

Có thể mô tả.

```text
GC

↓

Scan Stack

↓

Shrink Stack

↓

Update GC Metadata

↓

Continue
```

Điều này giúp GC luôn làm việc trên Stack mới.

---

# Stack Shrink có xảy ra thường xuyên không?

Không.

Trong phần lớn chương trình.

Stack của Goroutine:

- Không Grow.
- Hoặc chỉ Grow một vài lần.

Do đó.

Stack Shrink cũng không xảy ra nhiều.

Runtime cố tình thiết kế như vậy để:

- Giảm Allocation.
- Giảm Copy.
- Giảm Fragmentation.
- Giảm chi phí Scheduler.

---

# Chi phí của Stack Shrink

Mặc dù giúp tiết kiệm bộ nhớ.

Stack Shrink vẫn có chi phí.

Bao gồm.

- Allocate Stack mới.
- Copy Stack.
- Walk Stack.
- Adjust Pointer.
- Update Metadata.
- Free Stack cũ.

Do đó.

Runtime luôn cân bằng giữa:

```
Memory Saving
```

và.

```
CPU Cost
```

---

# Góc nhìn tổng thể

Có thể mô tả toàn bộ quá trình.

```text
Large Stack

↓

Garbage Collector

↓

Check Stack Usage

↓

Need Shrink?

     │

 ┌───┴────┐

 │        │

 No      Yes

 │        │

 ▼        ▼

Keep   Allocate Smaller Stack

              │

              ▼

        Copy Active Frames

              │

              ▼

        Adjust Pointers

              │

              ▼

         Update struct g

              │

              ▼

        Free Old Stack

              │

              ▼

      Continue Execution
```

---

# Dynamic Stack

Sau khi học xong bốn phần.

- Initial Stack.
- Stack Growth.
- Stack Copy.
- Stack Shrink.

Chúng ta có thể thấy toàn bộ vòng đời của Stack.

```text
Create Goroutine

↓

Allocate Initial Stack

↓

Running

↓

Need More Space?

↓

Stack Growth

↓

Stack Copy

↓

Running

↓

Garbage Collector

↓

Need Shrink?

↓

Stack Shrink

↓

Continue
```

Đây chính là cơ chế:

```
Dynamic Stack
```

của Go Runtime.

---

# So sánh với Thread truyền thống

| OS Thread | Goroutine |
|-----------|-----------|
| Stack cố định | Stack động |
| Thường không thu nhỏ | Có thể thu nhỏ |
| Khó thay đổi kích thước | Runtime tự quản lý |
| Lãng phí bộ nhớ khi Stack ít dùng | Sử dụng bộ nhớ hiệu quả |

Đây là một trong những ưu điểm lớn nhất của Goroutine.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Runtime sẽ thu nhỏ Stack ngay sau khi hàm trả về.

Sai.

Runtime thường đợi đến chu kỳ Garbage Collection để đánh giá việc thu nhỏ.

---

### Hiểu lầm 2

> Stack Shrink không tốn chi phí.

Sai.

Việc thu nhỏ cũng cần cấp phát vùng nhớ mới, sao chép dữ liệu và cập nhật Pointer giống như Stack Growth.

---

### Hiểu lầm 3

> Mọi Goroutine đều được Shrink về kích thước ban đầu.

Sai.

Runtime chỉ thu nhỏ khi điều đó mang lại lợi ích rõ ràng và vẫn đảm bảo đủ không gian cho Goroutine tiếp tục thực thi.

---

# Những điều cần ghi nhớ

- Stack của Goroutine có thể mở rộng và cũng có thể thu nhỏ.
- Stack Shrink thường được xem xét trong quá trình Garbage Collection.
- Runtime chỉ thu nhỏ khi lợi ích về bộ nhớ lớn hơn chi phí thực hiện.
- Quá trình Shrink cũng cần Copy Stack và điều chỉnh Pointer.
- Dynamic Stack là sự kết hợp của Initial Stack, Growth, Copy và Shrink.

---

# Tóm tắt

Stack Shrink là bước cuối cùng hoàn thiện cơ chế **Dynamic Stack** của Go Runtime.

Thay vì giữ nguyên một vùng Stack lớn sau khi nhu cầu sử dụng giảm xuống, Runtime có thể thu nhỏ Stack để trả lại bộ nhớ cho hệ thống, giúp giảm mức sử dụng RAM và cải thiện hiệu quả của Garbage Collector.

Việc thu nhỏ không diễn ra ngay lập tức mà được Runtime cân nhắc cẩn thận, thường trong chu kỳ Garbage Collection, nhằm cân bằng giữa chi phí CPU và lợi ích về bộ nhớ.

Sau khi hoàn thành bốn phần:

- **Initial Stack**
- **Stack Growth**
- **Stack Copy**
- **Stack Shrink**

chúng ta đã hiểu đầy đủ cách Go Runtime quản lý bộ nhớ Stack của Goroutine - một trong những thiết kế quan trọng nhất giúp Go có thể hỗ trợ hàng triệu Goroutine với mức tiêu thụ bộ nhớ rất thấp.

Trong phần tiếp theo, chúng ta sẽ nghiên cứu cơ chế bảo vệ Stack khỏi lỗi tràn bộ nhớ:

> **11.12 - Stack Overflow Detection** - cách Go Runtime sử dụng `stackguard`, `morestack()` và các kiểm tra ở mức Assembly để phát hiện và xử lý Stack Overflow trước khi nó xảy ra.

---

---

# 11.12 Stack Overflow Detection - Go Runtime phát hiện Stack Overflow như thế nào?

Sau khi hoàn thành bốn phần trước, chúng ta đã biết rằng Stack của Goroutine có thể:

- Khởi tạo với kích thước rất nhỏ.
- Tự động mở rộng.
- Sao chép sang vùng nhớ mới.
- Thu nhỏ khi không còn sử dụng.

Điều này dẫn đến một câu hỏi rất quan trọng.

> **Làm thế nào Go Runtime biết Stack sắp đầy?**

Runtime không thể đợi đến khi Stack thực sự tràn.

Nếu điều đó xảy ra.

CPU sẽ ghi dữ liệu ra ngoài vùng Stack.

Kết quả có thể là:

- Ghi đè bộ nhớ.
- Crash.
- Segmentation Fault.
- Hỏng dữ liệu.

Do đó.

Go Runtime phải phát hiện Stack Overflow **trước khi nó xảy ra**.

Đây chính là nhiệm vụ của cơ chế:

```
Stack Overflow Detection
```

---

# Stack Overflow là gì?

Stack Overflow xảy ra khi.

```
Current Stack

+

New Stack Frame

>

Stack Capacity
```

Ví dụ.

```go
func recursive() {

    recursive()

}
```

Nếu Runtime không kiểm tra.

Stack sẽ liên tục.

```
Push Frame

↓

Push Frame

↓

Push Frame
```

cho đến khi vượt quá giới hạn.

---

# Tại sao Runtime không kiểm tra sau khi Push?

Một ý tưởng tưởng chừng hợp lý là.

```
Push Frame

↓

Check Overflow
```

Nhưng cách này không an toàn.

Bởi vì.

Ngay khi.

```
Push Frame
```

đã có thể ghi đè lên vùng nhớ khác.

Lúc này.

Chương trình đã bị hỏng.

Do đó.

Runtime luôn kiểm tra.

```
Before Push
```

---

# Compiler hỗ trợ Runtime

Một trong những điểm thú vị nhất của Go là.

Compiler và Runtime phối hợp rất chặt chẽ.

Trước mỗi lời gọi hàm.

Compiler sẽ sinh thêm một đoạn Assembly.

Có thể mô tả.

```text
Compare SP

↓

stackguard

↓

Enough Space?
```

Nếu còn đủ Stack.

CPU tiếp tục.

Nếu không.

Runtime được gọi.

---

# `stackguard0`

Trong `struct g`.

Có trường.

```go
stackguard0
```

Đây là giá trị rất quan trọng.

Có thể hình dung.

```text
High Address

+----------------------------+

|                            |

|        Stack               |

|                            |

+----------------------------+

| stackguard0                |

+----------------------------+

Low Address
```

Nếu Stack Pointer đi xuống dưới:

```
stackguard0
```

Runtime biết rằng.

```
Stack gần đầy
```

---

# Minh họa

Giả sử.

```text
Stack

+----------------------------+

|                            |

|         Free               |

|                            |

+----------------------------+

SP

↓

+----------------------------+

| Current Frame              |

+----------------------------+

| stackguard0                |

+----------------------------+
```

Nếu tiếp tục gọi thêm hàm.

SP sẽ tiếp tục giảm.

Ngay trước khi vượt qua:

```
stackguard0
```

Runtime sẽ can thiệp.

---

# Đoạn kiểm tra điển hình

Có thể hình dung đoạn mã Compiler sinh ra như sau.

```go
if SP < stackguard0 {

    morestack()

}
```

Nếu điều kiện sai.

CPU tiếp tục gọi hàm.

Nếu điều kiện đúng.

Runtime chuyển sang:

```
morestack()
```

để mở rộng Stack.

---

# `morestack()`

Đây là một trong những hàm nổi tiếng nhất của Go Runtime.

Nó được viết bằng Assembly.

Ví dụ.

```
runtime.morestack
```

Vai trò của nó.

```
Pause Current Function

↓

Switch sang Runtime Stack

↓

Call newstack()
```

Điều quan trọng.

`morestack()` không trực tiếp mở rộng Stack.

Nó chỉ chuẩn bị môi trường để Runtime thực hiện việc đó.

---

# Tại sao `morestack()` viết bằng Assembly?

Lúc này.

Stack của Goroutine đã gần đầy.

Nếu tiếp tục gọi hàm Go thông thường.

Chính lời gọi đó cũng cần thêm Stack.

Điều này có thể gây ra.

```
Recursive Stack Growth
```

Để tránh.

Runtime sử dụng Assembly.

Giảm tối đa lượng Stack cần sử dụng.

---

# `stackguard1`

Ngoài.

```go
stackguard0
```

Runtime còn có.

```go
stackguard1
```

Trường này chủ yếu phục vụ:

- Debug.
- C Runtime.
- Một số trường hợp đặc biệt trong Runtime.

Đối với lập trình Go thông thường.

`stackguard0` là trường quan trọng nhất.

---

# Stack Guard không phải cuối Stack

Nhiều người tưởng.

```
stackguard

=

End Of Stack
```

Điều này không đúng.

Có thể hình dung.

```text
High Address

+-----------------------------+

|                             |

|         Stack               |

|                             |

+-----------------------------+

| Guard Zone                  |

+-----------------------------+

| stackguard0                 |

+-----------------------------+

Low Address
```

Runtime luôn chừa lại một vùng an toàn.

Điều này đảm bảo.

Ngay cả khi cần gọi.

```
morestack()
```

Runtime vẫn còn đủ Stack để thực hiện.

---

# Guard Zone

Guard Zone là vùng bảo vệ.

Ví dụ.

```
Current Frame

↓

Guard Zone

↓

Stack Bottom
```

Nhờ vùng này.

Runtime có thể:

- Gọi `morestack()`.
- Chuyển sang Runtime Stack.
- Khởi tạo Stack mới.

mà không bị tràn.

---

# Điều gì xảy ra khi Growth thất bại?

Thông thường.

Runtime sẽ luôn mở rộng Stack thành công.

Tuy nhiên.

Nếu.

- Hết bộ nhớ.
- Stack vượt quá giới hạn tối đa.

Runtime sẽ dừng chương trình.

Ví dụ.

```
runtime: goroutine stack exceeds ...
```

Đây là lỗi nghiêm trọng.

---

# Đệ quy vô hạn

Ví dụ.

```go
func f() {

    f()

}
```

Quá trình.

```text
2 KB

↓

4 KB

↓

8 KB

↓

16 KB

↓

32 KB

↓

64 KB

↓

...
```

Runtime sẽ liên tục mở rộng Stack.

Đến khi đạt giới hạn.

```
Maximum Stack Size
```

Runtime sẽ Panic.

Điều này ngăn Goroutine chiếm toàn bộ bộ nhớ hệ thống.

---

# Compiler không kiểm tra mọi lệnh

Một hiểu lầm phổ biến.

Nhiều người nghĩ.

```
Every Instruction

↓

Check Stack
```

Điều này không đúng.

Compiler chủ yếu chèn kiểm tra ở.

```
Function Prologue
```

tức là đầu mỗi lời gọi hàm.

Nhờ vậy.

Chi phí Runtime rất thấp.

---

# Góc nhìn từ Assembly

Một Function trong Go thường bắt đầu bằng.

```text
CMP SP, stackguard

↓

JLS morestack

↓

Continue Function
```

Có thể mô tả.

```text
Function Entry

↓

Enough Stack?

     │

 ┌───┴────┐

 │        │

Yes       No

 │        │

 ▼        ▼

Run   morestack()
```

Đây là đoạn mã mà hầu hết mọi Function Go đều có.

---

# Stack Overflow Detection và Scheduler

Scheduler không trực tiếp kiểm tra Stack Overflow.

Nhiệm vụ này thuộc về.

- Compiler.
- Runtime.
- Assembly Stub.

Sau khi Stack được mở rộng.

Scheduler thậm chí không biết điều đó vừa xảy ra.

Đối với Scheduler.

Goroutine vẫn tiếp tục chạy bình thường.

---

# Stack Overflow Detection và Garbage Collector

Sau khi Growth.

Garbage Collector cần biết.

```
Stack Changed
```

Runtime sẽ cập nhật.

- Stack Boundary.
- Stack Pointer.
- Metadata.

GC sẽ quét Stack mới ở chu kỳ tiếp theo.

---

# Góc nhìn tổng thể

Toàn bộ quá trình.

```text
Function Call

↓

Compiler Check

↓

SP < stackguard0 ?

      │

 ┌────┴────┐

 │         │

 No       Yes

 │         │

 ▼         ▼

Run   runtime.morestack()

              │

              ▼

        runtime.newstack()

              │

              ▼

       Allocate Bigger Stack

              │

              ▼

          Copy Stack

              │

              ▼

        Adjust Pointers

              │

              ▼

      Resume Function
```

Đây là toàn bộ luồng hoạt động của Stack Overflow Detection.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Runtime phát hiện Stack Overflow sau khi Stack bị tràn.

Sai.

Runtime luôn kiểm tra trước khi ghi thêm Stack Frame.

---

### Hiểu lầm 2

> `morestack()` chỉ được gọi khi xảy ra lỗi.

Sai.

`morestack()` là cơ chế bình thường để mở rộng Stack.

Nó được sử dụng trong hầu hết các chương trình có Stack tăng trưởng.

---

### Hiểu lầm 3

> `stackguard0` chính là cuối Stack.

Sai.

Runtime luôn dành một Guard Zone phía trên `stackguard0` để đảm bảo còn đủ không gian xử lý việc mở rộng Stack.

---

### Hiểu lầm 4

> Stack Overflow Detection thuộc về Scheduler.

Sai.

Cơ chế này là sự phối hợp giữa Compiler, Assembly và Runtime.

Scheduler chỉ tiếp tục thực thi Goroutine sau khi Stack đã sẵn sàng.

---

# Những điều cần ghi nhớ

- Runtime luôn phát hiện Stack Overflow trước khi nó xảy ra.
- Compiler chèn mã kiểm tra ở đầu mỗi lời gọi hàm.
- `stackguard0` là mốc kiểm tra quan trọng nhất.
- Khi Stack không đủ, Runtime gọi `morestack()`.
- `morestack()` chuyển sang `newstack()` để mở rộng Stack.
- Runtime luôn chừa một Guard Zone để tránh ghi đè bộ nhớ.
- Scheduler gần như không tham gia vào quá trình phát hiện Stack Overflow.

---

# Tổng kết Phần III - Dynamic Stack

Sau khi hoàn thành các Section từ **11.8** đến **11.12**, chúng ta đã hiểu đầy đủ cơ chế **Dynamic Stack** của Go Runtime.

Toàn bộ quá trình có thể mô tả như sau.

```text
Create Goroutine

↓

Allocate Initial Stack (~2 KB)

↓

Function Call

↓

Check stackguard

↓

Enough Stack?

     │

 ┌───┴────┐

 │        │

Yes       No

 │        │

 ▼        ▼

Run   morestack()

           │

           ▼

      newstack()

           │

           ▼

 Allocate New Stack

           │

           ▼

     Copy Stack

           │

           ▼

 Adjust Pointers

           │

           ▼

 Continue Running

           │

           ▼

 Garbage Collection

           │

           ▼

 Need Shrink?

     │

 ┌───┴────┐

 │        │

No       Yes

 │        │

 ▼        ▼

Keep   Shrink Stack
```

Đây là một trong những thiết kế quan trọng nhất giúp Go Runtime:

- Hỗ trợ hàng triệu Goroutine.
- Tiết kiệm bộ nhớ.
- Giảm áp lực lên Garbage Collector.
- Duy trì hiệu năng cao ngay cả với các chương trình có lời gọi hàm rất sâu.

---

# Chương tiếp theo

Đến thời điểm này, chúng ta đã hiểu:

- Goroutine được biểu diễn như thế nào.
- Execution Context được lưu ở đâu.
- Stack hoạt động ra sao.
- Dynamic Stack được triển khai như thế nào.

Nhưng vẫn còn một câu hỏi rất quan trọng.

> **Làm thế nào Go Runtime có thể tạm dừng một Goroutine trên một CPU và tiếp tục nó trên CPU khác mà chương trình vẫn chạy chính xác?**

Đó chính là nội dung của **Phần IV - Context Switching**, bắt đầu với:

> **11.13 - Context Switching**.

Ở phần này, chúng ta sẽ đi sâu vào các cơ chế lưu và khôi phục ngữ cảnh thực thi, phân tích vai trò của `gobuf`, các thanh ghi CPU và mã Assembly của Go Runtime trong quá trình chuyển đổi giữa các Goroutine.

---

---

# Phần IV - Context Switching

Trong các phần trước, chúng ta đã nghiên cứu rất nhiều thành phần của Goroutine.

Chúng ta đã biết:

- Goroutine được biểu diễn bằng `struct g`.
- `gobuf` lưu Execution Context.
- Stack lưu Stack Frame.
- Runtime có thể Grow và Shrink Stack.

Tuy nhiên, vẫn còn một câu hỏi rất quan trọng.

> **Làm thế nào Go Runtime có thể tạm dừng một Goroutine và tiếp tục một Goroutine khác?**

Ví dụ.

```go
go workerA()

go workerB()
```

Giả sử.

```
workerA()
```

đang chạy.

Sau vài mili giây.

Scheduler quyết định chuyển CPU sang.

```
workerB()
```

Một lúc sau.

CPU lại quay về.

```
workerA()
```

Điều kỳ lạ là.

`workerA()` tiếp tục chạy **đúng tại vị trí trước đó**.

Không chạy lại từ đầu.

Không mất dữ liệu.

Không mất biến cục bộ.

Làm thế nào Runtime thực hiện được điều đó?

Đó chính là nhiệm vụ của:

```
Context Switching
```

---

# 11.13 Context Switching

Context Switching là một trong những cơ chế quan trọng nhất của mọi hệ điều hành và Runtime.

Không có Context Switching.

Không thể tồn tại:

- Multitasking.
- Concurrency.
- Goroutine.
- Thread Scheduling.

Có thể nói.

> **Scheduler quyết định Goroutine nào sẽ chạy.**

Trong khi.

> **Context Switching quyết định làm thế nào để Goroutine đó có thể tiếp tục chạy đúng từ vị trí trước đó.**

Đây là hai khái niệm hoàn toàn khác nhau.

---

# Context là gì?

Trước khi nói đến Context Switching.

Chúng ta cần hiểu.

```
Context
```

là gì.

Có thể hiểu đơn giản.

> **Context là toàn bộ trạng thái cần thiết để CPU có thể tiếp tục thực thi một chương trình.**

Ví dụ.

Một Goroutine đang chạy.

```go
func calculate() {

    x := 10

    y := 20

    result := x + y

    fmt.Println(result)

}
```

Giả sử Scheduler muốn tạm dừng ngay sau.

```go
result := x + y
```

Runtime phải lưu.

- Đang thực thi đến dòng nào.
- Stack hiện tại ở đâu.
- Các Register đang chứa giá trị gì.
- Goroutine nào đang chạy.

Đó chính là Execution Context.

---

# Điều gì tạo nên Context?

Một Execution Context thường bao gồm.

```text
Program Counter

+

Stack Pointer

+

CPU Registers

+

Stack

+

Execution State
```

Trong Go Runtime.

Các thông tin này nằm rải rác trong.

- `gobuf`
- `struct g`
- Stack
- CPU Registers

---

# Một ví dụ đơn giản

Giả sử.

```go
func main() {

    A()

}
```

```go
func A() {

    B()

}
```

```go
func B() {

    C()

}
```

Trong lúc.

```
C()
```

đang chạy.

Execution Context có thể được mô tả.

```text
Program Counter

↓

Instruction trong C()
```

```
Stack Pointer

↓

Frame C()
```

```
Stack

↓

main()

↓

A()

↓

B()

↓

C()
```

Nếu Runtime lưu được toàn bộ các thông tin trên.

Nó có thể tiếp tục chương trình bất kỳ lúc nào.

---

# Context Switching là gì?

Context Switching là quá trình.

```text
Save Current Context

↓

Restore New Context

↓

Continue Execution
```

Ví dụ.

```text
G1

↓

Save Context

↓

Load Context G2

↓

Run G2
```

Một lúc sau.

```text
G2

↓

Save Context

↓

Load Context G1

↓

Continue G1
```

Đối với lập trình viên.

Hai Goroutine dường như đang chạy đồng thời.

Thực tế.

CPU chỉ đang chuyển đổi rất nhanh giữa các Context.

---

# Một ví dụ trực quan

Giả sử.

```
CPU
```

đang chạy.

```
G1
```

```text
Instruction 100

↓

Instruction 101

↓

Instruction 102
```

Scheduler quyết định.

```
Switch
```

Runtime lưu.

```text
PC = 103

SP = ...

Registers = ...
```

Sau đó.

CPU chạy.

```
G2
```

Một lúc sau.

Runtime đọc lại.

```text
PC = 103

↓

Continue
```

G1 tiếp tục từ.

```
Instruction 103
```

Không có gì bị mất.

---

# Góc nhìn của CPU

Đối với CPU.

Không tồn tại khái niệm:

```
Goroutine
```

CPU chỉ biết.

```
Registers
```

```
Program Counter
```

```
Stack Pointer
```

```
Instruction
```

Runtime phải chuyển đổi các thông tin này để CPU "tin" rằng nó đang tiếp tục cùng một luồng thực thi.

---

# Góc nhìn của Go Runtime

Có thể mô tả.

```text
          Goroutine A

                │

        Save Context

                │

                ▼

             gobuf

                │

                ▼

          Goroutine B

        Restore Context

                │

                ▼

              Running
```

`gobuf` chính là cầu nối giữa Scheduler và CPU.

---

# Scheduler và Context Switching

Hai khái niệm này thường bị nhầm lẫn.

Scheduler trả lời.

```
Ai chạy tiếp?
```

Context Switching trả lời.

```
Làm sao chạy tiếp?
```

Ví dụ.

```text
Scheduler

↓

Choose G2
```

Sau đó.

```text
Context Switch

↓

Save G1

↓

Restore G2
```

Nếu không có Scheduler.

Không ai quyết định Goroutine nào sẽ chạy.

Nếu không có Context Switching.

Scheduler cũng không thể chuyển CPU sang Goroutine khác.

---

# Khi nào Context Switching xảy ra?

Có rất nhiều tình huống.

Ví dụ.

## Scheduler

```text
Time Slice

↓

Switch
```

---

## Channel

```go
<-ch
```

```text
Running

↓

Waiting

↓

Switch
```

---

## Mutex

```go
mutex.Lock()
```

Nếu Mutex đang bị giữ.

```
Running

↓

Waiting

↓

Switch
```

---

## System Call

```go
read()
```

Thread bị Block.

Runtime sẽ.

```
Detach Processor

↓

Run Goroutine khác
```

---

## Preemption

Nếu một Goroutine chạy quá lâu.

Runtime chủ động.

```
Save Context

↓

Switch
```

---

## Garbage Collector

Một số giai đoạn của Garbage Collector cũng yêu cầu Runtime phối hợp với Context Switching để đưa Goroutine về trạng thái an toàn (Safe Point).

---

# Context Switching không phải lúc nào cũng giống nhau

Trong Go Runtime.

Có hai loại chính.

## Voluntary Context Switch

Goroutine tự nhường CPU.

Ví dụ.

```go
<-ch
```

```go
time.Sleep()
```

```go
runtime.Gosched()
```

Runtime biết chính xác.

```
Switch Now
```

---

## Involuntary Context Switch

Runtime chủ động lấy CPU.

Ví dụ.

```
Asynchronous Preemption
```

Đây là trường hợp phức tạp hơn nhiều.

Runtime phải đảm bảo.

- Không làm hỏng Stack.
- Không làm hỏng Registers.
- Không làm mất dữ liệu.

---

# Context Switching có tốn chi phí không?

Có.

Mỗi lần Switch.

Runtime phải.

- Lưu Register.
- Lưu Program Counter.
- Lưu Stack Pointer.
- Cập nhật `struct g`.
- Khôi phục Context mới.

Tuy nhiên.

Chi phí này nhỏ hơn rất nhiều so với việc Kernel phải chuyển đổi giữa hai OS Thread.

Đó là lý do Goroutine có khả năng mở rộng tốt hơn.

---

# User-space Context Switching

Đây là điểm khác biệt lớn nhất.

Thread truyền thống.

```text
Thread A

↓

Kernel

↓

Thread B
```

Go Runtime.

```text
G1

↓

Runtime

↓

G2
```

Không cần Kernel tham gia.

Không cần System Call.

Không cần chuyển sang Kernel Mode.

Điều này giúp giảm đáng kể thời gian Context Switching.

---

# Minh họa toàn bộ quá trình

```text
           Scheduler

                │

                ▼

      Chọn Goroutine tiếp theo

                │

                ▼

       Save Current Context

                │

                ▼

             gobuf

                │

                ▼

      Restore New Context

                │

                ▼

          Continue Running
```

Đây là luồng cơ bản của mọi lần Context Switching trong Go Runtime.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Scheduler chính là Context Switching.

Sai.

Scheduler quyết định **Goroutine nào** chạy.

Context Switching quyết định **chuyển từ Goroutine này sang Goroutine khác như thế nào**.

---

### Hiểu lầm 2

> CPU biết Goroutine là gì.

Sai.

CPU hoàn toàn không biết Goroutine.

CPU chỉ làm việc với.

- Program Counter.
- Stack Pointer.
- Registers.

Go Runtime chịu trách nhiệm ánh xạ Goroutine vào các trạng thái mà CPU có thể hiểu.

---

### Hiểu lầm 3

> Context chỉ gồm Stack.

Sai.

Stack chỉ là một phần của Execution Context.

Một Context hoàn chỉnh còn bao gồm.

- Program Counter.
- Stack Pointer.
- Registers.
- Metadata.

---

### Hiểu lầm 4

> Context Switching luôn xảy ra khi hết Time Slice.

Sai.

Context Switching còn xảy ra khi.

- Blocking I/O.
- Channel.
- Mutex.
- Sleep.
- Garbage Collector.
- Asynchronous Preemption.

---

# Những điều cần ghi nhớ

- Context là toàn bộ trạng thái thực thi của Goroutine.
- Context Switching gồm hai bước: Save Context và Restore Context.
- `gobuf` là nơi lưu phần quan trọng nhất của Execution Context.
- Scheduler quyết định **ai chạy**, Context Switching quyết định **làm sao để chạy tiếp**.
- CPU không biết Goroutine là gì.
- Go Runtime thực hiện User-space Context Switching nên nhanh hơn Thread truyền thống.

---

# Tóm tắt

Context Switching là cơ chế cho phép Go Runtime tạm dừng một Goroutine, lưu lại toàn bộ trạng thái thực thi và tiếp tục một Goroutine khác mà không làm thay đổi hành vi của chương trình.

Đây là nền tảng giúp Go Scheduler có thể quản lý hàng triệu Goroutine trên một số lượng rất nhỏ OS Thread.

Tuy nhiên, trong phần này chúng ta mới chỉ hiểu **khái niệm**.

Một câu hỏi quan trọng vẫn còn bỏ ngỏ:

> **Runtime thực sự lưu những gì khi Save Context?**

Hay nói cách khác.

> **CPU Register nào được lưu? Program Counter và Stack Pointer được lưu vào đâu?**

Đó sẽ là nội dung của phần tiếp theo:

> **11.14 - Save Context**

Trong phần đó, chúng ta sẽ đi xuống mức **CPU Register**, **Assembly** và **`gobuf`**, phân tích từng trường dữ liệu được Runtime lưu trong mỗi lần Context Switching.

---

---

# 11.14 Save Registers - Go Runtime lưu CPU Registers như thế nào?

Trong phần trước, chúng ta đã biết rằng Context Switching bao gồm hai bước:

```text
Save Context

↓

Restore Context
```

Chúng ta cũng biết rằng Runtime cần lưu:

- Program Counter.
- Stack Pointer.
- CPU Registers.
- Execution Context.

Nhưng vẫn còn một câu hỏi quan trọng.

> **Runtime thực sự lưu những Register nào của CPU?**

Đây là điểm mà rất nhiều lập trình viên Go chưa từng tìm hiểu.

Trong thực tế.

CPU không lưu trạng thái của Goroutine.

CPU chỉ có:

- Registers.
- Program Counter.
- Stack Pointer.

Nếu Runtime không lưu các Register này.

Sau khi chuyển sang Goroutine khác.

Toàn bộ phép tính đang thực hiện sẽ bị mất.

Đó là lý do bước:

```
Save Registers
```

là trái tim của Context Switching.

---

# CPU Register là gì?

Register là những vùng nhớ cực nhỏ nằm ngay bên trong CPU.

Có thể hình dung.

```text
CPU

+----------------------+

| Register AX          |

| Register BX          |

| Register CX          |

| Register DX          |

| Program Counter      |

| Stack Pointer        |

+----------------------+
```

Đây là nơi CPU lưu trữ dữ liệu mà nó đang xử lý.

Register nhanh hơn RAM rất nhiều.

Thông thường.

CPU sẽ luôn cố gắng thực hiện phép tính trực tiếp trên Register.

---

# Vì sao phải lưu Register?

Giả sử.

CPU đang thực hiện.

```go
result := a + b
```

Compiler có thể sinh mã máy như sau.

```text
MOV AX, a

ADD AX, b
```

Lúc này.

```
AX
```

đang chứa kết quả trung gian.

Nếu Scheduler đột ngột chuyển sang Goroutine khác.

```
AX
```

sẽ bị ghi đè.

Khi quay lại.

```
result
```

không còn chính xác.

Do đó.

Runtime phải lưu lại Register trước khi Context Switching.

---

# Một ví dụ trực quan

Giả sử.

```
G1
```

đang chạy.

CPU.

```text
AX = 10

BX = 20

CX = 30

PC = foo()

SP = ...
```

Scheduler quyết định.

```
Switch sang G2
```

Runtime phải lưu.

```text
AX

↓

gobuf
```

```text
BX

↓

gobuf
```

```text
PC

↓

gobuf
```

```text
SP

↓

gobuf
```

Sau đó.

CPU mới được phép chạy G2.

---

# Register nào cần lưu?

Đây là câu hỏi rất hay trong phỏng vấn.

Câu trả lời là:

> **Không phải mọi Register đều cần lưu.**

Điều này phụ thuộc vào:

- Kiến trúc CPU.
- Calling Convention.
- ABI của Go.

Ví dụ.

Trên kiến trúc AMD64.

Một số Register sẽ được Runtime bảo toàn.

Một số Register khác được phép ghi đè.

---

# Calling Convention

Calling Convention là tập hợp các quy tắc quy định.

- Hàm truyền tham số như thế nào.
- Hàm trả kết quả như thế nào.
- Register nào phải được bảo toàn.
- Register nào có thể thay đổi.

Ví dụ.

Một số Register được gọi là.

```
Caller Saved
```

Caller phải tự lưu nếu muốn giữ giá trị.

Ngược lại.

Một số Register là.

```
Callee Saved
```

Hàm được gọi phải khôi phục chúng trước khi trả về.

Go Runtime tuân theo Calling Convention của chính mình kết hợp với ABI của từng nền tảng.

---

# Program Counter

Program Counter.

```
PC
```

là Register quan trọng nhất.

Ví dụ.

```go
foo()

bar()

baz()
```

Nếu Runtime tạm dừng sau.

```go
bar()
```

PC sẽ lưu.

```
Địa chỉ của baz()
```

Sau khi Restore.

CPU tiếp tục.

```
baz()
```

Nếu PC không được lưu.

Chương trình sẽ:

- Chạy lại từ đầu.
- Hoặc nhảy tới vị trí sai.

---

# Stack Pointer

Register tiếp theo.

```
SP
```

SP luôn trỏ tới.

```
Current Stack Frame
```

Ví dụ.

```text
Stack

+----------------------+

| main                 |

+----------------------+

| foo                  |

+----------------------+

| bar                  |

+----------------------+

SP

↓

bar()
```

Sau khi Switch.

SP phải được lưu.

Nếu không.

CPU sẽ sử dụng Stack của Goroutine khác.

Điều này sẽ phá hỏng chương trình.

---

# Base Pointer

Một số kiến trúc còn sử dụng.

```
BP
```

Base Pointer.

BP hỗ trợ.

- Truy cập Stack Frame.
- Debug.
- Stack Walking.

Runtime sẽ lưu BP khi cần.

---

# Register chứa Goroutine hiện tại

Go Runtime còn có một Register đặc biệt.

Dùng để truy cập nhanh.

```
Current g
```

Có thể hình dung.

```text
CPU Register

↓

Current Goroutine
```

Nhờ vậy.

Runtime không phải tìm kiếm Goroutine hiện tại trong bộ nhớ.

Đây là một tối ưu rất quan trọng.

---

# Save vào đâu?

Tất cả các Register quan trọng sẽ được lưu vào.

```
gobuf
```

Có thể mô tả.

```text
CPU Registers

↓

gobuf

↓

struct g
```

Sau đó.

Scheduler có thể chuyển sang Goroutine khác.

---

# Quá trình Save

Có thể mô tả đơn giản.

```text
Running

↓

Save PC

↓

Save SP

↓

Save Registers

↓

Update gobuf

↓

Switch
```

Toàn bộ quá trình chỉ mất một khoảng thời gian rất ngắn.

---

# Ví dụ minh họa

Ban đầu.

```
CPU

↓

G1
```

```text
PC = foo()

SP = 0x1000

AX = 10

BX = 20
```

Sau khi Save.

```text
gobuf

PC = foo()

SP = 0x1000

AX = 10

BX = 20
```

CPU có thể sử dụng.

```
AX

BX

SP
```

cho Goroutine khác.

---

# Góc nhìn từ Assembly

Trong Runtime.

Việc Save Register thường được thực hiện bằng Assembly.

Có thể mô tả.

```text
MOV SP

↓

gobuf.sp
```

```text
MOV PC

↓

gobuf.pc
```

```text
MOV BP

↓

gobuf.bp
```

Sau đó.

Runtime tiếp tục lưu các Register cần thiết khác.

Đây là lý do Context Switching của Go có phần Assembly riêng cho từng kiến trúc CPU.

---

# Tại sao không lưu toàn bộ Register?

Một câu hỏi rất hay.

Nếu Runtime lưu toàn bộ Register.

Context Switching sẽ chậm hơn.

Do đó.

Runtime chỉ lưu:

```
Registers thật sự cần thiết
```

Các Register tạm thời.

```
Caller Saved
```

sẽ được Compiler xử lý theo Calling Convention.

Điều này giúp Context Switching nhanh hơn.

---

# Save Register và Scheduler

Scheduler không trực tiếp thao tác với Register.

Scheduler chỉ yêu cầu.

```
Switch G1

↓

G2
```

Việc lưu Register do các hàm Assembly của Runtime thực hiện.

Sau khi hoàn tất.

Scheduler tiếp tục công việc của mình.

---

# Save Register và Garbage Collector

Garbage Collector cũng cần biết.

Các Pointer hiện đang nằm trong Register.

Trước khi GC quét.

Runtime sẽ đưa Goroutine về Safe Point.

Lúc này.

Các Pointer quan trọng đã được ghi xuống Stack hoặc được Runtime quản lý.

Nhờ vậy.

GC có thể xác định chính xác các đối tượng còn sống.

---

# Minh họa toàn bộ quá trình

```text
Running Goroutine

↓

Scheduler Request

↓

Save PC

↓

Save SP

↓

Save BP

↓

Save Registers

↓

Store vào gobuf

↓

Context Saved
```

Sau bước này.

Runtime đã sẵn sàng chuyển sang Goroutine khác.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Runtime lưu toàn bộ Register của CPU.

Sai.

Runtime chỉ lưu các Register cần thiết theo ABI và Calling Convention.

---

### Hiểu lầm 2

> Register được lưu trên Stack.

Không hoàn toàn đúng.

Trong Context Switching.

Các Register quan trọng được lưu vào `gobuf`.

Một số giá trị khác có thể đã được Compiler ghi xuống Stack trước đó.

---

### Hiểu lầm 3

> Scheduler thao tác trực tiếp với Register.

Sai.

Việc Save và Restore Register được thực hiện bởi mã Assembly của Go Runtime.

Scheduler chỉ quyết định thời điểm chuyển đổi.

---

### Hiểu lầm 4

> Program Counter nằm trong Stack.

Sai.

Program Counter là một Register của CPU.

Go Runtime chỉ sao chép giá trị của nó vào `gobuf` khi thực hiện Context Switching.

---

# Những điều cần ghi nhớ

- Register là bộ nhớ tốc độ cao nằm trong CPU.
- Context Switching luôn phải lưu Program Counter và Stack Pointer.
- Các Register quan trọng được lưu vào `gobuf`.
- Runtime không lưu toàn bộ Register của CPU.
- Calling Convention quyết định Register nào cần được bảo toàn.
- Việc Save Register được thực hiện bằng Assembly để đạt hiệu năng cao.

---

# Tóm tắt

Save Registers là bước đầu tiên của Context Switching.

Trong bước này, Go Runtime lưu các Register quan trọng như Program Counter, Stack Pointer và một số Register khác vào `gobuf`, từ đó bảo toàn toàn bộ trạng thái thực thi của Goroutine trước khi chuyển CPU sang Goroutine khác.

Việc chỉ lưu những Register cần thiết giúp Go Runtime thực hiện Context Switching với chi phí rất thấp, đồng thời vẫn đảm bảo Goroutine có thể tiếp tục chạy chính xác khi được Scheduler chọn lại.

Tuy nhiên, lưu Context mới chỉ là một nửa của quá trình.

Sau khi chọn Goroutine mới, Runtime còn phải thực hiện bước ngược lại:

> **Khôi phục toàn bộ Register, Program Counter và Stack Pointer từ `gobuf` để CPU tiếp tục thực thi.**

Đó sẽ là nội dung của phần tiếp theo:

> **11.15 - Restore Context**.

---

---

# 11.15 Restore Context - Go Runtime khôi phục Execution Context như thế nào?

Trong phần trước, chúng ta đã tìm hiểu bước đầu tiên của Context Switching:

```text
Save Context

↓

Save Registers

↓

Store vào gobuf
```

Sau bước này.

Execution Context của Goroutine đã được lưu an toàn trong `gobuf`.

Nhưng vẫn còn một nửa câu chuyện.

Scheduler vừa chọn một Goroutine mới.

Ví dụ.

```
G2
```

Làm thế nào CPU có thể tiếp tục chạy G2 từ đúng vị trí trước đó?

Runtime phải thực hiện quá trình:

```
Restore Context
```

Có thể nói.

Nếu **Save Context** giống như việc lưu trạng thái của một trò chơi (Save Game).

Thì **Restore Context** chính là thao tác **Load Game**.

Toàn bộ trạng thái trước đó sẽ được khôi phục.

CPU tiếp tục chạy như chưa từng bị gián đoạn.

---

# Restore Context là gì?

Restore Context là quá trình.

```text
Read gobuf

↓

Restore Registers

↓

Restore Stack Pointer

↓

Restore Program Counter

↓

Continue Execution
```

Sau khi hoàn thành.

CPU tiếp tục chạy đúng tại lệnh tiếp theo của Goroutine.

Không chạy lại từ đầu.

Không mất biến.

Không mất Stack.

---

# Ví dụ đơn giản

Giả sử.

```
G1
```

đang chạy.

CPU.

```text
PC = foo()

SP = 0x1000

AX = 100
```

Scheduler quyết định chuyển sang.

```
G2
```

Sau khi Save.

CPU bắt đầu chạy G2.

Một lúc sau.

Scheduler muốn quay lại G1.

Runtime sẽ đọc.

```text
gobuf

↓

PC

↓

SP

↓

Registers
```

Sau đó.

CPU tiếp tục thực thi.

```
foo()
```

từ đúng vị trí đã dừng.

---

# Góc nhìn trực quan

Có thể hình dung.

```text
               Scheduler

                   │

                   ▼

          Chọn Goroutine G2

                   │

                   ▼

             struct g

                   │

                   ▼

               gobuf

                   │

                   ▼

         Restore Registers

                   │

                   ▼

          Continue Running
```

`gobuf` chính là cầu nối giữa Scheduler và CPU.

---

# Runtime khôi phục những gì?

Tương tự như Save Context.

Runtime cần khôi phục.

- Program Counter.
- Stack Pointer.
- Base Pointer.
- CPU Registers.
- Current Goroutine.

Đây là toàn bộ trạng thái mà CPU cần để tiếp tục thực thi.

---

# Khôi phục Program Counter

Đây là bước quan trọng nhất.

Giả sử.

```go
func worker() {

    step1()

    step2()

    step3()

}
```

Goroutine bị tạm dừng sau.

```go
step2()
```

Program Counter đã được lưu.

```
PC

↓

step3()
```

Khi Restore.

Runtime chỉ cần.

```text
PC

↓

CPU
```

CPU lập tức tiếp tục.

```
step3()
```

Nếu Program Counter sai.

Chương trình sẽ:

- Chạy sai.
- Chạy lại.
- Hoặc Crash.

---

# Khôi phục Stack Pointer

Tiếp theo.

Runtime khôi phục.

```
SP
```

Stack Pointer.

Ví dụ.

```text
Stack

+--------------------------+

| main                     |

+--------------------------+

| process                  |

+--------------------------+

| validate                 |

+--------------------------+

SP

↓

validate()
```

Sau khi Restore.

CPU tiếp tục sử dụng đúng Stack này.

Nếu SP không đúng.

CPU sẽ đọc:

- Sai biến.
- Sai Return Address.
- Sai Stack Frame.

---

# Khôi phục Base Pointer

Trên nhiều kiến trúc.

Runtime cũng khôi phục.

```
BP
```

Điều này giúp.

- Stack Walking.
- Debugging.
- Panic.
- Garbage Collection.

hoạt động chính xác.

---

# Khôi phục Current Goroutine

Runtime cũng cần khôi phục.

```
Current g
```

Có thể hình dung.

```text
CPU

↓

Current g

↓

G2
```

Từ thời điểm này.

Mọi lời gọi Runtime đều biết.

```
Current Goroutine

=

G2
```

---

# Restore Registers

Sau Program Counter và Stack Pointer.

Runtime khôi phục các Register khác.

Ví dụ.

```text
AX

BX

CX

DX
```

(tùy theo ABI và kiến trúc CPU.)

Sau bước này.

CPU gần như trở về trạng thái trước khi Goroutine bị tạm dừng.

---

# Góc nhìn từ Assembly

Quá trình Restore chủ yếu được viết bằng Assembly.

Có thể mô tả.

```text
MOV gobuf.sp

↓

SP
```

```text
MOV gobuf.pc

↓

PC
```

```text
MOV gobuf.bp

↓

BP
```

Sau đó.

Runtime khôi phục các Register còn lại.

Cuối cùng.

CPU thực hiện.

```
JMP PC
```

và tiếp tục chạy.

---

# Tại sao dùng JMP thay vì CALL?

Đây là chi tiết rất thú vị.

Nếu Runtime sử dụng.

```
CALL
```

CPU sẽ tạo thêm.

```
Stack Frame
```

Điều này làm thay đổi Stack hiện tại.

Runtime không muốn điều đó.

Do đó.

Sau khi khôi phục.

Runtime thường sử dụng.

```
Jump

↓

Continue
```

để tiếp tục đúng luồng thực thi cũ.

---

# Ví dụ toàn bộ quá trình

Ban đầu.

```
CPU

↓

G1
```

```text
PC = foo()

SP = 0x1000
```

Sau Context Switch.

```
CPU

↓

G2
```

Một lúc sau.

Scheduler quay lại.

```
G1
```

Runtime thực hiện.

```text
Read gobuf

↓

Restore PC

↓

Restore SP

↓

Restore Registers

↓

Jump

↓

Continue foo()
```

Đối với lập trình viên.

Không thể nhận ra Goroutine đã từng bị tạm dừng.

---

# Restore Context và Stack Growth

Một điểm thú vị.

Giả sử.

Trong lúc Goroutine bị tạm dừng.

Runtime vừa thực hiện.

```
Stack Growth
```

Khi Restore.

```
SP
```

không còn trỏ tới Stack cũ.

Mà trỏ tới.

```
New Stack
```

Điều này hoàn toàn bình thường.

Runtime đã cập nhật.

```
gobuf.sp
```

sau khi Stack được Copy.

CPU không hề biết Stack từng bị di chuyển.

---

# Restore Context và Scheduler

Scheduler chỉ làm hai việc.

```text
Choose G

↓

Call Runtime Switch
```

Phần còn lại.

- Restore Registers.
- Restore PC.
- Restore SP.

đều do Runtime thực hiện.

Sau khi hoàn tất.

Scheduler coi Goroutine đang ở trạng thái.

```
Running
```

---

# Restore Context và Garbage Collector

Garbage Collector yêu cầu.

```
Context Consistency
```

Điều này có nghĩa.

Sau khi Restore.

- Stack Pointer phải đúng.
- Stack Frame phải đúng.
- Pointer phải đúng.

Nếu không.

Garbage Collector sẽ quét sai Stack.

Đây là lý do Runtime phải hoàn tất Restore trước khi Goroutine tiếp tục chạy.

---

# Restore Context có tốn chi phí không?

Có.

Nhưng khá nhỏ.

Runtime chỉ cần.

- Đọc vài Register.
- Khôi phục Stack Pointer.
- Khôi phục Program Counter.

Chi phí này thấp hơn nhiều so với việc Kernel khôi phục toàn bộ trạng thái của một OS Thread.

---

# Minh họa toàn bộ quá trình

```text
Scheduler

↓

Choose Goroutine

↓

Read gobuf

↓

Restore PC

↓

Restore SP

↓

Restore Registers

↓

Set Current g

↓

Jump

↓

Continue Execution
```

Đây là toàn bộ quá trình Restore Context.

---

# Save và Restore luôn đi cùng nhau

Context Switching luôn bao gồm.

```text
Save Context

↓

Scheduler

↓

Restore Context
```

Không bao giờ chỉ có.

```
Save
```

hoặc.

```
Restore
```

riêng lẻ.

Đây là hai nửa của cùng một cơ chế.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Restore Context tạo lại Goroutine.

Sai.

Goroutine đã tồn tại.

Runtime chỉ khôi phục trạng thái thực thi của nó.

---

### Hiểu lầm 2

> Restore Context chạy lại hàm từ đầu.

Sai.

CPU tiếp tục từ Program Counter đã lưu.

---

### Hiểu lầm 3

> Restore Context cấp phát Stack mới.

Sai.

Stack đã tồn tại.

Runtime chỉ khôi phục Stack Pointer.

Nếu Stack từng được Grow hoặc Shrink, `gobuf` đã được cập nhật trước đó.

---

### Hiểu lầm 4

> Scheduler trực tiếp khôi phục Register.

Sai.

Scheduler chỉ quyết định Goroutine nào sẽ chạy.

Việc Restore Register do mã Assembly của Runtime thực hiện.

---

# Những điều cần ghi nhớ

- Restore Context là bước thứ hai của Context Switching.
- Runtime đọc Execution Context từ `gobuf`.
- Program Counter quyết định CPU tiếp tục từ đâu.
- Stack Pointer quyết định CPU sử dụng Stack nào.
- Runtime sử dụng Jump để tiếp tục thực thi thay vì tạo lời gọi hàm mới.
- Scheduler và Context Switching là hai cơ chế khác nhau nhưng luôn phối hợp chặt chẽ.

---

# Tóm tắt

Restore Context là quá trình Go Runtime khôi phục toàn bộ trạng thái thực thi của một Goroutine từ `gobuf`, bao gồm Program Counter, Stack Pointer, CPU Registers và các thông tin cần thiết khác.

Sau khi quá trình này hoàn tất, CPU tiếp tục thực thi Goroutine từ đúng vị trí mà nó đã dừng trước đó, tạo cảm giác như Goroutine chưa từng bị gián đoạn.

Kết hợp với Save Context, đây là nền tảng của toàn bộ cơ chế Context Switching và là yếu tố giúp Go Runtime có thể quản lý hàng triệu Goroutine trên một số lượng rất nhỏ OS Thread.

Trong phần tiếp theo, chúng ta sẽ đi sâu hơn vào cấp độ Assembly để trả lời câu hỏi:

> **Điều gì thực sự xảy ra khi CPU thực hiện một lời gọi hàm (`CALL`) và quay trở về (`RET`)?**

Đó sẽ là nội dung của:

> **11.16 - Function Call**.

Hiểu rõ cơ chế `CALL` và `RET` sẽ giúp chúng ta kết nối tất cả các kiến thức về Stack Frame, Program Counter, Return Address và Context Switching thành một bức tranh hoàn chỉnh về cách Go Runtime thực thi chương trình.

---

---

# 11.16 Function Call - Điều gì thực sự xảy ra khi gọi một hàm?

Trong các phần trước, chúng ta đã tìm hiểu:

- Stack
- Stack Frame
- Program Counter
- Stack Pointer
- Context Switching
- Save Context
- Restore Context

Tuy nhiên, vẫn còn một câu hỏi rất quan trọng.

> **Điều gì thực sự xảy ra bên trong CPU khi chúng ta gọi một hàm?**

Ví dụ.

```go
func main() {

    foo()

}
```

Đây là một đoạn code rất bình thường.

Nhưng phía sau lời gọi.

```go
foo()
```

CPU phải thực hiện rất nhiều công việc.

Hiểu được Function Call sẽ giúp chúng ta kết nối toàn bộ kiến thức đã học về:

- Stack
- Stack Frame
- Registers
- Program Counter
- Return Address
- Context Switching

Đây chính là nền tảng để đọc Assembly của Go Runtime.

---

# Một lời gọi hàm đơn giản

Ví dụ.

```go
func main() {

    foo()

    fmt.Println("Done")

}

func foo() {

    fmt.Println("Hello")

}
```

Ở mức Go.

Chúng ta chỉ thấy.

```go
foo()
```

Nhưng ở mức CPU.

Quá trình phức tạp hơn rất nhiều.

---

# Function Call gồm những bước nào?

Một lời gọi hàm thường bao gồm các bước sau.

```text
Prepare Parameters

↓

Save Return Address

↓

Create Stack Frame

↓

Jump tới Function

↓

Execute Function

↓

Return

↓

Restore Caller
```

Đây là quy trình chung của hầu hết mọi CPU hiện đại.

---

# Bước 1 - Chuẩn bị tham số

Giả sử.

```go
sum(10, 20)
```

CPU phải truyền.

```
10

20
```

cho hàm.

Tùy theo ABI.

Các tham số có thể được đặt vào:

- Register.
- Hoặc Stack.

Ví dụ.

```text
AX = 10

BX = 20
```

Hoặc.

```text
Stack

↓

10

20
```

Compiler sẽ quyết định cách tối ưu nhất.

---

# Bước 2 - Lưu Return Address

Đây là bước quan trọng nhất.

Giả sử.

```go
foo()

fmt.Println("Done")
```

Sau khi.

```
foo()
```

kết thúc.

CPU phải biết.

```
Quay về đâu?
```

CPU sẽ lưu địa chỉ của lệnh tiếp theo.

```
fmt.Println("Done")
```

Địa chỉ này được gọi là.

```
Return Address
```

---

# Minh họa

Có thể hình dung.

```text
main()

↓

CALL foo()

↓

Save Return Address

↓

foo()
```

Sau khi.

```
foo()
```

kết thúc.

CPU đọc.

```
Return Address
```

↓

Tiếp tục.

```
fmt.Println("Done")
```

---

# Bước 3 - Tạo Stack Frame

Sau khi lưu Return Address.

CPU tạo.

```
Stack Frame
```

mới.

Ví dụ.

```go
func foo() {

    x := 10

}
```

Stack lúc này.

```text
+--------------------------+

| Return Address           |

+--------------------------+

| Local Variable x         |

+--------------------------+
```

Mỗi lần gọi hàm.

Một Stack Frame mới được Push lên Stack.

---

# Bước 4 - Chuyển Program Counter

Lúc này.

Program Counter vẫn đang trỏ tới.

```go
foo()
```

CPU cập nhật.

```
PC

↓

Entry Address của foo()
```

Có thể hình dung.

```text
Old PC

↓

CALL

↓

New PC
```

Đây chính là lúc CPU bắt đầu thực thi hàm mới.

---

# Bước 5 - Thực thi Function

Sau khi Program Counter được cập nhật.

CPU bắt đầu.

```text
Instruction 1

↓

Instruction 2

↓

Instruction 3
```

cho đến khi gặp.

```
RET
```

---

# Bước 6 - RET

Lệnh.

```
RET
```

(Return)

có nhiệm vụ.

```text
Read Return Address

↓

Restore PC

↓

Continue
```

Sau đó.

CPU tiếp tục chạy.

```
Caller
```

---

# Toàn bộ quá trình

Có thể mô tả.

```text
main()

↓

CALL foo()

↓

Push Return Address

↓

Create Stack Frame

↓

Jump tới foo()

↓

Run foo()

↓

RET

↓

Pop Stack Frame

↓

Restore PC

↓

Continue main()
```

Đây là quy trình chuẩn của mọi lời gọi hàm.

---

# Minh họa Call Stack

Ban đầu.

```text
Stack

+--------------------------+

| Frame main()             |

+--------------------------+
```

Sau khi.

```
foo()
```

```text
Stack

+--------------------------+

| Frame main()             |

+--------------------------+

| Frame foo()              |

+--------------------------+
```

Sau khi.

```
foo()

↓

RET
```

```text
Stack

+--------------------------+

| Frame main()             |

+--------------------------+
```

---

# Program Counter thay đổi như thế nào?

Ban đầu.

```
PC

↓

CALL foo
```

Sau khi.

```
CALL
```

```
PC

↓

Instruction đầu tiên của foo()
```

Sau khi.

```
RET
```

```
PC

↓

Instruction sau CALL
```

Toàn bộ quá trình diễn ra hoàn toàn tự động.

---

# Stack Pointer thay đổi như thế nào?

Ban đầu.

```text
SP

↓

Frame main()
```

Sau khi.

```
CALL
```

```text
SP

↓

Frame foo()
```

Sau khi.

```
RET
```

```text
SP

↓

Frame main()
```

SP luôn trỏ tới Frame hiện tại.

---

# CALL và Context Switching

Một điểm rất thú vị.

Nhiều người nhầm lẫn.

```
CALL

=

Context Switch
```

Điều này hoàn toàn sai.

CALL chỉ chuyển.

```
Function

↓

Function
```

trong cùng một Goroutine.

Ví dụ.

```text
main()

↓

foo()

↓

bar()
```

CPU vẫn đang chạy cùng một Goroutine.

---

# Context Switching khác gì?

Context Switching là.

```text
G1

↓

Save Context

↓

G2

↓

Restore Context
```

Đây là hai Goroutine khác nhau.

Có thể hình dung.

```text
CALL

↓

Function Level
```

Trong khi.

```text
Context Switch

↓

Goroutine Level
```

Đây là hai tầng hoàn toàn khác nhau.

---

# Function Call và Scheduler

Scheduler không tham gia.

```
foo()

↓

bar()
```

Đây chỉ là lời gọi hàm bình thường.

Scheduler chỉ tham gia khi.

- Blocking.
- Preemption.
- Channel.
- Mutex.
- System Call.

---

# Function Call và Stack Growth

Trước mỗi.

```
CALL
```

Compiler sẽ kiểm tra.

```
Enough Stack?
```

Nếu.

```
No
```

Runtime gọi.

```
morestack()
```

Sau khi Growth.

CPU mới thực hiện.

```
CALL
```

Điều này đảm bảo.

Không bao giờ Push Stack Frame ra ngoài vùng Stack.

---

# Function Call trong Go Runtime

Một lời gọi hàm Go thực tế có thể được mô tả.

```text
Function Entry

↓

Check stackguard

↓

Need Growth?

↓

morestack()

↓

CALL

↓

Push Frame

↓

Execute

↓

RET

↓

Pop Frame
```

Đây là luồng hoạt động của hầu hết các Function trong Go.

---

# CALL và Assembly

Trong Assembly.

Một lời gọi hàm thường là.

```text
CALL foo
```

CPU tự động.

- Push Return Address.
- Cập nhật Program Counter.

Sau khi.

```
RET
```

CPU tự động.

- Pop Return Address.
- Khôi phục Program Counter.

Đây là cơ chế được hỗ trợ trực tiếp bởi phần cứng.

---

# Function Call và Return Address

Return Address luôn nằm trong Stack.

Ví dụ.

```text
Frame foo()

+--------------------------+

| Return Address           |

+--------------------------+

| Local Variables          |

+--------------------------+
```

Sau khi.

```
RET
```

Return Address bị Pop khỏi Stack.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> CALL tạo Goroutine mới.

Sai.

CALL chỉ gọi một Function trong cùng Goroutine.

---

### Hiểu lầm 2

> Scheduler tham gia mọi Function Call.

Sai.

Scheduler chỉ tham gia khi cần chuyển Goroutine.

Function Call thông thường hoàn toàn do CPU xử lý.

---

### Hiểu lầm 3

> RET chạy lại từ đầu hàm Caller.

Sai.

RET sử dụng Return Address để tiếp tục đúng tại lệnh ngay sau CALL.

---

### Hiểu lầm 4

> Context Switching và Function Call giống nhau.

Sai.

Function Call xảy ra trong phạm vi một Goroutine.

Context Switching chuyển CPU giữa nhiều Goroutine.

---

# Những điều cần ghi nhớ

- Function Call được CPU hỗ trợ trực tiếp thông qua lệnh `CALL`.
- `CALL` sẽ lưu Return Address và chuyển Program Counter đến hàm được gọi.
- Mỗi lời gọi hàm tạo ra một Stack Frame mới.
- `RET` sẽ khôi phục Program Counter từ Return Address.
- Function Call không liên quan trực tiếp đến Scheduler.
- Trước mỗi lời gọi hàm, Go Runtime kiểm tra Stack còn đủ hay không.

---

# Tóm tắt

Function Call là cơ chế cơ bản nhất của quá trình thực thi chương trình.

Mỗi lần gọi hàm, CPU sẽ lưu địa chỉ quay về, tạo Stack Frame mới, cập nhật Program Counter và bắt đầu thực thi hàm được gọi. Khi hàm kết thúc, lệnh `RET` sẽ khôi phục Program Counter và Stack Pointer để chương trình tiếp tục từ đúng vị trí trước đó.

Hiểu rõ cơ chế `CALL` và `RET` giúp chúng ta kết nối các khái niệm về Stack Frame, Program Counter, Return Address và Context Switching, đồng thời tạo nền tảng để đọc mã Assembly và source code của Go Runtime.

Trong phần tiếp theo, chúng ta sẽ tìm hiểu cơ chế ngược lại với Function Call:

> **11.17 - Function Return**.

Mặc dù lệnh `RET` chỉ gồm một vài chỉ thị Assembly, nhưng phía sau nó là sự phối hợp giữa Stack, Return Address, Program Counter và Calling Convention để đảm bảo chương trình quay trở lại đúng vị trí và tiếp tục thực thi một cách chính xác.

---

---

# 11.17 Defer - Cơ chế trì hoãn thực thi của Go Runtime

Trong các phần trước, chúng ta đã tìm hiểu cách một hàm được gọi và kết thúc.

Một lời gọi hàm thông thường diễn ra theo trình tự.

```text
CALL

↓

Execute Function

↓

RET
```

Tuy nhiên, Go cung cấp một cơ chế đặc biệt.

```go
func main() {

    defer fmt.Println("A")

    fmt.Println("B")

}
```

Kết quả.

```
B

A
```

Mặc dù `defer` được khai báo trước.

Nhưng lại được thực thi sau.

Điều này dẫn đến câu hỏi.

> Runtime lưu các `defer` ở đâu?

---

# Defer là gì?

`defer` là một cơ chế yêu cầu Runtime:

> **Thực thi một lời gọi hàm ngay trước khi Function hiện tại kết thúc.**

Ví dụ.

```go
func process() {

    file := open()

    defer file.Close()

}
```

Khi Function kết thúc.

Runtime sẽ tự động gọi.

```go
file.Close()
```

Điều này giúp lập trình viên tránh quên giải phóng tài nguyên.

---

# Runtime xử lý Defer như thế nào?

Mỗi lần gặp.

```go
defer
```

Runtime không thực thi ngay.

Thay vào đó.

Runtime tạo một đối tượng.

```
_defer
```

và liên kết nó với Goroutine hiện tại.

Có thể hình dung.

```text
struct g

↓

defer list

↓

Close()

↓

Unlock()

↓

Flush()
```

---

# Defer hoạt động theo LIFO

Ví dụ.

```go
defer A()

defer B()

defer C()
```

Thứ tự thực thi.

```
C()

↓

B()

↓

A()
```

Đây là nguyên tắc.

```
Last In

First Out
```

Tương tự như Stack.

---

# Runtime lưu Defer ở đâu?

Trong `struct g`.

Có trường.

```go
_defer
```

Trường này trỏ tới danh sách các Defer đang chờ thực thi.

Có thể hình dung.

```text
struct g

↓

_defer

↓

Close()

↓

Unlock()

↓

Flush()
```

---

# Khi nào Runtime thực thi Defer?

Runtime thực thi Defer khi.

- Function Return.
- Panic.
- Goroutine Exit.

Ngay trước khi Function thực sự rời khỏi Stack.

Runtime sẽ duyệt toàn bộ danh sách Defer.

---

# Defer và Context Switching

Nếu Goroutine bị Suspend.

Danh sách `_defer` vẫn nằm trong `struct g`.

Sau khi Restore Context.

Defer tiếp tục tồn tại.

Không bị ảnh hưởng bởi Scheduler.

---

# Những điều cần ghi nhớ

- Defer không chạy ngay.
- Runtime tạo `_defer`.
- Defer thực thi theo LIFO.
- `_defer` nằm trong `struct g`.
- Scheduler không quản lý Defer.

---

# Tóm tắt

Defer là một cơ chế được Runtime quản lý nhằm đảm bảo các thao tác dọn dẹp luôn được thực hiện trước khi Function kết thúc.

Mỗi `defer` được lưu trong `_defer` của Goroutine và được thực thi theo nguyên tắc LIFO.

---

# 11.18 Panic - Runtime xử lý lỗi nghiêm trọng như thế nào?

Trong Go.

Có hai cách báo lỗi.

```go
return err
```

và.

```go
panic(...)
```

Khác với Error.

`panic` không chỉ trả về.

Nó bắt đầu quá trình.

```
Stack Unwinding
```

---

# Panic là gì?

Ví dụ.

```go
func worker() {

    panic("database error")

}
```

Sau khi Panic.

Runtime sẽ.

```
Stop Normal Execution

↓

Execute Defer

↓

Walk Stack

↓

Exit
```

---

# Runtime tạo `_panic`

Tương tự `_defer`.

Runtime tạo.

```
_panic
```

được lưu trong `struct g`.

```text
struct g

↓

_panic

↓

"database error"
```

---

# Stack Unwinding

Giả sử.

```text
main()

↓

A()

↓

B()

↓

C()

↓

panic()
```

Runtime bắt đầu.

```text
Execute defer C()

↓

Pop Frame

↓

Execute defer B()

↓

Pop Frame

↓

Execute defer A()

↓

Pop Frame
```

Đây gọi là.

```
Stack Unwinding
```

---

# Panic và Scheduler

Panic không phải Context Switch.

Scheduler không tham gia.

Runtime xử lý toàn bộ.

---

# Nếu không Recover?

Runtime tiếp tục Unwind.

Cho đến.

```
main()
```

Sau đó.

```
Crash Program
```

---

# Những điều cần ghi nhớ

- Panic tạo `_panic`.
- Runtime bắt đầu Stack Unwinding.
- Defer luôn được thực thi.
- Nếu không Recover.
- Chương trình kết thúc.

---

# Tóm tắt

Panic là cơ chế Runtime sử dụng để xử lý các lỗi không thể tiếp tục thực thi.

Trong quá trình Panic, Runtime sẽ thực hiện Stack Unwinding, gọi toàn bộ Defer và chỉ dừng chương trình khi không còn Function nào có khả năng xử lý lỗi.

---

# 11.19 Recover - Khôi phục từ Panic

`recover()` là cơ chế duy nhất có thể dừng quá trình Panic.

Ví dụ.

```go
defer func() {

    recover()

}()
```

Nếu gọi đúng thời điểm.

Runtime sẽ.

```
Stop Unwinding
```

---

# Recover hoạt động như thế nào?

Runtime kiểm tra.

```
Current _panic
```

Nếu tồn tại.

Runtime đánh dấu.

```
Recovered
```

Sau đó.

```
Stop Panic

↓

Continue Return
```

---

# Recover chỉ hoạt động trong Defer

Ví dụ.

Sai.

```go
recover()
```

Đúng.

```go
defer func() {

    recover()

}()
```

Đây là quy tắc của Runtime.

---

# Vì sao?

Khi Panic.

Runtime đang duyệt.

```
_defer
```

Chỉ trong thời điểm này.

Recover mới có quyền dừng quá trình Unwind.

Ngoài thời điểm đó.

Recover luôn trả về.

```
nil
```

---

# Sau Recover

Giả sử.

```text
main()

↓

A()

↓

panic()

↓

recover()
```

Runtime.

```
Stop Unwind

↓

Return bình thường
```

Chương trình tiếp tục chạy.

---

# Những điều cần ghi nhớ

- Recover chỉ hoạt động trong Defer.
- Recover dừng Stack Unwinding.
- Recover đọc `_panic`.
- Nếu không có Panic.
- Recover trả về `nil`.

---

# Tóm tắt

Recover là cơ chế duy nhất cho phép chương trình khôi phục sau Panic.

Runtime chỉ cho phép Recover hoạt động trong Defer nhằm đảm bảo Stack vẫn còn ở trạng thái nhất quán trong quá trình Stack Unwinding.

---

# 11.20 Goroutine Exit - Điều gì xảy ra khi Goroutine kết thúc?

Cuối cùng.

Một Goroutine sẽ kết thúc.

Ví dụ.

```go
func worker() {

    fmt.Println("Done")

}
```

Sau khi Function cuối cùng Return.

Runtime không đơn giản chỉ xóa Goroutine.

Nó còn phải thực hiện rất nhiều công việc.

---

# Bước 1 - Execute Remaining Defer

Nếu còn.

```
_defer
```

Runtime sẽ thực thi toàn bộ.

---

# Bước 2 - Xóa Panic

Nếu Goroutine đang Panic.

Runtime giải phóng.

```
_panic
```

---

# Bước 3 - Đổi trạng thái

Runtime cập nhật.

```text
Running

↓

_Gdead
```

Scheduler sẽ không bao giờ chọn Goroutine này nữa.

---

# Bước 4 - Thu hồi Stack

Runtime đánh dấu.

```
Stack

↓

Free
```

Stack có thể được.

- Trả về Pool.
- Hoặc giải phóng.

Tùy Runtime quyết định.

---

# Bước 5 - Thu hồi struct g

`struct g` cũng không bị hủy ngay.

Thông thường.

Runtime đưa nó vào.

```
gFree

Pool
```

để tái sử dụng.

Điều này giảm Allocation.

---

# Bước 6 - Scheduler

Scheduler phát hiện.

```
Goroutine Dead
```

Tiếp tục chọn.

```
Goroutine khác
```

---

# Minh họa toàn bộ quá trình

```text
Function Return

↓

Execute Defer

↓

Cleanup Panic

↓

Status = _Gdead

↓

Release Stack

↓

Recycle struct g

↓

Scheduler Run Next G
```

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Goroutine kết thúc thì `struct g` bị hủy ngay.

Sai.

Runtime thường đưa `struct g` vào pool để tái sử dụng.

---

### Hiểu lầm 2

> Stack luôn được giải phóng ngay.

Không hoàn toàn đúng.

Runtime có thể giữ lại trong các pool nội bộ nhằm giảm chi phí cấp phát ở những lần tạo Goroutine tiếp theo.

---

### Hiểu lầm 3

> Scheduler xóa Goroutine.

Sai.

Scheduler chỉ ngừng lập lịch cho Goroutine.

Việc dọn dẹp do Runtime thực hiện.

---

# Tổng kết Phần V - Goroutine Runtime Lifecycle

Sau khi hoàn thành các Section từ **11.17** đến **11.20**, chúng ta đã hiểu toàn bộ vòng đời của một Goroutine ở mức Runtime.

```text
Create Goroutine

↓

Allocate Stack

↓

Running

↓

Function Call

↓

Defer

↓

Panic?

↓

Recover?

↓

Return

↓

Execute Remaining Defer

↓

_Gdead

↓

Recycle Stack

↓

Recycle struct g
```

Đây là toàn bộ chu trình sống của một Goroutine, từ khi được tạo ra cho đến khi được Runtime thu hồi.

---

# Chương tiếp theo

Đến thời điểm này, chúng ta đã hiểu gần như toàn bộ cấu trúc nội tại của Goroutine.

Tuy nhiên, vẫn còn một thành phần quan trọng chưa được khám phá:

> **Machine (`struct m`)** - đại diện cho OS Thread trong Go Runtime.

Ở các phần tiếp theo của Chapter 11, chúng ta sẽ chuyển từ **Goroutine (`g`)** sang **Machine (`m`)** và **Processor (`p`)**, hai thành phần còn lại trong mô hình **G-M-P**. Hiểu ba cấu trúc dữ liệu này sẽ giúp chúng ta nắm được toàn bộ cơ chế hoạt động của Scheduler và chuẩn bị cho các chương chuyên sâu về lập lịch của Go Runtime.

---

---

# Phần VI - Scheduler Interaction

Trong các phần trước của Chapter 11, chúng ta đã nghiên cứu Goroutine từ góc nhìn nội tại.

Chúng ta đã hiểu:

- `struct g`
- `gobuf`
- Stack
- Stack Frame
- Dynamic Stack
- Context Switching
- Defer
- Panic
- Recover
- Goroutine Exit

Tuy nhiên, tất cả những kiến thức đó mới chỉ trả lời câu hỏi:

> **Bên trong một Goroutine có những gì?**

Một câu hỏi quan trọng khác vẫn còn bỏ ngỏ.

> **Scheduler nhìn Goroutine như thế nào?**

Đối với Scheduler.

Điều quan trọng nhất không phải là:

- Stack.
- Program Counter.
- Local Variables.

Mà là:

```
Trạng thái hiện tại của Goroutine.
```

Chỉ khi biết Goroutine đang ở trạng thái nào.

Scheduler mới có thể quyết định:

- Có nên chạy nó hay không.
- Có nên đánh thức nó hay không.
- Có nên đưa nó vào Run Queue hay không.
- Có nên chuyển nó sang Machine khác hay không.

Đó chính là lý do Go Runtime xây dựng một **State Machine** dành riêng cho Goroutine.

Đây là nền tảng của toàn bộ Scheduler.

---

# 11.21 Goroutine States

Trong Chapter 10, chúng ta đã từng giới thiệu sơ lược các trạng thái của Goroutine.

Tuy nhiên.

Đó chỉ là góc nhìn ở mức khái niệm.

Trong phần này, chúng ta sẽ nghiên cứu:

- Các trạng thái thật sự trong Go Runtime.
- Giá trị của `atomicstatus`.
- Khi nào Runtime chuyển trạng thái.
- Thành phần nào chịu trách nhiệm chuyển trạng thái.
- Vì sao việc chuyển trạng thái phải được thực hiện bằng các thao tác Atomic.

Sau phần này, bạn sẽ có thể đọc và hiểu phần lớn mã nguồn trong:

```
runtime/proc.go
```

và.

```
runtime/runtime2.go
```

---

# State Machine của Goroutine

Có thể mô tả vòng đời của Goroutine như sau.

```text
                   New

                    │

                    ▼

              Runnable

                    │

                    ▼

               Running

      ┌─────────┼─────────┐

      │         │         │

      ▼         ▼         ▼

 Waiting   Syscall      Dead

      │         │

      ▼         ▼

   Runnable  Runnable
```

Trong Runtime.

Mỗi lần chuyển trạng thái đều được kiểm soát rất chặt chẽ.

Không có việc thay đổi trạng thái một cách tùy ý.

---

# Runtime lưu trạng thái ở đâu?

Trong `struct g`.

Có một trường rất quan trọng.

```go
atomicstatus
```

Đây là nơi Runtime lưu trạng thái hiện tại của Goroutine.

Ví dụ.

```go
type g struct {

    ...

    atomicstatus uint32

    ...

}
```

Scheduler gần như luôn đọc trường này trước khi quyết định xử lý một Goroutine.

---

# Tại sao gọi là atomicstatus?

Nhiều thành phần trong Runtime có thể cùng truy cập trạng thái của Goroutine.

Ví dụ.

- Scheduler.
- Netpoller.
- Garbage Collector.
- Channel.
- Timer.
- Syscall Handler.

Nếu nhiều Thread cùng cập nhật trạng thái mà không đồng bộ.

Có thể xảy ra.

```text
Race Condition
```

Ví dụ.

Thread A.

```
Runnable

↓

Running
```

Trong khi.

Thread B.

```
Runnable

↓

Waiting
```

Nếu không dùng Atomic Operation.

Kết quả cuối cùng sẽ không xác định.

Do đó.

Runtime sử dụng:

```
Atomic Compare-And-Swap (CAS)
```

để đảm bảo mỗi lần chuyển trạng thái đều an toàn.

---

# Các trạng thái chính

Trong Go Runtime.

Có khá nhiều trạng thái nội bộ.

Tuy nhiên.

Có một số trạng thái quan trọng nhất.

| Trạng thái | Ý nghĩa |
|------------|----------|
| `_Grunnable` | Sẵn sàng chạy |
| `_Grunning` | Đang chạy trên CPU |
| `_Gwaiting` | Đang chờ một sự kiện |
| `_Gsyscall` | Đang thực hiện System Call |
| `_Gdead` | Đã kết thúc |
| `_Gcopystack` | Đang sao chép Stack |

Các trạng thái khác chủ yếu phục vụ Garbage Collector và Debugger.

---

# `_Grunnable`

Đây là trạng thái mà Scheduler yêu thích nhất.

Một Goroutine ở trạng thái.

```
_Grunnable
```

có nghĩa là.

> **Đã sẵn sàng chạy nhưng chưa được cấp CPU.**

Có thể hình dung.

```text
Run Queue

↓

G1

↓

G2

↓

G3
```

Tất cả đều ở trạng thái.

```
Runnable
```

Scheduler chỉ cần chọn một Goroutine bất kỳ.

---

# `_Grunning`

Khi Scheduler gán Goroutine cho.

- Machine.
- Processor.

Runtime cập nhật.

```text
Runnable

↓

Running
```

Lúc này.

CPU bắt đầu thực thi các Instruction của Goroutine.

Một Goroutine chỉ có thể ở trạng thái Running nếu:

- Có một `P`.
- Có một `M`.
- Có CPU Core để thực thi.

---

# `_Gwaiting`

Đây là trạng thái phổ biến nhất sau Runnable.

Ví dụ.

```go
<-ch
```

```go
mutex.Lock()
```

```go
time.Sleep()
```

```go
cond.Wait()
```

Lúc này.

Runtime chuyển.

```text
Running

↓

Waiting
```

Điều quan trọng.

Một Goroutine Waiting:

- Không chiếm CPU.
- Không nằm trong Run Queue.
- Không được Scheduler chọn.

---

# `_Gsyscall`

Đây là trạng thái đặc biệt.

Ví dụ.

```go
syscall.Read(...)
```

Khi thực hiện System Call.

OS Thread có thể bị Block.

Runtime cập nhật.

```text
Running

↓

Syscall
```

Sau đó.

Processor (`P`) được tách khỏi Thread (`M`) để chạy Goroutine khác.

Đây là một trong những thiết kế quan trọng giúp Go tránh việc toàn bộ Scheduler bị chặn bởi một lời gọi hệ thống.

---

# `_Gdead`

Sau khi Function cuối cùng Return.

Runtime thực hiện.

```text
Running

↓

_Gdead
```

Từ thời điểm này.

Scheduler sẽ không bao giờ chọn Goroutine đó nữa.

Sau đó Runtime bắt đầu.

- Thu hồi Stack.
- Thu hồi `_defer`.
- Thu hồi `_panic`.
- Đưa `struct g` vào pool.

---

# `_Gcopystack`

Đây là trạng thái ít gặp.

Nó xuất hiện trong quá trình.

```
Stack Growth
```

hoặc.

```
Stack Shrink
```

Runtime đánh dấu Goroutine đang được sao chép Stack.

Trong trạng thái này.

Không thành phần nào khác được phép truy cập Stack của Goroutine.

Điều này đảm bảo tính nhất quán của dữ liệu.

---

# Chuyển đổi trạng thái

Một Goroutine không thể chuyển sang bất kỳ trạng thái nào tùy ý.

Ví dụ.

```text
Runnable

↓

Running
```

là hợp lệ.

Nhưng.

```text
Dead

↓

Running
```

là không hợp lệ.

Hay.

```text
Waiting

↓

Dead
```

thường cũng không xảy ra trực tiếp.

Runtime luôn tuân theo một tập các quy tắc chuyển trạng thái.

---

# Ví dụ với Channel

```go
value := <-ch
```

Quá trình.

```text
Runnable

↓

Running

↓

Waiting
```

Khi một Goroutine khác gửi dữ liệu.

```go
ch <- 100
```

Runtime.

```text
Waiting

↓

Runnable
```

Scheduler sau đó.

```text
Runnable

↓

Running
```

---

# Ví dụ với Mutex

```go
mutex.Lock()
```

Nếu Mutex đang bị giữ.

```text
Running

↓

Waiting
```

Khi.

```go
mutex.Unlock()
```

Runtime đánh thức Goroutine.

```text
Waiting

↓

Runnable
```

---

# Ví dụ với Time Sleep

```go
time.Sleep(time.Second)
```

Runtime.

```text
Running

↓

Waiting
```

Sau một giây.

Timer Manager.

```text
Waiting

↓

Runnable
```

Scheduler tiếp tục.

```text
Running
```

---

# Ví dụ với System Call

```go
file.Read(...)
```

Runtime.

```text
Running

↓

Syscall
```

Sau khi Kernel trả kết quả.

```text
Syscall

↓

Runnable
```

Scheduler tiếp tục thực thi.

---

# Ai thay đổi trạng thái?

Một điểm rất quan trọng.

Scheduler không phải là thành phần duy nhất thay đổi trạng thái Goroutine.

Các thành phần khác cũng có quyền.

| Thành phần | Thay đổi trạng thái |
|------------|--------------------|
| Scheduler | Runnable ↔ Running |
| Channel | Waiting → Runnable |
| Mutex | Waiting → Runnable |
| Timer | Waiting → Runnable |
| Netpoller | Waiting → Runnable |
| Syscall Handler | Syscall → Runnable |
| Runtime | Running → Dead |

Điều này giải thích vì sao `atomicstatus` phải là Atomic Variable.

---

# Atomic Compare-And-Swap (CAS)

Giả sử.

Hai Thread cùng muốn đánh thức một Goroutine.

Nếu cả hai đều thực hiện.

```text
Waiting

↓

Runnable
```

đồng thời.

Goroutine có thể bị đưa vào Run Queue hai lần.

Đây là lỗi nghiêm trọng.

Do đó.

Runtime sử dụng.

```text
CAS

↓

Success?

↓

Update

↓

Otherwise Retry
```

để đảm bảo chỉ có một Thread thay đổi trạng thái thành công.

---

# Góc nhìn tổng thể

Toàn bộ vòng đời của Goroutine có thể mô tả như sau.

```text
               New

                │

                ▼

          _Grunnable

                │

         Scheduler

                │

                ▼

          _Grunning

      ┌────────┼────────┐

      │        │        │

      ▼        ▼        ▼

_Gwaiting _Gsyscall _Gdead

      │        │

      ▼        ▼

 _Grunnable _Grunnable
```

Đây chính là State Machine của Goroutine.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Scheduler là thành phần duy nhất thay đổi trạng thái Goroutine.

Sai.

Channel, Timer, Netpoller, Syscall Handler và Garbage Collector cũng có thể thay đổi trạng thái trong những trường hợp cụ thể.

---

### Hiểu lầm 2

> Waiting Goroutine vẫn nằm trong Run Queue.

Sai.

Waiting Goroutine đã rời khỏi Run Queue.

Nó chỉ quay trở lại khi Runtime đánh thức nó.

---

### Hiểu lầm 3

> Dead Goroutine có thể chạy lại.

Sai.

Sau khi chuyển sang `_Gdead`, Goroutine kết thúc vĩnh viễn.

Nếu muốn thực hiện công việc mới, Runtime phải tạo một Goroutine mới.

---

### Hiểu lầm 4

> Chuyển trạng thái chỉ là phép gán biến.

Sai.

Runtime sử dụng các thao tác Atomic (CAS) để đảm bảo tính đúng đắn trong môi trường đa luồng.

---

# Những điều cần ghi nhớ

- Trạng thái Goroutine được lưu trong `atomicstatus`.
- Scheduler quyết định chủ yếu dựa trên trạng thái này.
- `_Grunnable` là trạng thái sẵn sàng chạy.
- `_Grunning` là trạng thái đang sử dụng CPU.
- `_Gwaiting` không tiêu thụ CPU.
- `_Gsyscall` đại diện cho Goroutine đang thực hiện lời gọi hệ thống.
- `_Gdead` là trạng thái cuối cùng.
- Mọi thay đổi trạng thái đều phải an toàn với nhiều Thread.

---

# Tóm tắt

Goroutine States là nền tảng của toàn bộ Go Scheduler.

Thông qua trường `atomicstatus`, Runtime luôn biết chính xác mỗi Goroutine đang ở trạng thái nào và có thể quyết định liệu Goroutine đó có được phép chạy, cần chờ một sự kiện hay đã kết thúc.

Hiểu rõ State Machine của Goroutine là bước quan trọng để đọc và phân tích mã nguồn của Scheduler trong `runtime/proc.go`, đồng thời giải thích được cách Go Runtime phối hợp giữa Scheduler, Netpoller, Timer, Channel và Garbage Collector.

Trong phần tiếp theo, chúng ta sẽ tìm hiểu cơ chế làm cho các trạng thái này thay đổi:

> **11.22 - Goroutine State Transitions**

Ở phần đó, chúng ta sẽ phân tích từng hàm trong Runtime như `casgstatus()`, `gopark()`, `goready()` và cách chúng phối hợp để chuyển Goroutine giữa các trạng thái một cách an toàn và hiệu quả.

---

---

# 11.22 Waiting Queue

Trong phần trước, chúng ta đã biết một Goroutine có thể chuyển sang trạng thái:

```
_Gwaiting
```

Khi đó.

Scheduler sẽ không còn chọn Goroutine này nữa.

Một câu hỏi rất quan trọng xuất hiện.

> **Nếu Goroutine không còn nằm trong Run Queue thì Runtime lưu nó ở đâu?**

Câu trả lời là:

```
Waiting Queue
```

Có thể nói.

Run Queue là nơi dành cho các Goroutine **có thể chạy**.

Trong khi.

Waiting Queue là nơi dành cho các Goroutine **đang chờ một điều kiện nào đó**.

---

# Run Queue và Waiting Queue

Có thể hình dung.

```text
                Scheduler

                     │

        ┌────────────┴────────────┐

        ▼                         ▼

   Run Queue                Waiting Queue

Runnable G              Waiting G
```

Một Goroutine chỉ có thể nằm ở **một Queue** tại một thời điểm.

Nó không bao giờ đồng thời tồn tại ở cả hai.

---

# Vì sao cần Waiting Queue?

Giả sử.

```go
<-ch
```

Nếu Channel chưa có dữ liệu.

Runtime có hai lựa chọn.

Lựa chọn thứ nhất.

```text
while(true){

    check channel

}
```

Đây gọi là.

```
Busy Waiting
```

CPU sẽ luôn bận.

Hiệu năng rất thấp.

Go Runtime không làm như vậy.

Thay vào đó.

Runtime đưa Goroutine vào Waiting Queue.

CPU lập tức được sử dụng để chạy Goroutine khác.

---

# Channel Waiting Queue

Ví dụ.

```go
value := <-ch
```

Nếu Channel rỗng.

Runtime thực hiện.

```text
Running

↓

_Gwaiting

↓

Channel Receive Queue
```

Có thể hình dung.

```text
Channel

+--------------------+

Receive Queue

↓

G1

↓

G2

↓

G3
```

Khi có Goroutine khác.

```go
ch <- 100
```

Runtime lấy.

```
G1
```

ra khỏi Queue.

```text
Receive Queue

↓

G2

↓

G3
```

Sau đó.

```
G1

↓

Runnable
```

Scheduler sẽ chạy G1.

---

# Send Queue

Nếu Channel đầy.

```go
ch <- value
```

Runtime cũng đưa Goroutine vào Queue.

```text
Channel

Send Queue

↓

G5

↓

G6
```

Sau khi Channel có chỗ.

Runtime đánh thức Goroutine đầu tiên.

---

# Mutex Waiting Queue

Ví dụ.

```go
mutex.Lock()
```

Nếu Mutex đã bị giữ.

Runtime.

```text
Running

↓

Waiting
```

Sau đó.

```text
Mutex Wait Queue

↓

G7

↓

G8
```

Khi.

```go
mutex.Unlock()
```

Runtime.

```text
Wake G7

↓

Runnable
```

---

# Sleep Waiting Queue

Ví dụ.

```go
time.Sleep(time.Second)
```

Runtime tạo.

```
Timer
```

Sau đó.

```text
Running

↓

Waiting

↓

Timer Queue
```

Sau một giây.

Timer Manager.

```text
Wake G

↓

Runnable
```

---

# Waiting Queue không thuộc Scheduler

Một hiểu lầm phổ biến.

Nhiều người nghĩ.

Scheduler quản lý Waiting Queue.

Thực tế.

Mỗi cơ chế có Queue riêng.

Ví dụ.

```text
Channel

↓

Channel Queue
```

```text
Mutex

↓

Mutex Queue
```

```text
Timer

↓

Timer Queue
```

```text
Netpoll

↓

Poll Queue
```

Scheduler chỉ quan tâm.

```
Runnable
```

---

# Tổng kết Waiting Queue

Có thể mô tả.

```text
Running

↓

Waiting

↓

Waiting Queue

↓

Wakeup

↓

Runnable

↓

Scheduler
```

Đây là vòng đời phổ biến nhất của Goroutine.

---

# 11.23 Park / Unpark

Sau khi hiểu Waiting Queue.

Một câu hỏi khác xuất hiện.

> Runtime đưa Goroutine vào Waiting bằng cách nào?

Runtime sử dụng hai hàm rất quan trọng.

```
gopark()
```

và.

```
goready()
```

Nhiều tài liệu gọi chúng là.

```
Park

↓

Unpark
```

---

# Park là gì?

Park có nghĩa.

```
Suspend Current Goroutine
```

Có thể mô tả.

```text
Running

↓

Save Context

↓

Waiting

↓

Scheduler Run Next
```

Runtime thực hiện việc này thông qua.

```
gopark()
```

---

# Ví dụ với Channel

```go
<-ch
```

Runtime.

```text
Receive?

↓

No

↓

gopark()

↓

Waiting
```

CPU lập tức chuyển sang Goroutine khác.

---

# Ví dụ với Mutex

```go
mutex.Lock()
```

Nếu Lock thất bại.

Runtime.

```text
gopark()

↓

Mutex Queue
```

---

# Ví dụ với Sleep

```go
time.Sleep()
```

Runtime.

```text
Create Timer

↓

gopark()
```

---

# Điều gì xảy ra trong gopark()?

Có thể mô tả.

```text
Save Registers

↓

Update Status

↓

Put vào Waiting Queue

↓

Call Scheduler
```

Sau khi.

```
gopark()
```

trả về.

CPU đã chạy Goroutine khác.

---

# Unpark là gì?

Ngược lại.

Runtime sử dụng.

```
goready()
```

để đánh thức Goroutine.

Có thể mô tả.

```text
Waiting

↓

Runnable

↓

Run Queue
```

Lưu ý.

```
goready()

≠

Run Immediately
```

Nó chỉ đưa Goroutine trở lại Run Queue.

Scheduler quyết định khi nào Goroutine thực sự chạy.

---

# Ví dụ

```text
G1

↓

Waiting
```

Channel nhận dữ liệu.

```text
goready(G1)

↓

Run Queue
```

Một lúc sau.

```text
Scheduler

↓

Running
```

---

# Park không phải Sleep

Nhiều người nhầm.

```
Park

=

Sleep
```

Sai.

Sleep chỉ là.

```
Một lý do để Park.
```

Channel.

Mutex.

Cond.

Netpoll.

Đều sử dụng.

```
gopark()
```

---

# Park và Context Switching

Mỗi lần.

```
gopark()
```

Runtime.

```text
Save Context

↓

Switch
```

Do đó.

Park luôn kéo theo Context Switching.

---

# Tổng kết Park / Unpark

```text
Running

↓

gopark()

↓

Waiting

↓

Event

↓

goready()

↓

Runnable

↓

Scheduler

↓

Running
```

Đây là vòng đời phổ biến nhất của Goroutine.

---

# 11.24 Preemption - Từ góc nhìn Goroutine

Cho đến bây giờ.

Chúng ta mới nghiên cứu các trường hợp.

```
Goroutine

↓

Tự nhường CPU
```

Ví dụ.

- Channel.
- Mutex.
- Sleep.

Đây gọi là.

```
Cooperative Scheduling
```

Nhưng nếu.

```go
for {

}
```

thì sao?

Goroutine không:

- Sleep.
- Lock.
- Receive.
- Yield.

Liệu nó có giữ CPU mãi mãi?

Câu trả lời là.

```
Không.
```

---

# Preemption là gì?

Preemption là khả năng.

> Runtime chủ động lấy CPU từ một Goroutine.

Không cần Goroutine đồng ý.

Có thể mô tả.

```text
Running

↓

Runtime Interrupt

↓

Save Context

↓

Runnable
```

Scheduler tiếp tục.

```
Run Goroutine khác
```

---

# Vì sao cần Preemption?

Giả sử.

```go
func worker() {

    for {

    }

}
```

Nếu không có Preemption.

```text
CPU

↓

Worker Forever
```

Toàn bộ Goroutine khác.

```
Starvation
```

Không bao giờ được chạy.

---

# Cooperative vs Preemptive

```text
Cooperative

↓

Goroutine tự nhường CPU
```

Ví dụ.

- Channel
- Mutex
- Sleep
- Gosched

---

```text
Preemptive

↓

Runtime lấy CPU
```

Ví dụ.

- Time Slice hết.
- Asynchronous Preemption.

---

# Quá trình Preemption

Có thể mô tả.

```text
Running

↓

Need Preempt?

↓

Yes

↓

Save Context

↓

Runnable

↓

Scheduler

↓

Run Next
```

Sau này.

Scheduler.

```text
Runnable

↓

Running
```

Goroutine tiếp tục.

---

# Goroutine có biết mình bị Preempt không?

Không.

Ví dụ.

```go
x++

y++
```

Runtime có thể.

```text
Preempt

↓

Run G2

↓

Restore

↓

Continue
```

Đối với lập trình viên.

Không có khác biệt.

---

# Asynchronous Preemption

Từ Go 1.14.

Runtime hỗ trợ.

```
Asynchronous Preemption
```

Điều này có nghĩa.

Runtime có thể.

```
Interrupt
```

một Goroutine ngay cả khi nó không gọi.

- Channel.
- Sleep.
- Lock.

Đây là một bước tiến rất lớn của Go Runtime.

---

# Safe Point

Runtime không thể Preempt ở mọi Instruction.

Ví dụ.

Trong lúc.

```
Move Stack
```

Hoặc.

```
Update Pointer
```

Preempt sẽ rất nguy hiểm.

Do đó.

Runtime chỉ Preempt tại.

```
Safe Point
```

Đây là vị trí mà.

- Stack nhất quán.
- Register nhất quán.
- Garbage Collector an toàn.

---

# Góc nhìn Goroutine

Đối với một Goroutine.

Toàn bộ vòng đời có thể mô tả.

```text
Running

↓

Preempt

↓

Runnable

↓

Run Queue

↓

Scheduler

↓

Restore Context

↓

Running
```

Điều thú vị.

Goroutine không hề biết.

Nó từng bị dừng.

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Waiting Queue do Scheduler quản lý.

Sai.

Mỗi thành phần như Channel, Mutex, Timer hay Netpoll đều có Waiting Queue riêng.

---

### Hiểu lầm 2

> `goready()` làm Goroutine chạy ngay.

Sai.

`goready()` chỉ đưa Goroutine về trạng thái Runnable.

Scheduler mới quyết định thời điểm chạy.

---

### Hiểu lầm 3

> Park và Sleep là một.

Sai.

Sleep chỉ là một trường hợp sử dụng `gopark()`.

---

### Hiểu lầm 4

> Preemption chỉ xảy ra khi Goroutine gọi `runtime.Gosched()`.

Sai.

Từ Go 1.14, Runtime hỗ trợ Asynchronous Preemption và có thể chủ động tạm dừng Goroutine tại Safe Point.

---

# Tổng kết Phần VI

Sau các phần từ **11.21** đến **11.24**, chúng ta đã hiểu Goroutine dưới góc nhìn của Scheduler.

```text
Runnable

↓

Running

↓

Waiting

↓

Waiting Queue

↓

goready()

↓

Runnable

↓

Scheduler

↓

Running

↓

Preemption

↓

Runnable
```

Đây chính là vòng đời thực tế của hầu hết Goroutine trong Go Runtime.

Đến thời điểm này, chúng ta đã hoàn thành gần như toàn bộ kiến thức về **Goroutine (`struct g`)**.

Ở phần tiếp theo, chúng ta sẽ chuyển sang thành phần thứ hai của mô hình **G-M-P**:

> **Machine (`struct m`)** - đại diện cho OS Thread trong Go Runtime.

Đây là nơi chúng ta sẽ tìm hiểu cách Go Runtime quản lý Thread, System Call, Thread Parking và mối quan hệ giữa Goroutine với OS Thread.

---

---

# Phần VII - Debugging Goroutine

Cho đến thời điểm này, chúng ta đã nghiên cứu gần như toàn bộ vòng đời của một Goroutine.

Chúng ta đã hiểu:

- `struct g`
- `gobuf`
- Stack
- Stack Frame
- Context Switching
- Dynamic Stack
- Scheduler States
- Waiting Queue
- Park / Unpark
- Preemption

Đây là góc nhìn từ bên trong Go Runtime.

Tuy nhiên.

Trong thực tế phát triển phần mềm.

Lập trình viên hiếm khi đọc trực tiếp `runtime/proc.go`.

Thay vào đó.

Họ thường gặp những vấn đề như.

- Chương trình ngày càng chậm.
- RAM liên tục tăng.
- Hàng chục nghìn Goroutine không biến mất.
- CPU thấp nhưng chương trình không phản hồi.
- Không biết Goroutine đang bị kẹt ở đâu.

Lúc này.

Điều quan trọng không còn là biết Scheduler hoạt động như thế nào.

Mà là:

> **Làm thế nào để quan sát Goroutine khi chương trình đang chạy?**

Đó chính là mục tiêu của phần này.

---

# 11.25 Goroutine Leak

Một trong những lỗi phổ biến nhất khi lập trình Go là:

```
Goroutine Leak
```

Đây cũng là một trong những câu hỏi xuất hiện rất nhiều trong các buổi phỏng vấn Senior Go.

---

# Goroutine Leak là gì?

Một Goroutine Leak xảy ra khi.

> **Một Goroutine vẫn còn tồn tại nhưng sẽ không bao giờ hoàn thành công việc của mình.**

Điều này khác với Memory Leak.

Trong Go.

Garbage Collector có thể giải phóng Object.

Nhưng.

Garbage Collector **không thể giải phóng Goroutine đang còn sống**.

Nếu Goroutine chưa kết thúc.

Runtime vẫn phải giữ lại.

- Stack.
- `struct g`.
- `_defer`.
- `_panic`.
- Metadata.

---

# Một ví dụ đơn giản

```go
func worker(ch chan int) {

    <-ch

}
```

```go
func main() {

    ch := make(chan int)

    go worker(ch)

}
```

Ở đây.

Không có Goroutine nào gửi dữ liệu.

```
worker()
```

sẽ chờ mãi mãi.

```text
Running

↓

Waiting

↓

Never Wake Up
```

Đây chính là Goroutine Leak.

---

# Leak khác Deadlock như thế nào?

Đây là câu hỏi rất hay.

## Deadlock

```text
Tất cả Goroutine

↓

Waiting
```

Chương trình dừng hoàn toàn.

Runtime sẽ Panic.

```
fatal error:

all goroutines are asleep
```

---

## Goroutine Leak

```text
Một vài Goroutine

↓

Waiting Forever
```

Trong khi.

Những Goroutine khác vẫn chạy bình thường.

Chương trình không Crash.

Nhưng RAM sẽ tăng dần.

---

# Những nguyên nhân phổ biến

## Channel

```go
<-ch
```

Không ai gửi.

---

## Send Channel

```go
ch <- value
```

Không ai Receive.

---

## Mutex

```go
mutex.Lock()
```

Không bao giờ Unlock.

---

## WaitGroup

```go
wg.Wait()
```

Không bao giờ Done.

---

## Context

```go
<-ctx.Done()
```

Context không bao giờ Cancel.

---

## Infinite Loop

```go
for {

}
```

Không bao giờ Exit.

---

# Vì sao Leak nguy hiểm?

Một Goroutine bị Leak vẫn giữ lại.

- Stack.
- Pointer.
- Heap Reference.

Garbage Collector sẽ coi toàn bộ Object mà Goroutine tham chiếu là:

```
Live Object
```

Kết quả.

```
RAM

↑

GC

↑

Latency

↑
```

---

# Cách phát hiện Leak

Có nhiều cách.

Ví dụ.

```go
runtime.NumGoroutine()
```

Nếu số lượng Goroutine liên tục tăng.

Có thể chương trình đang bị Leak.

---

# Nguyên tắc tránh Leak

- Luôn đóng Channel khi phù hợp.
- Luôn có Timeout.
- Luôn dùng Context.
- Không tạo Goroutine mà không có điều kiện kết thúc.
- Không bỏ quên `wg.Done()`.

---

# 11.26 Goroutine Dump

Khi nghi ngờ Goroutine Leak.

Điều đầu tiên lập trình viên thường làm là.

```
Dump toàn bộ Goroutine.
```

---

# Goroutine Dump là gì?

Goroutine Dump là ảnh chụp trạng thái của tất cả Goroutine tại một thời điểm.

Ví dụ.

```text
goroutine 1 [running]

goroutine 18 [chan receive]

goroutine 32 [IO wait]

goroutine 51 [sleep]
```

Mỗi Goroutine đều cho biết.

- ID.
- State.
- Stack Trace.

---

# Ví dụ

```go
pprof.Lookup("goroutine").WriteTo(...)
```

Hoặc.

```go
runtime.Stack(...)
```

Runtime sẽ in.

```text
goroutine 1 [running]:

main.main()

goroutine 18 [chan receive]:

worker()
```

---

# Ý nghĩa của State

Ví dụ.

```
running
```

Đang chạy.

---

```
chan receive
```

Đang chờ Channel.

---

```
IO wait
```

Đang chờ Network.

---

```
sleep
```

Đang Sleep.

---

```
semacquire
```

Đang chờ Mutex hoặc Semaphore.

---

# Stack Trace

Ví dụ.

```text
goroutine 18

↓

worker()

↓

receive()

↓

runtime.chanrecv()
```

Chỉ cần nhìn Stack Trace.

Chúng ta biết ngay.

```
Channel Receive
```

đang chặn Goroutine.

---

# Khi nào dùng Goroutine Dump?

Ví dụ.

- Server treo.
- Leak.
- Deadlock.
- CPU thấp nhưng Request Timeout.
- Không biết Goroutine đang ở đâu.

Đây là công cụ Debug mạnh nhất của Go Runtime.

---

# Đọc Goroutine Dump

Giả sử.

```text
goroutine 81 [chan receive]

worker()

↓

process()

↓

main()
```

Có thể kết luận.

```
worker()

↓

Waiting Channel
```

Không cần đọc toàn bộ chương trình.

---

# Những điều cần chú ý

Nếu thấy.

```
10000

goroutines

[chan receive]
```

Rất có thể.

Có Goroutine Leak.

---

# 11.27 Trace

Nếu Goroutine Dump giống như:

```
Ảnh chụp.
```

Thì.

```
Trace
```

giống như.

```
Video.
```

---

# Trace là gì?

Trace ghi lại toàn bộ hoạt động của Runtime.

Ví dụ.

- Scheduler.
- Goroutine.
- Channel.
- Network.
- Garbage Collector.
- Syscall.

Theo thời gian.

---

# Trace ghi lại điều gì?

Ví dụ.

```text
G1

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

Hoặc.

```text
P0

↓

Run G10

↓

Run G15

↓

Run G3
```

Toàn bộ quá trình đều được ghi lại.

---

# Runtime Trace

Go cung cấp package.

```go
runtime/trace
```

Ví dụ.

```go
trace.Start(file)

...

trace.Stop()
```

Sau đó.

Có thể xem bằng.

```
go tool trace
```

---

# Có thể quan sát gì?

Ví dụ.

## Scheduler Timeline

```text
Time

↓

P0

↓

G1

↓

G2

↓

G3
```

---

## Network

```text
Read

↓

Waiting

↓

Wakeup
```

---

## Channel

```text
Send

↓

Receive

↓

Wake
```

---

## Garbage Collector

```text
Mark

↓

Scan

↓

Sweep
```

---

## Syscall

```text
Enter

↓

Block

↓

Return
```

---

# Vì sao Trace quan trọng?

Profiler trả lời.

```
Tốn CPU ở đâu?
```

Memory Profiler trả lời.

```
RAM ở đâu?
```

Trace trả lời.

```
Runtime đang làm gì theo thời gian.
```

Đây là điểm khác biệt rất lớn.

---

# Trace và Scheduler

Trace cho phép quan sát.

- Goroutine nào đang chạy.
- Goroutine nào bị Park.
- Khi nào Preemption xảy ra.
- Machine nào đang chạy.
- Processor nào Idle.

Đây là công cụ gần nhất với việc "nhìn vào Scheduler".

---

# Trace và Production

Trace rất mạnh.

Nhưng cũng có chi phí.

Do đó.

Thông thường.

- Chỉ bật khi cần Debug.
- Hoặc thu thập trong thời gian ngắn.

Không nên bật liên tục trên hệ thống tải cao nếu không có lý do cụ thể.

---

# So sánh ba công cụ

| Công cụ | Mục đích |
|----------|----------|
| `runtime.NumGoroutine()` | Đếm số Goroutine |
| Goroutine Dump | Chụp trạng thái hiện tại |
| Runtime Trace | Theo dõi toàn bộ quá trình theo thời gian |

---

# Những hiểu lầm phổ biến

### Hiểu lầm 1

> Goroutine Leak là Memory Leak.

Sai.

Memory Leak liên quan đến Object.

Goroutine Leak liên quan đến Goroutine không bao giờ kết thúc.

---

### Hiểu lầm 2

> Goroutine Dump chỉ dùng khi Deadlock.

Sai.

Dump hữu ích cho cả Leak, Performance, Timeout và Debug thông thường.

---

### Hiểu lầm 3

> Trace thay thế Profiler.

Sai.

Trace và Profiler bổ sung cho nhau.

Profiler tập trung vào CPU hoặc Memory.

Trace tập trung vào hành vi của Runtime theo thời gian.

---

### Hiểu lầm 4

> Số lượng Goroutine lớn luôn là lỗi.

Sai.

Một chương trình có hàng trăm nghìn Goroutine vẫn có thể hoàn toàn bình thường nếu phần lớn chúng đang chờ I/O hoặc có vòng đời hợp lý.

Điều quan trọng là số lượng Goroutine có tiếp tục tăng không kiểm soát hay không.

---

# Những điều cần ghi nhớ

- Goroutine Leak là lỗi rất phổ biến trong Go.
- Garbage Collector không thể giải phóng Goroutine còn sống.
- Goroutine Dump cho biết trạng thái và Stack Trace của mọi Goroutine.
- Runtime Trace ghi lại toàn bộ hoạt động của Scheduler theo thời gian.
- Ba công cụ quan trọng nhất khi Debug Goroutine là:
  - `runtime.NumGoroutine()`
  - Goroutine Dump
  - Runtime Trace

---

# Tổng kết Phần VII

Sau ba Section cuối cùng này, chúng ta không chỉ hiểu cách Goroutine hoạt động bên trong Go Runtime mà còn biết cách quan sát và chẩn đoán chúng trong các hệ thống thực tế.

```text
Create Goroutine

↓

Run

↓

Waiting

↓

Wakeup

↓

Preemption

↓

Exit

↓

Debug

    ├── NumGoroutine
    ├── Goroutine Dump
    └── Runtime Trace
```

Đây là những công cụ mà hầu hết các Senior Go Engineer sử dụng hằng ngày khi phân tích các hệ thống Production.

---

---

# 11.28 Common Mistakes - Những sai lầm phổ biến về Goroutine và Go Runtime

Sau khi hoàn thành toàn bộ Chapter 11, chúng ta đã nghiên cứu gần như mọi khía cạnh của Goroutine, từ:

- `struct g`
- `gobuf`
- Stack
- Context Switching
- Scheduler Interaction
- Dynamic Stack
- Defer
- Panic
- Recover
- Goroutine States
- Waiting Queue
- Park / Unpark
- Preemption
- Debugging

Tuy nhiên.

Trong thực tế.

Rất nhiều lập trình viên Go, kể cả những người đã có nhiều năm kinh nghiệm, vẫn mắc phải những hiểu lầm về cách Goroutine hoạt động.

Phần này sẽ tổng hợp những sai lầm phổ biến nhất mà bạn có thể gặp trong:

- Phỏng vấn.
- Code Review.
- Thiết kế hệ thống.
- Debug Production.

---

# Sai lầm 1 - Goroutine là OS Thread

Đây có lẽ là hiểu lầm phổ biến nhất.

Nhiều người cho rằng.

```text
1 Goroutine

=

1 Thread
```

Điều này hoàn toàn sai.

Thực tế.

```text
100.000 Goroutines

↓

Có thể chạy trên

↓

8 Threads
```

Thông qua Scheduler.

Go Runtime ánh xạ.

```
G

↓

M

↓

CPU
```

chứ không tạo một Thread mới cho mỗi Goroutine.

---

# Sai lầm 2 - Tạo Goroutine rất tốn bộ nhớ

Nhiều lập trình viên đến từ Java hoặc C++ thường nghĩ.

```
go worker()
```

sẽ tiêu tốn.

```
1 MB

hoặc

8 MB
```

Stack.

Điều này không đúng.

Một Goroutine chỉ bắt đầu với.

```
≈ 2 KB Stack
```

Sau đó.

Runtime mới Grow khi cần.

Đây là lý do Go có thể tạo hàng triệu Goroutine.

---

# Sai lầm 3 - Goroutine chạy song song

Không phải lúc nào cũng đúng.

Ví dụ.

```
GOMAXPROCS = 1
```

```text
G1

↓

G2

↓

G3
```

Ba Goroutine.

Nhưng.

Chỉ có **một CPU** thực thi tại một thời điểm.

Đây là:

```
Concurrency
```

không phải:

```
Parallelism
```

---

# Sai lầm 4 - Goroutine luôn chạy ngay sau khi gọi `go`

Ví dụ.

```go
go worker()

fmt.Println("Done")
```

Không có gì đảm bảo.

```
worker()
```

được chạy trước.

Scheduler hoàn toàn quyết định thời điểm Goroutine bắt đầu thực thi.

---

# Sai lầm 5 - Goroutine có ID giống Thread ID

Go Runtime có.

```
goid
```

Đây là ID nội bộ.

Không phải.

```
Thread ID
```

Một Goroutine có thể chạy trên nhiều Thread khác nhau trong suốt vòng đời của nó.

---

# Sai lầm 6 - Scheduler chạy theo Round Robin đơn giản

Không đúng.

Go Scheduler sử dụng.

- Local Run Queue.
- Global Queue.
- Work Stealing.
- Spinning Thread.
- Netpoll.
- Timer Queue.
- Asynchronous Preemption.

Round Robin chỉ là một phần rất nhỏ của toàn bộ cơ chế lập lịch.

---

# Sai lầm 7 - Context Switching luôn cần Kernel

Đây là sự khác biệt lớn giữa Thread và Goroutine.

Thread.

```text
Thread A

↓

Kernel

↓

Thread B
```

Goroutine.

```text
G1

↓

Runtime

↓

G2
```

Hầu hết Context Switching của Goroutine diễn ra hoàn toàn trong User Space.

---

# Sai lầm 8 - Stack của Goroutine có kích thước cố định

Không đúng.

Stack của Goroutine.

```text
2 KB

↓

4 KB

↓

8 KB

↓

...

↓

Shrink
```

Đây là.

```
Dynamic Stack
```

Một trong những thiết kế quan trọng nhất của Go Runtime.

---

# Sai lầm 9 - Garbage Collector giải phóng Goroutine

Garbage Collector chỉ quản lý.

```
Heap Object
```

Không quản lý.

```
Running Goroutine
```

Nếu Goroutine bị Leak.

GC không thể xóa nó.

---

# Sai lầm 10 - Goroutine Leak cũng là Memory Leak

Hai khái niệm này liên quan nhưng không giống nhau.

## Memory Leak

Object không được giải phóng.

---

## Goroutine Leak

Goroutine không bao giờ kết thúc.

Do Goroutine vẫn giữ Pointer tới Heap.

Memory Leak thường xuất hiện sau đó.

---

# Sai lầm 11 - Waiting Goroutine vẫn dùng CPU

Không đúng.

Một Goroutine.

```text
Waiting
```

không sử dụng CPU.

CPU được Scheduler giao cho Goroutine khác.

---

# Sai lầm 12 - `time.Sleep()` giữ CPU

Ví dụ.

```go
time.Sleep(time.Second)
```

Nhiều người nghĩ.

CPU sẽ chờ một giây.

Thực tế.

Runtime.

```text
Running

↓

Waiting

↓

Run Goroutine khác
```

Không có CPU nào bị lãng phí.

---

# Sai lầm 13 - `goready()` chạy Goroutine ngay lập tức

Sai.

`goready()` chỉ làm.

```text
Waiting

↓

Runnable
```

Scheduler vẫn quyết định thời điểm Goroutine thực sự chạy.

---

# Sai lầm 14 - `gopark()` giống `time.Sleep()`

Không đúng.

`gopark()` chỉ là cơ chế:

```
Suspend Goroutine
```

Sleep.

Channel.

Mutex.

Netpoll.

Đều có thể sử dụng.

```
gopark()
```

---

# Sai lầm 15 - Preemption chỉ xảy ra khi gọi `runtime.Gosched()`

Sai.

Từ Go 1.14.

Runtime hỗ trợ.

```
Asynchronous Preemption
```

Runtime có thể chủ động tạm dừng Goroutine ngay cả khi Goroutine không tự Yield.

---

# Sai lầm 16 - Panic làm chương trình dừng ngay lập tức

Không đúng.

Runtime sẽ.

```text
Panic

↓

Execute Defer

↓

Stack Unwinding

↓

Recover?

↓

Exit
```

Panic là cả một quy trình.

Không phải chỉ là.

```
Crash
```

---

# Sai lầm 17 - Recover có thể gọi ở bất kỳ đâu

Sai.

Recover chỉ hoạt động.

```
Trong Defer.
```

Ngoài Defer.

```
recover()

↓

nil
```

---

# Sai lầm 18 - Defer luôn rất chậm

Đây là một hiểu lầm cũ.

Từ Go 1.14.

Compiler hỗ trợ.

```
Open-Coded Defer
```

Trong nhiều trường hợp.

Compiler Inline Defer.

Chi phí thấp hơn rất nhiều so với trước đây.

---

# Sai lầm 19 - Goroutine bị Block thì Thread cũng bị Block

Không đúng.

Ví dụ.

```go
<-ch
```

Runtime.

```text
Park Goroutine

↓

Run Goroutine khác
```

Thread vẫn tiếp tục làm việc.

Chỉ trong một số trường hợp như Blocking System Call, OS Thread mới thực sự bị chặn.

---

# Sai lầm 20 - Một Goroutine luôn chạy trên cùng một Thread

Sai.

Ví dụ.

```text
Time 1

G1

↓

M1
```

Sau Context Switching.

```text
Time 2

G1

↓

M4
```

Scheduler có thể đưa Goroutine sang Thread khác.

Miễn là Execution Context được Restore đầy đủ.

---

# Sai lầm 21 - Goroutine có thể bị Scheduler dừng ở mọi lệnh

Không hoàn toàn đúng.

Runtime chỉ Preempt tại.

```
Safe Point
```

Đây là nơi.

- Stack nhất quán.
- Register nhất quán.
- Garbage Collector an toàn.

---

# Sai lầm 22 - Goroutine Dump chỉ dùng khi Deadlock

Sai.

Dump còn dùng để.

- Leak.
- Performance.
- Timeout.
- Production Debug.

Đây là công cụ quan trọng nhất khi phân tích hệ thống Go.

---

# Sai lầm 23 - Runtime Trace thay thế Profiler

Không đúng.

Profiler trả lời.

```
CPU ở đâu?
```

Memory Profiler.

```
RAM ở đâu?
```

Trace.

```
Runtime đang làm gì?
```

Ba công cụ phục vụ ba mục đích khác nhau.

---

# Sai lầm 24 - Hàng trăm nghìn Goroutine luôn là lỗi

Không đúng.

Một Server sử dụng:

- Netpoll.
- Context.
- Event-driven.

có thể chạy.

```
500.000 Goroutines
```

hoàn toàn bình thường.

Điều quan trọng là.

- Chúng có bị Leak hay không.
- Có tiêu tốn tài nguyên bất thường hay không.

---

# Sai lầm 25 - Hiểu Scheduler là đủ để hiểu Go Runtime

Không đúng.

Scheduler chỉ là một phần của Runtime.

Để hiểu Go Runtime.

Cần hiểu thêm.

- Memory Allocator.
- Garbage Collector.
- Netpoll.
- Timer.
- Syscall.
- Synchronization.
- Compiler.
- Assembly.

Đây là lý do handbook này được chia thành nhiều Volume.

---

# Checklist tự đánh giá

Sau khi hoàn thành Chapter 11, hãy thử tự trả lời các câu hỏi sau mà không cần tra tài liệu.

- Bạn có thể giải thích vì sao Goroutine chỉ cần khoảng 2 KB Stack ban đầu?
- Bạn có thể mô tả cơ chế Stack Growth và Stack Copy?
- Bạn có thể giải thích vai trò của `struct g` và `gobuf`?
- Bạn có thể phân biệt Scheduler với Context Switching?
- Bạn có thể mô tả toàn bộ quá trình Park và Unpark?
- Bạn có thể giải thích sự khác nhau giữa `_Grunnable`, `_Grunning`, `_Gwaiting` và `_Gsyscall`?
- Bạn có thể giải thích cơ chế Preemption của Go 1.14+?
- Bạn có thể phân biệt Goroutine Leak với Memory Leak?
- Bạn có thể sử dụng Goroutine Dump và Runtime Trace để Debug?
- Bạn có thể mô tả vòng đời đầy đủ của một Goroutine?

Nếu bạn trả lời được phần lớn các câu hỏi trên mà không cần nhìn tài liệu, bạn đã có nền tảng rất vững về Goroutine và Go Runtime.

---

---

# Phần VIII - Engineering

Cho đến thời điểm này, chúng ta đã dành gần như toàn bộ Chapter 11 để tìm hiểu Goroutine từ góc nhìn của Go Runtime.

Chúng ta đã nghiên cứu:

- Cấu trúc `struct g`
- `gobuf`
- Dynamic Stack
- Context Switching
- Scheduler States
- Waiting Queue
- Park / Unpark
- Preemption
- Goroutine Leak
- Runtime Trace

Đây đều là những kiến thức ở mức Runtime.

Tuy nhiên.

Khi xây dựng một hệ thống Production.

Điều quan trọng hơn là:

> **Làm thế nào để sử dụng Goroutine đúng cách?**

Rất nhiều hệ thống Go gặp vấn đề không phải vì Go Runtime.

Mà vì cách lập trình viên sử dụng Goroutine.

Phần này tổng hợp những Best Practices quan trọng nhất được đúc kết từ:

- Source code của Go Runtime.
- Các bài viết của Go Team.
- Các hệ thống Production lớn.
- Kinh nghiệm thực tế của cộng đồng Go.

---

# 11.29 Best Practices

## Best Practice 1 - Không tạo Goroutine nếu chưa thật sự cần

Nhiều người có thói quen.

```go
go doSomething()
```

ở khắp mọi nơi.

Đây không phải lúc nào cũng tốt.

Hãy tự hỏi.

- Công việc có thực sự cần chạy song song?
- Có mang lại lợi ích về hiệu năng không?
- Có làm code khó Debug hơn không?

Concurrency là công cụ.

Không phải mục tiêu.

---

## Best Practice 2 - Mỗi Goroutine phải có điều kiện kết thúc

Đây là nguyên tắc quan trọng nhất.

Không bao giờ tạo Goroutine như.

```go
go func() {

    for {

    }

}()
```

nếu không có điều kiện dừng.

Mọi Goroutine đều nên có.

- Context.
- Channel.
- Signal.
- Exit Condition.

---

## Best Practice 3 - Luôn sử dụng `context.Context`

Thay vì.

```go
go worker()
```

Hãy ưu tiên.

```go
go worker(ctx)
```

Context giúp.

- Cancel.
- Timeout.
- Deadline.
- Trace.

Đây là chuẩn của hầu hết thư viện Go hiện đại.

---

## Best Practice 4 - Không Ignore Context Cancellation

Ví dụ.

```go
select {

case <-ctx.Done():

    return

}
```

Nếu Goroutine bỏ qua.

```
ctx.Done()
```

Rất dễ dẫn tới Goroutine Leak.

---

## Best Practice 5 - Không tạo Goroutine trong vòng lặp mà không giới hạn

Ví dụ.

Sai.

```go
for _, task := range tasks {

    go process(task)

}
```

Nếu.

```
tasks

=

1.000.000
```

Bạn vừa tạo.

```
1.000.000 Goroutines
```

Hãy cân nhắc.

- Worker Pool.
- Semaphore.
- Rate Limiter.

---

## Best Practice 6 - Dùng Worker Pool cho khối lượng công việc lớn

Thay vì.

```
1 Goroutine

↓

1 Task
```

Hãy.

```
N Workers

↓

Task Queue
```

Điều này.

- Giảm Allocation.
- Giảm Context Switching.
- Kiểm soát CPU.

---

## Best Practice 7 - Đóng Channel đúng thời điểm

Chỉ.

```
Sender
```

mới nên đóng Channel.

Không bao giờ.

```
Receiver Close
```

Điều này giúp tránh Panic.

---

## Best Practice 8 - Không đóng Channel nhiều lần

Sai.

```go
close(ch)

close(ch)
```

Sẽ Panic.

Nếu nhiều Goroutine có thể đóng.

Hãy sử dụng.

- `sync.Once`
- Hoặc cơ chế đồng bộ khác.

---

## Best Practice 9 - Không gửi vào Channel đã đóng

Ví dụ.

```go
close(ch)

ch <- value
```

Runtime sẽ Panic.

Đây là lỗi rất phổ biến.

---

## Best Practice 10 - Luôn có Timeout khi chờ I/O

Ví dụ.

```go
select {

case value := <-ch:

...

case <-time.After(...):

...
}
```

Hoặc.

```
Context Timeout
```

Điều này tránh Waiting Forever.

---

## Best Practice 11 - Không dùng `time.Sleep()` để đồng bộ

Sai.

```go
go worker()

time.Sleep(time.Second)
```

Đây là Race Condition tiềm ẩn.

Hãy dùng.

- WaitGroup.
- Channel.
- Context.

---

## Best Practice 12 - Luôn gọi `wg.Done()`

Ví dụ.

```go
defer wg.Done()
```

Đặt ngay đầu Goroutine.

Giúp tránh quên.

```
Done()
```

khi Return sớm.

---

## Best Practice 13 - Không Copy WaitGroup

Sai.

```go
wg2 := wg
```

WaitGroup không được Copy sau khi sử dụng.

Điều này có thể dẫn đến hành vi không xác định.

---

## Best Practice 14 - Giảm Blocking System Call

System Call dài.

Có thể làm tăng số lượng Thread.

Ví dụ.

- Database.
- File.
- Network.

Nên sử dụng.

- Context.
- Timeout.
- Connection Pool.

---

## Best Practice 15 - Không tạo Goroutine trong Library nếu không cần thiết

Library nên để.

```
Caller
```

quyết định Concurrency.

Điều này giúp.

- Dễ Debug.
- Dễ Test.
- Dễ kiểm soát tài nguyên.

---

## Best Practice 16 - Theo dõi số lượng Goroutine

Production nên ghi nhận.

```go
runtime.NumGoroutine()
```

Nếu số lượng tăng liên tục.

Có thể đang xảy ra.

```
Goroutine Leak
```

Đây là một chỉ số vận hành rất hữu ích.

---

## Best Practice 17 - Sử dụng Goroutine Dump khi Debug

Khi gặp.

- Deadlock.
- Leak.
- Timeout.
- CPU thấp.

Hãy tạo.

```
Goroutine Dump
```

trước khi sửa code.

Stack Trace thường cho bạn biết chính xác Goroutine đang bị kẹt ở đâu.

---

## Best Practice 18 - Dùng Runtime Trace để hiểu Scheduler

Khi gặp các vấn đề về.

- Latency.
- Wakeup.
- Park.
- Preemption.

Hãy sử dụng.

```
runtime/trace
```

Thay vì chỉ nhìn CPU Profiler.

Trace cho thấy hành vi của Runtime theo thời gian.

---

## Best Practice 19 - Hiểu sự khác nhau giữa Concurrency và Parallelism

Concurrency.

```
Nhiều Goroutine

↓

Có thể chạy xen kẽ
```

Parallelism.

```
Nhiều CPU

↓

Chạy cùng lúc
```

Không nên kỳ vọng rằng việc thêm Goroutine luôn làm chương trình nhanh hơn.

Nếu công việc bị giới hạn bởi:

- I/O.
- Lock.
- Database.

Việc tạo thêm Goroutine có thể không mang lại lợi ích.

---

## Best Practice 20 - Thiết kế Goroutine như một tài nguyên cần quản lý

Nhiều người xem Goroutine là "miễn phí".

Điều này không đúng.

Mỗi Goroutine đều tiêu tốn:

- Stack.
- `struct g`.
- Scheduler Metadata.
- Garbage Collector Scan.
- Context Switching.

Hãy coi Goroutine giống như:

- Database Connection.
- File Descriptor.
- Network Socket.

Đều là tài nguyên cần được quản lý cẩn thận.

---

# Bonus Best Practices

Ngoài 20 nguyên tắc trên, dưới đây là một số kinh nghiệm thường thấy trong các hệ thống lớn.

### Sử dụng `errgroup` thay cho `WaitGroup` khi cần trả lỗi

Nếu nhiều Goroutine cùng thực hiện một tác vụ và chỉ cần một lỗi để hủy toàn bộ công việc, `errgroup` thường phù hợp hơn `WaitGroup`.

---

### Tránh chia sẻ bộ nhớ khi có thể

Go có câu nói nổi tiếng.

> **"Do not communicate by sharing memory; instead, share memory by communicating."**

Ưu tiên.

- Channel.
- Immutable Data.

Thay vì Lock ở mọi nơi.

---

### Giảm thời gian giữ Mutex

Không nên.

```go
Lock()

↓

Database Query

↓

Unlock()
```

Hãy giữ Critical Section càng nhỏ càng tốt.

---

### Luôn Benchmark trước khi tối ưu

Không nên.

```
"Tạo thêm Goroutine sẽ nhanh hơn."
```

Hãy đo bằng.

- Benchmark.
- Trace.
- Profiler.

Runtime đôi khi tối ưu tốt hơn chúng ta nghĩ.

---

### Đọc Goroutine Dump trước khi đọc Source Code

Khi Production gặp sự cố.

Goroutine Dump thường giúp xác định vấn đề nhanh hơn việc đọc hàng nghìn dòng mã nguồn.

---

# Checklist Production

Trước khi triển khai một dịch vụ Go, hãy tự kiểm tra:

- [ ] Mọi Goroutine đều có điều kiện kết thúc.
- [ ] Có sử dụng `context.Context`.
- [ ] Không có Goroutine tạo không giới hạn.
- [ ] Có Timeout cho I/O.
- [ ] Không dùng `time.Sleep()` để đồng bộ.
- [ ] Channel được đóng đúng nơi.
- [ ] Không có khả năng Goroutine Leak.
- [ ] Có theo dõi `runtime.NumGoroutine()`.
- [ ] Có endpoint hoặc cơ chế lấy Goroutine Dump.
- [ ] Biết cách sử dụng Runtime Trace khi cần.

---

---

# 11.30 Interview Questions

Sau khi hoàn thành Chapter 11, bạn đã nắm gần như toàn bộ kiến thức về Goroutine ở mức Runtime.

Trong phần này, chúng ta sẽ tổng hợp khoảng **50 câu hỏi phỏng vấn** từ mức Junior đến Senior Staff.

Các câu hỏi được chia thành nhiều nhóm:

- Foundation
- Runtime
- Scheduler
- Performance
- Debugging
- Source Code
- System Design

Đây đều là những câu hỏi rất phổ biến trong các buổi phỏng vấn Go Backend Engineer.

---

# Phần I - Foundation

## Câu 1

Goroutine là gì?

---

## Câu 2

Goroutine khác Thread ở điểm nào?

---

## Câu 3

Tại sao Goroutine nhẹ hơn Thread?

---

## Câu 4

Một Goroutine ban đầu sử dụng bao nhiêu Stack?

---

## Câu 5

Tại sao Go Runtime không cấp phát 1MB Stack ngay từ đầu?

---

## Câu 6

Điều gì xảy ra khi Goroutine kết thúc?

---

## Câu 7

Có bao nhiêu Goroutine có thể chạy trên một OS Thread?

---

## Câu 8

Một Goroutine có luôn chạy trên cùng một Thread không?

---

## Câu 9

Concurrency khác Parallelism như thế nào?

---

## Câu 10

Điều gì xảy ra sau khi gọi từ khóa `go`?

---

# Phần II - Runtime Internals

## Câu 11

`struct g` là gì?

---

## Câu 12

`gobuf` dùng để làm gì?

---

## Câu 13

`atomicstatus` lưu thông tin gì?

---

## Câu 14

Execution Context gồm những thành phần nào?

---

## Câu 15

Program Counter có vai trò gì trong Context Switching?

---

## Câu 16

Stack Pointer được Runtime sử dụng như thế nào?

---

## Câu 17

Runtime lưu Execution Context ở đâu?

---

## Câu 18

Tại sao Go Runtime cần mã Assembly cho Context Switching?

---

## Câu 19

Return Address được lưu ở đâu?

---

## Câu 20

Điều gì xảy ra khi một Function Return?

---

# Phần III - Dynamic Stack

## Câu 21

Dynamic Stack là gì?

---

## Câu 22

Runtime phát hiện Stack đầy bằng cách nào?

---

## Câu 23

`stackguard0` là gì?

---

## Câu 24

`morestack()` có nhiệm vụ gì?

---

## Câu 25

Runtime mở rộng Stack như thế nào?

---

## Câu 26

Tại sao Runtime phải Copy Stack thay vì Realloc tại chỗ?

---

## Câu 27

Runtime cập nhật Pointer sau khi Copy Stack bằng cách nào?

---

## Câu 28

Stack Map là gì?

---

## Câu 29

Khi nào Runtime thu nhỏ Stack?

---

## Câu 30

Garbage Collector có vai trò gì trong Stack Shrink?

---

# Phần IV - Scheduler Interaction

## Câu 31

Các trạng thái chính của Goroutine là gì?

---

## Câu 32

Sự khác nhau giữa `_Grunnable` và `_Grunning`?

---

## Câu 33

Khi nào Goroutine chuyển sang `_Gwaiting`?

---

## Câu 34

Waiting Queue là gì?

---

## Câu 35

Channel đánh thức Goroutine bằng cách nào?

---

## Câu 36

`gopark()` làm gì?

---

## Câu 37

`goready()` làm gì?

---

## Câu 38

Park khác Sleep như thế nào?

---

## Câu 39

Preemption là gì?

---

## Câu 40

Asynchronous Preemption được giới thiệu từ phiên bản Go nào và giải quyết vấn đề gì?

---

# Phần V - Runtime Behavior

## Câu 41

Defer được Runtime quản lý như thế nào?

---

## Câu 42

Panic khác Error như thế nào?

---

## Câu 43

Recover hoạt động trong trường hợp nào?

---

## Câu 44

Stack Unwinding là gì?

---

## Câu 45

Runtime thực hiện những bước nào khi Goroutine Exit?

---

# Phần VI - Debugging

## Câu 46

Goroutine Leak là gì?

---

## Câu 47

Làm thế nào phát hiện Goroutine Leak?

---

## Câu 48

Goroutine Dump là gì?

---

## Câu 49

Runtime Trace dùng để làm gì?

---

## Câu 50

Nếu Production Server có 300.000 Goroutine, bạn sẽ bắt đầu Debug từ đâu?

---

# Bonus Questions (Senior)

Đây là các câu hỏi thường gặp trong các công ty lớn hoặc các vị trí Senior/Staff.

---

## Bonus 1

Giải thích toàn bộ vòng đời của một Goroutine từ khi tạo cho đến khi kết thúc.

---

## Bonus 2

Giải thích chi tiết quá trình Context Switching của Go Runtime.

---

## Bonus 3

Tại sao Context Switching của Goroutine nhanh hơn Thread?

---

## Bonus 4

Điều gì xảy ra nếu một Goroutine đệ quy vô hạn?

---

## Bonus 5

Nếu Runtime không có Dynamic Stack thì điều gì sẽ xảy ra?

---

## Bonus 6

Tại sao Go Runtime chuyển từ Segmented Stack sang Contiguous Stack?

---

## Bonus 7

Giải thích mối quan hệ giữa Goroutine, Scheduler và Garbage Collector.

---

## Bonus 8

Nếu một Goroutine bị Block trong System Call thì Scheduler xử lý như thế nào?

---

## Bonus 9

Giải thích sự khác nhau giữa Waiting Queue và Run Queue.

---

## Bonus 10

Điều gì xảy ra nếu `goready()` bị gọi hai lần cho cùng một Goroutine?

---

## Bonus 11

Runtime làm thế nào để đảm bảo việc chuyển trạng thái Goroutine không bị Race Condition?

---

## Bonus 12

Safe Point là gì? Tại sao Runtime chỉ Preempt tại Safe Point?

---

## Bonus 13

Runtime làm thế nào để Copy Stack mà không làm hỏng Pointer?

---

## Bonus 14

Nếu Runtime không cập nhật Stack Map thì Garbage Collector sẽ gặp vấn đề gì?

---

## Bonus 15

Giải thích mối liên hệ giữa:

- `struct g`
- `gobuf`
- Stack
- Stack Frame
- Program Counter
- Context Switching

---

## Bonus 16

Nếu bạn được thiết kế Goroutine từ đầu, bạn sẽ thay đổi điều gì trong Go Runtime hiện nay?

---

## Bonus 17

Theo bạn, phần khó nhất trong việc triển khai Goroutine là gì?

---

## Bonus 18

Tại sao Scheduler cần `atomicstatus` thay vì một biến thông thường?

---

## Bonus 19

Điều gì sẽ xảy ra nếu Runtime cho phép Preempt tại mọi Instruction?

---

## Bonus 20

Hãy mô tả kiến trúc Goroutine của Go Runtime mà không sử dụng bất kỳ hình minh họa nào.

---

# Coding Interview

Ngoài lý thuyết, rất nhiều công ty yêu cầu ứng viên giải thích hành vi của đoạn mã sau.

---

## Bài 1

```go
package main

import "fmt"

func main() {
    go fmt.Println("Hello")
}
```

**Câu hỏi**

Điều gì sẽ xảy ra?

Tại sao?

---

## Bài 2

```go
for i := 0; i < 10; i++ {
    go func() {
        fmt.Println(i)
    }()
}
```

**Câu hỏi**

Output có thể là gì?

Giải thích nguyên nhân.

---

## Bài 3

```go
go worker()

time.Sleep(time.Second)
```

**Câu hỏi**

Đây có phải cách đồng bộ Goroutine đúng không?

Nếu không, tại sao?

---

## Bài 4

```go
go func() {

    select {}

}()
```

**Câu hỏi**

Đây có phải Goroutine Leak không?

---

## Bài 5

```go
go func() {

    for {

    }

}()
```

**Câu hỏi**

Đoạn code này ảnh hưởng như thế nào đến Scheduler?

---

## Bài 6

```go
func worker(ch chan int) {

    <-ch

}
```

**Câu hỏi**

Trong Runtime, Goroutine này sẽ đi qua những trạng thái nào?

---

## Bài 7

```go
mutex.Lock()

defer mutex.Unlock()
```

**Câu hỏi**

Nếu xảy ra Panic thì Mutex có được Unlock không?

Giải thích theo cơ chế Runtime.

---

## Bài 8

```go
defer fmt.Println("A")

panic("x")
```

**Câu hỏi**

Runtime sẽ thực hiện các bước nào?

---

## Bài 9

```go
defer func() {

    recover()

}()
```

**Câu hỏi**

Recover hoạt động trong trường hợp nào?

---

## Bài 10

Một Production Server:

- CPU thấp
- RAM tăng liên tục
- Số Goroutine tăng từ 5.000 lên 150.000

**Câu hỏi**

Bạn sẽ Debug theo trình tự như thế nào?

---

# Lời khuyên khi phỏng vấn

Đối với các vị trí Middle hoặc Senior, nhà tuyển dụng thường không chỉ quan tâm đến việc bạn **biết** một khái niệm, mà còn muốn hiểu **bạn suy nghĩ như thế nào**.

Khi trả lời các câu hỏi về Go Runtime:

- Bắt đầu từ khái niệm tổng quát.
- Giải thích bằng ví dụ cụ thể.
- Mô tả luồng hoạt động của Runtime.
- Nếu phù hợp, liên hệ với `struct g`, Scheduler hoặc Garbage Collector.
- Chỉ đi xuống mức Assembly hoặc Source Code khi người phỏng vấn muốn đào sâu.

Một câu trả lời có cấu trúc và có khả năng kết nối các thành phần của Runtime thường tạo ấn tượng tốt hơn nhiều so với việc chỉ liệt kê định nghĩa.

---

# Tổng kết

50 câu hỏi trên bao quát gần như toàn bộ kiến thức về Goroutine trong Go Runtime, từ những khái niệm nền tảng cho đến các chủ đề chuyên sâu như Dynamic Stack, Context Switching, Scheduler, Debugging và Performance.

Nếu bạn có thể tự trả lời phần lớn các câu hỏi này một cách mạch lạc và giải thích được **tại sao Runtime được thiết kế như vậy**, bạn đã đạt đến mức hiểu biết đủ để tự tin trong các buổi phỏng vấn Go Backend từ Middle đến Senior, đồng thời có nền tảng vững chắc để tiếp tục nghiên cứu các chương tiếp theo về **Machine (`struct m`)**, **Processor (`struct p`)** và **G-M-P Scheduler**.

---

---

# 11.31 Tổng kết chương

Chúng ta vừa hoàn thành toàn bộ **Chapter 11 - Goroutine**.

Đây có lẽ là chương quan trọng nhất của toàn bộ Volume II - Go Runtime.

Bởi vì.

Mọi chương sau này đều xoay quanh Goroutine.

Nếu không hiểu Goroutine.

Sẽ rất khó hiểu:

- Scheduler
- Netpoll
- Garbage Collector
- Timer
- Memory Allocator
- Context
- Channel
- Synchronization

Có thể nói.

> **Goroutine chính là "đơn vị thực thi" (Execution Unit) của Go Runtime.**

Hiểu Goroutine đồng nghĩa với việc đã hiểu được trái tim của ngôn ngữ Go.

---

# Chúng ta đã học những gì?

Trong chương này.

Chúng ta không chỉ học cách sử dụng.

```go
go worker()
```

Mà còn đi sâu vào.

```
Go Runtime
```

để trả lời câu hỏi.

> **Điều gì thực sự xảy ra sau từ khóa `go`?**

Chúng ta đã lần lượt nghiên cứu từ mức khái niệm cho đến mức Runtime và Assembly.

---

# Phần I - Tổng quan

Chúng ta bắt đầu bằng những câu hỏi cơ bản.

- Goroutine là gì?
- Vì sao Goroutine nhẹ hơn Thread?
- Vai trò của Goroutine trong kiến trúc G-M-P.
- Mối quan hệ giữa Goroutine và Scheduler.

Sau phần này.

Chúng ta có được cái nhìn tổng quan về Goroutine như một đơn vị thực thi độc lập.

---

# Phần II - Cấu trúc nội bộ

Tiếp theo.

Chúng ta đi vào.

```
struct g
```

Đây là cấu trúc dữ liệu quan trọng nhất của Goroutine.

Chúng ta đã tìm hiểu.

- `struct g`
- `gobuf`
- Execution Context
- Program Counter
- Stack Pointer
- Registers

Sau phần này.

Chúng ta hiểu rằng.

> Goroutine không chỉ là một Stack.

Mà là một tập hợp lớn các thông tin Runtime dùng để quản lý quá trình thực thi.

---

# Phần III - Dynamic Stack

Đây là một trong những phần quan trọng nhất.

Chúng ta đã nghiên cứu.

- Initial Stack
- Stack Growth
- Stack Copy
- Stack Shrink
- Stack Overflow Detection

Đồng thời hiểu.

- Vì sao Goroutine chỉ bắt đầu với khoảng 2 KB Stack.
- Runtime mở rộng Stack như thế nào.
- Runtime cập nhật Pointer ra sao.
- Vì sao Go chuyển từ Segmented Stack sang Contiguous Stack.

Sau phần này.

Chúng ta hiểu rõ một trong những thiết kế quan trọng nhất của Go Runtime.

---

# Phần IV - Context Switching

Đây là phần kết nối Scheduler với CPU.

Chúng ta đã học.

- Context Switching
- Save Registers
- Restore Context
- Function Call
- Return Address

Đồng thời hiểu.

- Runtime lưu Program Counter ở đâu.
- Runtime Restore Execution Context như thế nào.
- Vì sao Context Switching của Goroutine nhanh hơn Thread.

Đây là nền tảng để đọc mã Assembly của Runtime.

---

# Phần V - Runtime Lifecycle

Tiếp theo.

Chúng ta nghiên cứu toàn bộ vòng đời thực thi của một Goroutine.

Bao gồm.

- Defer
- Panic
- Recover
- Goroutine Exit

Sau phần này.

Chúng ta hiểu Runtime xử lý.

- `_defer`
- `_panic`
- Stack Unwinding

và cách một Goroutine được thu hồi.

---

# Phần VI - Scheduler Interaction

Đây là lần đầu tiên chúng ta nhìn Goroutine từ góc nhìn của Scheduler.

Chúng ta đã nghiên cứu.

- Goroutine States
- Waiting Queue
- Park / Unpark
- Preemption

Sau phần này.

Chúng ta hiểu.

- Vì sao Goroutine Waiting không tiêu tốn CPU.
- Scheduler đánh thức Goroutine như thế nào.
- Runtime thực hiện Asynchronous Preemption ra sao.

---

# Phần VII - Debugging

Không chỉ hiểu Runtime.

Chúng ta còn học cách quan sát Runtime.

Bao gồm.

- Goroutine Leak
- Goroutine Dump
- Runtime Trace

Đây là những công cụ quan trọng trong Production.

Một Senior Go Engineer gần như sử dụng chúng hằng tuần.

---

# Phần VIII - Engineering

Sau khi hiểu Runtime.

Chúng ta chuyển sang.

```
Engineering Practices
```

Bao gồm.

- 20 Best Practices.
- 50 Interview Questions.

Đây là phần kết nối kiến thức Runtime với thực tế phát triển hệ thống.

---

# Bức tranh tổng thể

Sau khi học xong Chapter này.

Bạn có thể hình dung toàn bộ vòng đời của một Goroutine.

```text
go keyword

↓

newproc()

↓

Allocate struct g

↓

Allocate Initial Stack

↓

Runnable

↓

Scheduler

↓

Running

↓

Function Call

↓

Stack Growth

↓

Context Switch

↓

Waiting

↓

Wakeup

↓

Running

↓

Preemption

↓

Running

↓

Function Return

↓

Execute Defer

↓

Panic?

↓

Recover?

↓

Exit

↓

Release Stack

↓

Recycle struct g
```

Đây chính là hành trình đầy đủ của một Goroutine.

---

# Những điều quan trọng nhất cần nhớ

Nếu chỉ được giữ lại một vài kiến thức của chương này.

Đó nên là.

### 1. Goroutine không phải Thread

Go Runtime tự quản lý Goroutine.

Kernel hoàn toàn không biết Goroutine là gì.

---

### 2. Dynamic Stack là chìa khóa

Nếu Stack cố định như Thread.

Go sẽ không thể tạo hàng triệu Goroutine.

---

### 3. Scheduler và Context Switching là hai khái niệm khác nhau

Scheduler.

```
Chọn Goroutine.
```

Context Switching.

```
Lưu và khôi phục Execution Context.
```

---

### 4. Waiting Goroutine không sử dụng CPU

Điều này giải thích vì sao.

Một Server có thể duy trì.

```
500.000 Connections
```

mà CPU vẫn thấp.

---

### 5. Goroutine Leak nguy hiểm hơn Memory Leak

Một Goroutine bị Leak.

↓

Giữ Pointer.

↓

Giữ Heap Object.

↓

GC không giải phóng được.

↓

RAM tăng.

---

### 6. Runtime và Compiler luôn phối hợp

Rất nhiều cơ chế.

- Stack Growth.
- Stack Map.
- Defer.
- Safe Point.
- Preemption.

đều là kết quả của sự phối hợp giữa.

```
Compiler

+

Runtime
```

Không phải Runtime hoạt động một mình.

---

# Kiến thức còn thiếu

Mặc dù đã hiểu Goroutine.

Chúng ta vẫn chưa thể giải thích hoàn toàn Scheduler.

Bởi vì.

Scheduler không chỉ có.

```
G
```

Nó còn có.

```
M

↓

Machine
```

và.

```
P

↓

Processor
```

Đây là hai thành phần còn lại của mô hình.

```
G-M-P
```

---

# Chuẩn bị cho Chapter 12

Chapter tiếp theo sẽ nghiên cứu.

```
Machine

(struct m)
```

Đây là nơi Go Runtime biểu diễn.

```
OS Thread
```

Chúng ta sẽ trả lời các câu hỏi.

- Machine là gì?
- Vì sao Runtime cần `struct m`?
- Mối quan hệ giữa Thread và Machine?
- Machine lưu những gì?
- Machine được tạo và hủy như thế nào?
- Thread Parking là gì?
- Thread Spinning là gì?
- Thread Cache là gì?
- Runtime quản lý Thread Pool ra sao?

Sau chương này.

Chúng ta sẽ hiểu thêm một mảnh ghép quan trọng của kiến trúc G-M-P.

---

# Góc nhìn của toàn bộ Volume II

Sau Chapter 10 và Chapter 11.

Chúng ta đã hoàn thành.

```text
Volume II

Go Runtime

│

├── Chapter 10

│     Scheduler

│

└── Chapter 11

      Goroutine
```

Hai chương này giải thích.

- Scheduler hoạt động như thế nào.
- Goroutine là gì.
- Scheduler quản lý Goroutine ra sao.

Nhưng vẫn còn thiếu.

```
Machine

Processor
```

Đây sẽ là trọng tâm của các chương tiếp theo.

---

# Hành trình phía trước

Sau khi hoàn thành Goroutine.

Roadmap của Volume II sẽ tiếp tục như sau.

```text
Chapter 12

Machine (struct m)

↓

Chapter 13

Processor (struct p)

↓

Chapter 14

G-M-P Scheduler

↓

Chapter 15

Network Poller

↓

Chapter 16

Memory Allocator

↓

Chapter 17

Garbage Collector

↓

Chapter 18

Synchronization

↓

Chapter 19

Timer

↓

Chapter 20

Runtime Source Code Walkthrough
```

Đến cuối Volume II.

Bạn sẽ có khả năng.

- Đọc source code Go Runtime.
- Hiểu các quyết định thiết kế của Go Team.
- Phân tích các vấn đề Performance ở Production.
- Tự tin trả lời các câu hỏi Go Runtime ở mức Senior hoặc Staff.
- Xây dựng các hệ thống Go có khả năng mở rộng và hiệu năng cao.

---

# Lời kết

Nếu Chapter 10 giúp chúng ta hiểu **Scheduler** là "bộ não" của Go Runtime.

Thì Chapter 11 giúp chúng ta hiểu **Goroutine** chính là "tế bào" mà bộ não đó quản lý.

Mỗi Goroutine không chỉ là một hàm chạy song song.

Bên trong nó là cả một hệ thống phức tạp bao gồm:

- `struct g`
- `gobuf`
- Dynamic Stack
- Context Switching
- State Machine
- Defer
- Panic
- Recover
- Debugging Infrastructure

Tất cả những thành phần này phối hợp chặt chẽ để tạo nên một mô hình thực thi nhẹ, hiệu quả và có khả năng mở rộng mà rất ít ngôn ngữ lập trình hiện đại đạt được.

Đây chính là một trong những lý do khiến Go trở thành lựa chọn hàng đầu cho các hệ thống Backend, Cloud Native, Distributed Systems và Infrastructure Software.

**Chapter 11 kết thúc tại đây.**

Trong chương tiếp theo, chúng ta sẽ tiếp tục khám phá thành phần thứ hai của mô hình **G-M-P**:

> **Chapter 12 - Machine (`struct m`)** - nơi Go Runtime quản lý các OS Thread và xây dựng cầu nối giữa Goroutine với hệ điều hành.

---