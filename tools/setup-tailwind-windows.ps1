$ErrorActionPreference = 'Stop'

$RootDir = Resolve-Path (Join-Path $PSScriptRoot '..')
$ToolsDir = Join-Path $RootDir 'tools'
$BinPath = Join-Path $ToolsDir 'tailwindcss.exe'
$Version = 'v3.4.17'
$Url = "https://github.com/tailwindlabs/tailwindcss/releases/download/$Version/tailwindcss-windows-x64.exe"

New-Item -ItemType Directory -Force -Path $ToolsDir | Out-Null
Invoke-WebRequest -Uri $Url -OutFile $BinPath

Write-Host "Installed Tailwind CLI to $BinPath"
