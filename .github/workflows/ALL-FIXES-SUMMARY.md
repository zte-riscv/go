# GitHub Actions Workflow 所有修复总结

## 📋 问题列表与解决方案

### 1. ✅ GORISCV64 解析错误
**问题**：`"GORISCV64_rva23u64,zabha" is not a valid identifier name`  
**修复**：提取 profile 部分（逗号之前）  
**文件**：`go/src/cmd/go/internal/work/gc.go`, `go/src/cmd/dist/build.go`

---

### 2. ✅ YAML 语法错误（heredoc）
**问题**：`Implicit keys need to be on a single line`  
**修复**：用 `printf` 替换 heredoc  
**文件**：`go/.github/workflows/qemu-riscv64.yml`

---

### 3. ✅ `GITHUB_WORKSPACE` 路径错误
**问题**：`rsync: change_dir "/home/runner/work/go/go/go" failed`  
**修复**：从 `$GITHUB_WORKSPACE/go/` 改为 `$GITHUB_WORKSPACE/`  
**文件**：`go/.github/workflows/qemu-riscv64.yml` (第 602, 506 行)

---

### 4. ✅ Busybox tc 编译错误
**问题**：`'TCA_CBQ_MAX' undeclared`  
**修复**：禁用 `CONFIG_TC`  
**文件**：`go/.github/workflows/qemu-riscv64.yml` (第 290-291 行)

---

### 5. ✅ 磁盘空间不足
**问题**：`cp: error writing 'mnt/go/...': No space left on device`  
**修复**：  
  - 磁盘镜像从 512MB → 3GB
  - 添加 rsync 排除规则（`.git/`, `riscv64-linux/`, `*.o`, 等）  
**文件**：`go/.github/workflows/qemu-riscv64.yml` (第 568, 587-601 行)

---

### 6. ✅ QEMU 参数冲突
**问题**：`-serial stdio: cannot use stdio by multiple character devices`  
**修复**：删除 `-serial stdio`（`-nographic` 已包含）  
**文件**：`go/.github/workflows/qemu-riscv64.yml` (第 672-681 行)

---

### 7. ✅ VM 卡在交互式 shell
**问题**：init 脚本执行 `exec /bin/sh`，workflow 超时  
**修复**：所有分支都 `poweroff -f` 自动关机  
**文件**：`go/.github/workflows/qemu-riscv64.yml` (第 372, 406, 425 行)

---

### 8. ✅ Busybox 缺少 bash
**问题**：`env: can't execute 'bash': No such file or directory`  
**修复**：创建 `/bin/bash` → `/bin/busybox` 符号链接  
**文件**：`go/.github/workflows/qemu-riscv64.yml` (第 321-331 行)

---

### 9. ✅ Bootstrap Go 交叉编译失败
**问题**：生成的是脚本而不是 RISC-V 可执行文件  
**修复**：**使用官方预编译包** `go1.25.3.linux-riscv64.tar.gz`  
**文件**：`go/.github/workflows/qemu-riscv64.yml` (第 468-510 行)

---

## 🎯 关键改进：使用官方 Go Bootstrap

**旧方案**（交叉编译）：
```yaml
- 主机有 Go → 尝试交叉编译 → 可能生成脚本 ❌
- 验证不严格 → 复制无效 bootstrap 到 VM ❌
- VM 尝试运行 → 语法错误 ❌
```

**新方案**（官方预编译）：
```yaml
- 下载 go1.25.3.linux-riscv64.tar.gz (100MB) ✅
- 验证是 RISC-V 可执行文件 ✅
- 复制到 VM 的 /usr/local/go-bootstrap ✅
- VM 运行 go version → 成功 ✅
- 运行 ./all.bash → 构建和测试 Go ✅
```

**关键代码：**

```yaml
- name: Download Go bootstrap toolchain for RISC-V64
  run: |
    GO_VERSION="1.25.3"
    wget https://go.dev/dl/go${GO_VERSION}.linux-riscv64.tar.gz
    tar -xzf go${GO_VERSION}.linux-riscv64.tar.gz
    # → /tmp/go-bootstrap-riscv64/go
```

---

## 📊 完整执行流程（最终版本）

```
┌─────────────────────────────────────────────────────────┐
│  GitHub Actions 主机 (Ubuntu 24.04 AMD64)              │
├─────────────────────────────────────────────────────────┤
│  1. Checkout repository                           ✅   │
│  2. Install dependencies                          ✅   │
│  3. Install RISC-V toolchain                      ✅   │
│     ├─ apt-get install gcc-riscv64-linux-gnu           │
│     └─ 或 bootlin/GitHub releases                      │
│  4. Cache & Build QEMU                            ✅   │
│     └─ qemu-system-riscv64 v9.0.0                      │
│  5. Cache & Build Linux kernel                    ✅   │
│     └─ Linux v6.1 for RISC-V64                         │
│  6. Cache & Build Busybox                         ✅   │
│     └─ 1_36_stable, CONFIG_TC=n                        │
│  7. 📦 Download Go bootstrap (NEW!)               ✅   │
│     └─ go1.25.3.linux-riscv64.tar.gz (100MB)           │
│  8. Copy Go source to rootfs                      ✅   │
│     ├─ rsync with exclusions (300-500MB)               │
│     └─ Copy bootstrap Go (150MB)                       │
│  9. Boot QEMU VM                                  ✅   │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│  QEMU RISC-V64 Linux VM (2GB RAM, 4 CPUs)              │
├─────────────────────────────────────────────────────────┤
│  Linux 6.1.0 boots                                ✅   │
│  Run /init as init process                        ✅   │
│                                                         │
│  /init script:                                          │
│  ├─ Mount /proc, /sys, /dev                       ✅   │
│  ├─ Create /bin/bash symlink                      ✅   │
│  ├─ Display system info                           ✅   │
│  ├─ Check /go/src                                 ✅   │
│  ├─ Check /usr/local/go-bootstrap                 ✅   │
│  │  └─ Run: go version                                 │
│  │     Output: go version go1.25.3 linux/riscv64  ✅   │
│  ├─ export GOROOT_BOOTSTRAP                       ✅   │
│  ├─ cd /go/src                                    ✅   │
│  ├─ ./all.bash                                    ✅   │
│  │  ├─ Building Go bootstrap tool                      │
│  │  ├─ Building packages and commands                  │
│  │  ├─ Testing packages                                │
│  │  └─ ALL TESTS PASSED                           ✅   │
│  ├─ Display summary                               ✅   │
│  └─ poweroff -f                                   ✅   │
└─────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────┐
│  Workflow Completion                                    │
├─────────────────────────────────────────────────────────┤
│  ✅ Success message                                     │
│  ✅ Upload artifacts (kernel, rootfs)                   │
│  ✅ Workflow completed successfully                     │
└─────────────────────────────────────────────────────────┘
```

