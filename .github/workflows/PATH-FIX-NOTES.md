# 路径修复说明

## 问题原因

### 1. GitHub Workspace 路径问题

当 checkout Go 仓库时，`$GITHUB_WORKSPACE` 的实际路径是：
```
/home/runner/work/<repo-name>/<repo-name>/
```

对于这个仓库：
- 仓库名：`go`
- `$GITHUB_WORKSPACE` = `/home/runner/work/go/go/`

**错误代码：**
```bash
"$GITHUB_WORKSPACE/go/"  # → /home/runner/work/go/go/go/ ❌
```

**修复后：**
```bash
"$GITHUB_WORKSPACE/"     # → /home/runner/work/go/go/ ✅
```

### 2. Go 版本更新

- 旧版本：go1.21.0
- 新版本：go1.25.3

## 修复的代码位置

### 1. Copy Go source code to rootfs（第567-584行）

**修复前：**
```yaml
echo "Copying Go source from $GITHUB_WORKSPACE/go to rootfs..."
sudo rsync -a ... "$GITHUB_WORKSPACE/go/" mnt/go/
```

**修复后：**
```yaml
echo "Copying Go source from $GITHUB_WORKSPACE to rootfs..."
echo "GITHUB_WORKSPACE=$GITHUB_WORKSPACE"
ls -la "$GITHUB_WORKSPACE" | head -10  # 添加调试信息
sudo rsync -a ... "$GITHUB_WORKSPACE/" mnt/go/
```

### 2. Prepare Go bootstrap toolchain（第497-512行）

**修复前：**
```bash
if [ -d "$GITHUB_WORKSPACE/go/src" ]; then
  cp -r "$GITHUB_WORKSPACE/go" .
fi
wget -q "https://go.dev/dl/go1.21.0.src.tar.gz"
```

**修复后：**
```bash
if [ -d "$GITHUB_WORKSPACE/src" ]; then
  mkdir -p go
  cp -r "$GITHUB_WORKSPACE/"* go/
fi
wget -q "https://go.dev/dl/go1.25.3.src.tar.gz"
```

## 验证方法

workflow 运行时会输出：
```
Copying Go source from $GITHUB_WORKSPACE to rootfs...
GITHUB_WORKSPACE=/home/runner/work/go/go
total 120
drwxr-xr-x  8 runner docker  4096 ... .
drwxr-xr-x  3 runner docker  4096 ... ..
drwxr-xr-x  8 runner docker  4096 ... .git
-rw-r--r--  1 runner docker   123 ... README.md
drwxr-xr-x  2 runner docker  4096 ... src
...
```

## 测试建议

1. 检查 `$GITHUB_WORKSPACE` 的实际路径
2. 确认 Go 源代码正确复制到 `/go` 目录
3. 验证 VM 中能找到 `/go/src` 目录
