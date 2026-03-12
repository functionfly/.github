# Upload STAR_functionfly_com certs from Windows to edge VPS nodes.
# Requires: OpenSSL (Git for Windows or standalone) and OpenSSH (built into Windows 10/11).
#
# Usage (PowerShell):
#   .\upload-certs-windows.ps1 -CertFolder "C:\Users\YourName\Downloads\STAR_functionfly_com"
#   .\upload-certs-windows.ps1 -CertZip "C:\Users\YourName\Downloads\functionfly-certs.zip"
#   (With a zip: extract and use. Put privkey.pem in the zip or in the same folder as the zip.)
#   .\upload-certs-windows.ps1 -CertFolder "..." -SshUser root
#   If private keys are already on each VPS (no key in zip/folder):
#   .\upload-certs-windows.ps1 -CertZip "C:\...\certs.zip" -KeysOnServer

param(
    [string]$CertFolder,
    [string]$CertZip,
    [string]$SshUser = "root",
    [switch]$KeysOnServer  # Only upload fullchain.pem; privkey already on each VPS
)

$ErrorActionPreference = "Stop"
$NODES = @("217.160.124.206", "209.46.125.113")
$REMOTE_DIR = "/etc/ssl/functionfly"

if (-not $CertFolder -and -not $CertZip) {
    Write-Error "Use -CertFolder path\to\extracted\certs or -CertZip path\to\certs.zip"
}
if ($CertZip) {
    if (-not (Test-Path $CertZip)) { Write-Error "Zip not found: $CertZip" }
    $extractTo = Join-Path $PSScriptRoot "certs-in"
    $null = New-Item -ItemType Directory -Force -Path $extractTo
    Write-Host "Extracting $CertZip to $extractTo ..."
    Expand-Archive -Path $CertZip -DestinationPath $extractTo -Force
    # If zip contained a single folder, use that (e.g. STAR_functionfly_com)
    $subdirs = Get-ChildItem -Path $extractTo -Directory
    if ($subdirs.Count -eq 1) {
        $CertFolder = $subdirs[0].FullName
    } else {
        $CertFolder = $extractTo
    }
    # Private key: allow in extracted folder or next to the zip
    $zipDir = [System.IO.Path]::GetDirectoryName($CertZip)
    if (-not (Get-ChildItem -Path $CertFolder -File | Where-Object { $_.Name -eq 'privkey.pem' -or $_.Extension -eq '.key' })) {
        $keyNextToZip = Get-ChildItem -Path $zipDir -File -ErrorAction SilentlyContinue | Where-Object { $_.Name -eq 'privkey.pem' -or $_.Extension -eq '.key' } | Select-Object -First 1
        if ($keyNextToZip) {
            Copy-Item $keyNextToZip.FullName (Join-Path $CertFolder $keyNextToZip.Name)
            Write-Host "Using private key from zip folder: $($keyNextToZip.Name)"
        }
    }
}
if (-not (Test-Path $CertFolder)) {
    Write-Error "Cert folder not found: $CertFolder"
}

# Find OpenSSL (Git for Windows or in PATH)
$openssl = $null
foreach ($p in @(
    "${env:ProgramFiles}\Git\usr\bin\openssl.exe",
    "${env:ProgramFiles(x86)}\Git\usr\bin\openssl.exe",
    "openssl"
)) {
    if ($p -eq "openssl") {
        $o = Get-Command openssl -ErrorAction SilentlyContinue
        if ($o) { $openssl = $o.Source; break }
    } elseif (Test-Path $p) {
        $openssl = $p
        break
    }
}
if (-not $openssl) {
    Write-Error "OpenSSL not found. Install Git for Windows (https://git-scm.com) or OpenSSL and ensure openssl is in PATH."
}

# Find scp/ssh (Windows OpenSSH)
$scp = Get-Command scp -ErrorAction SilentlyContinue
$ssh = Get-Command ssh -ErrorAction SilentlyContinue
if (-not $scp -or -not $ssh) {
    Write-Error "scp/ssh not found. Enable OpenSSH Client: Settings > Apps > Optional features > Add OpenSSH Client."
}

