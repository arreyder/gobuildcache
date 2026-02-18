# gobuildcache improvements (arreyder fork)

## Done

- [x] **Touch-on-GET lifecycle refresh** — async S3 CopyObject self-to-self on
  backend cache hits resets `LastModified`, preventing lifecycle policy expiry
  for frequently-accessed entries. In-process dedup avoids redundant touches.
- [x] **Machine-readable stats** (`-stats-machine`) — single-line `key=value`
  output to stderr for Buildkite/CI log parsing.
- [x] **Conditional PUT** (`-conditional-put`) — `HeadObject` check before
  uploading skips redundant PUTs when S3 already has the object. Saves
  bandwidth and S3 write costs on ephemeral CI agents where local cache is cold.
- [x] **Debounced touch with staleness threshold** (`-touch-age-threshold`) —
  when set (e.g. `84h`), Touch does a `HeadObject` first and skips the
  `CopyObject` if the object was modified more recently than the threshold.
  Reduces unnecessary S3 calls when builds run frequently.
- [ ] **Lifecycle-aware metrics** — distinguish S3 misses (never existed) from
  entries that likely expired, to help tune lifecycle policy duration.
