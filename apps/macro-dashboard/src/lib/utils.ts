import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
export const cn = (...inputs: ClassValue[]) => twMerge(clsx(inputs));
export const formatNumber = (value: number | undefined, digits = 2) =>
  value === undefined || !Number.isFinite(value)
    ? "—"
    : value.toLocaleString("vi-VN", { maximumFractionDigits: digits });
export const formatDate = (value: string) =>
  new Date(value).toLocaleDateString("vi-VN", { timeZone: "Asia/Ho_Chi_Minh" });
