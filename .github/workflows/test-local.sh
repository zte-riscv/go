#!/bin/bash
# 本地验证 GitHub Actions Workflow 脚本

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
WORKFLOW_DIR="$SCRIPT_DIR"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

echo "═══════════════════════════════════════════════════════════"
echo "🔍 GitHub Actions Workflow 本地验证工具"
echo "═══════════════════════════════════════════════════════════"
echo ""

# 检查 act 是否安装
if ! command -v act &> /dev/null; then
    echo "❌ act 未安装"
    echo ""
    echo "安装方法："
    echo "  curl https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash"
    echo ""
    echo "或者使用包管理器："
    echo "  sudo apt-get install act  # Ubuntu/Debian"
    echo "  brew install act          # macOS"
    echo ""
    exit 1
fi

echo "✅ act 已安装: $(act --version)"
echo ""

# 检查 Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker 未安装"
    exit 1
fi

if ! docker ps &> /dev/null; then
    echo "❌ Docker 未运行，请启动 Docker 服务"
    exit 1
fi

echo "✅ Docker 正在运行"
echo ""

# 切换到项目根目录
cd "$PROJECT_ROOT/go"

# 列出可用的 workflows
echo "📋 可用的 workflows:"
act -l 2>/dev/null | grep -E "qemu-riscv64|Event|Job ID" || echo "  未找到 qemu-riscv64 workflows"
echo ""

# 询问用户要运行哪个 workflow
echo "请选择要运行的 workflow:"
echo "  1) qemu-riscv64-simple.yml (简化版本，快速)"
echo "  2) qemu-riscv64.yml (完整版本，较慢)"
echo ""
read -p "请输入选项 (1 或 2，默认 1): " choice
choice=${choice:-1}

case $choice in
    1)
        WORKFLOW_FILE=".github/workflows/qemu-riscv64-simple.yml"
        JOB_NAME="qemu-riscv64-simple"
        echo "✅ 选择: 简化版本"
        ;;
    2)
        WORKFLOW_FILE=".github/workflows/qemu-riscv64.yml"
        JOB_NAME="qemu-riscv64"
        echo "✅ 选择: 完整版本"
        ;;
    *)
        echo "❌ 无效选项"
        exit 1
        ;;
esac

echo ""
echo "🚀 开始运行 workflow..."
echo "═══════════════════════════════════════════════════════════"
echo ""

# 运行 workflow
# 使用 dry-run 模式先检查
if [ "${1:-}" = "--dry-run" ]; then
    echo "🔍 干运行模式（仅检查，不实际执行）"
    act workflow_dispatch -W "$WORKFLOW_FILE" --dry-run
else
    # 实际运行
    echo "⚠️  注意: 完整运行可能需要很长时间（特别是完整版本）"
    echo "   按 Ctrl+C 可以随时中断"
    echo ""
    read -p "按 Enter 继续，或 Ctrl+C 取消..."
    
    act workflow_dispatch \
        -W "$WORKFLOW_FILE" \
        -j "$JOB_NAME" \
        -P ubuntu-latest=catthehacker/ubuntu:act-latest \
        --container-architecture linux/amd64
    
    if [ $? -eq 0 ]; then
        echo ""
        echo "═══════════════════════════════════════════════════════════"
        echo "✅ ✅ ✅ Workflow 本地验证成功！ ✅ ✅ ✅"
        echo "═══════════════════════════════════════════════════════════"
    else
        echo ""
        echo "═══════════════════════════════════════════════════════════"
        echo "❌ Workflow 运行失败，请检查上面的错误信息"
        echo "═══════════════════════════════════════════════════════════"
        exit 1
    fi
fi

