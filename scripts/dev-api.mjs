import { spawn, spawnSync } from 'node:child_process';
import { once } from 'node:events';
import { mkdtemp, rename, rm } from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import chokidar from 'chokidar';

const root = path.resolve(import.meta.dirname, '..');
const apiDir = path.join(root, 'apps/api');
const tempDir = await mkdtemp(path.join(os.tmpdir(), 'noir-api-dev-'));
const binaryPath = path.join(tempDir, 'server');
let server;
let restarting = false;
let restartQueued = false;
let shuttingDown = false;
let debounce;

function stopServer() {
  if (!server || server.exitCode !== null) return Promise.resolve();
  const child = server;
  child.kill('SIGTERM');
  return once(child, 'exit').then(() => {
    if (server === child) server = undefined;
  });
}

async function rebuildAndRestart() {
  if (shuttingDown) return;
  if (restarting) {
    restartQueued = true;
    return;
  }

  restarting = true;
  do {
    restartQueued = false;
    console.log('[api] Building...');
    const result = spawnSync('go', ['build', '-o', `${binaryPath}.next`, './cmd/server'], {
      cwd: apiDir,
      stdio: 'inherit',
      env: { ...process.env, CGO_ENABLED: '0', GOCACHE: path.join(tempDir, 'go-cache') },
    });

    if (result.status !== 0) {
      console.error('[api] Build failed; waiting for the next change.');
      continue;
    }

    await stopServer();
    await rename(`${binaryPath}.next`, binaryPath);
    console.log('[api] Starting server.');
    server = spawn(binaryPath, [], { cwd: apiDir, stdio: 'inherit', env: process.env });
    server.on('exit', (code, signal) => {
      if (!shuttingDown) console.log(`[api] Server exited (${signal ?? code}).`);
    });
  } while (restartQueued && !shuttingDown);
  restarting = false;
}

const watcher = chokidar.watch(apiDir, {
  ignored: (watchedPath) => watchedPath.includes(`${path.sep}vendor${path.sep}`),
  ignoreInitial: true,
});

watcher.on('all', (_event, changedPath) => {
  if (!/\.(go|json|sql|yaml|yml|toml)$/.test(changedPath)) return;
  clearTimeout(debounce);
  debounce = setTimeout(() => {
    console.log(`[api] Change detected: ${path.relative(root, changedPath)}`);
    void rebuildAndRestart();
  }, 250);
});

await once(watcher, 'ready');
console.log('[api] Watching apps/api for changes.');
await rebuildAndRestart();

async function shutdown() {
  if (shuttingDown) return;
  shuttingDown = true;
  clearTimeout(debounce);
  await watcher.close();
  await stopServer();
  await rm(tempDir, { recursive: true, force: true });
  process.exit(0);
}

process.on('SIGINT', shutdown);
process.on('SIGTERM', shutdown);
