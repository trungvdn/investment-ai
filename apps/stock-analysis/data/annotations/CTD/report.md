# CTD — Ví dụ phân tích Wyckoff tại 30/10/2026

> Bản AI soạn, chờ người dùng duyệt. review.status=in_review; eligible_for_ai=false. Dựa trên OHLCV, chưa đối chiếu nhãn trên ảnh.

## 1. Thông tin và dữ liệu

- Khung 1D; cutoff 2026-10-30.
- [candles_CTD_1D.csv](candles_CTD_1D.csv): 497 phiên từ 2024-10-31 đến 2026-10-30, cột event trống.
- [CTD_nhan.png](CTD_nhan.png): ảnh để người duyệt đối chiếu.
- source_commit: 84b3251a4a28111bc9510332b6ea2b4a8a387561.
- Close cuối 64 theo đơn vị CSV. Chưa xác minh đơn vị, điều chỉnh giá và lịch giao dịch.
- RV20=volume/trung bình 20 phiên liền trước, loại phiên hiện tại.

## 2. Kết luận nhanh

**Ưu tiên giả thuyết tích lũy đang thử chuyển sang pha D, tin cậy trung bình định tính.** Breakout rộng 20/10 vượt 65 nhưng 29–30/10 quay xuống dưới biên; LPS chưa được xác nhận. Có thể phải hạ về pha B nếu không lấy lại vùng. Không kết luận markup bền vững chỉ vì đã có một nhịp vượt đỉnh.

## 3. Bối cảnh

- High 17/06=73.9; low 23/07=54.7 sau nhịp giảm tháng 7.\n- Hồi lên high 05/08=65; kiểm tra 17/09 low 56.8 giữ trên 54.7.\n- 20/10 close 66 vượt 65; 29/10 close 64.7 và 30/10 close 64 quay dưới vùng breakout.

Không suy diễn nguyên nhân doanh nghiệp hay hoạt động tổ chức từ OHLCV; chưa bổ sung ngành/vĩ mô vì thiếu dữ liệu trong mẫu.

## 4. Vùng giao dịch

| structure_id | Thời gian | Hỗ trợ | Kháng cự | Trạng thái |
|---|---|---|---|---|
| CTD_CONTEXT_202607 | Từ 22/07 | 54.7–56 | 63.2–65 | provisional; breakout đang kiểm tra |
| CTD_TR_202609 | 11/09–07/10 | 56.8–57.5 | 59.9–60.7 | provisional; breakout nhỏ |

Biên 65 có từ high 05/08; không thay bằng đỉnh mới 68.4 để xóa breakout 20/10. Ngày 07/10 vượt 60.7 của vùng nhỏ nhưng chưa vượt 65 của vùng rộng. end_date vùng nhỏ là mốc breakout ứng viên.

## 5. Nhật ký nhãn sự kiện

Mọi label_verdict=unreviewed, final_label=null. invalidated là đề xuất bác bỏ của AI. Nhãn cùng phiên có thể khác kết luận nếu structure_id khác.

### E001 — SC: 2026-07-22 → 2026-07-23

- structure_id: CTD_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 22/07 low=close=56; 23/07 low 54.7, close 56.9 ở 57.9% biên độ.
- Bằng chứng volume: 22/07 volume 661800 RV20 1.25; 23/07 458000 RV20 0.85.
- Phản ứng đến cutoff: 27/07 low 55 chưa phá 54.7; sau đó hồi lên 65 ngày 05/08.
- Bằng chứng phản đối/giới hạn: Volume không climax rõ, có thể chỉ là stopping action.
- Điều kiện xác nhận: Cần nhịp hồi và kiểm tra giữ đáy; người duyệt có thể bỏ SC nếu thiếu climax.
- Điều kiện bác bỏ: Không ép SC chỉ vì đáy; mất bền 54.7 làm suy yếu giả thuyết giữ đáy.
- event_status: uncertain; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E002 — AR: 2026-07-24 → 2026-08-05

- structure_id: CTD_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Sau kiểm tra low 55 ngày 27–28/07, hồi lên high 65, close 63.2 ngày 05/08.
- Bằng chứng volume: 05/08 volume 405300 RV20 0.82.
- Phản ứng đến cutoff: Sau đỉnh 65, giá điều chỉnh dần về vùng 56.8 tháng 9.
- Bằng chứng phản đối/giới hạn: Nhịp bắt đầu có kiểm tra đáy xen kẽ; phạm vi AR cần duyệt.
- Điều kiện xác nhận: Hồi sau stopping action tạo cạnh trên và phản ứng điều chỉnh.
- Điều kiện bác bỏ: Nếu thay đổi SC hoặc phân đoạn nhịp hồi thì sửa phạm vi AR.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E003 — ST: 2026-09-11 → 2026-09-17

