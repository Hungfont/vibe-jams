import { readdir, readFile } from "node:fs/promises";
import path from "node:path";

import { evaluateShadcnConformance, type ShadcnExceptionEntry, type SourceFileInput } from "../src/lib/ui/governance";

interface PrimitiveInventory {
  approvedPrimitives: string[];
}

interface ExceptionRegistry {
  exceptions: ShadcnExceptionEntry[];
}

const root = process.cwd();
const srcDir = path.join(root, "src");
const uiDir = path.join(srcDir, "components", "ui");
const inventoryPath = path.join(root, "config", "shadcn-primitive-inventory.json");
const exceptionsPath = path.join(root, "config", "shadcn-exceptions.json");

function toProjectPath(absPath: string): string {
  return path.relative(root, absPath).replaceAll("\\", "/");
}

function isSourceFile(filePath: string): boolean {
  return /\.(ts|tsx|js|jsx)$/.test(filePath);
}

function isExcludedFile(filePath: string): boolean {
  return /\.(test|spec)\.(ts|tsx|js|jsx)$/.test(filePath) || filePath.endsWith(".d.ts");
}

async function walk(dirPath: string): Promise<string[]> {
  const entries = await readdir(dirPath, { withFileTypes: true });
  const files: string[] = [];

  for (const entry of entries) {
    const fullPath = path.join(dirPath, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await walk(fullPath)));
      continue;
    }
    files.push(fullPath);
  }

  return files;
}

async function loadJson<T>(filePath: string): Promise<T> {
  const raw = await readFile(filePath, "utf8");
  return JSON.parse(raw) as T;
}

function printProblems(title: string, items: string[]): void {
  if (items.length === 0) {
    return;
  }
  console.error(`\n${title}`);
  for (const item of items) {
    console.error(`- ${item}`);
  }
}

async function main(): Promise<void> {
  const inventory = await loadJson<PrimitiveInventory>(inventoryPath);
  const registry = await loadJson<ExceptionRegistry>(exceptionsPath);
  const allSourceFiles = (await walk(srcDir)).filter((filePath) => isSourceFile(filePath) && !isExcludedFile(filePath));
  const uiComponentFiles = (await walk(uiDir)).filter((filePath) => isSourceFile(filePath) && !isExcludedFile(filePath)).map(toProjectPath);

  const sourceFiles: SourceFileInput[] = [];
  for (const filePath of allSourceFiles) {
    sourceFiles.push({
      path: toProjectPath(filePath),
      content: await readFile(filePath, "utf8"),
    });
  }

  const result = evaluateShadcnConformance({
    approvedPrimitiveNames: inventory.approvedPrimitives,
    exceptions: registry.exceptions,
    uiComponentFiles,
    sourceFiles,
  });

  if (result.configErrors.length > 0 || result.violations.length > 0) {
    printProblems("shadcn exception config errors:", result.configErrors);
    printProblems("shadcn conformance violations:", result.violations);
    console.error("\nFix violations or update config/shadcn-exceptions.json with full approved metadata.");
    process.exit(1);
  }

  console.log("shadcn conformance passed.");
}

main().catch((error: unknown) => {
  console.error("shadcn conformance check failed unexpectedly:", error);
  process.exit(1);
});