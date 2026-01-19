# 最新修复：Bash 符号链接 + 缓存问题

## 🐛 问题

虽然使用了官方 Go bootstrap，但仍然出现：
```
env: can't execute 'bash': No such file or directory
VM ready for interactive use. Type exit to shutdown.
/go/src #
```

## 🔍 原因

1. **旧的 Busybox rootfs 被缓存了** 
   - 包含旧的 init 脚本（有 "VM ready for interactive use" 消息）
   - 没有正确创建 `/bin/bash` 符号链接

2. **符号链接创建逻辑不够强**
   - 条件检查 `[ ! -e /bin/bash ]` 可能失败
   - 缺少调试输出

## ✅ 修复

### 1. 增强 bash 符号链接创建（第 323-347 行）

```bash
# 始终创建（移除条件检查）
ln -sf /bin/busybox /bin/bash
ln -sf /bin/env /usr/bin/env

# 添加详细调试输出
echo "Created /bin/bash -> /bin/busybox"
echo "Created /usr/bin/env -> /bin/env"

# 验证 bash 可用
command -v bash && echo "bash is available: $(which bash)"
```

### 2. 更新 Busybox 缓存 key（第 241 行）

```yaml
# 从
key: busybox-riscv64-1.36_stable-${{ runner.os }}

# 改为
key: busybox-riscv64-1.36_stable-v2-${{ runner.os }}
```

强制重新构建，使用新的 init 脚本。

## 🚀 下次运行预期

```
Creating bash and env symlinks...
Created /bin/bash -> /bin/busybox
Created /usr/bin/env -> /bin/env
bash is available: /bin/bash

Using bootstrap Go: /usr/local/go-bootstrap
go version go1.25.3 linux/riscv64

Building Go toolchain...

##### Building Go bootstrap tool.
##### Building packages and commands for linux/riscv64.
##### Testing packages.
...
ALL TESTS PASSED

Go build process completed. Shutting down...
✅ QEMU VM completed successfully
```

## 📝 所有已实施修复总结

1. ✅ GORISCV64 解析错误 → 提取 profile 部分
2. ✅ YAML heredoc 错误 → 使用 `printf`
3. ✅ GITHUB_WORKSPACE 路径 → 修正路径
4. ✅ Busybox tc 编译错误 → 禁用 CONFIG_TC
5. ✅ 磁盘空间不足 → 3GB + rsync 排除规则
6. ✅ QEMU 参数冲突 → 删除 `-serial stdio`
7. ✅ VM 卡在 shell → 所有分支 `poweroff -f`
8. ✅ **Bootstrap 交叉编译失败 → 使用官方预编译包** ⭐
9. ✅ **Bash 符号链接 → 强制创建 + 调试输出 + v2 缓存** ⭐ (NEW)

## 📊 关键改进时间线

```
首次运行 → 多个错误（GORISCV64, YAML, 路径等）
  ↓ 修复
第二次 → 磁盘空间不足、QEMU 冲突
  ↓ 修复
第三次 → Bootstrap Go 交叉编译失败
  ↓ 使用官方预编译包 (你的建议！)
第四次 → Bash 符号链接 + 旧缓存
  ↓ 增强创建逻辑 + 更新缓存 v2
第五次 → 应该成功！🎉
```

## 🎯 预计下次运行

**状态：** 应该完全成功！

**关键指标：**
- [ ] Bootstrap Go 下载成功（~100MB）
- [ ] Busybox 重新构建（使用 v2 缓存）
- [ ] Bash 符号链接创建成功
- [ ] `./all.bash` 执行成功
- [ ] 所有测试通过
- [ ] 自动关机（无交互式 shell）

**预计时间：** 40-60 分钟（首次使用 v2 缓存）

