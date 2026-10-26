# Provenance

`internal/claudetui` is ported from
[claude-tui-boost](https://github.com/Guuuu233/claude-tui-boost) v0.2.16,
commit `f57ee457cf7f41ecc18ed8284519ab565f46ad16`: its Go runtime's bridge
(`runtime/internal/bridge`, itself a fork of magpie's
`internal/gateway` Claude subscription bridge) and its vendored wrapper
(`vendor/claude-bridge-magpie`, from desplega.sh's claude-bridge).

- claude-tui-boost: MIT, `LICENSE.claude-tui-boost`
- claude-bridge (the wrapper): MIT, `wrapper/LICENSE`; its adaptations,
  the plugin's and magpie's, in `wrapper/ADAPTATIONS.md`
- magpie itself: MIT, the repository's `LICENSE`

The plugin's precompiled runtime and helper binaries are not used. What
was left out and what was changed is in
[docs/subsystems/claude-tui-bridge.md](../../docs/subsystems/claude-tui-bridge.md).
