# BCM — Ví dụ phân tích Wyckoff tại 30/10/2026

> Bản AI soạn đã được người dùng review và xác nhận ngày 31/10/2026. review.status=approved; eligible_for_ai=true. Dựa trên OHLCV, chưa đối chiếu nhãn trên ảnh.

## 1. Thông tin và dữ liệu

- Khung 1D; cutoff 2026-10-30.
- [candles_BCM_1D.csv](candles_BCM_1D.csv): 497 phiên từ 2024-10-31 đến 2026-10-30; event trống.
- [BCM_nhan.png](BCM_nhan.png): ảnh để người duyệt đối chiếu.
- source_commit: 58cded756648bbe4b7766167a840aa0c2c3df784.
- Close cuối 38.4 theo đơn vị CSV; chưa xác minh lịch giao dịch, đơn vị và điều chỉnh giá.
- RV20=volume/trung bình 20 phiên liền trước, loại phiên hiện tại.

## 2. Kết luận nhanh

**Ưu tiên tích lũy tiềm năng pha B, cấu trúc chưa xác nhận; tin cậy trung bình định tính.** SC ứng viên 27/07, AR đạt đỉnh 13/08; ST 07/10 và 14/10 giữ vùng kiểm tra gần. Nhịp tăng tháng 10 chưa vượt kháng cự rộng 45.5–47 nên chưa có SOS/LPS rộng hợp lệ. Breakout vùng nhỏ tháng 10 đang bị kiểm tra thất bại ở cutoff.

## 3. Bối cảnh

- Close 01/07=52.8; 27/07=35.1, ngày 28/07 low 33.9.\n- Nhịp hồi đạt high 47 và close 45.5 ngày 13/08.\n- 07/10 và 14/10 low 37, giữ trên low 06/10=36.5 và trên cụm đáy tháng 7.

Không suy diễn nguyên nhân doanh nghiệp hoặc hoạt động tổ chức; chưa thêm ngành/vĩ mô do thiếu dữ liệu trong mẫu.

## 4. Vùng giao dịch và cực trị

| Phạm vi | Hỗ trợ | Kháng cự | Thời gian |
|---|---|---|---|
| BCM_CONTEXT_202607 — cực trị vùng rộng | 33.9–35.1 | 45.5–47 | Từ 27/07 |
| Vùng tham chiếu làm việc của vùng rộng | 36.5 | 45.5 | Ghi nhận tại cutoff; 36.5 quan sát từ 06/10 |
| BCM_TR_202610 — vùng nhỏ | 36.5–37 | 39.7–40.5 | Từ 06/10 |

Đáy 27/07=35.1 bị xuyên ngày 28/07=33.9; AR 13/08 high 47 nhưng close 45.5. Vùng tham chiếu 36.5–45.5 là quy ước diễn giải, không thay thế cực trị. Không dùng hỗ trợ 36.5 để gán Spring trước 06/10. Biên nhỏ 40.5 hình thành ngày 19/10 trước breakout 20/10.

## 5. Nhật ký nhãn sự kiện

Người dùng đã duyệt: PS/LPS còn uncertain; các nhãn candidate giữ mức ứng viên; xác nhận bác bỏ SOS vùng lớn E006. Sự kiện có diễn ra nhưng kịch bản thành công thất bại không đồng nghĩa mọi nhãn gốc đều sai.

### E001 — PS: 2026-07-23

- structure_id: BCM_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 36.7, close 38.4 sau nhịp giảm mạnh; đóng ở 68% biên độ.
- Bằng chứng volume: Volume 610200 RV20 1.68.
- Phản ứng đến cutoff: 27–28/07 tiếp tục tạo đáy thấp hơn.
- Bằng chứng phản đối/giới hạn: Phản ứng cầu tạm thời chưa đủ khẳng định PS tách biệt.
- Điều kiện xác nhận: Cần quan hệ với cụm SC sau đó và thay đổi hành vi.
- Điều kiện bác bỏ: Nếu chỉ là hồi trong markdown hoặc thuộc cụm SC thì bỏ PS riêng.
- event_status: uncertain; label_verdict: uncertain.
- Nhãn cuối: chưa chốt (uncertain); người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E002 — SC: 2026-07-27

- structure_id: BCM_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: O/H/L/C=37.7/38/35.1/35.1; phiên giảm mạnh đóng đáy.
- Bằng chứng volume: Volume 934600 RV20 2.41.
- Phản ứng đến cutoff: 28/07 low 33.9, close 35.55 rồi hồi; đến 13/08 high 47.
- Bằng chứng phản đối/giới hạn: 27/07 không phải đáy cuối cùng; đọc cùng phản ứng 28/07, không bỏ qua lower low.
- Điều kiện xác nhận: Nhịp bán mạnh cuối markdown được theo sau bởi dừng giảm/AR và kiểm tra giữ đáy.
- Điều kiện bác bỏ: Nếu markdown tiếp tục bền dưới 33.9, xem lại SC mở đầu cấu trúc hiện tại.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: SC; người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E003 — AR: 2026-07-28 → 2026-08-13

