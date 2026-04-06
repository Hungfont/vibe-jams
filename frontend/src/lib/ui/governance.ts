export type ExceptionReviewStatus = "approved" | "temporary";

export interface ShadcnExceptionEntry {
  componentPath: string;
  owner: string;
  rationale: string;
  reviewStatus: ExceptionReviewStatus;
}

export interface SourceFileInput {
  path: string;
  content: string;
}

export interface ConformanceInput {
  approvedPrimitiveNames: string[];
  exceptions: ShadcnExceptionEntry[];
  uiComponentFiles: string[];
  sourceFiles: SourceFileInput[];
}

export interface ConformanceResult {
  configErrors: string[];
  violations: string[];
}

const IMPORT_REGEX = /^\s*import(?:[\s\w{},*]+from\s+)?["']@\/components\/ui\/([^"']+)["'];?/gm;

function normalizePath(path: string): string {
  return path.replaceAll("\\", "/").replace(/^\.\//, "").trim();
}

function fileBaseName(path: string): string {
  const normalized = normalizePath(path);
  const name = normalized.split("/").at(-1) ?? "";
  return name.replace(/\.(ts|tsx|js|jsx)$/, "").toLowerCase();
}

function isSourceFile(path: string): boolean {
  return /\.(ts|tsx|js|jsx)$/.test(path);
}

function isTestFile(path: string): boolean {
  return /\.(test|spec)\.(ts|tsx|js|jsx)$/.test(path);
}

export function validateExceptionRegistry(entries: ShadcnExceptionEntry[]): string[] {
  const errors: string[] = [];

  for (const entry of entries) {
    const componentPath = normalizePath(entry.componentPath);
    if (!componentPath.startsWith("src/components/ui/") || !isSourceFile(componentPath)) {
      errors.push(`Exception ${entry.componentPath} must target src/components/ui/* source file`);
    }
    if (!entry.owner?.trim()) {
      errors.push(`Exception ${entry.componentPath} is missing owner`);
    }
    if (!entry.rationale?.trim()) {
      errors.push(`Exception ${entry.componentPath} is missing rationale`);
    }
    if (entry.reviewStatus !== "approved" && entry.reviewStatus !== "temporary") {
      errors.push(`Exception ${entry.componentPath} has invalid reviewStatus: ${entry.reviewStatus}`);
    }
  }

  return errors;
}

function exceptionPathSet(exceptions: ShadcnExceptionEntry[]): Set<string> {
  const set = new Set<string>();
  for (const entry of exceptions) {
    set.add(normalizePath(entry.componentPath));
  }
  return set;
}

function isExceptionForImport(exceptionPaths: Set<string>, importPath: string): boolean {
  const normalizedImport = normalizePath(importPath).replace(/^src\//, "");
  for (const extension of [".tsx", ".ts", ".jsx", ".js"]) {
    const fullPath = normalizePath(`src/components/ui/${normalizedImport}${extension}`);
    if (exceptionPaths.has(fullPath)) {
      return true;
    }
  }
  return false;
}

export function evaluateShadcnConformance(input: ConformanceInput): ConformanceResult {
  const approved = new Set(input.approvedPrimitiveNames.map((name) => name.trim().toLowerCase()).filter(Boolean));
  const configErrors = validateExceptionRegistry(input.exceptions);
  const exceptionPaths = exceptionPathSet(input.exceptions);
  const violations = new Set<string>();

  for (const componentPath of input.uiComponentFiles) {
    const normalized = normalizePath(componentPath);
    const primitiveName = fileBaseName(normalized);
    if (!approved.has(primitiveName) && !exceptionPaths.has(normalized)) {
      violations.add(
        `Unapproved primitive ${primitiveName} at ${normalized}. Reuse approved primitive or add documented exception.`,
      );
    }
  }

  for (const file of input.sourceFiles) {
    const normalized = normalizePath(file.path);
    if (!isSourceFile(normalized) || isTestFile(normalized)) {
      continue;
    }

    const isUiPrimitiveFile = normalized.startsWith("src/components/ui/");
    const isComponentFile = normalized.startsWith("src/components/");
    const primitiveName = fileBaseName(normalized);

    if (isComponentFile && !isUiPrimitiveFile && approved.has(primitiveName)) {
      violations.add(
        `Duplicate primitive filename at ${normalized}. Primitive ${primitiveName} must live in src/components/ui/.`,
      );
    }

    for (const match of file.content.matchAll(IMPORT_REGEX)) {
      const importPath = match[1]?.trim();
      if (!importPath) {
        continue;
      }

      const importPrimitive = fileBaseName(importPath);
      if (approved.has(importPrimitive)) {
        continue;
      }

      if (isExceptionForImport(exceptionPaths, importPath)) {
        continue;
      }

      violations.add(
        `Unapproved ui import '@/components/ui/${importPath}' in ${normalized}. Use approved primitive or add exception metadata.`,
      );
    }
  }

  return {
    configErrors,
    violations: [...violations].sort(),
  };
}