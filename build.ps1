
param(
    [Parameter(Position = 0)]
    [string]$Task = "all",
    [Parameter(Position = 1)]
    [string]$Arg
)

$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

$BinDir = "bin"
$Mihan = Join-Path $BinDir "mihan.exe"
$Server = Join-Path $BinDir "mihan-ls.exe"

function Say($msg) { Write-Host "==> $msg" -ForegroundColor Cyan }
function Bad($msg) { Write-Host "    $msg" -ForegroundColor Red }
function Good($msg) { Write-Host "    $msg" -ForegroundColor Green }

function Build {
    Say "Building the Mihan compiler"
    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
    go build -o $Mihan ./cmd/mihan
    if ($LASTEXITCODE -ne 0) { throw "the build failed" }
    Good "$Mihan"
    go build -o $Server ./cmd/mihan-ls
    if ($LASTEXITCODE -ne 0) { throw "the language server did not build" }
    Good "$Server"
}

function BuildExtension {
    Say "Building the editor extension"
    $ext = Join-Path $PSScriptRoot "extension"
    if (-not (Test-Path (Join-Path $ext "node_modules"))) {
        Push-Location $ext
        try { npm install --silent } finally { Pop-Location }
    }
    Push-Location $ext
    try {
        npx tsc -p ./
        if ($LASTEXITCODE -ne 0) { throw "the extension did not compile" }
    } finally {
        Pop-Location
    }
    Good "extension\out\extension.js"
}

function TestGo {
    Say "Running the Go tests"
    go test ./...
    if ($LASTEXITCODE -ne 0) { throw "a Go test failed" }
}

function TestDocs {
    Say "Running the documentation's own examples"
    & $Mihan test
    if ($LASTEXITCODE -ne 0) { throw "a documented example does not do what it says" }
    $py = Get-Command python -ErrorAction SilentlyContinue
    if ($py) {
        Say "Running every sample of the course (docs/tutorial/*.md)"
        & $py.Source tools/tutorial_check.py --mihan $Mihan
        if ($LASTEXITCODE -ne 0) { throw "a lesson's sample does not print what the lesson says" }
    } else {
        Bad "python is not installed, so the course's samples were not checked"
    }
}

function Check {
    Say "Vetting"
    go vet ./...
    if ($LASTEXITCODE -ne 0) { throw "go vet had something to say" }
    $unformatted = gofmt -l . 2>$null
    if ($unformatted) {
        foreach ($f in $unformatted) { Bad "gofmt wants to tidy $f" }
        throw "run .\build.ps1 fmt"
    }
    Good "vet is quiet and every file is tidy"
}

function Fmt {
    Say "Tidying the Go source"
    gofmt -w .
    Say "Tidying the Mihan programs"
    $programs = Get-ChildItem -Recurse -Include *.mhn -Path docs,tests,std -ErrorAction SilentlyContinue
    foreach ($p in $programs) {
        & $Mihan fmt $p.FullName
    }
    Good "done"
}

function Bench {
    Say "Running the benchmarks"
    go test -bench=. -benchmem ./benchmarks/... 2>&1 | Tee-Object -FilePath (Join-Path $BinDir "bench.txt")
    if (Test-Path (Join-Path $BinDir "bench.txt")) {
        Good "written to $BinDir\bench.txt"
    }
}

function Clean {
    Say "Cleaning up"
    Remove-Item -Recurse -Force $BinDir -ErrorAction SilentlyContinue
    Good "done"
}

switch ($Task) {
    "build" { Build }
    "extension" { BuildExtension }
    "install" { Build; & $Mihan install }
    "test"  { Build; TestGo; TestDocs }
    "go"    { TestGo }
    "docs"  { Build; TestDocs }
    "fmt"   { Build; Fmt }
    "check" { Check }
    "bench" { Bench }
    "clean" { Clean }
    "run" {
        Build
        if (-not $Arg) { throw "give me a program: .\build.ps1 run my.mhn" }
        & $Mihan run $Arg
    }
    "repl" { Build; & $Mihan repl }
    default {
        Say "Building, testing and checking Mihan"
        Build
        TestGo
        TestDocs
        Check
        Good "everything is in order"
    }
}
