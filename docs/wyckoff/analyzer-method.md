# Phương pháp analyzer: bối cảnh trước, sự kiện sau

Trạng thái: phương pháp luật thử nghiệm; chưa phải mô hình đã hiệu chỉnh xác suất. Tài liệu này mô tả **luồng chạy mặc định** trong `apps/stock-analysis/internal/wyckoff/context_first.go`, không phải cam kết mọi sơ đồ Wyckoff đều có thể tự động nhận diện.

## 1. Vấn đề của phương pháp trước

- Quét nến để gán SC/BC/Spring trước khi xác định chiến dịch; một nến khối lượng cao có thể bị gọi sai vai trò.
- Các cửa sổ 12–20 phiên có thể chiếm quyền quyết định dù không đủ bối cảnh.
- Dùng đỉnh breakout để thay kháng cự AR hoặc dùng đáy Spring để thay hỗ trợ cũ, xóa mất chính điều kiện cần kiểm tra.
- Gộp ứng viên của nhiều vùng vào một dòng sự kiện; xuất hiện SC và BC trái ngữ cảnh, nhiều SOS/LPS trên cùng biểu đồ.
- Pha và swing chưa là đầu ra chính thức; thử nghiệm cuối chuỗi dễ che việc chưa thể nhận diện ở thời điểm trước.

Luồng gán nhãn mặc định cũ đã được loại bỏ. Không lấy kết quả cũ rồi chỉ lọc marker cho đẹp.

## 2. Trình tự quyết định

1. Kiểm tra và sắp xếp OHLCV; không nhận ngày trùng, NaN/Inf hoặc OHLC bất hợp lệ.
2. Khởi đầu với **40 phiên gần nhất**. Dưới 40 phiên trả lỗi thiếu dữ liệu.
3. Đọc swing xác nhận, hướng đi trước đó và vùng dừng xu hướng.
4. Dựng chiến dịch từ cụm dừng bán/mua và phản ứng AR; hoặc vùng tiếp diễn có ít nhất 40 quan sát hình thành.
5. Đóng băng biên của từng chiến dịch trước khi kiểm tra Spring/UTAD và SOS/SOW.
6. Theo dõi ST, kiểm tra biên, breakout/breakdown và pullback; mọi sự kiện có cùng `structure_id` và ngày đủ dữ liệu `observed_at`.
7. Chọn chiến dịch phù hợp **bối cảnh hiện tại** rồi đánh giá giai đoạn đang hình thành, bằng chứng phản đối và phương án khác.
8. Chỉ vẽ sự kiện của chiến dịch chính. Cấu trúc lớn/nhỏ và ứng viên khác nằm riêng trong `structures`.

## 3. Mở rộng và thu hẹp

`analysisWindow.baseBars=40` luôn là điểm xuất phát. Bộ tìm kiếm mở rộng tối đa 160 phiên để tìm cụm dừng xu hướng, và có thể đọc thêm bối cảnh 40 phiên trước cụm đó. `usedBars`, `start`, `end`, `reasons` công bố phạm vi thực dùng; không tự gọi một vùng 12 phiên là toàn bộ bối cảnh.

Vùng sự kiện có thể thu hẹp xuống 12/20 phiên **chỉ sau khi** một cấu trúc lớn đã được xác lập từ bối cảnh trên. Vùng nhỏ phải nằm trong biên vùng lớn, có `parent_structure_id`, không mất hỗ trợ hai phiên trước breakout. Một chuyển tiếp mới có bằng chứng Spring/SOS được ưu tiên hơn một đỉnh cũ chưa xác nhận phân phối. Không ghép SOS vùng nhỏ thành breakout của vùng lớn.

Các mức hiện tại lấy từ cực trị OHLC của SC/AR. Đây có thể khác vùng vận hành do con người chọn theo thân nến/lần kiểm tra lặp lại. Ví dụ BCM có cực trị 33.9–47.0 trong CSV, khác vùng thủ công 36.5–45.5. Giữ nguyên nhãn thủ công và báo khác biệt; không hardcode giá để tăng điểm.

## 4. Swing và HH/HL, LH/LL

