# GEX — Ví dụ phân tích Wyckoff tại 30/10/2026

> Bản AI soạn đã được người dùng review và xác nhận ngày 31/10/2026. review.status=approved; eligible_for_ai=true. Phân tích từ OHLCV; chưa đối chiếu nhãn trên ảnh.

## 1. Thông tin và dữ liệu

- Khung 1D; cutoff 2026-10-30.
- [candles_GEX_1D.csv](candles_GEX_1D.csv): 497 phiên từ 2024-10-31 đến 2026-10-30; cột event trống.
- [GEX_nhan.png](GEX_nhan.png): ảnh để người duyệt đối chiếu.
- source_commit: eca550eb32245f9ad0e4f9690c8dd3a95239e545.
- Close cuối 22.2 theo đơn vị CSV; chưa xác minh lịch giao dịch, đơn vị và điều chỉnh giá.
- RV20=volume/trung bình 20 phiên liền trước, loại phiên hiện tại.

## 2. Kết luận nhanh

**Cấu trúc chưa rõ; ưu tiên tích lũy tiềm năng đang xây nền pha B của vùng lớn, tin cậy thấp.** Đáy 19.7 tháng 7 chưa bị phá nhưng các nhịp hồi không vượt 26.85–27.2. Phải giữ giả thuyết cạnh tranh là tái phân phối trong markdown. Chưa có SOS/LPS vùng lớn được xác nhận; vùng nhỏ tháng 10 cũng chưa đủ điều kiện kết luận pha D.

## 3. Bối cảnh

- High 17/06=32.25; low 28/07=19.7 sau markdown tháng 7.\n- Hồi tới high 13/08=26.85 và 28/08=27.2; tháng 9–10 điều chỉnh nhưng chưa xuyên 19.7.\n- 30/10 close 22.2 nằm trong vùng rộng; chưa có breakout trên 27.2.

Không suy ra hoạt động tổ chức hay nguyên nhân doanh nghiệp từ OHLCV; chưa có dữ liệu ngành/vĩ mô trong bộ mẫu.

## 4. Vùng giao dịch

| structure_id | Thời gian | Hỗ trợ | Kháng cự | Trạng thái |
|---|---|---|---|---|
| GEX_CONTEXT_202607 | Từ 27/07 | 19.7–20.3 | 26.85–27.2 | provisional |
| GEX_TR_202610 | Từ 08/10 | 21.35–21.8 | 24.2–24.35 | provisional, ngắn |

Biên rộng hình thành theo thời gian; không giả vờ biết 27.2 từ đầu tháng 8. Biên nhỏ 24.35 chỉ được quan sát ngày 26/10; không dùng ngược để gán breakout 21/10. Đáy nhỏ 21.35 vẫn cao hơn đáy rộng 19.7.

## 5. Nhật ký nhãn sự kiện

Người dùng đã duyệt: PS/SC còn uncertain; AR/ST vùng lớn giữ mức ứng viên; E005–E008 được xác nhận bác bỏ, gồm AR vùng nhỏ E006 theo sửa trực tiếp trong JSON. Không ép đủ chuỗi sự kiện hoặc suy luận ngược từ kết quả tương lai.

### E001 — PS: 2026-07-22 → 2026-07-23

- structure_id: GEX_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 22/07 close=low=20.75; 23/07 low 20.2 rồi close=high=22.2.
- Bằng chứng volume: 23/07 volume 21127600 RV20 1.75.
- Phản ứng đến cutoff: Sau đó 28/07 vẫn tạo đáy thấp hơn 19.7.
- Bằng chứng phản đối/giới hạn: Có phản ứng cầu nhưng chưa chắc vai trò PS; cũng có thể là một phần cụm dừng giảm.
- Điều kiện xác nhận: Cần thay đổi hành vi cung–cầu trước cụm SC theo cùng cấu trúc.
- Điều kiện bác bỏ: Nếu người duyệt chọn SC bao gồm cả cụm này thì bỏ nhãn PS riêng.
- event_status: uncertain; label_verdict: uncertain.
- Nhãn cuối: chưa chốt (uncertain); người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E002 — SC: 2026-07-27 → 2026-07-28