$workDir = Join-Path $PSScriptRoot "certs-out"
$null = New-Item -ItemType Directory -Force -Path $workDir

# Build fullchain.pem: leaf first, then intermediates
$certFiles = Get-ChildItem -Path $CertFolder -File | Where-Object {
    $_.Extension -match '\.(cer|crt|pem)$' -and $_.Name -notmatch 'privkey|\.key$'
}
if ($certFiles.Count -eq 0) {
    Write-Error "No certificate files (.cer, .crt, .pem) found in $CertFolder"
}

# Order: STAR/functionfly leaf first, then others
$leaf = $certFiles | Where-Object { $_.Name -match 'STAR|functionfly' } | Select-Object -First 1
if (-not $leaf) { $leaf = $certFiles | Select-Object -First 1 }
$others = $certFiles | Where-Object { $_.FullName -ne $leaf.FullName }

$fullchainPath = Join-Path $workDir "fullchain.pem"
$fullchainPath = $fullchainPath -replace '\\', '/'
$chainParts = @()

function Convert-ToPem {
    param([string]$InPath)
    $ext = [System.IO.Path]::GetExtension($InPath).ToLower()
    $tmp = Join-Path $workDir "tmp_$(Get-Random).pem"
    if ($ext -eq '.pem') {
        Get-Content $InPath -Raw | Set-Content $tmp -NoNewline
    } else {
        & $openssl x509 -in $InPath -out $tmp -inform DER 2>$null
        if ($LASTEXITCODE -ne 0) {
            & $openssl x509 -in $InPath -out $tmp -inform PEM 2>$null
        }
    }
    if (Test-Path $tmp) {
        Get-Content $tmp -Raw
        Remove-Item $tmp -Force
    }
}

$chainParts += Convert-ToPem $leaf.FullName
foreach ($f in $others) {
    $chainParts += Convert-ToPem $f.FullName
}
# Write single fullchain (one newline between cert blocks)
$fullchainContent = ($chainParts -join "`n").Trim() + "`n"
[System.IO.File]::WriteAllText($fullchainPath, $fullchainContent, [System.Text.Encoding]::ASCII)

# Private key (skip if keys already on each VPS)
$uploadKey = -not $KeysOnServer
if ($uploadKey) {
    $keyFile = Get-ChildItem -Path $CertFolder -File | Where-Object {
        $_.Name -eq 'privkey.pem' -or $_.Extension -eq '.key'
    } | Select-Object -First 1
    if (-not $keyFile) {
        Write-Error "No privkey.pem or .key in $CertFolder. Add the key, or use -KeysOnServer if keys are already on each VPS."
    }
    $privkeyPath = Join-Path $workDir "privkey.pem"
    Copy-Item $keyFile.FullName $privkeyPath -Force
    Write-Host "Prepared fullchain.pem and privkey.pem in $workDir"
} else {
    Write-Host "Prepared fullchain.pem (keys already on VPS; not uploading privkey)"
}
Write-Host "Uploading to edge VPS (user=$SshUser, dir=$REMOTE_DIR)..."

foreach ($node in $NODES) {
    Write-Host "  -> $node"
    & ssh "${SshUser}@${node}" "mkdir -p $REMOTE_DIR"
    & scp $fullchainPath "${SshUser}@${node}:${REMOTE_DIR}/fullchain.pem"
    if ($uploadKey) {
        & scp $privkeyPath "${SshUser}@${node}:${REMOTE_DIR}/privkey.pem"
        & ssh "${SshUser}@${node}" "chmod 644 ${REMOTE_DIR}/fullchain.pem && chmod 600 ${REMOTE_DIR}/privkey.pem"
    } else {
        & ssh "${SshUser}@${node}" "chmod 644 ${REMOTE_DIR}/fullchain.pem"
    }
}

Write-Host "Done. Restart Caddy on each node to pick up certs: ssh root@<ip> 'sudo systemctl restart caddy'"
