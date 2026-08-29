# ==============================================================================
# BOB Gemini Free - Authenticated Windows Setup Script
# Break Ordinary Boundaries | Powered by ABCsteps (https://abcsteps.com)
# Author: Divyanshu Singh Chouhan (@div197)
# ==============================================================================

$ErrorActionPreference = "Stop"
$UpdatePublicKeyHex = "ba7854781bca2a14da4f1ec5e931ff45f458ac9377c42ac127c349e5ecad2dff"
$ReleaseDownloadBase = "https://github.com/div197/BOB-Gemini-Free/releases/latest/download"
$MaxManifestBytes = 1MB
$MaxSignatureBytes = 4KB
$MaxReleaseAssetBytes = 512MB
$TempRoot = $null

function Fail-Install {
    param([Parameter(Mandatory = $true)][string]$Message)
    throw "BOB Gemini Free installation stopped: $Message"
}

function Convert-HexToBytes {
    param([Parameter(Mandatory = $true)][string]$Hex)
    $normalized = $Hex.Trim()
    if (($normalized.Length % 2) -ne 0 -or $normalized -notmatch '^[0-9a-fA-F]+$') {
        Fail-Install "the embedded update public key is malformed"
    }
    $bytes = New-Object byte[] ($normalized.Length / 2)
    for ($index = 0; $index -lt $bytes.Length; $index++) {
        $bytes[$index] = [Convert]::ToByte($normalized.Substring($index * 2, 2), 16)
    }
    return $bytes
}

function Get-FileSizeBytes {
    param([Parameter(Mandatory = $true)][string]$Path)
    return ([System.IO.FileInfo]$Path).Length
}

