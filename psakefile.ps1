# https://psake.readthedocs.io/en/latest/structure-of-a-psake-build-script/
#
# overide with properties flag:
#  Invoke-psake ? -properties @{"Verbose"=$true}

Properties {
    $golangcilintPath = 'golangci-lint.exe'
    $GoPath = 'go.exe'
    # $GoOS = 'windows'
    $FuzzTime = '20s'

    $Verbose = $true
}

# todo: pre-task to validate GoOS
# todo: add go build and test flags, tags (functional, admin, etc.), and output directories
# todo: allow building and (fuzz) testing individual package

Task default -depends Lint

Task ? -Description 'show documentation' {
    Write-Output $test
    WriteDocumentation }
Task ?? -Description 'show detailed documentation' { WriteDocumentation($true) }

TaskSetup {
    if ( $Verbose ) {
        $script:VerbosePreference = 'Continue'
    }
    # script scope stays valid during entire PS session, so state is preserved between instances
    if ( -not $Script:go ) {
        $Script:go = Confirm-Path $GoPath 'go.exe'
    }
    if ( -not $Script:linter ) {
        $Script:linter = Confirm-Path $golangcilintPath 'golangci-lint.exe'
    }
}

Task Validate `
    -alias PR `
    -depends Lint, GoGen, Test, TestFuzz `
    -Description 'Run checks needed before creating a PR' `
{
    MyExec { & $Script:go mod tidy }
}

Task Lint -Description 'Lint the repo' {
    MyExec {
        & $Script:linter run `
            --verbose --timeout=5m `
            --config=.golangci.yml `
            --max-issues-per-linter=0 --max-same-issues=0 `
            --modules-download-mode=readonly
    }
}

Task GoGen -Description "Run 'go generate' on the repo" {
    MyExec { & $Script:go generate -x ./... }
}

Task Test -Description 'Run all go tests in the repo' {
    MyExec { & $Script:go test -gcflags='all=-d=checkptr' -v ./... }
}

# can only  call `go test -fuzz ..` per package, not on entire repo
Task Fuzz -Description 'Run all go fuzzing tests in the repo' {
    if ( (Get-GoVersion) -lt '1.18' ) {
        Write-Warning 'Fuzzing not supported for go1.17 or less'
        return
    }

    Get-GoTestDirs -Package '.' |
        ForEach-Object {
            MyExec { & $Script:go test -gcflags='all=-d=checkptr' -v -run='^#' -fuzz='.' -fuzztime="$FuzzTime" $_ }
        }
}

Task BuildTools -Alias Tools -Description 'Run all go fuzzing tests in the repo' {
    $out = '.\bin'
    New-Item -ItemType Directory -Force $out > $null
    @(
        './pkg/etw/sample/'
        './tools/etw-provider-gen/'
        './tools/mkwinsyscall/'
        './wim/validate/'
    ) | ForEach-Object {
        $p = Resolve-Path $_
        MyExec { & $Script:go build -gcflags='all=-d=checkptr' -o $out $p }
    }
}

Task Clean -Description 'Remove binaries' {
    if ( Get-Item '.\bin' -ErrorAction SilentlyContinue ) {
        Remove-Item -Recurse -Force '.\bin'
    }
}

function MyExec {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [scriptblock]$cmd
    )
    #Set-PSDebug -Trace 1
    #Set-PSDebug -Trace 0

    # creating a new scriptblock and invoking $cmd from in it causes scoping and recursion concerns ....
    $cmd = [scriptblock]::Create("
    Write-Verbose `"Executing block:`n$($cmd.ToString())`"
    $cmd
    ")
    Exec -cmd $cmd
}

function Get-GoTestDirs {
    [CmdletBinding()]
    [OutputType([string[]])]
    param (
        [Parameter(Position = 0)]
        [string]
        $Package = '.',

        [string]
        $Tags,

        [string]
        $go = $Script:go
    )
    $Package = Resolve-Path $Package
    $listcmd = @('list', "-tags=`'$tags`'", '-f' )

    $ModulePath = & $go @listcmd '{{ .Root }}' "$Package"
    & $go @listcmd  `
        '{{ if .TestGoFiles }}{{ .Dir }}{{ \"\n\" }}{{ end }}' `
        "$ModulePath/..."
}

function Get-GoVersion {
    [CmdletBinding()]
    [OutputType([version])]
    param (
        [string]
        $go = $Script:go
    )
    [version]((& $go env GOVERSION) -replace 'go', '')
}

function Confirm-Path {
    [OutputType([string])]
    [CmdletBinding()]
    param(
        [Parameter(Position = 0, Mandatory = $true)]
        [string]
        $Path,

        [Parameter(Position = 1, Mandatory = $false)]
        [string]
        $Name
    )

    $p = (Get-Command $Path -ErrorAction SilentlyContinue).Source
    $s = 'Invalid path'
    if ( $Name ) {
        $s += " to $Name"
    }

    Assert ([bool]$p) "${s}: $Path"

    $o = "Using `"$p`""
    if ( $Name) {
        $o += " for $Name"
    }
    # WriteColoredOutput $o -foregroundcolor "green"
    Write-Verbose $o
    $p
}
