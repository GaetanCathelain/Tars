---
name: containerized-service-operations
description: Use when operating or verifying containerized services over SSH.
---

# Containerized service operations

Operate the service the user named, then prove that the service—not merely its container—is healthy.

## Procedure

1. **Discover the runtime boundary.** Over SSH, identify the exact container, its state, network mode, the listener on the requested port, and the listener PID. Treat container uptime and service uptime as separate facts.
2. **Identify the service lifecycle.** Read the container entrypoint/command and current process tree. If PID 1 is the service, restart the container. If PID 1 is only a keeper process such as `sleep infinity` and the service was launched with `docker exec`, restart the service process itself.
3. **Name the target.** Before stopping anything, bind the action to the exact container, port, process, and PID discovered in step 1.
4. **Restart once.** Stop the old service process gracefully, confirm its PID exited, then start the documented service command inside the existing container. Preserve the container and its mounted state unless the user explicitly asked for broader lifecycle work.
5. **Verify the replacement.** Require all of:
   - the container is in the expected state;
   - the requested address and port are listening;
   - the listener PID differs from the stopped PID;
   - the service health endpoint or equivalent application probe succeeds;
   - when credentials/config already exist and the task permits it, one minimal real authenticated application request succeeds. A TCP connect, unauthenticated catalog/status call, or health endpoint alone does not prove the upstream request path works.
6. **Re-read final state.** Immediately before reporting, re-read the deployed version/revision, container state, listener PID, and application health. For deployments tied to a remote branch or release, re-read the authoritative remote revision and verify the deployed release commit is in its ancestry.
7. **Report the verdict.** State the old and new PID, listener address, application-health result, and the real request's status plus one non-secret response fact. A restarted container without a healthy application is a failed service restart.

## Pitfalls

- A successful `docker restart` proves only that the container lifecycle restarted. It can silently remove a detached `docker exec` service when PID 1 is a keeper process.
- An empty port immediately after restart may be startup latency; use a short bounded readiness loop before diagnosing failure.
- Avoid broad process matching. Prefer the PID bound to the requested listener, and confirm that PID has exited before launching its replacement. Never use `pkill -f` with a pattern copied into the same shell command: the wrapper can match and terminate itself.
- A port number is not a service identity. The same numeric port can be owned simultaneously on loopback, a Tailscale address, a published Docker address, and inside another network namespace. Resolve the full chain—public/Tailscale entrypoint, proxy target, bind address, port, PID, container/network namespace, command, and cwd—before diagnosing or restarting. If a reverse proxy points to `127.0.0.1:<port>`, do not substitute a listener bound to `<tailscale-ip>:<port>` merely because the port matches. Verify both the intended backend listener and the user-facing ingress after a restart.
- Keep credentials out of evidence. Inspect only the fields needed for lifecycle discovery; do not print whole environment blocks.
- Do not infer the launch command from the process name alone. Read image/container metadata and the executable's own help or package metadata. For a relative executable path, reconstruct both the listener's NUL-delimited `/proc/<pid>/cmdline` and `/proc/<pid>/cwd`; relaunch with `docker exec -w <cwd>` so the same argv resolves against the same checkout. Immediately before signaling, revalidate that the recorded PID still owns the port and still belongs to the recorded container.
- For a secret-safe authenticated proxy probe, an existing API key may be read from the listener's NUL-delimited `/proc/<pid>/environ` inside a privileged on-host script and kept only in memory. Never print it, put it in shell argv, or copy it off-host. Use it for one minimal real completion/message request and report only status plus a deliberately non-secret response fact; a successful models/catalog request is not enough.
- Validate a proxy/config image through its **actual image entrypoint contract**. Official images may prepend the binary only when the first argument begins with `-`; passing the binary name again can exercise a different command path and produce a false-positive check. For HAProxy's official image, use `IMAGE -c -f /path/haproxy.cfg`, then also inspect the long-form parser error from the real service logs if startup differs.
- Do not use full HAProxy config validation as the recurring container healthcheck when an inactive blue/green server may be absent from DNS. `haproxy -c` can spend longer than the health timeout resolving the intentionally absent slot and mark a functioning proxy unhealthy. Validate config separately during deployment; make runtime health probe the actual local dataplane listener and active backend, with enough startup grace for initial resolution.
- A bind mount being present does not mean the container runtime UID can read or write it. Probe the mounted file/directory from a disposable container running as the production UID before startup. This is especially important for rootless proxy images, runtime socket directories, and files under host directories with restrictive traversal modes.
- Atomic replacement (`mktemp` + `mv`) creates a new inode and can silently reset the intended mode. Give public configuration/map files their required mode on the temporary inode before `mv`; keep secret/state writers separate rather than reusing a universal mode-0600 helper. Re-check mode from inside the production container.
- Do not bind-mount an atomically replaced file directly. A single-file bind mount remains pinned to the old inode, so later `mv` replacements may be correct on the host but invisible to the running container and its restarted workers. Put public mutable config/map files in a dedicated non-secret directory and bind-mount that directory read-only; atomic replacement inside the directory then remains visible. Recreate once when migrating from the old mount and verify readability as the production UID.
- Docker `internal: true` networks block internet egress as well as ingress. For an internet-calling service behind a private proxy, use three explicit planes when needed: an internal proxy-to-backend network, a normal egress network attached only to application slots, and a normal frontend network attached to the proxy for the narrowly bound published listener. Keep backend application ports unpublished and verify upstream DNS/API access with a real request.
- Read an image entrypoint before using it for a one-time ownership repair. If it always invokes an application CLI and appends `"$@"`, pass a known-success CLI subcommand (for example `help`) rather than an executable name such as `node`; verify post-drop read/write access separately as the final non-root UID.
- Dario's public `accounts add` command refuses an existing alias, but deleting first creates an unnecessary credential gap. For an in-place refresh, call the underlying start/complete OAuth flow that persists the exact alias only after the exchange succeeds. Run the OAuth program from a temporary file or safe `-e` argument—not source piped into `bun -`, because that consumes stdin before the authorization-code prompt. Keep the same native PTY alive, use a fresh browser profile per account, and verify the exact alias has a future expiry afterward. See `references/interactive-oauth-refresh.md`.

