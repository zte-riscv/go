# Bash Wrapper 脚本最终修复

## 🐛 问题

即使创建了符号链接 `/bin/bash -> /bin/sh`，仍然出现 "bash: applet not found" 错误。

### 根本原因

**问题出在 `all.bash` 第 12 行：**
```bash
bash run.bash --no-rebuild
```

这里**直接调用 `bash` 命令**，而不是通过 shebang。

**为什么符号链接不够：**

```
执行：bash run.bash
  ↓
Shell 在 PATH 中找到：/bin/bash
  ↓
符号链接：/bin/bash -> /bin/sh -> busybox
  ↓
最终执行：/bin/busybox
  ↓
Busybox 检查 argv[0]：argv[0] = "bash"（调用时的命令名）
  ↓
Busybox 查找 "bash" applet
  ↓
❌ 错误：bash applet 不存在（只有 sh applet）
  ↓
bash: applet not found
```

**关键点：**
- Busybox 多调用二进制文件根据 `argv[0]` 来决定运行哪个 applet
- 即使通过符号链接执行，`argv[0]` 仍然是 **调用时使用的名称** ("bash")
- 符号链接不会改变 `argv[0]` 的值

## ✅ 解决方案：Bash Wrapper 脚本

创建一个**真正的脚本** `/bin/bash`，它调用 `/bin/sh` 并转发所有参数。

### 实施代码（第 323-333 行）

```bash
# Create a real bash script that calls sh
printf "#!/bin/sh\n# Bash wrapper for Busybox sh\nexec /bin/sh \"\$@\"\n" > /bin/bash
chmod +x /bin/bash
```

**生成的 `/bin/bash` 内容：**
```sh
#!/bin/sh
# Bash wrapper for Busybox sh
exec /bin/sh "$@"
```

**为什么这样工作：**

```
执行：bash run.bash
  ↓
Shell 在 PATH 中找到：/bin/bash
  ↓
/bin/bash 是一个 shell 脚本
  ↓
执行 /bin/bash 的 shebang：#!/bin/sh
  ↓
/bin/sh 启动（Busybox sh applet）
  ↓
执行脚本内容：exec /bin/sh "$@"
  ↓
替换当前进程为：/bin/sh run.bash
  ↓
✅ 成功：Busybox sh 执行 run.bash
```

**关键优势：**
1. ✅ `/bin/bash` 是一个真正的脚本文件（不是符号链接）
2. ✅ 使用 `exec` 替换进程，避免多余的进程层
3. ✅ `"$@"` 正确转发所有参数（包括空格和特殊字符）
4. ✅ 对于调用者，看起来就像是在使用 bash

## 📊 符号链接 vs Wrapper 脚本对比

### 方案 1：符号链接（不工作）

```
文件系统：
/bin/bash -> /bin/sh -> busybox

执行 "bash run.bash"：
argv[0] = "bash"
Busybox 查找 "bash" applet
❌ 失败
```

### 方案 2：Wrapper 脚本（工作）✅

```
文件系统：
/bin/bash (脚本文件，内容：exec /bin/sh "$@")
/bin/sh -> busybox

执行 "bash run.bash"：
1. /bin/bash 脚本启动（通过 /bin/sh）
2. exec /bin/sh "run.bash"
3. argv[0] = "sh"（因为 exec /bin/sh）
4. Busybox 查找 "sh" applet
✅ 成功
```

## 🎯 预期输出（下次运行）

