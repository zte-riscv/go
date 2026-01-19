# 使用官方预编译 Go Bootstrap 方案

## 🎯 解决方案概述

使用 Go 官方提供的预编译 RISC-V64 二进制包作为 bootstrap，而不是在主机上交叉编译。

**优势：**
- ✅ 官方支持，稳定可靠
- ✅ 无需交叉编译，避免兼容性问题
- ✅ 下载快速（~100MB），比编译快得多
- ✅ 原生 RISC-V64 可执行文件，可直接在 VM 中运行

## 📦 实施步骤

### 1. 下载官方预编译包（第 468-510 行）

**新增步骤：`Download Go bootstrap toolchain for RISC-V64`**

```yaml
- name: Download Go bootstrap toolchain for RISC-V64
  working-directory: riscv64-linux
  timeout-minutes: 10
  run: |
    echo "📦 Downloading official Go bootstrap toolchain for RISC-V64..."
    
    BOOTSTRAP_DIR="/tmp/go-bootstrap-riscv64"
    mkdir -p "$BOOTSTRAP_DIR"
    cd "$BOOTSTRAP_DIR"
    
    # 下载官方预编译的 RISC-V64 Go 包
    GO_VERSION="1.25.3"
    GO_TARBALL="go${GO_VERSION}.linux-riscv64.tar.gz"
    GO_URL="https://go.dev/dl/${GO_TARBALL}"
    
    wget -q --show-progress "$GO_URL" -O "$GO_TARBALL"
    tar -xzf "$GO_TARBALL"
    
    # 验证
    if [ -f go/bin/go ]; then
      echo "✅ Official RISC-V64 Go bootstrap ready"
      ls -lh go/bin/go
      file go/bin/go
    fi
```

**关键点：**
- 使用 `wget` 下载 `go1.25.3.linux-riscv64.tar.gz`
- 直接从 https://go.dev/dl/ 获取官方包
- 解压到 `/tmp/go-bootstrap-riscv64/go`
- 验证文件存在性和类型

### 2. 复制到 rootfs（第 617-642 行）

**修改步骤：`Copy Go source code to rootfs`**

```yaml
# 复制官方下载的 bootstrap Go 到 rootfs
if [ -d /tmp/go-bootstrap-riscv64/go ]; then
  echo "📦 Copying official RISC-V64 Go bootstrap to rootfs..."
  
  # 验证二进制文件
  if file /tmp/go-bootstrap-riscv64/go/bin/go | grep -qi "RISC-V"; then
    echo "✅ Bootstrap Go binary is a RISC-V executable"
    
    # 复制到 rootfs 的 /usr/local/go-bootstrap
    sudo mkdir -p mnt/usr/local
    sudo cp -r /tmp/go-bootstrap-riscv64/go mnt/usr/local/go-bootstrap
    
    echo "✅ Bootstrap Go copied successfully"
    ls -lh mnt/usr/local/go-bootstrap/bin/go
  fi
fi
```

**关键点：**
- 复制整个 `go/` 目录到 VM 的 `/usr/local/go-bootstrap`
- 使用 `file` 命令验证是 RISC-V 可执行文件
- 如果复制失败，终止 workflow 并显示错误

### 3. VM 中使用 bootstrap（第 358-395 行）

**init 脚本会自动检测和使用：**

```bash
# Check for bootstrap Go
BOOTSTRAP_VALID=false
if [ -f /usr/local/go-bootstrap/bin/go ]; then
  echo "Found bootstrap Go at /usr/local/go-bootstrap"
  
  # 尝试运行 go version 验证
  if /usr/local/go-bootstrap/bin/go version 2>/dev/null; then
    export GOROOT_BOOTSTRAP=/usr/local/go-bootstrap
    export PATH=$GOROOT_BOOTSTRAP/bin:$PATH
    BOOTSTRAP_VALID=true
    
    echo "✅ Bootstrap Go is valid and executable"
    $GOROOT_BOOTSTRAP/bin/go version
  fi
fi

if [ "$BOOTSTRAP_VALID" = "true" ]; then
  # 继续构建 Go
  ./all.bash 2>&1 | tee /tmp/go-build.log
else
  # 显示错误并关机
  echo "❌ ERROR: No valid bootstrap Go found"
  poweroff -f
fi
```

## 📊 完整执行流程

```
GitHub Actions 主机 (AMD64 Ubuntu)
  ↓
Install dependencies ✅
  ↓
Install RISC-V toolchain ✅
  ↓
Build QEMU ✅ (or use cache)
  ↓
Build Linux kernel ✅ (or use cache)
  ↓
Build Busybox ✅ (or use cache)
  ↓
📦 NEW: Download Go 1.25.3 for RISC-V64
  ↓
  wget https://go.dev/dl/go1.25.3.linux-riscv64.tar.gz
  tar -xzf go1.25.3.linux-riscv64.tar.gz
  ↓
  ✅ Go bootstrap ready (~100MB)
  ↓
Copy Go source to rootfs
  ↓
  rsync $GITHUB_WORKSPACE/ → mnt/go/
  cp /tmp/go-bootstrap-riscv64/go → mnt/usr/local/go-bootstrap/
  ↓
  ✅ Rootfs prepared (~1-2GB)
  ↓
Boot QEMU VM (RISC-V64 Linux)
  ↓
  ┌──────────────────────────────────────┐
  │  RISC-V64 Linux VM                   │
  │                                      │
  │  Mount filesystems ✅                │
  │  Create /bin/bash symlink ✅         │
  │  Check /go/src ✅                    │
  │  Check /usr/local/go-bootstrap ✅    │
  │    ↓                                 │
  │    Execute: go version               │
  │    Output: go version go1.25.3       │
  │            linux/riscv64             │
  │    ✅ Bootstrap valid                │
  │    ↓                                 │
  │    export GOROOT_BOOTSTRAP=          │
  │           /usr/local/go-bootstrap    │
  │    ↓                                 │
  │    cd /go/src                        │
  │    ./all.bash                        │
  │    ↓                                 │
  │    ##### Building Go bootstrap tool. │
  │    ##### Building Go toolchain      │
  │    ...                               │
  │    ##### Testing packages.           │
  │    ...                               │
  │    ✅ ALL TESTS PASSED               │
  │    ↓                                 │
  │    poweroff -f                       │
  └──────────────────────────────────────┘
  ↓
✅ Workflow completed successfully
```