- structure_id: BCM_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Nhịp hồi bắt đầu với low 33.9 ngày 28/07; đỉnh 13/08 high 47, close 45.5.
- Bằng chứng volume: 13/08 volume 1809000 RV20 2.30.
- Phản ứng đến cutoff: Sau 13/08 điều chỉnh về vùng kiểm tra tháng 9–10.
- Bằng chứng phản đối/giới hạn: Ngày đầu AR còn có lower low; phạm vi nhịp phải được người duyệt kiểm tra.
- Điều kiện xác nhận: Phản ứng hồi sau SC hình thành cạnh trên và điều chỉnh tiếp theo.
- Điều kiện bác bỏ: Nếu sửa cụm SC/điểm bắt đầu hồi, điều chỉnh ngày AR tương ứng.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: AR; người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E004 — ST: 2026-10-07

- structure_id: BCM_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Sau low 06/10=36.5, ngày 07/10 low 37, close 38.5; chưa phá đáy rộng 33.9.
- Bằng chứng volume: Volume 387600 RV20 0.70, bằng 41.5% volume SC 27/07.
- Phản ứng đến cutoff: 08/10 high 39.7, sau đó kiểm tra lần nữa 14/10.
- Bằng chứng phản đối/giới hạn: Đây là kiểm tra nông so đáy tháng 7; cần xét cả cụm 06–07/10.
- Điều kiện xác nhận: Giữ đáy kiểm tra gần, cung giảm và có phản ứng hồi.
- Điều kiện bác bỏ: Nếu mất 36.5 không hồi, nhịp ST gần suy yếu; mất bền 33.9 đe dọa vùng lớn.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: ST; người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E005 — ST: 2026-10-14

- structure_id: BCM_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 37 bằng low 07/10; close 37.9, trên 36.5.
- Bằng chứng volume: Volume 258400 RV20 0.54, thấp hơn 07/10 khoảng 33.3%.
- Phản ứng đến cutoff: 20–23/10 hồi tới close 41.65; cuối tháng quay xuống 38.4.
- Bằng chứng phản đối/giới hạn: ST có phản ứng hồi nhưng chưa dẫn đến breakout vùng lớn.
- Điều kiện xác nhận: Kiểm tra giữ đáy với cung giảm và hồi sau đó.
- Điều kiện bác bỏ: Nếu giá chấp nhận dưới 36.5, hạ đánh giá ST thành công; không tự gọi Spring nếu chưa xuyên đúng biên đã có.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: ST; người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E006 — SOS: 2026-10-23

- structure_id: BCM_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Close 41.65, high 41.9 vẫn dưới kháng cự rộng 45.5–47.
- Bằng chứng volume: Volume 1298500 RV20 2.60; volume lớn riêng lẻ không tạo breakout rộng.
- Phản ứng đến cutoff: 26/10 high 42.4 rồi 30/10 close 38.4.
- Bằng chứng phản đối/giới hạn: Có cầu trên nhịp hồi nhưng không phải SOS vùng lớn.
- Điều kiện xác nhận: Cần vượt đúng kháng cự vùng lớn, giữ vùng và có follow-through.
- Điều kiện bác bỏ: Chưa vượt 45.5–47 nên đề xuất bác bỏ SOS rộng.
- event_status: invalidated; label_verdict: incorrect.
- Nhãn cuối: no_event; người duyệt: trungvdn; ngày duyệt: 2026-10-31.\n- Ghi chú duyệt: Người dùng xác nhận bác bỏ SOS rộng 23/10 vì chưa vượt kháng cự 45.5–47; không phủ nhận cầu trên nhịp hồi.

### E007 — SOS: 2026-10-20 → 2026-10-21

- structure_id: BCM_TR_202610; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 20/10 close 40.9 vượt high 19/10=40.5; 21/10 close 41.2.
- Bằng chứng volume: 20/10 volume 857100; 21/10 979900; 23/10 tiếp tục tăng volume 1298500.
- Phản ứng đến cutoff: 23/10 close 41.65; 28–30/10 quay lại dưới 40.5.
- Bằng chứng phản đối/giới hạn: TR nhỏ ngắn; breakout không giữ được đến cutoff. Không dùng thành công tạm thời để xác nhận pha D rộng.
- Điều kiện xác nhận: Cần giữ/lấy lại 40.5 sau kiểm tra để củng cố breakout nhỏ.
- Điều kiện bác bỏ: Chấp nhận lại trong vùng và mất 36.5 sẽ phản đối SOS nhỏ thành công.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: SOS; người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E008 — LPS: 2026-10-27 → 2026-10-30

