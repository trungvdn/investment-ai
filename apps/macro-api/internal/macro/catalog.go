package macro

func Catalog() []Indicator {
	return []Indicator{
		{"VN_DEPOSIT_RATE", "Lãi suất huy động VND · 6–12 tháng", "Liquidity", "VN", "%", "MONTHLY", "RANGE", true, 75, "NHNN: khoảng lãi suất tiền gửi VND bình quân của nhóm NHTM nhà nước và NHTM cổ phần, khoản tiền gửi mới và cũ, kỳ hạn từ 6 đến 12 tháng. Hiển thị cận dưới–cận trên (%/năm), không tính trung điểm."},
		{"VN_M2_GROWTH", "Tăng trưởng M2", "Liquidity", "VN", "% YoY", "MONTHLY", "YoY", true, 75, "NHNN: YoY chỉ khi có cùng tháng năm trước theo phương pháp từ 10/2025; không so qua mốc đổi phương pháp."},
		{"VN_M2_LEVEL", "Tổng phương tiện thanh toán M2", "Liquidity", "VN", "tỷ đồng", "MONTHLY", "LEVEL", true, 75, "NHNN: số dư tổng phương tiện thanh toán theo phương pháp từ 10/2025. Không nối với chuỗi cũ."},
		{"VN_M2_YTD", "Tăng trưởng M2 từ đầu năm", "Liquidity", "VN", "% YTD", "MONTHLY", "YTD", true, 75, "NHNN: mức tăng được công bố so cuối năm trước; không gán nhãn YoY."},
		{"VN_CREDIT_GROWTH", "Tăng trưởng tín dụng", "Credit", "VN", "% YTD", "MONTHLY", "YTD", true, 75, "NHNN: tăng trưởng tổng dư nợ so với cuối năm trước, không phải YoY."},
		{"VN_INTERBANK_ON", "Lãi suất liên ngân hàng O/N", "Liquidity", "VN", "%", "DAILY", "LEVEL", true, 7, "NHNN: lãi suất VND qua đêm, lấy đúng ngày giao dịch của dòng O/N."},
		{"VN_SBV_TBILL", "Tín phiếu NHNN", "Liquidity", "VN", "%", "DAILY", "LEVEL", true, 7, "NHNN: lãi suất tín phiếu trúng thầu bình quân gia quyền theo khối lượng trong phiên; cơ cấu kỳ hạn có thể thay đổi. Không phải lãi suất repo, tín phiếu Kho bạc hoặc lượng lưu hành. Phiên không công bố lãi suất không được gán 0."},
		{"VN_USD_VND", "Tỷ giá trung tâm USD/VND", "External", "VN", "VND/USD", "DAILY", "LEVEL", true, 7, "NHNN: tỷ giá trung tâm, khác tỷ giá mua/bán và tỷ giá tính thuế; ngày áp dụng khác ngày ban hành."},
		{"VN_CPI_YOY", "CPI Việt Nam", "Inflation", "VN", "% YoY", "MONTHLY", "YoY", true, 60, "CPI cả nước tháng so cùng kỳ; SDMX được tính từ cùng chuỗi chỉ số gốc."},
		{"VN_CORE_CPI_YOY", "Lạm phát cơ bản", "Inflation", "VN", "% YoY", "MONTHLY", "YoY", true, 60, "Chỉ nhận số tháng so cùng kỳ, không nhận số bình quân lũy kế."},
		{"VN_PPI_YOY", "PPI công nghiệp", "Inflation", "VN", "% YoY", "QUARTERLY", "YoY", true, 120, "NSO: PPI tổng sản phẩm công nghiệp B–E, cùng quý năm trước = 100; lấy chỉ số trừ 100, không chia tiếp cho năm trước."},
		{"VN_PMI", "PMI sản xuất", "Growth", "VN", "index", "MONTHLY", "LEVEL", true, 60, "S&P Global: PMI sản xuất Việt Nam toàn phần đã điều chỉnh mùa vụ; 50 là ngưỡng không thay đổi. Thu thập từ bản công bố công khai được cấu hình qua PMI_RELEASE_URLS."},
		{"VN_IIP_YOY", "Sản xuất công nghiệp", "Growth", "VN", "% YoY", "MONTHLY", "YoY", true, 60, "SDMX AIP_ISIC4_IX; tính YoY từ cùng chuỗi tổng toàn ngành."},
		{"VN_RETAIL_YOY", "Bán lẻ & dịch vụ tiêu dùng", "Growth", "VN", "% YoY", "MONTHLY", "YoY", true, 60, "Tăng trưởng tháng theo giá hiện hành; không phải mức tăng thực đã loại trừ giá."},
		{"VN_EXPORT_YOY", "Xuất khẩu hàng hóa", "External", "VN", "% YoY", "MONTHLY", "YoY", true, 60, "SDMX TXG_FOB_USD, tổng xuất khẩu hàng hóa tháng; tính YoY."},
		{"US_DXY", "ICE US Dollar Index (DXY)", "External", "US", "index Mar 1973=100", "DAILY", "LEVEL", true, 7, "ICE DXY giao ngay; cần cấu hình nguồn xuất dữ liệu DXY_FEED_URL. Không dùng futures DX hoặc FRED Broad USD thay thế."},
		{"US_EFFR", "Lãi suất Fed hiệu dụng", "External", "US", "%", "DAILY", "LEVEL", true, 7, "FRED DFF; không phải trần hoặc sàn lãi suất mục tiêu Fed."},
		{"US_BROAD_USD", "Broad USD Index", "External", "US", "index Jan 2006=100", "DAILY", "LEVEL", true, 14, "FRED DTWEXBGS; chỉ số USD rộng, KHÔNG phải ICE DXY."},
		{"US_10Y", "Lợi suất US Treasury 10Y", "External", "US", "%", "DAILY", "LEVEL", true, 7, "FRED DGS10; lợi suất constant maturity 10 năm."},
	}
}
func Find(code string) (Indicator, bool) {
	for _, i := range Catalog() {
		if i.Code == code {
			return i, true
		}
	}
	return Indicator{}, false
}
