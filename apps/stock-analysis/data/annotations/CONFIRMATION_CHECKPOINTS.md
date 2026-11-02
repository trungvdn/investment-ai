# Mốc bằng chứng cho MSB, HDG, NLG

Các ngày sau là **mốc quan sát bằng chứng**, không khẳng định ngày xác nhận thị trường sớm nhất. Không thay đổi trạng thái ứng viên/chưa chắc chắn đã duyệt.

| Mã / sự kiện | Ngày bằng chứng | Quan sát OHLCV | Ý nghĩa giới hạn |
|---|---|---|---|
| MSB E001 Spring | 2026-08-28 | Close 13.35 > biên cũ 12.808 | Phục hồi sau cụm xuyên đáy; điều chỉnh giá vẫn chưa xác minh |
| MSB E003 SOS | 2026-09-24 | Close 14.25 > 13.726 | Follow-through sau breakout 23/09 |
| MSB E004 LPS | 2026-10-08 | Low 05/10 = 13.85 > 13.726; close 08/10 = 14.75 | Pullback giữ vùng và hồi; củng cố markup sau reaccumulation |
| HDG E005 | 2026-10-05 | Close 02/10 = 15.05 và 05/10 = 14.65 đều < 15.2 | Hai phiên đóng dưới vùng cũ; tích lũy thất bại, không tự gán SOW của cấu trúc phân phối |
| HDG E005 đã duyệt | 2026-10-30 | `confirmed_as_of` có sẵn trong JSON | Giữ nguyên mốc xác nhận thị trường đã duyệt |
| NLG E003 vùng nhỏ | 2026-10-15 | Low 21.0 < 21.3 nhưng > 19.15 | Xuyên hỗ trợ vùng nhỏ; chưa xuyên hỗ trợ vùng lớn |
| NLG E005 / E008 | 2026-10-23 | Close 23.3 > 22.5; high 23.75 < 25.95 | Follow-through vùng nhỏ không phải breakout vùng bối cảnh lớn |

Trong ba JSON, `label_confirmed_as_of=2026-10-30` chỉ là mốc xác nhận **nhãn hồi cứu bảo thủ**, khác với `confirmed_as_of` của sự kiện thị trường. Không dùng ngày người duyệt 31/10 làm ngày thị trường xác nhận. Các nhãn uncertain tiếp tục để trống ngày xác nhận. Muốn đo độ trễ xác nhận sớm nhất phải duyệt từng prefix OHLCV và xác minh điều chỉnh giá.