- structure_id: GEX_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 27/07 low=close=20.3; 28/07 low 19.7, close 20.9 ở 68.6% biên độ.
- Bằng chứng volume: 27/07 volume 15477400 RV20 1.18; 28/07 volume 11534000 RV20 0.85.
- Phản ứng đến cutoff: Hồi tới vùng 26.85–27.2 trong tháng 8.
- Bằng chứng phản đối/giới hạn: Không có volume climax nổi bật tại đáy; SC cần duyệt, không chọn tự động vì đây là low thấp nhất.
- Điều kiện xác nhận: Stopping action có AR và các lần kiểm tra giữ đáy về sau.
- Điều kiện bác bỏ: Nếu thiếu bằng chứng climax, đổi sang stopping action/no_event; mất bền 19.7 làm suy yếu cấu trúc hiện tại.
- event_status: uncertain; label_verdict: uncertain.
- Nhãn cuối: chưa chốt (uncertain); người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E003 — AR: 2026-07-29 → 2026-08-13

- structure_id: GEX_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Nhịp hồi sau đáy: high 13/08=26.85 nhưng close 25.4 ở đáy phiên. 28/08 mở rộng high lên 27.2, theo dõi riêng thay vì tự kéo dài AR.
- Bằng chứng volume: 13/08 volume 16682000 RV20 1.22; 28/08 volume 15194000 RV20 1.31.
- Phản ứng đến cutoff: Sau đỉnh tháng 8, giá điều chỉnh tháng 9.
- Bằng chứng phản đối/giới hạn: Nhịp hồi kéo dài và có mở rộng sau đó; cần người duyệt xác nhận phạm vi AR.
- Điều kiện xác nhận: Hồi sau dừng giảm tạo cạnh trên, sau đó có phản ứng kiểm tra.
- Điều kiện bác bỏ: Nếu xác định AR kết thúc sớm hơn hoặc bao gồm tới 28/08 thì cập nhật ngày và biên tương ứng.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: AR; người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E004 — ST: 2026-09-14 → 2026-10-09

- structure_id: GEX_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 14/09 low=close=22.3; 09/10 low 21.35, close 22.25; đáy cụm còn cao hơn 19.7 khoảng 8.38%.
- Bằng chứng volume: 14/09 volume 24534400 RV20 2.73; 09/10 6410900 RV20 0.77.
- Phản ứng đến cutoff: 21/10 hồi close 23.85; đến 30/10 chưa xuyên 21.35.
- Bằng chứng phản đối/giới hạn: Đây là nhịp kiểm tra kéo dài gồm nhiều sóng; 14/09 cung mạnh, không gọi toàn cụm cạn cung hoặc ST đã hoàn tất.
- Điều kiện xác nhận: Giữ đáy rộng, có dấu hiệu cung giảm về cuối và phản ứng hồi; duyệt ST nông hay chỉ pullback.
- Điều kiện bác bỏ: Nếu markdown tiếp tục bền dưới 19.7, bác bỏ ST thành công; mất 21.35 là cảnh báo sớm.
- event_status: candidate; label_verdict: correct.
- Nhãn cuối: ST; người duyệt: trungvdn; ngày duyệt: 2026-10-31.

### E005 — Spring: 2026-10-09

- structure_id: GEX_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 21.35 cao hơn đáy vùng lớn 19.7, không có xuyên hỗ trợ rộng.
- Bằng chứng volume: Volume thấp hơn trung bình không bù điều kiện thiếu xuyên đáy.
- Phản ứng đến cutoff: Có hồi sau 09/10 nhưng không biến phiên này thành Spring của vùng lớn.
- Bằng chứng phản đối/giới hạn: Nếu dùng vùng nhỏ khác phải chỉ rõ structure_id và biên trước sự kiện.
- Điều kiện xác nhận: Spring của vùng lớn cần xuyên hỗ trợ đã hình thành rồi phục hồi.
- Điều kiện bác bỏ: Đề xuất bác bỏ Spring 09/10 ở GEX_CONTEXT_202607 do không xuyên 19.7.
- event_status: invalidated; label_verdict: incorrect.
- Nhãn cuối: no_event; người duyệt: trungvdn; ngày duyệt: 2026-10-31.\n- Kết quả duyệt: Bác bỏ nhãn

