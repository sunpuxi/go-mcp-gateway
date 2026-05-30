#!/bin/bash
# ============================================================
#  一键部署脚本：提交代码 → SSH 远程拉取 → 重启 Docker 服务
#  用法: bash deploy.sh [commit message]
#  示例: bash deploy.sh "feat: 新增xxx功能"
# ============================================================
set -e

# ---------- 配置区（请根据实际情况修改）----------
REMOTE_HOST="8.141.97.141"          # 远程服务器 IP 或域名
REMOTE_USER="root"                    # 远程服务器用户名
REMOTE_PORT="22"                      # SSH 端口
REMOTE_PROJECT_DIR="/opt/go-mcp-gateway" # 远程项目目录
# -------------------------------------------------

# 获取提交信息，默认使用当前时间
COMMIT_MSG="${1:-deploy: $(date '+%Y-%m-%d %H:%M:%S')}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[INFO]${NC}  $1"; }
log_warn()  { echo -e "${YELLOW}[WARN]${NC}  $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

echo "============================================"
echo "   Go MCP Gateway 一键部署"
echo "============================================"
echo ""

# ---- Step 1: 提交代码到远程仓库 ----
log_info "Step 1/4: 提交代码到远程仓库..."
if [ -z "$(git status --porcelain)" ]; then
    log_warn "没有需要提交的变更，跳过 commit"
else
    git add -A
    git commit -m "$COMMIT_MSG"
    log_info "已提交: $COMMIT_MSG"
fi

# 获取当前分支名
BRANCH=$(git rev-parse --abbrev-ref HEAD)
log_info "推送分支 [$BRANCH] 到远程仓库..."
git push origin "$BRANCH"
log_info "代码推送完成 ✓"

# ---- Step 2: SSH 远程操作 ----
log_info "Step 2/4: 连接远程服务器 ${REMOTE_USER}@${REMOTE_HOST}..."
log_info "Step 3/4: 拉取最新代码..."
log_info "Step 4/4: 重启 Docker 服务..."

ssh -p "$REMOTE_PORT" "${REMOTE_USER}@${REMOTE_HOST}" << 'ENDSSH'
set -e

# ---------- 远程服务器配置 ----------
PROJECT_DIR="/opt/go-mcp-gateway"
BRANCH="master"
# -----------------------------------

cd "$PROJECT_DIR" || { echo "目录 $PROJECT_DIR 不存在！"; exit 1; }

echo "[远程] 当前目录: $(pwd)"
echo "[远程] 拉取分支: $BRANCH"

# 拉取最新代码
git fetch origin
git checkout "$BRANCH"
git reset --hard "origin/$BRANCH"

echo "[远程] 代码拉取完成"
echo "[远程] 最新提交: $(git log -1 --oneline)"

echo "[远程] 停止正在运行的服务..."
docker compose down 2>/dev/null || docker-compose down 2>/dev/null

echo "[远程] 重新构建并启动服务..."
docker compose up -d --build 2>/dev/null || docker-compose up -d --build 2>/dev/null

echo "[远程] 等待服务启动..."
sleep 3

echo ""
echo "[远程] 容器状态:"
docker compose ps 2>/dev/null || docker-compose ps 2>/dev/null

echo ""
echo "[远程] 服务部署完成 ✓"
ENDSSH

echo ""
log_info "============================================"
log_info "  部署全部完成！"
log_info "============================================"
