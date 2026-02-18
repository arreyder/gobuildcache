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
- [x] **Lifecycle-aware metrics** — tracks the age of backend cache hits
  (hours since original PUT) using DDSketch quantile estimation. Reports
  p50/p90/p99/max age in human-readable stats and `entry_age_p50_hours` /
  `entry_age_max_hours` in machine-readable output. If entries are
  approaching the lifecycle duration, the policy is too short.