### E006 — AR: 2026-10-12 → 2026-10-26

- structure_id: GEX_TR_202610; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 21/10 close 23.85, 22/10 high 24.2; 26/10 high 24.35 nhưng close 23.3.
- Bằng chứng volume: 21/10 volume 18324000 RV20 2.98.
- Phản ứng đến cutoff: 27–30/10 điều chỉnh, close cuối 22.2.
- Bằng chứng phản đối/giới hạn: Vùng mới ngắn; có thể chỉ là hồi kỹ thuật trong markdown, không đủ khẳng định tích lũy mới.
- Điều kiện xác nhận: Nhịp hồi sau low 21.35 tạo cạnh trên và có pullback tiếp theo.
- Điều kiện bác bỏ: Nếu không hình thành nền và tiếp tục giảm, xem lại AR của giả thuyết tích lũy mới.
- event_status: invalidated; label_verdict: incorrect.
- Nhãn cuối: no_event; người duyệt: trungvdn; ngày duyệt: 2026-10-31.\n- Kết quả duyệt: Người dùng bác bỏ nhãn AR cho nhịp hồi 12–26/10 trong GEX_TR_202610.

- Diễn giải sau review: giữ quan sát nhịp hồi giá 12–26/10, bỏ nhãn AR. Các điều kiện phía trên là giả thuyết gốc để lưu lịch sử.

### E007 — SOS: 2026-10-21

- structure_id: GEX_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Close 23.85 dưới kháng cự rộng 26.85–27.2.
- Bằng chứng volume: Volume 18324000 RV20 2.98; volume lớn riêng lẻ không tạo breakout vùng rộng.
- Phản ứng đến cutoff: Nhịp hồi cao nhất 26/10=24.35 vẫn thấp hơn đỉnh tháng 8.
- Bằng chứng phản đối/giới hạn: Chỉ bác bỏ SOS của vùng rộng, không phủ nhận phiên có cầu tăng hoặc breakout vùng cực nhỏ khác.
- Điều kiện xác nhận: Phải vượt đúng biên TR đã xác định và có follow-through.
- Điều kiện bác bỏ: Chưa vượt 26.85–27.2 nên bác bỏ SOS vùng lớn.
- event_status: invalidated; label_verdict: incorrect.
- Nhãn cuối: no_event; người duyệt: trungvdn; ngày duyệt: 2026-10-31.\n- Kết quả duyệt: Bác bỏ nhãn

### E008 — LPS: 2026-10-27 → 2026-10-30

- structure_id: GEX_TR_202610; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Pullback về close=low=22.2 ngày 30/10, chưa có SOS hợp lệ của TR mới trước đó.
- Bằng chứng volume: 30/10 volume 6788200 RV20 0.89; volume giảm không tự tạo LPS.
- Phản ứng đến cutoff: Không có dữ liệu hồi sau 30/10.
- Bằng chứng phản đối/giới hạn: Có thể là kiểm tra bên trong chưa xác nhận; không gán LPS khi thiếu SOS.
- Điều kiện xác nhận: Cần SOS trước và pullback giữ breakout có phản ứng cầu.
- Điều kiện bác bỏ: Thiếu SOS và hồi xác nhận: đề xuất bác bỏ LPS ở cutoff.
- event_status: invalidated; label_verdict: incorrect.
- Nhãn cuối: no_event; người duyệt: trungvdn; ngày duyệt: 2026-10-31.\n- Kết quả duyệt: Bác bỏ nhãn.

## 6. Cung–cầu và pha

