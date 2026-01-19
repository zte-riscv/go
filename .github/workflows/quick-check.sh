#!/bin/bash
# 快速检查 workflow 语法和基本配置

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "🔍 快速检查 workflow 配置..."
echo ""

# 检查文件是否存在
for file in qemu-riscv64.yml qemu-riscv64-simple.yml; do
    if [ -f "$file" ]; then
        echo "✅ $file 存在"
        
        # 检查基本语法
        if grep -q "pull_request:" "$file"; then
            echo "   ✅ 包含 pull_request 触发器"
        else
            echo "   ⚠️  未找到 pull_request 触发器"
        fi
        
        if grep -q "workflow_dispatch:" "$file"; then
            echo "   ✅ 包含 workflow_dispatch 触发器"
        else
            echo "   ⚠️  未找到 workflow_dispatch 触发器"
        fi
        
        # 检查是否有成功消息
        if grep -q "Success message" "$file"; then
            echo "   ✅ 包含成功消息步骤"
        else
            echo "   ⚠️  未找到成功消息步骤"
        fi
    else
        echo "❌ $file 不存在"
    fi
    echo ""
done

# 检查 yamllint（如果可用）
if command -v yamllint &> /dev/null; then
    echo "🔍 使用 yamllint 检查 YAML 语法..."
    yamllint qemu-riscv64*.yml 2>&1 || echo "⚠️  yamllint 发现一些问题"
else
    echo "💡 提示: 安装 yamllint 可以进行 YAML 语法检查"
    echo "   sudo apt-get install yamllint"
fi

echo ""
echo "✅ 快速检查完成！"
echo ""
echo "📝 下一步："
echo "   1. 安装 act: curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash"
echo "   2. 运行测试: ./test-local.sh"
echo "   3. 或手动运行: act workflow_dispatch -W .github/workflows/qemu-riscv64-simple.yml"

