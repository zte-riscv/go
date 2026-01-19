# Bootstrap Go 问题修复

## 🐛 当前问题

从 QEMU 输出看到以下错误：

```
Using bootstrap Go: /usr/local/go-bootstrap
/usr/local/go-bootstrap/bin/go: line 2: syntax error: unexpected word (expecting ")")
Building Go toolchain...
GOROOT_BOOTSTRAP=/usr/local/go-bootstrap
env: can't execute 'bash': No such file or directory
```

### 问题分析

#### 问题 1：Bootstrap Go 二进制文件无效
- `/usr/local/go-bootstrap/bin/go` 是一个脚本而不是 RISC-V 可执行二进制文件
- 这是因为在主机上的交叉编译没有正确生成 RISC-V 原生二进制

**根本原因：**
- 在 GitHub Actions 主机（AMD64 Linux）上使用 `GOOS=linux GOARCH=riscv64 ./make.bash` 交叉编译时
- 虽然可以编译，但生成的可能是包装脚本而不是原生 RISC-V 可执行文件
- 或者交叉编译失败了，但错误被忽略了

#### 问题 2：Busybox 没有 bash
- Busybox 默认只提供 `sh`（ash shell），不是完整的 `bash`
- Go 的 `all.bash` 脚本需要 bash (`#!/usr/bin/env bash`)

**影响：**
```bash
env: can't execute 'bash': No such file or directory
```

## ✅ 修复方案

### 修复 1：验证 Bootstrap Go 二进制（第 539-553 行）

**修改前：**
```bash
if [ -f ../../bin/go ]; then
  echo "✅ Bootstrap Go built successfully!"
  ../../bin/go version
else
  echo "⚠️  Bootstrap binary not found after build"
fi
```

**修改后：**
```bash
if [ -f ../go/bin/go ]; then
  # 检查是否是 RISC-V 二进制文件
  if file ../go/bin/go | grep -qi "RISC-V" && file ../go/bin/go | grep -qi "executable"; then
    echo "✅ Bootstrap Go built successfully!"
    file ../go/bin/go
    ls -lh ../go/bin/go
  else
    echo "⚠️  bin/go is not a RISC-V executable binary:"
    file ../go/bin/go || echo "file command failed"
    echo "Skipping bootstrap (will need to build in VM or provide manually)"
    rm -rf ../go
    exit 0
  fi
else
  echo "⚠️  Bootstrap binary not found after build"
  exit 0
fi
```

**改进：**
- ✅ 使用 `file` 命令验证二进制文件类型
- ✅ 确保是 RISC-V 可执行文件（不是脚本）
- ✅ 如果无效，删除并跳过（不复制到 rootfs）

### 修复 2：创建 bash 符号链接（第 321-331 行）

**在 init 脚本中添加：**
```bash
# Create bash symlink (Go scripts need bash)
if [ ! -e /bin/bash ] && [ -e /bin/busybox ]; then
  ln -sf /bin/busybox /bin/bash
fi
if [ ! -e /usr/bin/env ] && [ -e /bin/env ]; then
  mkdir -p /usr/bin
  ln -sf /bin/env /usr/bin/env
fi
```

**说明：**
- Busybox 的 `sh` 支持大部分 bash 特性
- 创建 `/bin/bash` → `/bin/busybox` 符号链接
- 创建 `/usr/bin/env` → `/bin/env` 符号链接（Go 脚本使用 `#!/usr/bin/env bash`）

### 修复 3：改进 Bootstrap Go 验证（第 358-395 行）

