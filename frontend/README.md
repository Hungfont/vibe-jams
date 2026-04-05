# Frontend (Bun)

This frontend is configured to use Bun as the package manager and script runner.

## 1) Install Bun

### Windows (PowerShell)

```powershell
irm bun.sh/install.ps1 | iex
```

If your antivirus blocks the script, allow Bun installer execution in your security settings and run the command again.

Verify installation:

```powershell
bun --version
```

## 2) Install dependencies

From the `frontend/` directory:

```bash
bun install
```

This will generate `bun.lock` and install dependencies in `node_modules/`.

## 3) Run project commands

```bash
bun dev
bun run build
bun run start
bun run lint
bun test
bun run test:watch
bun run test:coverage
```

App URL: [http://localhost:3000](http://localhost:3000)
