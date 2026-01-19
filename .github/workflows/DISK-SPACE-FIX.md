# 磁盘空间不足问题修复

## 🐛 问题描述

在 "Copy Go source code to rootfs" 步骤中出现：
```
cp: error writing 'mnt/go/...': No space left on device
```

### 根本原因分析

1. **磁盘镜像太小**：原来只有 512MB，无法容纳 Go 完整源代码和构建环境
2. **复制了不必要的文件**：
   - 编译产物（.o, .a, .so 文件）
   - 构建目录 `riscv64-linux/`（包含 QEMU、Kernel、Busybox 的源码和编译产物）
   - 测试数据目录 `testdata/`
   - 临时文件（.cmd 文件）
3. **Go 源代码库很大**：完整的 Go 源代码加上构建产物可能超过 1GB

## ✅ 修复方案

### 1. 增大磁盘镜像（第557行）

```yaml
# 修改前：
dd if=/dev/zero of=busybox.img bs=1M count=512  # 512MB

# 修改后：
dd if=/dev/zero of=busybox.img bs=1M count=3072  # 3GB (3072MB)
```

### 2. 优化文件排除规则（第575-596行）

**增强的排除规则：**

```yaml
sudo rsync -a --info=progress2 \
  --exclude='.git/' \              # Git 仓库（约 100-200MB）
  --exclude='.github/' \            # GitHub Actions 配置
  --exclude='riscv64-linux/' \      # 🔥 关键：临时构建目录（可能 1-2GB）
  --exclude='*.o' \                 # 编译的目标文件
  --exclude='*.a' \                 # 静态库文件
  --exclude='*.so' \                # 动态库文件
  --exclude='*.test' \              # 测试二进制文件
  --exclude='*.exe' \               # Windows 可执行文件
  --exclude='**/testdata/' \        # 测试数据目录
  --exclude='**/bin/' \             # 二进制目录
  --exclude='**/pkg/' \             # 包构建缓存
  --exclude='test/mytest/' \        # 自定义测试目录
  --exclude='**/.cmd' \             # Busybox 构建命令文件
  --exclude='**/*.cmd' \            # 其他 .cmd 文件
  "$GITHUB_WORKSPACE/" mnt/go/
```

**关键排除项说明：**

| 排除项 | 说明 | 预计节省空间 |
|--------|------|--------------|
| `riscv64-linux/` | 包含 QEMU、Kernel、Busybox 源码和编译产物 | ~1-2GB |
| `.git/` | Git 版本控制历史 | ~100-200MB |
| `*.o`, `*.a`, `*.so` | 编译产物（如果有） | ~50-100MB |
| `**/testdata/` | 测试数据文件 | ~20-50MB |
| `**/*.cmd` | Busybox 编译命令文件 | ~10-20MB |

### 3. 添加调试信息（第575-577行）

```yaml
# 显示 GITHUB_WORKSPACE 下的内容（便于排查）
echo "Contents of GITHUB_WORKSPACE:"
ls -la "$GITHUB_WORKSPACE" | head -20

# 显示 rsync 进度
sudo rsync -a --info=progress2 ...
```

### 4. 备用复制方案（第598-615行）

如果 rsync 失败，使用 `find` 命令只复制源代码文件：

```bash
sudo find . -type f \( 
  -name '*.go' -o 
  -name '*.c' -o 
  -name '*.h' -o 
  -name '*.s' -o 
  -name '*.S' -o 
  -name '*.sh' -o 
  -name 'Makefile' -o 
  -name '*.mod' -o 
  -name '*.sum' 
\) -not -path './.git/*' \
  -not -path './riscv64-linux/*' \
  -exec cp --parents {} .../mnt/go/ \;
```

## 📊 空间使用预估

| 组件 | 预估大小 |
|------|----------|
| Busybox rootfs | ~50MB |
| Go 源代码（过滤后） | ~300-500MB |
| Go bootstrap（如果复制） | ~100-200MB |
| 预留空间（编译产物） | ~1.5-2GB |
| **总计** | **~2-2.8GB** |

**磁盘镜像大小：3GB** ✅ 足够使用

## 🔍 验证步骤

修复后的 workflow 将输出：

```
Contents of GITHUB_WORKSPACE:
total 120
drwxr-xr-x  8 runner docker  4096 ... .
drwxr-xr-x  3 runner docker  4096 ... ..
drwxr-xr-x  2 runner docker  4096 ... api
...
(应该看不到 riscv64-linux 目录被列出)

Starting rsync with exclusions...
          0   0%    0.00kB/s    0:00:00  (xfr#0, to-chk=0/N)
         ...进度显示...

✅ Go source code copied to rootfs
Disk usage:
Filesystem      Size  Used Avail Use% Mounted on
/dev/loop0      3.0G  800M  2.1G  28% .../mnt
```

## 🎯 预期结果

- ✅ 磁盘镜像 3GB，足够容纳所有必要文件
- ✅ 只复制源代码文件，不复制编译产物和临时文件
- ✅ 排除 `riscv64-linux/` 目录，避免重复复制构建环境
- ✅ 磁盘使用率约 25-30%，留有充足空间供编译使用
- ✅ 复制时间大幅减少（从复制 2-3GB 减少到 500-800MB）

## 📝 相关代码位置

- **第 557 行**：磁盘镜像大小配置
- **第 575-603 行**：rsync 复制命令和排除规则
- **第 605-612 行**：复制 bootstrap Go（如果存在）
- **第 614-621 行**：设置权限和显示磁盘使用情况