**在 VM 的 init 脚本中：**
```bash
# Check for bootstrap Go
BOOTSTRAP_VALID=false
if [ -f /usr/local/go-bootstrap/bin/go ]; then
  echo "Found bootstrap Go at /usr/local/go-bootstrap"
  echo "Checking if it is a valid RISC-V binary..."
  if /usr/local/go-bootstrap/bin/go version 2>/dev/null; then
    export GOROOT_BOOTSTRAP=/usr/local/go-bootstrap
    export PATH=$GOROOT_BOOTSTRAP/bin:$PATH
    BOOTSTRAP_VALID=true
    echo "✅ Bootstrap Go is valid and executable"
    echo "Using bootstrap Go: $GOROOT_BOOTSTRAP"
    $GOROOT_BOOTSTRAP/bin/go version
  else
    echo "⚠️  Bootstrap Go binary exists but cannot execute"
    echo "File info:"
    ls -lh /usr/local/go-bootstrap/bin/go
    file /usr/local/go-bootstrap/bin/go 2>/dev/null || echo "file command not available"
    head -5 /usr/local/go-bootstrap/bin/go 2>/dev/null || true
  fi
fi

if [ "$BOOTSTRAP_VALID" = "false" ]; then
  echo ""
  echo "❌ ERROR: No valid bootstrap Go found"
  echo ""
  echo "Bootstrap Go is required to build the Go toolchain."
  echo "Please provide a pre-compiled Go 1.21+ for RISC-V64."
  # ... (详细错误信息和指导)
  poweroff -f
  exit 1
fi
```

**改进：**
- ✅ 尝试执行 `go version` 来验证二进制有效性
- ✅ 如果无效，显示详细的调试信息（file 类型、文件头等）
- ✅ 提供清晰的错误信息和手动构建指导

### 修复 4：只复制有效的 Bootstrap（第 632-650 行）

**修改后：**
```bash
# 复制 bootstrap Go（如果已构建并验证）
if [ -d /tmp/go-bootstrap-riscv64/go ] && [ -f /tmp/go-bootstrap-riscv64/go/bin/go ]; then
  echo "Validating bootstrap Go binary..."
  if file /tmp/go-bootstrap-riscv64/go/bin/go | grep -qi "RISC-V" && \
     file /tmp/go-bootstrap-riscv64/go/bin/go | grep -qi "executable"; then
    echo "✅ Bootstrap Go binary is valid RISC-V executable"
    echo "Copying bootstrap Go to rootfs..."
    sudo mkdir -p mnt/usr/local
    sudo cp -r /tmp/go-bootstrap-riscv64/go mnt/usr/local/go-bootstrap
    # ...
  else
    echo "⚠️  Bootstrap Go binary is not a valid RISC-V executable, skipping copy"
    file /tmp/go-bootstrap-riscv64/go/bin/go || echo "file command failed"
  fi
else
  echo "⚠️  No bootstrap Go found in /tmp/go-bootstrap-riscv64"
  echo "VM will need a pre-compiled RISC-V64 Go bootstrap to build"
fi
```

**改进：**
- ✅ 在复制前验证二进制有效性
- ✅ 只复制有效的 RISC-V 可执行文件
- ✅ 如果无效，给出清晰的警告

## 📊 执行流程（修复后）

### 主机端（GitHub Actions）

```
Install RISC-V toolchain ✅
  ↓
Build QEMU ✅
  ↓
Build Linux kernel ✅
  ↓
Build Busybox ✅
  ↓
Prepare Go bootstrap toolchain
  ↓
  ├─ 尝试交叉编译 Go for RISC-V64
  ├─ 验证生成的 bin/go 是否是 RISC-V 可执行文件
  │  ├─ ✅ 是 → 继续
  │  └─ ❌ 否 → 删除，跳过（不复制到 VM）
  └─ 如果验证失败 → 显示警告，需要手动提供 bootstrap
  ↓
Copy Go source to rootfs ✅
  ↓
  ├─ 验证 bootstrap Go（如果存在）
  │  ├─ ✅ 有效 RISC-V 二进制 → 复制到 rootfs
  │  └─ ❌ 无效 → 跳过，显示警告
  └─ 复制 Go 源代码（排除不必要的文件）
  ↓
Boot QEMU VM
```

### VM 端（RISC-V64 Linux）

