import { describe, expect, it } from "vitest";

import {
  evaluateShadcnConformance,
  validateExceptionRegistry,
  type ShadcnExceptionEntry,
} from "@/lib/ui/governance";

const APPROVED = ["alert", "avatar", "badge", "button", "card", "dialog", "dropdown-menu", "input", "scroll-area", "separator", "skeleton", "slider", "tabs", "toast", "tooltip", "use-toast"];

describe("validateExceptionRegistry", () => {
  it("returns config errors when required metadata is missing", () => {
    const invalid = [
      {
        componentPath: "src/components/ui/legacy-chip.tsx",
        owner: "",
        rationale: "",
        reviewStatus: "approved",
      },
    ] as unknown as ShadcnExceptionEntry[];

    const result = validateExceptionRegistry(invalid);

    expect(result).toHaveLength(2);
    expect(result[0]).toContain("owner");
    expect(result[1]).toContain("rationale");
  });
});

describe("evaluateShadcnConformance", () => {
  it("flags unknown primitive in components/ui without approved exception", () => {
    const result = evaluateShadcnConformance({
      approvedPrimitiveNames: APPROVED,
      exceptions: [],
      uiComponentFiles: ["src/components/ui/button.tsx", "src/components/ui/legacy-chip.tsx"],
      sourceFiles: [],
    });

    expect(result.configErrors).toEqual([]);
    expect(result.violations).toHaveLength(1);
    expect(result.violations[0]).toContain("legacy-chip");
  });

  it("allows unknown primitive when an approved exception exists", () => {
    const result = evaluateShadcnConformance({
      approvedPrimitiveNames: APPROVED,
      exceptions: [
        {
          componentPath: "src/components/ui/legacy-chip.tsx",
          owner: "@frontend-guild",
          rationale: "No equivalent primitive exists yet",
          reviewStatus: "approved",
        },
      ],
      uiComponentFiles: ["src/components/ui/button.tsx", "src/components/ui/legacy-chip.tsx"],
      sourceFiles: [],
    });

    expect(result.configErrors).toEqual([]);
    expect(result.violations).toEqual([]);
  });

  it("flags duplicate primitive file outside components/ui", () => {
    const result = evaluateShadcnConformance({
      approvedPrimitiveNames: APPROVED,
      exceptions: [],
      uiComponentFiles: ["src/components/ui/button.tsx"],
      sourceFiles: [
        {
          path: "src/components/auth/button.tsx",
          content: "export function Button(){return null}",
        },
      ],
    });

    expect(result.violations).toHaveLength(1);
    expect(result.violations[0]).toContain("src/components/auth/button.tsx");
  });

  it("flags imports to non-approved primitives", () => {
    const result = evaluateShadcnConformance({
      approvedPrimitiveNames: APPROVED,
      exceptions: [],
      uiComponentFiles: ["src/components/ui/button.tsx"],
      sourceFiles: [
        {
          path: "src/components/auth/login-form.tsx",
          content: "import { LegacyChip } from \"@/components/ui/legacy-chip\";",
        },
      ],
    });

    expect(result.violations).toHaveLength(1);
    expect(result.violations[0]).toContain("legacy-chip");
  });
});