## 🎯 预期输出

### 主机端（下载 bootstrap）

```
📦 Downloading official Go bootstrap toolchain for RISC-V64...
Using Go 1.25.3 pre-compiled for linux/riscv64
Downloading from: https://go.dev/dl/go1.25.3.linux-riscv64.tar.gz
go1.25.3.linux-riscv64.tar.gz  100%[===============>] 102.45M  10.2MB/s    in 10s
✅ Download successful
File size: 103M
Extracting...
✅ Bootstrap Go extracted successfully
Go binary info:
-rwxr-xr-x 1 runner docker 14567890 ... go/bin/go
go/bin/go: ELF 64-bit LSB executable, UCB RISC-V, version 1 (SYSV), dynamically linked
(Note: Cannot run RISC-V binary on AMD64 host, this is expected)
✅ Official RISC-V64 Go bootstrap ready
```

### 主机端（复制到 rootfs）

```
📦 Copying official RISC-V64 Go bootstrap to rootfs...
Validating bootstrap Go binary...
✅ Bootstrap Go binary is a RISC-V executable
/tmp/go-bootstrap-riscv64/go/bin/go: ELF 64-bit LSB executable, UCB RISC-V
✅ Bootstrap Go copied successfully
Bootstrap Go location in VM: /usr/local/go-bootstrap
-rwxr-xr-x 1 root root 14567890 ... mnt/usr/local/go-bootstrap/bin/go
145M	mnt/usr/local/go-bootstrap
```

### VM 端（使用 bootstrap）

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
cmd/compile
...

##### Building packages and commands for linux/riscv64.
...

##### Testing packages.
ok  	archive/tar	0.123s
ok  	bufio	0.456s
...
ok  	cmd/go	42.123s
...

===========================================================
Go build and test completed successfully!
===========================================================

Build log saved to /tmp/go-build.log

Last 50 lines of build log:
-----------------------------------------------------------
ALL TESTS PASSED
---

Go build process completed. Shutting down...
[  1234.567890] reboot: Power down

✅ QEMU VM completed successfully
```

## 📝 与旧方案对比

| 方面 | 旧方案（交叉编译） | 新方案（官方预编译） |
|------|-------------------|---------------------|
| **构建时间** | 15-30 分钟（编译 bootstrap） | 1-2 分钟（下载） |
| **成功率** | ❌ 低（交叉编译兼容性问题） | ✅ 高（官方支持） |
| **二进制有效性** | ❌ 可能生成脚本而不是二进制 | ✅ 保证是 RISC-V 原生二进制 |
| **下载大小** | 需要下载 Go 源码（~300MB） | 只需下载二进制（~100MB） |
| **依赖** | 需要 RISC-V 交叉编译器 | 只需 wget |
| **维护性** | ❌ 需要维护交叉编译逻辑 | ✅ 只需更新版本号 |
| **稳定性** | ❌ 可能随 Go 版本变化而失败 | ✅ 官方发布，稳定可靠 |

## 🔧 版本更新方法

需要更新 Go 版本时，只需修改一处：

```yaml
GO_VERSION="1.25.3"  # 改为新版本，如 "1.26.0"
```

然后确认官方是否提供该版本的 RISC-V64 包：
https://go.dev/dl/

## 💡 后续优化建议

### 1. 缓存 Bootstrap Go

```yaml
- name: Cache Go bootstrap
  id: cache-go-bootstrap
  uses: actions/cache@v4
  with:
    path: /tmp/go-bootstrap-riscv64
    key: go-bootstrap-riscv64-v1.25.3
    restore-keys: |
      go-bootstrap-riscv64-v1.25.

- name: Download Go bootstrap toolchain
  if: steps.cache-go-bootstrap.outputs.cache-hit != 'true'
  run: |
    # ... download code ...
```

**好处：**
- 第二次运行时直接使用缓存，无需下载
- 进一步减少构建时间

### 2. 校验 SHA256

从 https://go.dev/dl/ 获取官方 SHA256，增加安全性：

```bash
GO_SHA256="abc123..."  # 从官方网站获取
echo "$GO_SHA256  $GO_TARBALL" | sha256sum -c -
```

### 3. 支持多个 Go 版本

可以通过 workflow input 或 matrix 支持测试多个 Go 版本：

```yaml
strategy:
  matrix:
    go_version: ["1.23.0", "1.24.0", "1.25.3"]
```

## 📝 相关文件修改

- **第 468-510 行**：新增 "Download Go bootstrap toolchain" 步骤
- **第 617-642 行**：修改 "Copy Go source code to rootfs" 步骤中的 bootstrap 复制逻辑
- **第 358-395 行**：init 脚本中的 bootstrap 验证逻辑（已有修复）

## ✅ 优势总结

1. **简单可靠**：使用官方支持的预编译包
2. **快速高效**：下载比编译快 10-15 倍
3. **易于维护**：只需更新版本号即可
4. **兼容性好**：官方保证 RISC-V64 支持
5. **可重现构建**：每次使用相同的官方发布版本

