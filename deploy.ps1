# ============================================================
#  Deploy Script: git push -> SSH pull -> docker compose restart
#  Usage: .\deploy.ps1 -Message "chore: update"
# ============================================================
param(
    [string]$Message = "deploy: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
)

# ---------- Config ----------
$REMOTE_HOST   = "8.141.97.141"
$REMOTE_USER   = "root"
$REMOTE_PORT   = 22
$REMOTE_DIR    = "/opt/go-mcp-gateway"
$BRANCH        = "master"
# ---------------------------

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "   Go MCP Gateway Deploy" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# ---- Step 1: git commit & push ----
Write-Host "[INFO] Step 1/4: Commit & push..." -ForegroundColor Green

$hasChanges = git status --porcelain
if (-not $hasChanges) {
    Write-Host "[WARN] No changes, skip commit." -ForegroundColor Yellow
} else {
    git add -A
    git commit -m $Message
    Write-Host "[INFO] Committed: $Message" -ForegroundColor Green
}

Write-Host "[INFO] Pushing branch [$BRANCH] to remote..." -ForegroundColor Green
git push origin $BRANCH
Write-Host "[INFO] Push done." -ForegroundColor Green

# ---- Step 2-4: SSH remote deploy ----
Write-Host "[INFO] Step 2/4: SSH ${REMOTE_USER}@${REMOTE_HOST} ..." -ForegroundColor Green
Write-Host "[INFO] Step 3/4: git pull on remote ..." -ForegroundColor Green
Write-Host "[INFO] Step 4/4: docker compose restart ..." -ForegroundColor Green

# Build bash script using .NET StringBuilder (ASCII-safe)
$sb = New-Object System.Text.StringBuilder
[void]$sb.AppendLine("set -e")
[void]$sb.AppendLine("cd $REMOTE_DIR || { echo 'dir $REMOTE_DIR not found'; exit 1; }")
[void]$sb.AppendLine("echo '[REMOTE] pwd: $(pwd)'")
[void]$sb.AppendLine("echo '[REMOTE] branch: $BRANCH'")
[void]$sb.AppendLine("git fetch origin")
[void]$sb.AppendLine("git checkout $BRANCH")
[void]$sb.AppendLine("git reset --hard origin/$BRANCH")
[void]$sb.AppendLine("echo '[REMOTE] latest commit: $(git log -1 --oneline)'")
[void]$sb.AppendLine("echo '[REMOTE] stopping services ...'")
[void]$sb.AppendLine("docker compose down 2>/dev/null || docker-compose down 2>/dev/null || true")
[void]$sb.AppendLine("echo '[REMOTE] building and starting ...'")
[void]$sb.AppendLine("docker compose up -d --build 2>/dev/null || docker-compose up -d --build 2>/dev/null")
[void]$sb.AppendLine("echo '[REMOTE] waiting 3s ...'")
[void]$sb.AppendLine("sleep 3")
[void]$sb.AppendLine("echo ''")
[void]$sb.AppendLine("echo '[REMOTE] container status:'")
[void]$sb.AppendLine("docker compose ps 2>/dev/null || docker-compose ps 2>/dev/null || true")
[void]$sb.AppendLine("echo ''")
[void]$sb.AppendLine("echo '[REMOTE] deploy done.'")

$tempFile = "$env:TEMP\go-mcp-deploy.sh"
[System.IO.File]::WriteAllText($tempFile, $sb.ToString())

Get-Content -Raw $tempFile | ssh -p $REMOTE_PORT "$REMOTE_USER@$REMOTE_HOST" "bash -s"

Remove-Item $tempFile -Force -ErrorAction SilentlyContinue

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "============================================" -ForegroundColor Cyan
    Write-Host "   Deploy SUCCESS!" -ForegroundColor Green
    Write-Host "============================================" -ForegroundColor Cyan
} else {
    Write-Host ""
    Write-Host "============================================" -ForegroundColor Red
    Write-Host "   Deploy FAILED! Check remote logs." -ForegroundColor Red
    Write-Host "============================================" -ForegroundColor Red
}