- Pivot cần hai quan sát bên trái và hai quan sát bên phải.
- Bỏ các dao động rất nhỏ: tối thiểu hai phiên giữa hai đầu swing, biên dịch chuyển tối thiểu 2%.
- Đỉnh/đáy cùng loại được so với đỉnh/đáy trước đó; ngưỡng phân biệt là 0.3% để tránh đếm mức gần bằng nhau thành HH/HL.
- Hai đỉnh và hai đáy đủ điều kiện mới xác nhận `HH_HL` hoặc `LH_LL`.
- Nhịp cuối chưa đủ hai phiên bên phải không trở thành swing xác nhận. Xu hướng tăng/giảm đang hình thành khi chưa đủ swing được ghi rõ là chưa xác nhận.
- HH/HL trong vùng đi ngang không tự động là markup: cần thoát vùng hoặc bằng chứng tiếp diễn phù hợp.

UI có lớp HH/HL–LH/LL riêng và chỉ vẽ swing đã có đủ dữ liệu xác nhận tới cutoff.

## 5. Quy tắc sự kiện

| Sự kiện | Điều kiện chính | Điều không được làm |
|---|---|---|
| PS / PSY | Nỗ lực dừng xu hướng trước climax, có volume và vị trí đóng cửa hỗ trợ giả thuyết | Không ép mọi chiến dịch phải có |
| SC | Cụm đáy mới của bối cảnh 40 phiên sau nhịp giảm; kết hợp mức giảm và RV20; giữ trường hợp hấp thụ volume lớn | Không gọi mọi HL là SC; không chọn phiên giảm nhỏ muộn nhất |
| BC | Cụm đỉnh mới của bối cảnh 40 phiên sau nhịp tăng; cần effort/result đủ mạnh | Một đỉnh hồi hoặc breakout đóng mạnh không tự động là BC |
| AR | Cực trị phản ứng đầu sau cụm SC/BC trong tối đa 15 phiên có sẵn; cần hai phiên sau để xác nhận pivot | Không thay AR bằng đỉnh breakout về sau |
| ST | Kiểm tra biên dừng xu hướng sau AR; so effort với climax, biên hẹp hơn và có phản ứng tiếp theo | Không chỉ dựa vào RV20 < 1; không gán ST cùng phiên xuyên biên thành Spring |
| Spring | Xuyên hỗ trợ đã tồn tại, hồi lại trong tối đa ba phiên sau; phiên effort cao cần phản ứng phục hồi ở phiên sau | Không hồi tố hạ hỗ trợ xuống đáy Spring; HL cao hơn đáy Spring cũ không thành một Spring mới |
| Test | Sau Spring/UTAD, giữ cực trị và biên gốc, volume giảm, gần vùng kiểm tra, trong 15 phiên | Không ép có Test; mất cực trị sau recovery sẽ bác bỏ chuỗi Test |
| SOS | Đóng vượt kháng cự gốc, vị trí đóng cửa mạnh và RV20 tăng; vừa rời vùng gần đây | Không gọi mọi phiên tăng trong markup là SOS mới |
| LPS | Pullback thực sau SOS trong 15 phiên, gần kháng cự cũ, effort giảm; hai phiên phản ứng kế tiếp giữ vùng | Không gọi nến cuối đang giảm hoặc pullback xa vùng cũ là LPS đã đủ điều kiện |
| UT / UTAD | Xuyên kháng cự rồi quay vào vùng phân phối; UT sớm, UTAD sau quá trình kiểm tra lặp lại | Không lấy mọi đỉnh mới trong markup làm UTAD |
| SOW | Breakdown của vùng phân phối hoặc vùng sau một xu hướng tăng có bằng chứng yếu đi | Không đổi việc tích lũy thất bại thành SOW một cách máy móc |
| LPSY | Pullback yếu sau SOW tới hỗ trợ cũ, có phản ứng giảm tiếp xác nhận | Không gán trước SOW |

RV20 dùng **20 phiên trước**, không bao gồm phiên đang xét. Các ngưỡng RV, CLV, độ xuyên biên, khoảng cách và thời gian trong code là tham số nghiên cứu, không phải định luật Wyckoff đã được kiểm định.

