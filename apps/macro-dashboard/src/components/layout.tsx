import { useEffect, useState } from "react";
import { NavLink, Outlet, useLocation } from "react-router-dom";
import {
  ArrowUpRight,
  ChartCandlestick,
  ChartNoAxesCombined,
  ChevronLeft,
  ChevronRight,
  Globe2,
  Menu,
  Moon,
  Sun,
  X,
  ShieldCheck,
  BookOpen,
} from "lucide-react";
import { Button } from "./ui/button";
import { cn } from "../lib/utils";
export function Layout() {
  const [dark, setDark] = useState(() => {
    try {
      return localStorage.getItem("investment-theme") === "dark";
    } catch {
      return false;
    }
  });
  const [collapsed, setCollapsed] = useState(false),
    [mobile, setMobile] = useState(false);
  const { pathname } = useLocation();
  useEffect(() => {
    document.documentElement.classList.toggle("dark", dark);
    try {
      localStorage.setItem("investment-theme", dark ? "dark" : "light");
    } catch {
      /* Storage may be disabled. */
    }
  }, [dark]);
  useEffect(() => {
    setMobile(false);
    window.scrollTo(0, 0);
  }, [pathname]);
  return (
    <div className={cn("app-layout", collapsed && "is-collapsed")}>
      {mobile && (
        <button
          className="sidebar-backdrop"
          aria-label="Đóng menu"
          onClick={() => setMobile(false)}
        />
      )}
      <aside className={cn("sidebar", mobile && "mobile-open")}>
        <NavLink to="/" className="logo" aria-label="Investment AI">
          <span className="logo-mark">
            <ArrowUpRight size={24} />
          </span>
          <span className="sidebar-copy">
            Investment<span className="logo-ai">AI</span>
            <small>RESEARCH WORKSPACE</small>
          </span>
        </NavLink>
        <div className="sidebar-section sidebar-copy">
          KHÔNG GIAN NGHIÊN CỨU
        </div>
        <nav aria-label="Điều hướng chính">
          <NavLink
            end
            to="/"
            className={({ isActive }: { isActive: boolean }) =>
              cn("nav-item", isActive && "active")
            }
            title="Phân tích cổ phiếu"
          >
            <ChartCandlestick size={20} />
            <span className="sidebar-copy">Phân tích cổ phiếu</span>
          </NavLink>
          <NavLink
            to="/macro"
            className={({ isActive }: { isActive: boolean }) =>
              cn("nav-item", isActive && "active")
            }
            title="Bối cảnh vĩ mô"
          >
            <Globe2 size={20} />
            <span className="sidebar-copy">Bối cảnh vĩ mô</span>
          </NavLink>
        </nav>
        <div className="sidebar-insight sidebar-copy">
          <BookOpen size={20} />
          <p>
            Bằng chứng trước.
            <br />
            Nhận định sau.
          </p>
          <span>Hiểu thị trường qua cấu trúc giá và dòng tiền.</span>
          <div className="sidebar-mini-chart">
            <i />
            <i />
            <i />
            <i />
            <i />
            <i />
            <i />
            <i />
          </div>
        </div>
        <div className="sidebar-bottom">
          <div className="sidebar-trust sidebar-copy">
            <ShieldCheck size={15} />
            <span>Explainability by design</span>
          </div>
          <button
            className="theme-switch"
            onClick={() => setDark(!dark)}
            aria-label={dark ? "Chuyển giao diện sáng" : "Chuyển giao diện tối"}
          >
            {dark ? <Sun size={18} /> : <Moon size={18} />}
            <span className="sidebar-copy">
              {dark ? "Giao diện sáng" : "Giao diện tối"}
            </span>
          </button>
          <button
            className="collapse-button"
            onClick={() => setCollapsed(!collapsed)}
            aria-label={collapsed ? "Mở rộng sidebar" : "Thu gọn sidebar"}
          >
            {collapsed ? <ChevronRight size={16} /> : <ChevronLeft size={16} />}
          </button>
        </div>
      </aside>
      <div className="workspace">
        <header className="topbar">
          <Button
            variant="ghost"
            size="icon"
            className="mobile-menu"
            aria-label={mobile ? "Đóng menu" : "Mở menu"}
            onClick={() => setMobile(!mobile)}
          >
            {mobile ? <X size={20} /> : <Menu size={20} />}
          </Button>
          <div className="breadcrumb">
            <span>Workspace</span>
            <ChevronRight size={13} />
            <strong>
              {pathname === "/macro" ? "Bối cảnh vĩ mô" : "Phân tích cổ phiếu"}
            </strong>
          </div>
          <div className="topbar-right">
            <span className="research-pill">
              <span /> Research edition
            </span>
            <span className="avatar">TI</span>
          </div>
        </header>
        <main>
          <Outlet context={{ dark }} />
        </main>
        <footer>
          <span>
            <ChartNoAxesCombined size={13} /> INVESTMENT AI{" "}
            <span className="footer-slash">/</span> Research Workspace
          </span>
          <span>Phân tích có bằng chứng · Quyết định có cơ sở</span>
        </footer>
      </div>
    </div>
  );
}