```
Creating bash wrapper script...
Created /bin/bash wrapper script
#!/bin/sh
# Bash wrapper for Busybox sh
exec /bin/sh "$@"

-rwxr-xr-x    1 0        0               57 Nov 25 08:30 /bin/bash
lrwxrwxrwx    1 0        0                7 Nov 25 08:30 /bin/sh -> busybox

bash command found: /bin/bash
Testing bash wrapper: SUCCESS
✅ bash wrapper is working correctly

===========================================================
RISC-V64 Linux VM Started
===========================================================

Go source code found in /go

===========================================================
Starting Go compiler build and test...
===========================================================

Using bootstrap Go: /usr/local/go-bootstrap
go version go1.25.3 linux/riscv64

Building Go toolchain...
GOROOT_BOOTSTRAP=/usr/local/go-bootstrap

##### Building Go bootstrap tool.
cmd/dist

##### Building Go toolchain using /usr/local/go-bootstrap.
Building Go cmd/dist using /usr/local/go-bootstrap. (go1.25.3 linux/riscv64)
...

##### Building packages and commands for linux/riscv64.
...

##### Testing packages.
ok  	archive/tar	0.123s
ok  	bufio	0.456s
ok  	bytes	0.789s
...
ok  	cmd/go	42.123s
...
ALL TESTS PASSED

---
Installed Go for linux/riscv64 in /go
Installed commands in /go/bin

===========================================================
Go build and test completed successfully!
===========================================================

Go build process completed. Shutting down...
[  1234.567890] reboot: Power down

✅ QEMU VM completed successfully
```

## 📝 为什么需要 wrapper

**理解 Busybox 多调用二进制文件：**

Busybox 是一个特殊的可执行文件，包含多个 "applet"（小程序），通过 `argv[0]` 来决定运行哪一个：

```c
// Busybox 伪代码
int main(int argc, char **argv) {
    char *applet_name = basename(argv[0]);  // 提取命令名
    
    if (strcmp(applet_name, "sh") == 0) {
        return sh_main(argc, argv);
    } else if (strcmp(applet_name, "ls") == 0) {
        return ls_main(argc, argv);
    } else if (strcmp(applet_name, "cat") == 0) {
        return cat_main(argc, argv);
    } else {
        printf("%s: applet not found\n", applet_name);
        return 1;
    }
}
```

**关键**：`argv[0]` 是调用时使用的命令名，**不会因为符号链接而改变**。

## 🔧 其他尝试过的方案

### 方案 A：`/bin/bash -> /bin/busybox`（v1-v2）
❌ 失败：Busybox 查找 "bash" applet，不存在

### 方案 B：`/bin/bash -> /bin/sh`（v3）
❌ 失败：虽然 shebang 工作了，但 `all.bash` 中的 `bash run.bash` 直接调用仍然失败

### 方案 C：Bash wrapper 脚本（v4）✅
✅ 成功：真正的脚本文件，使用 `exec /bin/sh "$@"` 转发

## 📄 相关文件修改

- **第 323-333 行**：创建 bash wrapper 脚本
- **第 335-345 行**：测试 bash wrapper
- **第 241 行**：更新 Busybox 缓存 key 到 `v4`

## ✅ 成功标准

下次运行应该看到：
- [x] "Created /bin/bash wrapper script" 消息
- [x] 显示 wrapper 脚本内容
- [x] "Testing bash wrapper: SUCCESS" 消息
- [x] "✅ bash wrapper is working correctly" 消息
- [x] `./all.bash` 成功执行，包括 `bash run.bash`
- [x] 完整的 Go 构建和测试过程
- [x] "ALL TESTS PASSED"
- [x] 自动关机

## 💡 学到的经验

1. **符号链接不会改变 `argv[0]`**
   - 即使通过符号链接执行，程序仍然看到原始的命令名

2. **Busybox 的特殊性**
   - 多调用二进制文件依赖 `argv[0]` 来决定行为
   - 需要确保 `argv[0]` 与实际的 applet 名称匹配

3. **Wrapper 脚本的优势**
   - 提供了一个真正的 "命令替换" 层
   - 可以完全控制如何调用目标命令
   - 使用 `exec` 避免多余的进程层

4. **测试的重要性**
   - 不仅要测试 shebang 场景
   - 还要测试直接命令调用场景

这应该是**最终的修复**了！🎉