## Parallel image verification

When validating a new container image on a host that already runs the service, treat the live deployment as read-only:

1. Snapshot the live container ID, image ID, start time, listener, and application health before testing.
2. Use exact, unique names for the temporary container, volume, image tag, source directory, and unused loopback port. Refuse to overwrite unexpected pre-existing resources.
3. Prefer pulling the immutable registry reference. If registry access is unavailable, stage the exact authoritative source revision and build that revision locally; report registry pull as unverified rather than implying the fallback tested it.
4. Copy only the minimum state needed into an isolated temporary volume. Never mount the live writable state into the test container, print credentials, or reuse the live listener.
5. Run the temporary image with production hardening, then verify image metadata, runtime UID, CLI/doctor behavior, image-defined healthcheck, and one real authenticated application request. Record only non-secret response facts.
6. Clean up only exact resources created by the test, ideally with a failure trap. Re-read the live container identity, start time, listener, and health afterward and prove they are unchanged.

## References

- For a validated side-by-side production image smoke-test recipe, read `references/parallel-image-verification.md`.
- For the validated detached-service pattern observed with a keeper container, read `references/detached-service-restart.md`.
- For source-package staging, atomic in-container replacement, and credential-safe end-to-end verification, read `references/source-package-deploy.md`.
- For Dario pool OAuth diagnosis, replacement-alias recovery, and secret-safe post-reauth checks, read `references/dario-pool-reauth.md`.
- For safe in-place interactive OAuth refreshes with a native PTY, fresh browser profiles, localized magic-link mail, and post-write expiry proof, read `references/interactive-oauth-refresh.md`.
- For HAProxy atomic maps, Compose network-plane separation, health-vs-config checks, deterministic zero-failure drain proof, and idempotent upgrade verification, read `references/haproxy-compose-bluegreen-safety.md`.