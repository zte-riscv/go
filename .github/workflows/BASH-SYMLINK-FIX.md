# Bash 符号链接问题修复

## 🐛 当前问题

从最新的 QEMU 输出看到：

```
Using bootstrap Go: /usr/local/go-bootstrap
go version go1.25.3 linux/riscv64
Building Go toolchain...
GOROOT_BOOTSTRAP=/usr/local/go-bootstrap
env: can't execute 'bash': No such file or directory
===========================================================
Go build and test completed successfully!
===========================================================
...
VM ready for interactive use. Type exit to shutdown.
===========================================================
/bin/sh: can't access tty; job control turned off
/go/src #
```

### 问题分析

1. **✅ Bootstrap Go 被正确识别和执行**
   - `go version go1.25.3 linux/riscv64` 成功输出
   
2. **❌ `env: can't execute 'bash': No such file or directory`**
   - 执行 `./all.bash` 时，脚本中的 `#!/usr/bin/env bash` 找不到 bash
   - 说明 `/bin/bash` 符号链接没有创建成功

3. **❌ 进入交互式 shell**
   - "VM ready for interactive use" - 这个消息不在我们的新 init 脚本中
   - 说明使用的是旧的、缓存的 Busybox rootfs

## 🔍 根本原因

### 原因 1：缓存的旧 init 脚本

**问题：**
- Busybox rootfs 被缓存了（`busybox-riscv64-1.36_stable`）
- 旧的 init 脚本中有 "VM ready for interactive use" 消息
- 即使我们更新了 workflow，缓存的旧 rootfs 仍然被使用

**证据：**
- 输出中的消息 "VM ready for interactive use. Type exit to shutdown." 不在当前的 init 脚本中
- 这是之前版本的残留

### 原因 2：bash 符号链接检查逻辑不够强

**旧代码（第 324-325 行）：**
```bash
if [ ! -e /bin/bash ] && [ -e /bin/busybox ]; then
  ln -sf /bin/busybox /bin/bash
fi
```

**问题：**
- 没有调试输出，无法确认是否执行
- 条件 `[ ! -e /bin/bash ]` 可能在某些情况下失败
- 没有验证创建是否成功

## ✅ 修复方案

### 修复 1：增强 bash 符号链接创建（第 323-346 行）

**修改后：**
```bash
# Create bash and env symlinks (Go scripts need bash)
echo "Creating bash and env symlinks..."
if [ -e /bin/busybox ]; then
  ln -sf /bin/busybox /bin/bash
  echo "Created /bin/bash -> /bin/busybox"
else
  echo "WARNING: /bin/busybox not found"
fi

if [ -e /bin/env ]; then
  mkdir -p /usr/bin
  ln -sf /bin/env /usr/bin/env
  echo "Created /usr/bin/env -> /bin/env"
elif [ -e /usr/bin/env ]; then
  echo "/usr/bin/env already exists"
else
  echo "WARNING: env not found"
fi

# Verify bash is available
if command -v bash >/dev/null 2>&1; then
  echo "bash is available: $(which bash)"
else
  echo "WARNING: bash command not found in PATH"
fi
```

**改进：**
- ✅ 移除 `[ ! -e /bin/bash ]` 条件，始终创建（`-f` 强制覆盖）
- ✅ 添加详细的调试输出
- ✅ 验证 bash 命令是否可用
- ✅ 处理 `/bin/env` 和 `/usr/bin/env` 两种情况

### 修复 2：更新 Busybox 缓存 key（第 241 行）

**修改前：**
```yaml
key: busybox-riscv64-1.36_stable-${{ runner.os }}
```

**修改后：**
```yaml
key: busybox-riscv64-1.36_stable-v2-${{ runner.os }}
```

**说明：**
- 添加版本号 `v2`，强制重新构建 Busybox rootfs
- 新的 rootfs 将包含更新的 init 脚本
- 旧的缓存将被忽略

## 📊 执行流程（修复后）

