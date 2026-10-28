import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MacroScorePanel } from "./macro-score-panel";
import { fetchMacroScore } from "../data/macro-score-client";
vi.mock("../data/macro-score-client", () => ({ fetchMacroScore: vi.fn() }));
afterEach(cleanup);
function renderPanel() { render(<QueryClientProvider client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}><MacroScorePanel /></QueryClientProvider>); }
describe("macro score panel", () => {
  it("shows no fabricated history and requests the selected frequency", async () => {
    vi.mocked(fetchMacroScore).mockResolvedValue({ current: null, history: [] } as any);
    renderPanel(); await screen.findByText(/Chưa có snapshot/);
    await userEvent.click(screen.getByRole("button", { name: "Tuần" }));
    await waitFor(() => expect(fetchMacroScore).toHaveBeenCalledWith("weekly", expect.any(AbortSignal)));
  });
  it("isolates an unavailable score endpoint", async () => {
    vi.mocked(fetchMacroScore).mockRejectedValue(new Error("Kho điểm không sẵn sàng"));
    renderPanel(); expect(await screen.findByRole("alert", {}, { timeout: 3000 })).toHaveTextContent("Kho điểm không sẵn sàng");
  });
});