- structure_id: BCM_TR_202610; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Close giảm từ 40.55 ngày 27/10 về 38.4 ngày 30/10; mất biên breakout 40.5.
- Bằng chứng volume: Volume giảm đến 492800 ngày 30/10 RV20 0.85, nhưng close=low.
- Phản ứng đến cutoff: Chưa có hồi sau 30/10 trong dữ liệu.
- Bằng chứng phản đối/giới hạn: Volume giảm có thể do thiếu cầu; chưa chứng minh cạn cung hoặc LPS thành công.
- Điều kiện xác nhận: Chờ giữ vùng hỗ trợ và hồi lấy lại 40.5 với bằng chứng cầu.
- Điều kiện bác bỏ: Giảm tiếp không có hồi xác nhận thì không xác nhận LPS; mất 36.5 làm xấu vùng nhỏ.
- event_status: uncertain; label_verdict: uncertain.
- Nhãn cuối: chưa chốt (uncertain); người duyệt: trungvdn; ngày duyệt: 2026-10-31.

## 6. Cung–cầu và pha

| Quan sát | Diễn giải | Giới hạn |
|---|---|---|
| 27/07 giảm mạnh, RV20 2.41; 28/07 tạo low thấp hơn | SC ứng viên cần đọc theo cụm dừng giảm | 27/07 không phải đáy cuối cùng |
| 13/08 high 47, close 45.5, RV20 2.30 | Nhịp AR tạo cạnh trên | Không tự biến hồi sau markdown thành SOS |
| 07/10 và 14/10 cùng low 37; volume lần sau giảm 33.3% | Cung giảm trên kiểm tra gần | Chưa có breakout rộng xác nhận tích lũy |
| 23/10 RV20 2.60 nhưng close 41.65 | Có cầu hồi bên trong vùng | Chưa vượt 45.5–47 |
| 30/10 close=low=38.4 | Pullback còn yếu | Volume giảm không tự xác nhận LPS |

Pha B là giả thuyết vùng lớn; không gán pha D rộng chỉ từ breakout vùng nhỏ. Chưa gán Spring/Test hoặc LPS cũ từ chu kỳ khác vào vùng hiện tại.

## 7. Kịch bản

- **S1: Tiếp tục tích lũy tiềm năng pha B.** structure_id: BCM_CONTEXT_202607. Xác nhận: Giữ 33.9, phản ứng tích cực quanh 36.5–37 và chưa có breakout bền khỏi 45.5–47. Bác bỏ: Thoát vùng bền vững theo một hướng; mất 36.5 làm suy yếu ST gần, chưa tự mất vùng rộng. Bằng chứng nền: E002, E003, E004, E005.

- **S2: Breakout vùng lớn, chuyển pha.** structure_id: BCM_CONTEXT_202607. Xác nhận: Hai close trên 47, có follow-through và kiểm tra giữ 45.5–47; vượt 45.5 đơn lẻ chỉ là tín hiệu sớm. Bác bỏ: Quay vào vùng, không lấy lại kháng cự và phá đáy pullback. Bằng chứng nền: E003, E005.

- **S3: Giả thuyết tích lũy thất bại, markdown tiếp diễn.** structure_id: BCM_CONTEXT_202607. Xác nhận: Mất 36.5 là cảnh báo; hai close dưới 33.9, không lấy lại trong ba phiên sau, phản đối giữ đáy toàn vùng. Bác bỏ: Hồi nhanh vào vùng và kiểm tra giữ đáy; chỉ khi dữ liệu thực tế có mới đánh giá Spring. Bằng chứng nền: E002, E004, E005.

- **S4: Khôi phục breakout nhỏ.** structure_id: BCM_TR_202610. Xác nhận: Lấy lại 40.5 qua hai close, kiểm tra giữ được rồi vượt 42.4 có follow-through. Bác bỏ: Không lấy lại vùng và mất 36.5 bền. Bằng chứng nền: E007, E008.

Ngưỡng hai/ba phiên là minh họa chưa backtest; không phải quy tắc Wyckoff bắt buộc. Không gán xác suất hoặc mục tiêu giá thiếu cơ sở.

## 8. Duyệt mẫu

- Trạng thái: approved; eligible_for_ai=true.
- Người duyệt: trungvdn; ngày duyệt: 2026-10-31.
- Phạm vi: BCM_CONTEXT_202607, BCM_TR_202610, 8 sự kiện và 4 kịch bản.
- E001 PS và E008 LPS giữ uncertain; chưa chốt nhãn cuối.
- E002 SC 27/07, E003 AR đến 13/08, E004 ST 07/10, E005 ST 14/10 và E007 SOS nhỏ giữ nhãn ứng viên. Breakout nhỏ có phản ứng thất bại, không tự chuyển confirmed.
- E006 SOS rộng 23/10 bị bác bỏ; no_event chỉ áp dụng nhãn/phạm vi này.
- Giữ giả thuyết tích lũy tiềm năng pha B và phân biệt vùng tham chiếu 36.5–45.5 với cực trị 33.9–47.
- Giới hạn xác minh ảnh, điều chỉnh giá và ngưỡng chưa backtest giữ nguyên; kịch bản tương lai vẫn là giả thuyết.

## 9. Kết quả tương lai — loại khỏi đầu vào AI

Không có dữ liệu sau 30/10/2026; evaluation_date=null, observations=[]. Kịch bản chưa phải kết quả đã xảy ra.
