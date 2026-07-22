# RFC-002 — Market Intelligence Information Model

Status: Draft

Version: 1.0

Owner: Investment AI Team

Parent Capability: Market Intelligence

---

# 1. Purpose

## Objective

Information Model định nghĩa toàn bộ Business Objects được sử dụng trong Market Intelligence Capability.

Tài liệu này mô tả:

- Ý nghĩa nghiệp vụ của từng đối tượng.
- Quan hệ giữa các đối tượng.
- Luồng chuyển đổi tri thức.
- Ngôn ngữ chung (Ubiquitous Language).

Information Model **không định nghĩa**:

- JSON Schema
- Go Struct
- Database Schema
- API Contract
- YAML
- Implementation

---

# 2. Design Principles

Information Model tuân theo các nguyên tắc sau.

## Business First

Mọi đối tượng đều phản ánh khái niệm nghiệp vụ.

Ví dụ:

✓ Market Trend

✓ Liquidity

✓ Risk Environment

Không sử dụng tên kỹ thuật như:

× DTO

× Response

× Payload

× Entity

---

## Immutable Knowledge

Mỗi Business Object đại diện cho tri thức tại một thời điểm phân tích.

Sau khi được tạo, nội dung không bị chỉnh sửa.

Nếu dữ liệu thay đổi, hệ thống tạo một phiên bản mới.

---

## Traceability

Mọi kết luận đều phải truy vết được.

Market Summary

↓

Market Context

↓

Evidence

↓

Indicators

---

## Explainability

Không có kết luận nào tồn tại nếu không có Evidence hỗ trợ.

---

# 3. Information Hierarchy

Market Intelligence tạo ra tri thức theo mô hình sau.

Raw Market Data

↓

Indicators

↓

Evidence

↓

Signals

↓

Analysis Result

↓

Market Context

↓

Market Analysis Result

Trong đó:

Indicators là dữ liệu quan sát.

Evidence là diễn giải dữ liệu.

Signals là kết luận cục bộ.

Analysis Result là kết quả của từng Analysis Module.

Market Context là bức tranh tổng hợp.

Market Analysis Result là sản phẩm cuối cùng của Capability.

---

# 4. Core Business Objects
### Definition

Indicator là dữ liệu đầu vào phản ánh trạng thái của thị trường.

Indicator chưa mang ý nghĩa đầu tư.

Nó chỉ là một quan sát.

Ví dụ:

- VNINDEX
- Trading Volume
- Trading Value
- Advance/Decline Ratio
- PMI
- CPI
- Credit Growth
- Interest Rate

Indicator có thể đến từ:

- Market Data
- Macro Data
- Alternative Data

Indicator không đưa ra kết luận.
### Definition

Evidence là diễn giải của Indicator dựa trên Business Rules.

Evidence trả lời câu hỏi:

"Dữ liệu này đang nói lên điều gì?"

Ví dụ:

Indicator

Trading Volume tăng 25%

↓

Evidence

Liquidity Improving

---

Indicator

PMI > 50

↓

Evidence

Manufacturing Expansion

---

Một Indicator có thể tạo nhiều Evidence.
Nhiều Indicator cũng có thể cùng tạo một Evidence.

### Definition

Signal là kết luận cục bộ được sinh ra từ một hoặc nhiều Evidence.

Signal mang ý nghĩa đầu tư.

Ví dụ

Evidence

Liquidity Improving

Foreign Buying

↓

Signal

Bullish Liquidity

---

Evidence

PMI Falling

Interest Rate Rising

↓

Signal

Economic Slowdown

Signal luôn có:

- Confidence
- Supporting Evidence
- Source Module

### Definition

Analysis Result là kết quả của một Analysis Module.

Ví dụ

Liquidity Analysis

↓

Liquidity Analysis Result

Macro Analysis

↓

Macro Analysis Result

Market Cycle Analysis

↓

Market Cycle Analysis Result

Mỗi Analysis Result bao gồm:

- Summary
- Score
- Confidence
- Supporting Evidence
- Signals

### Definition

Market Context là bức tranh tổng hợp về trạng thái của thị trường.

Đây là tri thức quan trọng nhất được Capability tạo ra.

Market Context tổng hợp:

- Market Trend
- Liquidity
- Breadth
- Capital Flow
- Macro Environment
- Market Cycle
- Risk Environment

Market Context không đưa ra khuyến nghị mua hoặc bán.

Nó chỉ mô tả môi trường đầu tư hiện tại.

### Definition

Market Analysis Result là sản phẩm cuối cùng của Market Intelligence Capability.

Đây là đối tượng mà các Capability khác sử dụng.

Ví dụ:

Financial Analysis

↓

Market Analysis Result

↓

Điều chỉnh giả định tăng trưởng

---

Valuation

↓

Market Analysis Result

↓

Điều chỉnh Discount Rate

---

Investment Decision

↓

Market Analysis Result

↓

Điều chỉnh Risk Score
