# baki — Xiaomi Mod State Manager

**Version control for Android partitions.** Track, snapshot, and rollback partition changes during Xiaomi modding.

```
baki init      Initialize tracking
baki commit    Snapshot partition state
baki log       View snapshot history
baki status    Check current state vs last snapshot
baki diff      Show partition changes
baki checkout  Rollback to a snapshot
baki recommend Smart backup suggestions
```

## Why baki?

Before flashing a kernel, magisk module, or custom ROM:

```
baki commit "Before flashing kernel X"
# ... flash stuff ...
baki diff           # what changed?
baki checkout 1     # rollback if something broke
```

## Requirements

- Android device with **USB debugging** enabled
- `adb` installed on your computer
- Xiaomi device (works best, but any Android with fastboot partitions works)

## Install

```bash
git clone https://github.com/soe1hom-arch/baki.git
cd baki
go build -o baki .
sudo mv baki /usr/local/bin/
```

## Commands

| Command | Description |
|---|---|
| `baki init` | Initialize baki in current directory |
| `baki commit <msg>` | Snapshot partition checksums |
| `baki commit --backup <msg>` | Snapshot + backup critical partitions |
| `baki log` | Show snapshot history |
| `baki status` | Show partitions & changes |
| `baki diff` | Show what changed since last snapshot |
| `baki checkout <id>` | Restore partitions from backup |
| `baki recommend` | Show which partitions to backup |

## What makes baki different?

- **Partition-aware**: Knows which partitions are critical (boot, persist, efs) vs optional
- **Git-like workflow**: Commit, log, diff, checkout — familiar if you use git
- **Zero dependencies**: Single Go binary, only needs `adb`
- **Context-aware**: Tracks what you did ("Before flashing kernel X"), not just partition hashes

## Credits

Created and maintained by **soe1hom-arch**.

## License

MIT &mdash; &copy; 2026 soe1hom-arch
