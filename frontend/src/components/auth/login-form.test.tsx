import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LoginForm } from "@/components/auth/login-form";

const routerMocks = vi.hoisted(() => ({
  push: vi.fn(),
}));

const authClientMocks = vi.hoisted(() => ({
  loginWithPassword: vi.fn(),
}));

vi.mock("next/navigation", () => ({
  useRouter: () => ({
    push: routerMocks.push,
  }),
}));

vi.mock("@/lib/auth/client", () => ({
  loginWithPassword: authClientMocks.loginWithPassword,
}));

describe("LoginForm", () => {
  beforeEach(() => {
    routerMocks.push.mockReset();
    authClientMocks.loginWithPassword.mockReset();
  });

  it("blocks submit when client validation fails", async () => {
    render(<LoginForm />);

    await userEvent.click(screen.getByRole("button", { name: "Sign in" }));

    expect(await screen.findByText("Enter a valid email")).toBeInTheDocument();
    expect(screen.getByText("Password must be at least 8 characters")).toBeInTheDocument();
    expect(authClientMocks.loginWithPassword).not.toHaveBeenCalled();
  });

  it("maps server error envelope to actionable UI message", async () => {
    authClientMocks.loginWithPassword.mockResolvedValue({
      success: false,
      error: {
        code: "unauthorized",
        message: "Invalid email or password",
      },
    });

    render(<LoginForm />);

    await userEvent.type(screen.getByLabelText("Email"), "premium@example.com");
    await userEvent.type(screen.getByLabelText("Password"), "wrong-password");
    await userEvent.click(screen.getByRole("button", { name: "Sign in" }));

    const alert = await screen.findByRole("alert");
    expect(alert).toHaveTextContent("Invalid email or password");
    expect(routerMocks.push).not.toHaveBeenCalled();
  });

  it("redirects after successful login", async () => {
    authClientMocks.loginWithPassword.mockResolvedValue({
      success: true,
      data: {
        expiresAt: "2026-04-06T10:00:00Z",
        claims: {
          userId: "user-premium-1",
          plan: "premium",
          sessionState: "valid",
          scope: ["jam:read"],
        },
      },
    });

    render(<LoginForm />);

    await userEvent.type(screen.getByLabelText("Email"), "premium@example.com");
    await userEvent.type(screen.getByLabelText("Password"), "premium-pass");
    await userEvent.click(screen.getByRole("button", { name: "Sign in" }));

    await waitFor(() => {
      expect(routerMocks.push).toHaveBeenCalledWith("/");
    });
  });
});