- structure_id: CTD_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 11/09 low 57.5; 17/09 low 56.8 rồi close 58.8; cao hơn SC low 54.7 khoảng 3.84%.
- Bằng chứng volume: 17/09 volume 157100 RV20 0.66, bằng 34.3% volume 23/07.
- Phản ứng đến cutoff: Sau đó đi ngang và breakout vùng nhỏ 07/10.
- Bằng chứng phản đối/giới hạn: ST muộn và chưa sát đáy rộng; có thể là kiểm tra bên trong.
- Điều kiện xác nhận: Giữ đáy, cung giảm và phản ứng hồi, đánh giá trong bối cảnh vùng lớn.
- Điều kiện bác bỏ: Mất 54.7 bền phản đối ST thành công; thiếu quan hệ với SC thì bỏ nhãn.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E004 — SOS: 2026-10-20

- structure_id: CTD_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: O/H/L/C=63.3/66.2/63.3/66; close vượt 65 khoảng 1.54%.
- Bằng chứng volume: Volume 1272800 RV20 3.80.
- Phản ứng đến cutoff: 27/10 high 68.4; 29–30/10 close 64.7 và 64 quay dưới 65.
- Bằng chứng phản đối/giới hạn: Breakout có xảy ra nhưng giữ vùng chưa thành công; không đồng nhất thất bại sau đó với nhãn quan sát sai.
- Điều kiện xác nhận: Cần giữ/lấy lại 65 và kiểm tra có cầu hỗ trợ để củng cố pha D.
- Điều kiện bác bỏ: Giá chấp nhận lại trong vùng và mất 63.2 không phục hồi sẽ phản đối SOS thành công.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E005 — LPS: 2026-10-29 → 2026-10-30

- structure_id: CTD_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: 29/10 low 64.6 close 64.7; 30/10 low=close=64, dưới biên 65.
- Bằng chứng volume: 29/10 volume 883800 RV20 1.91; 30/10 419300 RV20 0.85.
- Phản ứng đến cutoff: Chưa có phiên hồi sau cutoff.
- Bằng chứng phản đối/giới hạn: Một phiên volume giảm sau phiên bán mạnh chưa chứng minh cạn cung; chưa giữ lại breakout.
- Điều kiện xác nhận: Chờ hồi lấy lại 65 và kiểm tra giữ vùng với cung giảm.
- Điều kiện bác bỏ: Giảm tiếp dưới 63.2 không lấy lại vùng thì không xác nhận LPS.
- event_status: uncertain; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E006 — Spring: 2026-09-17

- structure_id: CTD_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Low 56.8 cao hơn hỗ trợ đáy rộng 54.7, không xuyên đáy này.
- Bằng chứng volume: Volume thấp không thay thế điều kiện xuyên hỗ trợ.
- Phản ứng đến cutoff: Phản ứng hồi sau đó phù hợp ST ứng viên E003 hơn Spring rộng.
- Bằng chứng phản đối/giới hạn: Nếu xét vùng nhỏ khác cần biên đã tồn tại và structure_id riêng.
- Điều kiện xác nhận: Spring rộng cần xuyên hỗ trợ rộng rồi hồi lại.
- Điều kiện bác bỏ: Bác bỏ Spring vùng lớn 17/09 do không xuyên 54.7.
- event_status: invalidated; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E007 — SOS: 2026-10-07

- structure_id: CTD_CONTEXT_202607; nguồn: AI_OHLCV_example.
- Bằng chứng giá: Close 61.9 dưới kháng cự rộng 65.
- Bằng chứng volume: Volume 685300 RV20 2.82.
- Phản ứng đến cutoff: Sau đó mới tới breakout rộng ngày 20/10.
- Bằng chứng phản đối/giới hạn: Có breakout vùng nhỏ E008; chỉ bác bỏ SOS của vùng rộng cùng ngày.
- Điều kiện xác nhận: Vượt đúng kháng cự của structure_id đang xét và có follow-through.
- Điều kiện bác bỏ: Chưa vượt 65 nên không phải breakout rộng.
- event_status: invalidated; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

### E008 — SOS: 2026-10-07