function Download-ReleaseFile {
    param(
        [Parameter(Mandatory = $true)][string]$Url,
        [Parameter(Mandatory = $true)][string]$Destination,
        [Int64]$MaximumBytes = $MaxReleaseAssetBytes
    )

    if (Get-Command curl.exe -ErrorAction SilentlyContinue) {
        & curl.exe --fail --silent --show-error --location `
            --proto "=https" --proto-redir "=https" `
            --max-time 120 --max-filesize $MaximumBytes --retry 2 --retry-delay 1 `
            --output $Destination $Url
        if ($LASTEXITCODE -ne 0) {
            Fail-Install "could not download the signed release file"
        }
        return
    }

    Fail-Install "curl.exe is required for the authenticated bootstrap path; install Windows curl or build from source"
}

function Invoke-OpenSslEd25519Verification {
    param(
        [Parameter(Mandatory = $true)][string]$ManifestPath,
        [Parameter(Mandatory = $true)][string]$SignaturePath,
        [Parameter(Mandatory = $true)][string]$PublicKeyPath
    )

    $openssl = Get-Command openssl.exe -ErrorAction SilentlyContinue
    if (-not $openssl) {
        return $false
    }

    $arguments = @(
        "pkeyutl", "-verify", "-rawin", "-pubin", "-inkey", $PublicKeyPath,
        "-in", $ManifestPath, "-sigfile", $SignaturePath
    )
    $quotedArguments = ($arguments | ForEach-Object {
        '"' + ($_ -replace '"', '\"') + '"'
    }) -join ' '
    $process = Start-Process -FilePath $openssl.Source -ArgumentList $quotedArguments -Wait -PassThru -WindowStyle Hidden
    return $process.ExitCode -eq 0
}

function Verify-Ed25519Manifest {
    param(
        [Parameter(Mandatory = $true)][string]$ManifestPath,
        [Parameter(Mandatory = $true)][string]$SignaturePath
    )

    $publicKey = Convert-HexToBytes $UpdatePublicKeyHex
    if ($publicKey.Length -ne 32) {
        Fail-Install "the embedded update public key has an invalid length"
    }
    $spkiPrefix = [byte[]](0x30, 0x2a, 0x30, 0x05, 0x06, 0x03, 0x2b, 0x65, 0x70, 0x03, 0x21, 0x00)
    $spki = New-Object byte[] ($spkiPrefix.Length + $publicKey.Length)
    [Array]::Copy($spkiPrefix, 0, $spki, 0, $spkiPrefix.Length)
    [Array]::Copy($publicKey, 0, $spki, $spkiPrefix.Length, $publicKey.Length)
    $publicKeyPath = Join-Path $TempRoot "update-public-key.pem"
    $decodedSignaturePath = Join-Path $TempRoot "SHA256SUMS.sig.decoded"
    $publicKeyPem = "-----BEGIN PUBLIC KEY-----`n$([Convert]::ToBase64String($spki))`n-----END PUBLIC KEY-----`n"
    [System.IO.File]::WriteAllText($publicKeyPath, $publicKeyPem, [Text.Encoding]::ASCII)

    try {
        $signatureText = ([System.IO.File]::ReadAllText($SignaturePath)).Trim()
        $signature = [Convert]::FromBase64String($signatureText)
    } catch {
        Fail-Install "the release signature is not valid base64"
    }
    if ($signature.Length -ne 64) {
        Fail-Install "the release signature has an invalid Ed25519 length"
    }
    [System.IO.File]::WriteAllBytes($decodedSignaturePath, $signature)

    if (Invoke-OpenSslEd25519Verification $ManifestPath $decodedSignaturePath $publicKeyPath) {
        return
    }

    # Newer .NET runtimes expose Ed25519. Use it only when the host provides
    # the API; older Windows PowerShell versions fail closed below.
    try {
        Add-Type -TypeDefinition @"
using System.Security.Cryptography;
public static class BobGeminiEd25519Verifier {
    public static bool Verify(byte[] signature, byte[] message, byte[] publicKey) {
        return Ed25519.Verify(signature, message, publicKey);
    }
}
"@ -ErrorAction Stop
        $message = [System.IO.File]::ReadAllBytes($ManifestPath)
        if ([BobGeminiEd25519Verifier]::Verify($signature, $message, $publicKey)) {
            return
        }
    } catch {
        # Fall through to one explicit, actionable fail-closed error.
    }

    Fail-Install "no available Ed25519 verifier accepted the release; install OpenSSL/Go or use a publisher-signed desktop package"
}

function Get-Sha256Hex {
    param([Parameter(Mandatory = $true)][string]$Path)
    return (Get-FileHash -Algorithm SHA256 -LiteralPath $Path).Hash.ToLowerInvariant()
}

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "    BOB Gemini Free - Break Ordinary Boundaries                " -ForegroundColor Green
Write-Host "    Powered by ABCsteps.com | Divyanshu Singh Chouhan (@div197) " -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan
Write-Host ""

$AppName = "bob-gemini-free.exe"
$ConfigDir = Join-Path $HOME ".config\bob-gemini-free"
$TargetBin = Join-Path (Get-Location) $AppName
$sourceMode = if ($env:BOB_GEMINI_FREE_INSTALL_FROM_SOURCE) { $env:BOB_GEMINI_FREE_INSTALL_FROM_SOURCE } else { "0" }
if ($sourceMode -notin @("0", "1")) {
    Fail-Install "BOB_GEMINI_FREE_INSTALL_FROM_SOURCE must be 0 or 1"
}
$sourceBuild = $sourceMode -eq "1"

try {
    if ($sourceBuild) {
        if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
            Fail-Install "Go is required for the explicit source-build path"
        }
        if (-not (Test-Path "go.mod" -PathType Leaf)) {
            Fail-Install "explicit source-build path requires a repository checkout"
        }
        $moduleLine = Select-String -Path "go.mod" -Pattern '^module github\.com/div197/bob-gemini-free$' -Quiet -ErrorAction SilentlyContinue
        if (-not $moduleLine) {
            Fail-Install "current directory is not the BOB Gemini Free source module"
        }
        Write-Host "[*] Explicit source-build mode enabled; compiling the checked-out module..." -ForegroundColor Cyan
        $env:CGO_ENABLED = "0"
        $ldflags = "-s -w -X github.com/div197/bob-gemini-free/internal/updater.BuildUpdatePublicKey=$UpdatePublicKeyHex"
        & go build -ldflags $ldflags -o $AppName .
        if ($LASTEXITCODE -ne 0) {
            Fail-Install "the local source build failed"
        }
        Write-Host "[✔] Successfully built $AppName from local source." -ForegroundColor Green
    } else {
        Write-Host "[*] Fetching and authenticating the latest signed Windows release..." -ForegroundColor Cyan
        $TempRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("bob-gemini-free-install-" + [Guid]::NewGuid().ToString("N"))
        New-Item -ItemType Directory -Force -Path $TempRoot | Out-Null
        $manifestPath = Join-Path $TempRoot "SHA256SUMS"
        $signaturePath = Join-Path $TempRoot "SHA256SUMS.sig"
        $architecture = [Environment]::GetEnvironmentVariable("PROCESSOR_ARCHITECTURE")
        switch ($architecture.ToUpperInvariant()) {
            "AMD64" { $architectureName = "amd64" }
            "ARM64" { $architectureName = "arm64" }
            default { Fail-Install "unsupported Windows architecture $architecture" }
        }
        $assetName = "bob-gemini-free-windows-$architectureName.exe"
        $assetPath = Join-Path $TempRoot $assetName

        Download-ReleaseFile "$ReleaseDownloadBase/SHA256SUMS" $manifestPath $MaxManifestBytes
        Download-ReleaseFile "$ReleaseDownloadBase/SHA256SUMS.sig" $signaturePath $MaxSignatureBytes
        if ((Get-FileSizeBytes $manifestPath) -gt $MaxManifestBytes) {
            Fail-Install "the signed release manifest exceeds the safe size limit"
        }
        if ((Get-FileSizeBytes $signaturePath) -gt $MaxSignatureBytes) {
            Fail-Install "the signed release signature exceeds the safe size limit"
        }
        Verify-Ed25519Manifest $manifestPath $signaturePath

        $manifestText = [Text.Encoding]::UTF8.GetString([System.IO.File]::ReadAllBytes($manifestPath))
        $entries = @{}
        $expectedDigest = $null
        foreach ($line in ($manifestText -split "`n")) {
            $cleanLine = $line.TrimEnd("`r")
            if ([string]::IsNullOrWhiteSpace($cleanLine)) { continue }
            if ($cleanLine -notmatch '^(?<digest>[0-9a-fA-F]{64})  (?<name>[^\s]+)$') {
                Fail-Install "the signed release manifest contains an invalid entry"
            }
            $name = $Matches.name
            if ($entries.ContainsKey($name)) {
                Fail-Install "the signed release manifest contains a duplicate entry"
            }
            $entries[$name] = $Matches.digest.ToLowerInvariant()
            if ($name -eq $assetName) {
                $expectedDigest = $entries[$name]
            }
        }
        if (-not $expectedDigest) {
            Fail-Install "the signed release manifest has no Windows $architectureName asset"
        }

        Download-ReleaseFile "$ReleaseDownloadBase/$assetName" $assetPath $MaxReleaseAssetBytes
        if ((Get-FileSizeBytes $assetPath) -gt $MaxReleaseAssetBytes) {
            Fail-Install "the Windows release asset exceeds the safe size limit"
        }
        $actualDigest = Get-Sha256Hex $assetPath
        if ($actualDigest -ne $expectedDigest) {
            Fail-Install "the downloaded Windows asset failed its signed SHA-256 check"
        }

        $InstallDir = Join-Path $HOME "bob-gemini-free"
        New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
        $TargetBin = Join-Path $InstallDir $AppName
        if (Test-Path -LiteralPath $TargetBin -PathType Container) {
            Fail-Install "the install target is a directory: $TargetBin"
        }
        $stagedTarget = Join-Path $InstallDir ("." + $AppName + "." + [Guid]::NewGuid().ToString("N") + ".tmp")
        try {
            Copy-Item -LiteralPath $assetPath -Destination $stagedTarget -Force
            Move-Item -LiteralPath $stagedTarget -Destination $TargetBin -Force
        } finally {
            if (Test-Path -LiteralPath $stagedTarget) {
                Remove-Item -LiteralPath $stagedTarget -Force -ErrorAction SilentlyContinue
            }
        }
        Write-Host "[✔] Installed authenticated release to $TargetBin" -ForegroundColor Green
    }

    New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null
    $configPath = Join-Path $ConfigDir "config.json"
    if (-not (Test-Path $configPath) -and (Test-Path "config.example.json")) {
        Copy-Item "config.example.json" $configPath
        Write-Host "[✔] Created default configuration at $configPath" -ForegroundColor Green
    }
} finally {
    if ($TempRoot -and (Test-Path $TempRoot)) {
        Remove-Item -LiteralPath $TempRoot -Recurse -Force -ErrorAction SilentlyContinue
    }
}

Write-Host ""
Write-Host "================================================================" -ForegroundColor Green
Write-Host "    INSTALLATION COMPLETE!" -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green
Write-Host ""
Write-Host "To launch the gateway and open the Web Studio:" -ForegroundColor White
Write-Host "  `"$TargetBin`" --port 9610" -ForegroundColor Cyan
Write-Host ""
Write-Host "API Base URL: http://127.0.0.1:9610/v1" -ForegroundColor DarkGray
Write-Host "UI Dashboard: http://127.0.0.1:9610/playground" -ForegroundColor DarkGray
Write-Host ""

exit 0