```
启动 Linux ✅
  ↓
执行 /init 脚本
  ↓
挂载文件系统 ✅
  ↓
创建 /bin/bash → /bin/busybox 符号链接 ✅
创建 /usr/bin/env → /bin/env 符号链接 ✅
  ↓
检查 Go 源代码 (/go/src) ✅
  ↓
检查 Bootstrap Go (/usr/local/go-bootstrap)
  ↓
  ├─ 尝试执行 go version
  │  ├─ ✅ 成功 → BOOTSTRAP_VALID=true
  │  │   ↓
  │  │   运行 ./all.bash
  │  │   ↓
  │  │   显示结果
  │  │   ↓
  │  │   poweroff ✅
  │  │
  │  └─ ❌ 失败 → BOOTSTRAP_VALID=false
  │      ↓
  │      显示详细错误信息（file 类型、文件头）
  │      ↓
  │      显示手动构建指导
  │      ↓
  │      poweroff ✅
  │
  └─ 不存在 → BOOTSTRAP_VALID=false
      ↓
      显示错误和指导
      ↓
      poweroff ✅
```

## 🎯 当前状态

### 如果 Bootstrap 交叉编译成功

```
✅ Bootstrap Go built successfully!
✅ Bootstrap Go binary is valid RISC-V executable
✅ Bootstrap Go copied to rootfs

(在 VM 中)
Found bootstrap Go at /usr/local/go-bootstrap
Checking if it is a valid RISC-V binary...
✅ Bootstrap Go is valid and executable
Using bootstrap Go: /usr/local/go-bootstrap
go version go1.25.3 linux/riscv64

Building Go toolchain...
##### Building Go bootstrap tool.
...
```

### 如果 Bootstrap 交叉编译失败（当前情况）

```
⚠️  bin/go is not a RISC-V executable binary:
/tmp/go-bootstrap-riscv64/go/bin/go: Bourne-Again shell script, ASCII text executable
Skipping bootstrap (will need to build in VM or provide manually)

⚠️  No bootstrap Go found in /tmp/go-bootstrap-riscv64
VM will need a pre-compiled RISC-V64 Go bootstrap to build

(在 VM 中)
⚠️  Bootstrap Go binary exists but cannot execute
File info:
-rwxr-xr-x 1 root root 123 ... /usr/local/go-bootstrap/bin/go
/usr/local/go-bootstrap/bin/go: Bourne-Again shell script, ASCII text executable
#!/usr/bin/env bash
...

❌ ERROR: No valid bootstrap Go found

Bootstrap Go is required to build the Go toolchain.
Please provide a pre-compiled Go 1.21+ for RISC-V64.

To manually build bootstrap Go:
  1. On a machine with Go installed:
     cd go/src
     GOOS=linux GOARCH=riscv64 ./bootstrap.bash
  2. Copy the generated bootstrap package to the VM

Shutting down VM...
```

## 🔧 手动提供 Bootstrap Go 的方法

### 方法 1：使用 bootstrap.bash（推荐）

在有 Go 的机器上：

```bash
cd go/src
GOOS=linux GOARCH=riscv64 ./bootstrap.bash
```

这会生成 `../../go-linux-riscv64-bootstrap.tbz`

然后：
1. 解压到 `riscv64-linux/mnt/usr/local/go-bootstrap`
2. 或者修改 workflow，在 "Copy Go source code to rootfs" 步骤前解压

### 方法 2：在 Docker 中使用 RISC-V 模拟器

```bash
docker run --rm -v $(pwd):/workspace \
  riscv64/ubuntu:22.04 \
  bash -c "cd /workspace/go/src && ./make.bash"
```

### 方法 3：下载预编译的 RISC-V64 Go（如果可用）

检查是否有社区提供的预编译包：
- https://github.com/carlosedp/riscv-bringup
- https://github.com/golang/go/issues?q=riscv64

## 📝 相关代码位置

- **第 539-553 行**：验证 bootstrap Go 二进制（主机端）
- **第 632-650 行**：只复制有效的 bootstrap（主机端）
- **第 321-331 行**：创建 bash 符号链接（VM init 脚本）
- **第 358-395 行**：验证 bootstrap Go（VM init 脚本）

## 💡 后续改进建议

1. **使用预构建的 Bootstrap Go**
   - 在 workflow 中下载预编译的 RISC-V64 Go 包
   - 或者使用 Docker 镜像中的 Go

2. **改进交叉编译**
   - 使用 `bootstrap.bash` 而不是 `make.bash`
   - 或者在 QEMU 用户模式中编译

3. **缓存 Bootstrap Go**
   - 使用 GitHub Actions cache 缓存成功构建的 bootstrap
   - 避免每次都重新编译