- structure_id: CTD_TR_202609; nguồn: AI_OHLCV_example.
- Bằng chứng giá: O/H/L/C=59/62.4/59/61.9, vượt biên nhỏ 60.7 đã hình thành ngày 28/09.
- Bằng chứng volume: Volume 685300 RV20 2.82.
- Phản ứng đến cutoff: 08/10 close 62.3; 14/10 close 63.6; sau đó vượt vùng lớn 20/10.
- Bằng chứng phản đối/giới hạn: Chỉ áp dụng vùng nhỏ; không dùng để tuyên bố pha D rộng ngay 07/10.
- Điều kiện xác nhận: Breakout giữ 60.7 và có follow-through; kiểm tra xác nhận thêm chất lượng vùng nhỏ.
- Điều kiện bác bỏ: Quay bền dưới 59.9–60.7 sẽ làm suy yếu breakout nhỏ.
- event_status: candidate; label_verdict: unreviewed.
- Nhãn cuối/người duyệt/ngày duyệt: chưa điền.

## 6. Cung–cầu và pha

| Quan sát | Diễn giải | Giới hạn |
|---|---|---|
| 23/07 low 54.7, volume RV20 0.85 | Dừng giảm/SC ứng viên | Không có volume climax rõ |
| 17/09 low 56.8, RV20 0.66 | Cung giảm trên nhịp kiểm tra cao hơn đáy | Không phải Spring rộng |
| 07/10 close 61.9, RV20 2.82 | Breakout vùng nhỏ | Không đồng nghĩa vượt 65 |
| 20/10 close 66, RV20 3.80 | SOS rộng ứng viên | Cần giữ vùng sau breakout |
| 29/10 giảm mạnh, RV20 1.91; 30/10 close=low=64 | Cung quay lại, chưa có LPS xác nhận | Volume giảm 30/10 không tự chứng minh cạn cung |

Pha D là giả thuyết đang kiểm tra. Mất 65 chỉ làm suy yếu breakout trước; mất 54.7 bền mới đe dọa đáy toàn vùng. Không đồng nhất thất bại kịch bản với mọi nhãn sự kiện lịch sử đều sai.

## 7. Kịch bản

- **S1: Giữ breakout, củng cố pha D.** structure_id: CTD_CONTEXT_202607. Xác nhận: Lấy lại 65 qua hai close rồi kiểm tra giữ vùng với cung giảm; vượt 68.4 có follow-through là bằng chứng bổ sung. Bác bỏ: Chấp nhận dưới 63.2 và không lấy lại 65. Bằng chứng nền: E004, E005.

- **S2: Quay lại xây nền pha B.** structure_id: CTD_CONTEXT_202607. Xác nhận: Ở dưới 65 nhưng giữ 54.7; tiếp tục dao động chưa có hướng thoát vùng bền. Bác bỏ: Thoát bền trên 65 hoặc dưới 54.7. Bằng chứng nền: E002, E003, E004.

- **S3: Giả thuyết tích lũy thất bại.** structure_id: CTD_CONTEXT_202607. Xác nhận: Hai close dưới 54.7 và không phục hồi trong ba phiên sau; các nhịp hồi bị từ chối củng cố. Bác bỏ: Phục hồi nhanh vào vùng và kiểm tra giữ đáy. Bằng chứng nền: E001, E003.

- **S4: Breakout nhỏ giữ hiệu lực khi vùng rộng điều chỉnh.** structure_id: CTD_TR_202609. Xác nhận: Nếu kiểm tra 59.9–60.7, vùng giữ được với cung giảm và có nhịp hồi. Bác bỏ: Chấp nhận dưới 59.9 và không lấy lại 60.7. Bằng chứng nền: E008.

Các ngưỡng hai/ba phiên là minh họa chưa backtest, không phải quy tắc Wyckoff bắt buộc. Không gán xác suất hay mục tiêu giá thiếu cơ sở.

## 8. Duyệt mẫu

- Trạng thái: in_review; eligible_for_ai=false.
- Người duyệt/ngày duyệt: chưa điền; chưa có nhãn đúng/sai được người dùng xác nhận.
- Cần duyệt: SC cụm 22–23/07 còn uncertain; phạm vi AR; ST tháng 9; SOS nhỏ 07/10; SOS rộng 20/10 và pha D đang kiểm tra.
- LPS cuối tháng chưa đủ bằng chứng. Đề xuất bác bỏ Spring 17/09 và SOS 07/10 của vùng rộng.
- Đối chiếu ảnh và điều chỉnh giá; cập nhật final_label, label_verdict, correction_reason, reviewer trong JSON cùng báo cáo.

## 9. Kết quả tương lai — loại khỏi đầu vào AI

Chưa có dữ liệu sau 30/10/2026; evaluation_date=null, observations=[]. Không tự khẳng định kịch bản đã xảy ra.
