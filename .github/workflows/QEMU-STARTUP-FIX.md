# QEMU 启动问题修复

## 🐛 问题描述

### 问题 1：QEMU 启动失败
```
qemu-system-riscv64: -serial stdio: cannot use stdio by multiple character devices
qemu-system-riscv64: -serial stdio: could not connect serial device to character backend 'stdio'
⚠️  QEMU exited with code 1
```

### 问题 2：看不到 all.bash 执行信息
即使 QEMU 启动成功，也没有看到 VM 内部运行 `all.bash` 的输出。

## 🔍 根本原因分析

### 原因 1：参数冲突

QEMU 命令行同时使用了：
- `-nographic` - 隐含地将串口连接到 stdio
- `-serial stdio` - 显式地将串口连接到 stdio

**冲突**：两个参数都试图将串口连接到 stdio，导致"多个字符设备使用 stdio"错误。

### 原因 2：init 脚本会进入交互式 shell

init 脚本在多个地方会执行 `exec /bin/sh`：
1. 没有找到 Go 源代码时
2. 没有找到 bootstrap Go 时
3. 完成 all.bash 后

**问题**：`exec /bin/sh` 会进入交互式 shell，VM 不会自动退出，导致：
- GitHub Actions 超时（120分钟后）
- 无法看到完整的执行流程

## ✅ 修复方案

### 修复 1：删除 `-serial stdio` 参数（第 677 行）

**修改前：**
```yaml
$QEMU_CMD \
  -nographic \
  ...
  -no-reboot \
  -serial stdio || QEMU_EXIT=$?
```

**修改后：**
```yaml
$QEMU_CMD \
  -nographic \
  ...
  -no-reboot || QEMU_EXIT=$?
```

**说明**：`-nographic` 已经自动将串口连接到 stdio，无需显式指定。

### 修复 2：init 脚本自动关机（第 354-380、404-422 行）

#### 2.1 有 bootstrap Go 的情况

**修改后（第 404-413 行）：**
```bash
  # Show summary
  if [ -f /tmp/go-build.log ]; then
    echo "Last 50 lines of build log:"
    echo "-----------------------------------------------------------"
    tail -50 /tmp/go-build.log
  fi
  echo ""
  echo "Go build process completed. Shutting down..."
  sleep 2
  poweroff -f
```

#### 2.2 没有 bootstrap Go 的情况

**修改前（第 363-380 行）：**
```bash
  else
    echo "No pre-built bootstrap found..."
    ...
    exec /bin/sh  # ❌ 进入交互式 shell，导致超时
  fi
```

**修改后（第 363-377 行）：**
```bash
  else
    echo "❌ ERROR: No pre-built bootstrap Go found"
    echo "This requires a working Go compiler or cross-compiled bootstrap"
    echo ""
    echo "Please ensure the bootstrap Go is built on the host and copied to rootfs"
    echo ""
    echo "Shutting down VM..."
    sleep 2
    poweroff -f
    exit 1
  fi
```

#### 2.3 没有 Go 源代码的情况

**修改前（第 411-421 行）：**
```bash
else
  echo "Go source code not found in /go"
  echo "Available directories:"
  ls -la / | head -10
fi
echo ""
echo "VM ready for interactive use. Type exit to shutdown."
exec /bin/sh  # ❌ 进入交互式 shell，导致超时
```

**修改后（第 415-429 行）：**
```bash
else
  echo "❌ Go source code not found in /go"
  echo "Available directories in /:"
  ls -la / | head -20
  echo ""
  echo "Checking /go directory:"
  ls -la /go/ 2>/dev/null || echo "/go does not exist"
  echo ""
  echo "Disk usage:"
  df -h
fi
echo ""
echo "==========================================================="
echo "Workflow completed. Shutting down VM..."
echo "==========================================================="
echo ""
sleep 2
poweroff -f
```

**改进**：
1. ✅ 显示更详细的调试信息（`/go` 目录内容、磁盘使用情况）
2. ✅ 自动关机而不是进入交互式 shell

## 📊 执行流程图

### 修复前
```
启动 QEMU
  ↓
❌ 参数冲突 → 启动失败 (exit code 1)
```

或者（如果启动成功）：

```
启动 QEMU
  ↓
启动 VM
  ↓
执行 init 脚本
  ↓
检查 /go/src
  ├─ 存在 → 检查 bootstrap
  │           ├─ 存在 → 运行 all.bash → exec /bin/sh ⏱️ 超时
  │           └─ 不存在 → exec /bin/sh ⏱️ 超时
  └─ 不存在 → exec /bin/sh ⏱️ 超时
```

### 修复后
```
启动 QEMU ✅
  ↓
启动 VM ✅
  ↓
执行 init 脚本 ✅
  ↓
挂载文件系统 ✅
  ↓
显示系统信息 ✅
  ↓
检查 /go/src
  ├─ 存在 → 检查 bootstrap
  │           ├─ 存在 → 运行 all.bash → 显示结果 → poweroff ✅
  │           └─ 不存在 → 显示错误 → poweroff ✅
  └─ 不存在 → 显示调试信息 → poweroff ✅
```

## 🎯 预期结果

修复后，你应该能看到：

```
🚀 Starting QEMU RISC-V64 Linux virtual machine...
This will boot the VM and automatically build Go compiler
Note: This may take 30-60 minutes depending on system performance

[    0.000000] Linux version 6.1.0 (runner@...) ...
[    0.234567] Run /init as init process
===========================================================
RISC-V64 Linux VM Started
===========================================================

System information:
Linux buildkitsandbox 6.1.0 #1 SMP ... riscv64 GNU/Linux

Available memory:
MemTotal:        2097152 kB

CPU information:
processor       : 0
...

Go source code found in /go

===========================================================
Starting Go compiler build and test...
===========================================================

Using bootstrap Go: /usr/local/go-bootstrap
go version go1.25.3 linux/riscv64

Building Go toolchain...
GOROOT_BOOTSTRAP=/usr/local/go-bootstrap

##### Building Go bootstrap tool.
...
##### Building packages and commands for linux/riscv64.
...
##### Testing packages.
...

===========================================================
Go build and test completed successfully!
===========================================================

Build log saved to /tmp/go-build.log

Last 50 lines of build log:
-----------------------------------------------------------
...

Go build process completed. Shutting down...
[  1234.567890] reboot: Power down

✅ QEMU VM completed successfully
```

或者，如果出错：

```
❌ Go source code not found in /go
Available directories in /:
total 56
drwxr-xr-x   18 root     root          4096 ... .
drwxr-xr-x   18 root     root          4096 ... ..
drwxr-xr-x    2 root     root          4096 ... bin
...

Checking /go directory:
/go does not exist

Disk usage:
Filesystem      Size  Used Avail Use% Mounted on
/dev/vda        5.0G  800M  4.0G  17% /

===========================================================
Workflow completed. Shutting down VM...
===========================================================

[   12.345678] reboot: Power down

⚠️  QEMU exited with code 0
```

## 📝 相关代码位置

- **第 677 行**：删除 `-serial stdio` 参数
- **第 363-377 行**：没有 bootstrap Go 时自动关机
- **第 404-413 行**：all.bash 完成后自动关机
- **第 415-429 行**：没有 Go 源代码时自动关机并显示调试信息

## 🔧 下一步

如果 VM 启动成功但找不到 Go 源代码，检查：
1. "Copy Go source code to rootfs" 步骤是否成功
2. `/go` 目录是否正确复制到 rootfs
3. 磁盘镜像是否有足够空间
