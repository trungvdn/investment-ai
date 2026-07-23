Infrastructure Layer

Tổng quan

Infrastructure Layer là tầng cung cấp các dịch vụ kỹ thuật phục vụ cho việc thực thi Workflow.

Tầng này đóng vai trò cầu nối giữa Execution Layer và các hệ thống bên ngoài như Repository, HTTP API, Large Language Model (LLM) và hệ thống ghi log.

Infrastructure Layer không chứa business logic, không điều phối Workflow và không đưa ra quyết định trong quá trình thực thi. Mục tiêu duy nhất của tầng này là cung cấp các khả năng kỹ thuật (Technical Services) để Execution Layer và Capability Layer hoạt động độc lập với hạ tầng bên dưới.

⸻

Mục tiêu

Infrastructure Layer được thiết kế nhằm các mục tiêu sau:

* Cung cấp các dịch vụ kỹ thuật dùng chung cho toàn bộ hệ thống.
* Tách biệt Execution Layer khỏi các chi tiết triển khai.
* Trừu tượng hóa việc kết nối với các hệ thống bên ngoài.
* Giảm sự phụ thuộc giữa Business Logic và Infrastructure.
* Cho phép thay thế hoặc mở rộng các thành phần hạ tầng mà không ảnh hưởng đến Workflow Engine.

⸻

Vị trí trong kiến trúc

                        Client
                           │
                           ▼
                    Execution Layer
                           │
                           ▼
                 Infrastructure Layer
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
 Definition Repository  Handler Registry  LLM Gateway
        │                  │                  │
        └──────────────────┼──────────────────┘
                           ▼
                   Capability Layer
                           │
                           ▼
                  External Systems

Infrastructure Layer nằm giữa Execution Layer và Capability Layer, cung cấp các dịch vụ kỹ thuật cần thiết để thực thi Workflow và giao tiếp với các hệ thống bên ngoài.

⸻

Trách nhiệm

Infrastructure Layer chịu trách nhiệm:

* Nạp Workflow Definition từ nguồn lưu trữ.
* Ánh xạ Activity đến Handler tương ứng.
* Cung cấp Capability cho Handler.
* Giao tiếp với các mô hình LLM.
* Gọi các dịch vụ HTTP bên ngoài.
* Ghi nhận log trong quá trình thực thi.

Infrastructure Layer không chịu trách nhiệm:

* Điều phối Workflow.
* Quản lý Runtime.
* Quản lý State.
* Thực thi Business Logic.
* Áp dụng Business Rules.
* Định nghĩa Execution Policy.

⸻

Các thành phần chính

Trong phiên bản MVP, Infrastructure Layer bao gồm sáu thành phần.

Infrastructure
├── Definition Repository
├── Handler Registry
├── Capability Registry
├── LLM Gateway
├── HTTP Client
└── Logger

Definition Repository

Chịu trách nhiệm nạp Workflow Definition từ nguồn dữ liệu.

Trong phiên bản MVP, Workflow Definition được lưu dưới dạng YAML.

⸻

Handler Registry

Quản lý ánh xạ giữa Activity Definition và Activity Handler.

Execution Engine chỉ cần biết Activity cần thực thi, còn Handler Registry sẽ chịu trách nhiệm tìm đúng Handler tương ứng.

⸻

Capability Registry

Quản lý các Business Capability được sử dụng bởi Handler.

Capability Registry giúp Handler không cần biết cách khởi tạo hoặc quản lý vòng đời của Capability.

⸻

LLM Gateway

Cung cấp một giao diện thống nhất để làm việc với các mô hình ngôn ngữ lớn.

Nhờ đó có thể thay đổi giữa OpenAI, Gemini, Claude hoặc Local LLM mà không cần sửa Business Logic.

⸻

HTTP Client

Đóng vai trò giao tiếp với các dịch vụ HTTP bên ngoài.

Ví dụ:

* API dữ liệu tài chính
* API tin tức
* API nội bộ
* Các dịch vụ của bên thứ ba

⸻

Logger

Ghi nhận các sự kiện trong quá trình thực thi Workflow.

Ví dụ:

* Execution Started
* Activity Started
* Activity Completed
* Execution Completed

⸻

Nguyên tắc thiết kế

Infrastructure Layer được xây dựng dựa trên các nguyên tắc sau.

Tách biệt trách nhiệm (Separation of Concerns)

Infrastructure chỉ cung cấp các dịch vụ kỹ thuật.

Business Logic thuộc Capability Layer.

Điều phối Workflow thuộc Execution Layer.

⸻

Đảo ngược phụ thuộc (Dependency Inversion)

Execution Layer chỉ phụ thuộc vào các interface của Infrastructure.

Các implementation cụ thể có thể thay đổi mà không ảnh hưởng đến Workflow Engine.

⸻

Có thể thay thế (Replaceable Components)

Mỗi thành phần Infrastructure đều có thể được thay thế.

Ví dụ:

* YAML → Database
* OpenAI → Local LLM
* HTTP → gRPC

Execution Layer không cần thay đổi khi Infrastructure được thay thế.

⸻

Giao diện ổn định (Stable Interfaces)

Các interface của Infrastructure cần ổn định theo thời gian.

Điều này giúp giảm sự phụ thuộc giữa các Layer và hỗ trợ mở rộng hệ thống trong tương lai.

⸻

Cấu trúc tài liệu

infrastructure/
README.md
architecture.md
specification.md
walkthrough.md
definition-repository.md
handler-registry.md
capability-registry.md
llm-gateway.md
http-client.md
logger.md

⸻

Thứ tự đọc

Để hiểu đầy đủ Infrastructure Layer, nên đọc theo thứ tự sau:

1. README.md
2. architecture.md
3. specification.md
4. walkthrough.md
5. definition-repository.md
6. handler-registry.md
7. capability-registry.md
8. llm-gateway.md
9. http-client.md
10. logger.md

⸻

Tài liệu liên quan

Infrastructure Layer là một phần của Workflow Engine và hoạt động cùng với:

* Capability Layer
* Execution Layer
* Workflow Definition
* Runtime Model
* Activity Executor

Ba tầng Capability – Execution – Infrastructure phối hợp với nhau để tạo thành một kiến trúc Workflow Engine hoàn chỉnh, trong đó:

* Capability Layer định nghĩa năng lực nghiệp vụ.
* Execution Layer điều phối quá trình thực thi.
* Infrastructure Layer cung cấp các dịch vụ kỹ thuật hỗ trợ cho toàn bộ hệ thống.