Breakout liên tiếp được gộp và ghi ngày ở phiên cuối còn đạt điều kiện, không tạo một cột SOS trùng nhau. Hệ thống không bắt buộc Spring trong mọi tích lũy, cũng không bắt buộc đủ SC/AR/ST của vùng nhỏ nếu các sự kiện đó thuộc vùng lớn.

## 6. Giai đoạn đang hình thành

| Giai đoạn | Cách diễn giải |
|---|---|
| A | Nỗ lực dừng xu hướng/AR đang hình thành; climax volume lớn có thể là ứng viên trước khi có AR |
| B | Kiểm tra hai biên và xây vùng; khi chưa phân biệt tích lũy với tái phân phối, ghi rõ bias chưa chốt |
| C | Spring/Test hoặc UTAD trong đúng vùng; chưa phải xác nhận breakout |
| D | SOS/SOW và kiểm tra mức phá vùng; quay lại biên là bằng chứng phản đối, không tự sinh một BC mới |
| E | Markup/markdown với swing hoặc breakout và pullback giữ mức đủ mạnh; nếu thiếu HH/HL xác nhận phải ghi đang hình thành |

Các chiến dịch được xếp theo độ bao phủ dữ liệu gần nhất, đầy đủ ngữ cảnh dừng xu hướng, kiểm tra biên, chuyển tiếp gần đây và khả năng biên còn hiệu lực. Cấu trúc cũ có SOW và LH/LL còn tiếp diễn có thể là bối cảnh lớn; cụm dừng giảm mới với effort rõ ràng được phép trở thành giả thuyết hiện tại. Một vùng chồng lấn chưa rõ hướng không được đánh bại recovery có bằng chứng chỉ vì rộng hơn.

Hai lần đóng vượt biên theo hướng bất lợi làm thất bại giả thuyết. Giữ sự kiện lịch sử, đánh dấu `failed_outcome` và ký hiệu ×, thay vì xóa lịch sử hoặc tiếp tục coi breakout thành công.

## 7. Đầu ra và khả năng giải thích

- `analysisWindow`: gốc 40, phạm vi thực dùng và lý do thay đổi.
- `priceStructure`: hướng, mẫu swing và từng pivot với `confirmedAt`.
- `context`: stage, trạng thái developing, priorTrend, rationale, counterEvidence, alternatives.
- `events`: dòng sự kiện **của chiến dịch chính**; `date` khác `observed_at`.
- `structures`: các chiến dịch/vùng lớn–nhỏ tách riêng cùng lịch sử trạng thái.
- `phase`: pha hiện tại đang được đánh giá; không phải nhãn được con người duyệt.

Không có mã chứng khoán, ngày nhãn hoặc giá từ annotation được đưa vào hàm suy luận. Bộ mẫu chỉ dùng khi đánh giá.

## 8. Kiểm chứng và giới hạn

Kiểm thử gồm chuỗi tích lũy và phân phối đối xứng, PS/PSY, SC/BC, AR, ST, Spring/Test, UT/UTAD, SOS/SOW, LPS/LPSY; chuyển pha B→C→D→E ở các cutoff; swing HH/HL và LH/LL; tối thiểu 40 phiên; SZC không phát sinh BC giả cuối tháng 10; 9 mẫu đã duyệt; prefix replay không nhận dữ liệu sau cutoff.

Đây vẫn là rule engine thử nghiệm. Điểm heuristic không là xác suất. Chín mẫu không phải tập kiểm định ngoài mẫu; ngưỡng chưa backtest và điều chỉnh giá chưa được xác minh. Pivots và chiến dịch đang hình thành có thể được đánh giá lại khi có dữ liệu mới. `observed_at` là mốc khả dụng của quy tắc hiện tại, không chứng minh thời điểm người giao dịch có thể kết luận chắc chắn. Không có lưu trữ vintage bất biến giữa các lần gọi.

Benchmark cũ đã quét mọi ứng viên nhiều vùng; `events` mới là chiến dịch chính. Cần đọc phạm vi đánh giá trước khi so điểm, không dùng số lượng nhãn ít hơn để tuyên bố precision tăng. Các bất đồng với nhãn thủ công phải được giữ trong báo cáo, không sửa ground truth để làm xanh kiểm thử.