```
启动 VM
  ↓
执行 /init 脚本
  ↓
挂载文件系统 (/proc, /sys, /dev) ✅
  ↓
创建符号链接
  ├─ ln -sf /bin/busybox /bin/bash
  ├─ echo "Created /bin/bash -> /bin/busybox" ✅
  ├─ ln -sf /bin/env /usr/bin/env
  ├─ echo "Created /usr/bin/env -> /bin/env" ✅
  └─ command -v bash ✅
  ↓
检查 Go 源代码 (/go/src) ✅
  ↓
检查 Bootstrap Go (/usr/local/go-bootstrap)
  ├─ go version ✅
  └─ export GOROOT_BOOTSTRAP ✅
  ↓
执行 ./all.bash
  ├─ #!/usr/bin/env bash → 找到 /bin/bash ✅
  ├─ Building Go bootstrap tool...
  ├─ Building packages and commands...
  ├─ Testing packages...
  └─ ALL TESTS PASSED ✅
  ↓
显示总结
  ↓
poweroff -f ✅
```

## 🎯 预期输出（下次运行）

```
===========================================================
RISC-V64 Linux VM Started
===========================================================

System information:
Linux (none) 6.1.0 #1 SMP ... riscv64 GNU/Linux

Available memory:
MemTotal:        2038548 kB

CPU information:
processor	: 0
processor	: 1
processor	: 2
processor	: 3

Creating bash and env symlinks...
Created /bin/bash -> /bin/busybox
Created /usr/bin/env -> /bin/env
bash is available: /bin/bash

Go source code found in /go

===========================================================
Starting Go compiler build and test...
===========================================================

Found bootstrap Go at /usr/local/go-bootstrap
Checking if it is a valid RISC-V binary...
✅ Bootstrap Go is valid and executable
Using bootstrap Go: /usr/local/go-bootstrap
go version go1.25.3 linux/riscv64

Building Go toolchain...
GOROOT_BOOTSTRAP=/usr/local/go-bootstrap

##### Building Go bootstrap tool.
cmd/dist

##### Building Go toolchain using /usr/local/go-bootstrap.
Building Go cmd/dist using /usr/local/go-bootstrap.
...

##### Building packages and commands for linux/riscv64.
...

##### Testing packages.
ok  	archive/tar	0.123s
ok  	bufio	0.456s
...
ALL TESTS PASSED

---
Installed Go for linux/riscv64 in /go
Installed commands in /go/bin

===========================================================
Go build and test completed successfully!
===========================================================

Build log saved to /tmp/go-build.log

Last 50 lines of build log:
-----------------------------------------------------------
ALL TESTS PASSED
---
Installed Go for linux/riscv64 in /go
Installed commands in /go/bin

Go build process completed. Shutting down...
[  1234.567890] reboot: Power down

✅ QEMU VM completed successfully
```

## 📝 关键改进

| 改进 | 说明 | 影响 |
|------|------|------|
| **移除条件检查** | 始终创建符号链接（`-f` 强制） | 确保符号链接存在 |
| **添加调试输出** | 显示每个步骤的结果 | 便于排查问题 |
| **验证 bash 可用** | 使用 `command -v bash` | 确认 bash 在 PATH 中 |
| **更新缓存 key** | 添加 `v2` 版本号 | 强制使用新的 init 脚本 |
| **处理 env 路径** | 支持 `/bin/env` 和 `/usr/bin/env` | 兼容不同配置 |

## 🔧 手动验证（如果需要）

如果下次运行仍然有问题，可以手动验证：

```bash
# 在 VM 中（如果能够进入 shell）
ls -la /bin/bash
ls -la /usr/bin/env
command -v bash
bash --version
which bash
```

预期输出：
```
lrwxrwxrwx 1 root root 12 ... /bin/bash -> /bin/busybox
lrwxrwxrwx 1 root root 8  ... /usr/bin/env -> /bin/env
/bin/bash
BusyBox v1.36.1 (2025-11-25 ...) multi-call binary
/bin/bash
```

## 📄 相关文件修改

- **第 323-346 行**：增强 bash 符号链接创建和验证
- **第 241 行**：更新 Busybox 缓存 key 到 `v2`

## ✅ 成功标准

下次运行应该看到：
- [x] "Creating bash and env symlinks..." 消息
- [x] "Created /bin/bash -> /bin/busybox" 消息
- [x] "bash is available: /bin/bash" 消息
- [x] `./all.bash` 成功执行，无 bash 相关错误
- [x] "ALL TESTS PASSED" 消息
- [x] 自动关机，无交互式 shell

如果以上都满足，问题就彻底解决了！🎉