| Quan sát | Diễn giải | Giới hạn |
|---|---|---|
| 28/07 low 19.7, close 20.9; RV20 0.85 | Phản ứng dừng giảm ứng viên | Chưa đủ volume climax cho SC chắc chắn |
| 14/09 RV20 2.73, close=low=22.3 | Cung còn mạnh trên nhịp kiểm tra | Không gọi cả cụm ST cạn cung |
| 09/10 volume thấp hơn trung bình, low 21.35 | Có giảm nỗ lực bán so 14/09 | Không xuyên 19.7 nên không phải Spring vùng rộng |
| 21/10 RV20 2.98, close 23.85 | Có cầu trên nhịp hồi | Dưới kháng cự rộng nên chưa là SOS rộng |
| 30/10 close=low=22.2 | Pullback chưa có cầu xác nhận | Volume giảm không đủ tạo LPS |

Pha B chỉ là giả thuyết ưu tiên, không kết luận accumulation đã xác nhận. Nếu mất đáy rộng hoặc breakout phía trên, phải cập nhật cấu trúc/pha theo bằng chứng mới.

## 7. Kịch bản

- **S1: Tích lũy rộng tiếp tục xây nền.** structure_id: GEX_CONTEXT_202607. Xác nhận: Giữ 19.7; kiểm tra sâu/nông có cung giảm và phản ứng hồi; chưa vượt bền 27.2. Bác bỏ: Thoát vùng bền theo một hướng; mất 19.7 phản đối tích lũy giữ đáy. Bằng chứng nền: E002, E003, E004.

- **S2: Breakout vùng rộng và chuyển pha.** structure_id: GEX_CONTEXT_202607. Xác nhận: Hai close trên 27.2, có follow-through và kiểm tra lại giữ 26.85–27.2; sau đó mới xem xét SOS/LPS rộng. Bác bỏ: Quay nhanh vào vùng, không lấy lại biên và mất đáy kiểm tra. Bằng chứng nền: E003, E004.

- **S3: Markdown/tái phân phối tiếp diễn.** structure_id: GEX_CONTEXT_202607. Xác nhận: Mất 21.35 là cảnh báo; hai close dưới 19.7 không lấy lại trong ba phiên sau sẽ bác bỏ giả thuyết giữ đáy. Bác bỏ: Phục hồi nhanh vào vùng và kiểm tra giữ đáy; đánh giá khả năng Spring mới bằng dữ liệu thực tế. Bằng chứng nền: E002, E004.

- **S4: Vùng nhỏ giữ nền và hồi phục.** structure_id: GEX_TR_202610. Xác nhận: Giữ 21.35–21.8 rồi hồi vượt 24.35 và kiểm tra giữ được; điều này chưa phải breakout vùng rộng. Bác bỏ: Mất 21.35 bền, không có phản ứng cầu. Bằng chứng nền: E006, E008.

Các ngưỡng hai/ba phiên là minh họa chưa backtest, không phải quy tắc Wyckoff bắt buộc. Chưa gán xác suất hay mục tiêu giá.

## 8. Duyệt mẫu

- Trạng thái: approved; eligible_for_ai=true.
- Người duyệt: trungvdn; ngày duyệt: 2026-10-31.
- Phạm vi: GEX_CONTEXT_202607, GEX_TR_202610, 8 sự kiện và 4 kịch bản.
- E001 PS, E002 SC: giữ uncertain. E003 AR và E004 ST vùng lớn: giữ nhãn ứng viên.
- Xác nhận bác bỏ E005 Spring vùng lớn, E006 AR vùng nhỏ, E007 SOS vùng lớn và E008 LPS vùng nhỏ theo chỉnh sửa người dùng.
- no_event chỉ áp dụng nhãn và phạm vi bị bác bỏ, không xóa quan sát OHLCV. Tham chiếu E006 trong kịch bản là bằng chứng của nhịp hồi, không phải nhãn AR đã được chấp nhận.
- Assessment giữ structure=unclear, pha B là giả thuyết tích lũy tiềm năng. Giới hạn xác minh ảnh/điều chỉnh dữ liệu và ngưỡng chưa backtest giữ nguyên; các kịch bản tương lai vẫn là giả thuyết.

## 9. Kết quả tương lai — loại khỏi đầu vào AI

Không có dữ liệu sau 30/10/2026. evaluation_date=null và observations=[]; kịch bản chưa phải diễn biến đã xảy ra.