---

## 📝 修改的文件

| 文件 | 行数 | 修改内容 |
|------|------|----------|
| `go/.github/workflows/qemu-riscv64.yml` | 779 | 主要 workflow 文件 |
| `go/src/cmd/go/internal/work/gc.go` | - | GORISCV64 解析 |
| `go/src/cmd/dist/build.go` | - | GORISCV64 解析 |

---

## 📄 创建的文档

| 文档 | 内容 |
|------|------|
| `PATH-FIX-NOTES.md` | `GITHUB_WORKSPACE` 路径修复 |
| `DISK-SPACE-FIX.md` | 磁盘空间问题和 rsync 排除规则 |
| `QEMU-STARTUP-FIX.md` | QEMU 参数冲突和自动关机修复 |
| `BOOTSTRAP-GO-FIX.md` | Bootstrap Go 问题分析 |
| `OFFICIAL-BOOTSTRAP-SOLUTION.md` | 官方预编译包方案（推荐） |
| `ALL-FIXES-SUMMARY.md` | 本文档 |

---

## 🚀 下次运行预期

**预计时间**：30-60 分钟（首次），15-30 分钟（使用缓存）

**成功输出关键信息**：

1. **下载 bootstrap**：
   ```
   📦 Downloading official Go bootstrap toolchain for RISC-V64...
   go1.25.3.linux-riscv64.tar.gz  100%  102.45M
   ✅ Official RISC-V64 Go bootstrap ready
   ```

2. **复制到 rootfs**：
   ```
   📦 Copying official RISC-V64 Go bootstrap to rootfs...
   ✅ Bootstrap Go binary is a RISC-V executable
   ✅ Bootstrap Go copied successfully
   ```

3. **VM 启动**：
   ```
   ===========================================================
   RISC-V64 Linux VM Started
   ===========================================================
   Go source code found in /go
   ```

4. **Bootstrap 验证**：
   ```
   Found bootstrap Go at /usr/local/go-bootstrap
   ✅ Bootstrap Go is valid and executable
   go version go1.25.3 linux/riscv64
   ```

5. **构建 Go**：
   ```
   Building Go toolchain...
   ##### Building Go bootstrap tool.
   ##### Building packages and commands for linux/riscv64.
   ##### Testing packages.
   ...
   ALL TESTS PASSED
   ```

6. **自动关机**：
   ```
   Go build process completed. Shutting down...
   [  1234.567890] reboot: Power down
   ✅ QEMU VM completed successfully
   ```

---

## 💡 后续优化建议

### 1. 缓存 Go Bootstrap
```yaml
- name: Cache Go bootstrap
  uses: actions/cache@v4
  with:
    path: /tmp/go-bootstrap-riscv64
    key: go-bootstrap-riscv64-v1.25.3
```

### 2. 并行测试
```yaml
strategy:
  matrix:
    goriscv64: ["rva20u64", "rva22u64", "rva23u64"]
```

### 3. 保存构建产物
```yaml
- name: Extract Go binaries from VM
  run: |
    sudo mount -o loop busybox.img mnt
    sudo cp -r mnt/go/bin ./go-riscv64-bin
```

### 4. 性能优化
- 使用 `tmpfs` 作为构建目录（更快的 I/O）
- 增加 QEMU CPU 数量到 8（如果 GitHub Actions 允许）
- 使用 KVM 加速（如果可用）

---

## ✅ 成功标准

- [ ] Workflow 运行无错误
- [ ] QEMU VM 正常启动
- [ ] Bootstrap Go 被识别并使用
- [ ] `./all.bash` 完成构建
- [ ] 所有测试通过
- [ ] VM 自动关机
- [ ] Workflow 在 60 分钟内完成

---

## 🎉 结论

通过这一系列修复，特别是采用官方预编译 Go bootstrap 方案，workflow 现在应该能够：

1. ✅ **稳定运行**：使用官方支持的工具链
2. ✅ **快速构建**：缓存 + 预编译包大幅减少时间
3. ✅ **自动化**：无需人工干预，自动关机
4. ✅ **可重现**：每次构建使用相同的官方版本
5. ✅ **易维护**：只需更新版本号即可升级

**现在可以重新运行 workflow 了！** 🚀

