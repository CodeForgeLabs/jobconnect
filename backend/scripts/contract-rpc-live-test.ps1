[CmdletBinding()]
param(
    [string]$CasesFile = "",
    [string]$SummaryOut = "",
    [switch]$SkipComposeBuild,
    [switch]$SkipMigrations
)

Set-StrictMode -Version 1
$ErrorActionPreference = "Stop"

$script:Results = New-Object System.Collections.Generic.List[object]
$script:Context = @{}

if ([string]::IsNullOrWhiteSpace($CasesFile)) {
    $CasesFile = Join-Path $PSScriptRoot "contract-rpc-cases.json"
}
if ([string]::IsNullOrWhiteSpace($SummaryOut)) {
    $SummaryOut = Join-Path $PSScriptRoot "contract-rpc-live-test-summary.json"
}

function Write-Step {
    param([string]$Message)
    Write-Host "`n=== $Message ===" -ForegroundColor Cyan
}

function Write-Info {
    param([string]$Message)
    Write-Host "[info] $Message" -ForegroundColor Gray
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[warn] $Message" -ForegroundColor Yellow
}

function Write-Ok {
    param([string]$Message)
    Write-Host "[ok] $Message" -ForegroundColor Green
}

function Require-Command {
    param([string]$Name)
    if (-not (Get-Command $Name -ErrorAction SilentlyContinue)) {
        throw "Required command '$Name' is not available in PATH."
    }
}

function Parse-EnvFile {
    param([string]$Path)
    $envMap = @{}
    if (-not (Test-Path $Path)) {
        return $envMap
    }
    Get-Content $Path | ForEach-Object {
        $line = $_.Trim()
        if (-not $line -or $line.StartsWith("#")) {
            return
        }
        $idx = $line.IndexOf("=")
        if ($idx -lt 1) {
            return
        }
        $key = $line.Substring(0, $idx).Trim()
        $val = $line.Substring($idx + 1).Trim().Trim("`"").Trim("'")
        if ($key) {
            $envMap[$key] = $val
        }
    }
    return $envMap
}

function Get-PostmanEnvValue {
    param(
        [string]$Path,
        [string]$Key
    )
    if (-not (Test-Path $Path)) {
        return $null
    }
    $json = Get-Content $Path -Raw | ConvertFrom-Json
    foreach ($item in $json.values) {
        if ($item.key -eq $Key) {
            return [string]$item.value
        }
    }
    return $null
}

function Wait-ForHttp {
    param(
        [string]$Url,
        [int]$Attempts = 60,
        [int]$DelaySeconds = 2
    )
    for ($i = 1; $i -le $Attempts; $i++) {
        try {
            $resp = Invoke-WebRequest -Uri $Url -Method Get -TimeoutSec 10 -UseBasicParsing
            if ($resp.StatusCode -ge 200 -and $resp.StatusCode -lt 300) {
                return
            }
        } catch {
            Start-Sleep -Seconds $DelaySeconds
            continue
        }
        Start-Sleep -Seconds $DelaySeconds
    }
    throw "Timed out waiting for HTTP endpoint: $Url"
}

function Wait-ForTcpPort {
    param(
        [int]$Port,
        [int]$Attempts = 60,
        [int]$DelaySeconds = 2
    )
    for ($i = 1; $i -le $Attempts; $i++) {
        try {
            $ok = Test-NetConnection -ComputerName "localhost" -Port $Port -WarningAction SilentlyContinue
            if ($ok -and $ok.TcpTestSucceeded) {
                return
            }
        } catch {
            Start-Sleep -Seconds $DelaySeconds
            continue
        }
        Start-Sleep -Seconds $DelaySeconds
    }
    throw "Timed out waiting for localhost:$Port"
}

function Read-ErrorResponseBody {
    param([System.Exception]$Exception)
    try {
        $resp = $Exception.Response
        if ($null -eq $resp) {
            return ""
        }
        $stream = $resp.GetResponseStream()
        if ($null -eq $stream) {
            return ""
        }
        $reader = New-Object System.IO.StreamReader($stream)
        $body = $reader.ReadToEnd()
        $reader.Dispose()
        return $body
    } catch {
        return ""
    }
}

function Invoke-HttpJson {
    param(
        [ValidateSet("GET", "POST", "PATCH", "PUT", "DELETE")]
        [string]$Method,
        [string]$Url,
        [object]$Body,
        [hashtable]$Headers
    )

    $params = @{
        Method      = $Method
        Uri         = $Url
        TimeoutSec  = 30
        ErrorAction = "Stop"
    }
    if ($Headers) {
        $params["Headers"] = $Headers
    }
    if ($Body -ne $null) {
        $params["ContentType"] = "application/json"
        $params["Body"] = ($Body | ConvertTo-Json -Depth 40)
    }

    try {
        $resp = Invoke-RestMethod @params
        return @{
            ok     = $true
            status = 200
            body   = $resp
            raw    = ""
        }
    } catch {
        $code = 0
        try {
            if ($_.Exception.Response -and $_.Exception.Response.StatusCode) {
                $code = [int]$_.Exception.Response.StatusCode.value__
            }
        } catch {
            $code = 0
        }
        $raw = Read-ErrorResponseBody -Exception $_.Exception
        $parsed = $null
        if ($raw) {
            try {
                $parsed = $raw | ConvertFrom-Json
            } catch {
                $parsed = $null
            }
        }
        return @{
            ok     = $false
            status = $code
            body   = $parsed
            raw    = $raw
            error  = $_.Exception.Message
        }
    }
}

function Decode-JwtPayload {
    param([string]$Token)
    if (-not $Token) {
        throw "Cannot decode empty token"
    }
    $parts = $Token.Split('.')
    if ($parts.Length -lt 2) {
        throw "Invalid JWT token format"
    }
    $payload = $parts[1].Replace('-', '+').Replace('_', '/')
    switch ($payload.Length % 4) {
        2 { $payload += "==" }
        3 { $payload += "=" }
    }
    $json = [System.Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($payload))
    return $json | ConvertFrom-Json
}

function Get-PathValue {
    param(
        [object]$Object,
        [string]$Path
    )
    if ($null -eq $Object -or -not $Path) {
        return $null
    }
    $current = $Object
    foreach ($segment in $Path.Split('.')) {
        if ($null -eq $current) {
            return $null
        }
        if ($segment -match '^\d+$') {
            $idx = [int]$segment
            if ($current -is [System.Array]) {
                if ($idx -lt $current.Count) {
                    $current = $current[$idx]
                    continue
                }
                return $null
            }
            if ($current -is [System.Collections.IList]) {
                if ($idx -lt $current.Count) {
                    $current = $current[$idx]
                    continue
                }
                return $null
            }
            return $null
        }

        if ($current -is [System.Collections.IDictionary]) {
            if ($current.Contains($segment)) {
                $current = $current[$segment]
                continue
            }
            $alt = Get-AlternateSegmentName -Segment $segment
            if ($alt -and $current.Contains($alt)) {
                $current = $current[$alt]
                continue
            }
            return $null
        }

        $prop = $current.PSObject.Properties[$segment]
        if ($null -eq $prop) {
            $alt = Get-AlternateSegmentName -Segment $segment
            if ($alt) {
                $prop = $current.PSObject.Properties[$alt]
            }
            if ($null -eq $prop) {
                return $null
            }
        }
        $current = $prop.Value
    }
    return $current
}

function Get-AlternateSegmentName {
    param([string]$Segment)
    if (-not $Segment) {
        return ""
    }
    if ($Segment.Contains("_")) {
        return [regex]::Replace($Segment, '_([a-zA-Z0-9])', {
            param($m)
            $m.Groups[1].Value.ToUpperInvariant()
        })
    }
    $snake = [regex]::Replace($Segment, '([a-z0-9])([A-Z])', '$1_$2').ToLowerInvariant()
    if ($snake -ne $Segment) {
        return $snake
    }
    return ""
}

function Get-ObjectPropertyValue {
    param(
        [object]$Object,
        [string]$Name
    )
    if ($null -eq $Object -or -not $Name) {
        return $null
    }
    $prop = $Object.PSObject.Properties[$Name]
    if ($null -eq $prop) {
        return $null
    }
    return $prop.Value
}

function Coerce-Scalar {
    param([string]$Value)
    if ($null -eq $Value) {
        return $null
    }
    $trim = $Value.Trim()
    if ($trim -match '^-?\d+$') {
        return [int64]$trim
    }
    if ($trim -match '^-?\d+\.\d+$') {
        return [double]$trim
    }
    if ($trim -eq "true") {
        return $true
    }
    if ($trim -eq "false") {
        return $false
    }
    return $Value
}

function Resolve-TemplateString {
    param(
        [string]$Template,
        [hashtable]$Context
    )
    if ($null -eq $Template) {
        return $null
    }

    $resolved = [System.Text.RegularExpressions.Regex]::Replace(
        $Template,
        '\{\{([a-zA-Z0-9_\-]+)\}\}',
        {
            param($m)
            $key = $m.Groups[1].Value
            if ($Context.ContainsKey($key)) {
                return [string]$Context[$key]
            }
            throw "Missing template key: $key"
        }
    )
    return (Coerce-Scalar -Value $resolved)
}

function Resolve-Object {
    param(
        [object]$Value,
        [hashtable]$Context
    )
    if ($null -eq $Value) {
        return $null
    }

    if ($Value -is [string]) {
        return Resolve-TemplateString -Template $Value -Context $Context
    }

    if ($Value -is [System.Collections.IDictionary]) {
        $out = @{}
        foreach ($k in $Value.Keys) {
            $out[$k] = Resolve-Object -Value $Value[$k] -Context $Context
        }
        return $out
    }

    if (($Value -is [System.Array]) -or ($Value -is [System.Collections.IList])) {
        $arr = @()
        foreach ($item in $Value) {
            $arr += ,(Resolve-Object -Value $item -Context $Context)
        }
        return $arr
    }

    if ($Value.PSObject -and $Value.PSObject.Properties.Count -gt 0) {
        $out = @{}
        foreach ($p in $Value.PSObject.Properties) {
            $out[$p.Name] = Resolve-Object -Value $p.Value -Context $Context
        }
        return $out
    }

    return $Value
}

function Parse-GrpcCode {
    param([string]$OutputText)
    if (-not $OutputText) {
        return "Unknown"
    }
    $m = [regex]::Match($OutputText, 'Code:\s*([A-Za-z]+)')
    if ($m.Success) {
        return $m.Groups[1].Value
    }
    return "Unknown"
}

function Invoke-GrpcJson {
    param(
        [string]$Address,
        [string]$Method,
        [object]$Request,
        [hashtable]$Headers
    )

    $requestJson = "{}"
    if ($Request -ne $null) {
        $requestJson = $Request | ConvertTo-Json -Depth 60 -Compress
    }

    $requestId = [Guid]::NewGuid().ToString("N")
    $allHeaders = @{}
    if ($Headers) {
        foreach ($k in $Headers.Keys) {
            $allHeaders[$k] = [string]$Headers[$k]
        }
    }
    $allHeaders["x-request-id"] = $requestId

    $args = @("-plaintext")
    foreach ($key in ($allHeaders.Keys | Sort-Object)) {
        $args += @("-H", ("{0}: {1}" -f $key, $allHeaders[$key]))
    }
    $args += @("-d", "@", $Address, $Method)

    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $prevNativePreference = $false
    $hadNativePreference = $false
    try {
        if (Get-Variable -Name PSNativeCommandUseErrorActionPreference -Scope Global -ErrorAction SilentlyContinue) {
            $hadNativePreference = $true
            $prevNativePreference = $global:PSNativeCommandUseErrorActionPreference
            $global:PSNativeCommandUseErrorActionPreference = $false
        }
        $prevErrPref = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        $output = ($requestJson | & grpcurl @args 2>&1)
        $exitCode = $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $prevErrPref
        if ($hadNativePreference) {
            $global:PSNativeCommandUseErrorActionPreference = $prevNativePreference
        }
        $sw.Stop()
    }

    $text = ($output | Out-String).Trim()
    $grpcCode = if ($exitCode -eq 0) { "OK" } else { Parse-GrpcCode -OutputText $text }

    $responseObj = $null
    if ($exitCode -eq 0 -and $text) {
        try {
            $responseObj = $text | ConvertFrom-Json
        } catch {
            $responseObj = $null
        }
    }

    return @{
        ok          = ($exitCode -eq 0)
        code        = $grpcCode
        output      = $text
        response    = $responseObj
        request_id  = $requestId
        elapsed_ms  = [int]$sw.ElapsedMilliseconds
        method      = $Method
        request_json = $requestJson
    }
}

function Get-ActorHeaders {
    param(
        [string]$Actor,
        [hashtable]$Ctx
    )

    if (-not $Actor) {
        return @{}
    }

    switch -Regex ($Actor) {
        '^client$' {
            return @{ authorization = "Bearer $($Ctx.client_token)" }
        }
        '^freelancer$' {
            return @{ authorization = "Bearer $($Ctx.freelancer_token)" }
        }
        '^internal:(.+)$' {
            $caller = $Matches[1]
            return @{
                "x-jobconnect-internal" = $caller
                "x-jobconnect-internal-secret" = [string]$Ctx.internal_secret
            }
        }
        default {
            return @{}
        }
    }
}

function Merge-Headers {
    param(
        [hashtable]$A,
        [hashtable]$B
    )
    $out = @{}
    if ($A) {
        foreach ($k in $A.Keys) { $out[$k] = $A[$k] }
    }
    if ($B) {
        foreach ($k in $B.Keys) { $out[$k] = $B[$k] }
    }
    return $out
}

function Add-Result {
    param(
        [string]$Rpc,
        [string]$Scenario,
        [string]$Actor,
        [string[]]$ExpectedCodes,
        [hashtable]$Call
    )

    $expectedHit = $false
    foreach ($code in $ExpectedCodes) {
        if ($Call.code -eq $code) {
            $expectedHit = $true
            break
        }
    }

    $script:Results.Add([pscustomobject]@{
        rpc            = $Rpc
        scenario       = $Scenario
        actor          = $Actor
        expected_codes = ($ExpectedCodes -join ",")
        actual_code    = $Call.code
        success        = $expectedHit
        elapsed_ms     = $Call.elapsed_ms
        request_id     = $Call.request_id
        detail         = if ($Call.output) { $Call.output } else { "" }
    })

    if ($expectedHit) {
        Write-Ok "$Rpc [$Scenario] -> $($Call.code)"
    } else {
        Write-Warn "$Rpc [$Scenario] -> expected: $($ExpectedCodes -join ',') | actual: $($Call.code)"
    }

    return $expectedHit
}

function Invoke-CaseScenario {
    param(
        [string]$Rpc,
        [string]$Scenario,
        [string]$Actor,
        [object]$Request,
        [object]$ExtraHeaders,
        [string[]]$ExpectedCodes,
        [hashtable]$Ctx
    )

    $resolvedReq = Resolve-Object -Value $Request -Context $Ctx
    $baseHeaders = Get-ActorHeaders -Actor $Actor -Ctx $Ctx
    $resolvedExtra = @{}
    if ($ExtraHeaders) {
        $resolvedExtra = Resolve-Object -Value $ExtraHeaders -Context $Ctx
    }
    $headers = Merge-Headers -A $baseHeaders -B $resolvedExtra
    $call = Invoke-GrpcJson -Address $Ctx.contract_address -Method ("contract.v1.ContractService/{0}" -f $Rpc) -Request $resolvedReq -Headers $headers
    $ok = Add-Result -Rpc $Rpc -Scenario $Scenario -Actor $Actor -ExpectedCodes $ExpectedCodes -Call $call
    return @{ ok = $ok; call = $call }
}

function Get-Int64 {
    param([object]$Value)
    if ($null -eq $Value) {
        return [int64]0
    }
    return [int64]$Value
}

function Invoke-LoginWithFallbackRegister {
    param(
        [string]$BaseUrl,
        [string]$Role,
        [string]$Email,
        [string]$Password
    )

    $loginBody = @{ email = $Email; password = $Password }

    Start-Sleep -Milliseconds 1200
    $login = Invoke-HttpJson -Method POST -Url "$BaseUrl/api/v1/auth/login" -Body $loginBody -Headers @{}
    if ($login.ok -and $login.body -and $login.body.access_token) {
        return [string]$login.body.access_token
    }

    Write-Warn "Login failed for $Role ($Email). Attempting register and retry login."

    $regBody = @{
        email        = $Email
        password     = $Password
        first_name   = "Live"
        last_name    = "Tester"
        role         = $Role
        accept_terms = $true
    }

    Start-Sleep -Milliseconds 1200
    $reg = Invoke-HttpJson -Method POST -Url "$BaseUrl/api/v1/auth/register" -Body $regBody -Headers @{}
    if (-not $reg.ok -and $reg.status -ne 409) {
        Write-Warn "Register did not succeed for $Role. status=$($reg.status) raw=$($reg.raw)"
    }

    Start-Sleep -Milliseconds 1200
    $retry = Invoke-HttpJson -Method POST -Url "$BaseUrl/api/v1/auth/login" -Body $loginBody -Headers @{}
    if ($retry.ok -and $retry.body -and $retry.body.access_token) {
        return [string]$retry.body.access_token
    }

    throw "Unable to login $Role user '$Email'. Last status: $($retry.status) body: $($retry.raw)"
}

function New-JobFixture {
    param(
        [string]$ClientToken,
        [string]$Kind,
        [string]$Tag
    )

    $deadline = [DateTimeOffset]::UtcNow.AddDays(14).ToUnixTimeSeconds()
    $payload = @{
        title                 = "Live Test Job $Tag"
        description           = "Contract service live test job ($Tag)"
        required_skills       = @("grpc", "docker", "qa")
        deadline_unix_seconds = $deadline
    }

    if ($Kind -eq "hourly") {
        $payload["job_type_enum"] = "JOB_TYPE_HOURLY"
        $payload["hourly_rate"] = 55
    } else {
        $payload["job_type_enum"] = "JOB_TYPE_FIXED"
        $payload["budget_fixed"] = 1200
    }

    $call = Invoke-GrpcJson -Address "localhost:50053" -Method "job.v1.JobService/CreateJob" -Request $payload -Headers @{ authorization = "Bearer $ClientToken" }
    if ($call.code -ne "OK") {
        throw "CreateJob gRPC fixture failed for $Tag. code=$($call.code) output=$($call.output)"
    }

    $jobId = Get-Int64 (Get-PathValue -Object $call.response -Path "job.id")
    if ($jobId -le 0) {
        throw "CreateJob gRPC fixture missing job.id for $Tag"
    }

    return $jobId
}

function New-ProposalFixture {
    param(
        [int64]$JobId,
        [string]$Kind,
        [string]$FreelancerToken
    )

    $bidType = if ($Kind -eq "hourly") { "hourly" } else { "fixed" }
    $proposalReq = @{
        job_id         = $JobId
        cover_letter   = "Submitting proposal for live contract test"
        bid_type       = $bidType
        bid_amount     = if ($Kind -eq "hourly") { 55 } else { 1200 }
        estimated_days = 10
        connects_spent = 6
    }

    $call = Invoke-GrpcJson -Address "localhost:50054" -Method "proposal.v1.ProposalService/SubmitProposal" -Request $proposalReq -Headers @{ authorization = "Bearer $FreelancerToken" }
    if ($call.code -ne "OK") {
        throw "SubmitProposal failed for job $JobId. code=$($call.code) output=$($call.output)"
    }

    $proposalId = Get-Int64 (Get-PathValue -Object $call.response -Path "proposal.id")
    if ($proposalId -le 0) {
        throw "SubmitProposal response missing proposal.id for job $JobId"
    }
    return $proposalId
}

function New-JobProposalPair {
    param(
        [string]$Kind,
        [string]$Tag,
        [hashtable]$Ctx
    )

    $jobId = New-JobFixture -ClientToken $Ctx.client_token -Kind $Kind -Tag $Tag
    $proposalId = New-ProposalFixture -JobId $jobId -Kind $Kind -FreelancerToken $Ctx.freelancer_token
    return [pscustomobject]@{
        job_id = $jobId
        proposal_id = $proposalId
        kind = $Kind
        tag = $Tag
    }
}

function New-ContractOffer {
    param(
        [int64]$JobId,
        [int64]$ProposalId,
        [string]$Kind,
        [int]$MilestoneCount,
        [hashtable]$Ctx
    )

    $req = @{
        freelancer_id = $Ctx.freelancer_user_id
        job_id = $JobId
        proposal_id = $ProposalId
    }

    if ($Kind -eq "hourly") {
        $req.contract_type = "CONTRACT_TYPE_HOURLY"
        $req.title = "Hourly Contract $($Ctx.run_id)"
        $req.description = "Hourly live test contract"
        $req.hourly_rate_minor = 5500
        $req.weekly_hour_limit = 30
    } else {
        $req.contract_type = "CONTRACT_TYPE_FIXED"
        $req.title = "Fixed Contract $($Ctx.run_id)"
        $req.description = "Fixed live test contract"
        $req.fixed_total_minor = 240000
        if ($MilestoneCount -lt 1) { $MilestoneCount = 1 }
        $milestones = @()
        $amountPer = [int64](240000 / $MilestoneCount)
        for ($i = 1; $i -le $MilestoneCount; $i++) {
            $milestones += @{
                title = "Milestone $i"
                description = "Milestone $i for live test"
                amount_minor = $amountPer
                due_at_unix_seconds = [DateTimeOffset]::UtcNow.AddDays(7 + $i).ToUnixTimeSeconds()
            }
        }
        $req.milestones = $milestones
    }

    $call = Invoke-GrpcJson -Address $Ctx.contract_address -Method "contract.v1.ContractService/CreateContract" -Request $req -Headers @{ authorization = "Bearer $($Ctx.client_token)" }
    if ($call.code -ne "OK") {
        throw "CreateContract fixture failed (job=$JobId proposal=$ProposalId kind=$Kind). code=$($call.code) output=$($call.output)"
    }

    $contractId = Get-Int64 (Get-PathValue -Object $call.response -Path "contract.id")
    if ($contractId -le 0) {
        throw "CreateContract fixture missing contract.id"
    }

    $milestoneIds = @()
    $milestones = Get-PathValue -Object $call.response -Path "contract.milestones"
    if ($milestones) {
        foreach ($m in $milestones) {
            $id = Get-Int64 (Get-PathValue -Object $m -Path "id")
            if ($id -gt 0) { $milestoneIds += $id }
        }
    }

    return [pscustomobject]@{
        contract_id = $contractId
        milestone_ids = $milestoneIds
    }
}

function Accept-ContractFixture {
    param(
        [int64]$ContractId,
        [hashtable]$Ctx
    )

    $call = Invoke-GrpcJson -Address $Ctx.contract_address -Method "contract.v1.ContractService/AcceptContract" -Request @{ contract_id = $ContractId } -Headers @{ authorization = "Bearer $($Ctx.freelancer_token)" }
    if ($call.code -ne "OK") {
        throw "AcceptContract fixture failed for contract $ContractId. code=$($call.code) output=$($call.output)"
    }
}

function Invoke-RequiredFixtureRpc {
    param(
        [string]$Method,
        [object]$Request,
        [hashtable]$Headers,
        [string[]]$ExpectedCodes,
        [string]$Label,
        [hashtable]$Ctx
    )

    $call = Invoke-GrpcJson -Address $Ctx.contract_address -Method $Method -Request $Request -Headers $Headers
    $hit = $false
    foreach ($code in $ExpectedCodes) {
        if ($call.code -eq $code) { $hit = $true; break }
    }
    if (-not $hit) {
        throw "$Label failed. Expected $($ExpectedCodes -join ',') but got $($call.code). output=$($call.output)"
    }
    return $call
}

Write-Step "Contract RPC Live Test - Preflight"
$workspaceRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
$backendDir = Resolve-Path (Join-Path $PSScriptRoot "..")
$envPath = Join-Path $backendDir ".env"
$postmanEnvPath = Join-Path $backendDir "docs/postman/proposal-local.postman_environment.json"

Require-Command -Name "docker"
Require-Command -Name "grpcurl"
Require-Command -Name "curl.exe"

$null = & docker compose version 2>$null
if ($LASTEXITCODE -ne 0) {
    throw "docker compose is not available or not running properly."
}

if (-not (Test-Path $CasesFile)) {
    throw "Cases file not found: $CasesFile"
}

$envMap = Parse-EnvFile -Path $envPath
if ((-not $envMap.ContainsKey("JOBCONNECT_INTERNAL_CALLER_SECRET")) -or [string]::IsNullOrWhiteSpace($envMap["JOBCONNECT_INTERNAL_CALLER_SECRET"])) {
    if (-not [string]::IsNullOrWhiteSpace($env:JOBCONNECT_INTERNAL_CALLER_SECRET)) {
        $envMap["JOBCONNECT_INTERNAL_CALLER_SECRET"] = $env:JOBCONNECT_INTERNAL_CALLER_SECRET
    }
}
if ((-not $envMap.ContainsKey("JOBCONNECT_INTERNAL_CALLER_SECRET")) -or [string]::IsNullOrWhiteSpace($envMap["JOBCONNECT_INTERNAL_CALLER_SECRET"])) {
    throw "JOBCONNECT_INTERNAL_CALLER_SECRET must be set in backend/.env or process environment before running live RPC tests."
}

$script:Context.run_id = [DateTimeOffset]::UtcNow.ToString("yyyyMMddHHmmss")
$script:Context.contract_address = "localhost:50055"
$script:Context.base_url = "http://localhost:8080"
$script:Context.internal_secret = $envMap["JOBCONNECT_INTERNAL_CALLER_SECRET"]
$script:Context.future_due_unix = [DateTimeOffset]::UtcNow.AddDays(21).ToUnixTimeSeconds()

Write-Info "Workspace: $workspaceRoot"
Write-Info "Backend: $backendDir"
Write-Info "Cases file: $CasesFile"

Write-Step "Start Docker Compose Stack"
Push-Location $backendDir
try {
    if ($SkipComposeBuild) {
        & docker compose up -d
    } else {
        & docker compose up --build -d
    }
    if ($LASTEXITCODE -ne 0) {
        throw "docker compose up failed"
    }

    & docker compose ps

    if (-not $SkipMigrations) {
        Write-Step "Run Database Migrations"
        & "$backendDir/scripts/migrate-all.ps1"
        if ($LASTEXITCODE -ne 0) {
            throw "migrate-all.ps1 failed"
        }
    } else {
        Write-Info "Skipping migrations (-SkipMigrations)"
    }
} finally {
    Pop-Location
}

Write-Step "Wait For Services"
Wait-ForHttp -Url "$($script:Context.base_url)/healthz" -Attempts 90 -DelaySeconds 2
Wait-ForTcpPort -Port 8080 -Attempts 45 -DelaySeconds 2
Wait-ForTcpPort -Port 50055 -Attempts 45 -DelaySeconds 2
Write-Ok "Gateway and Contract gRPC ports are reachable"

Write-Step "Auth Bootstrap"
$defaultClientEmail = Get-PostmanEnvValue -Path $postmanEnvPath -Key "clientEmail"
$defaultFreelancerEmail = Get-PostmanEnvValue -Path $postmanEnvPath -Key "freelancerEmail"
$defaultPassword = Get-PostmanEnvValue -Path $postmanEnvPath -Key "password"

# Use fresh users per run so proposal connects balance is deterministic.
$defaultClientEmail = "itest.client.$($script:Context.run_id)@jobconnect.test"
$defaultFreelancerEmail = "itest.freel.$($script:Context.run_id)@jobconnect.test"
if (-not $defaultPassword) { $defaultPassword = "Passw0rd!23" }

$script:Context.client_email = $defaultClientEmail
$script:Context.freelancer_email = $defaultFreelancerEmail
$script:Context.test_password = $defaultPassword

$script:Context.client_token = Invoke-LoginWithFallbackRegister -BaseUrl $script:Context.base_url -Role "client" -Email $script:Context.client_email -Password $script:Context.test_password
$script:Context.freelancer_token = Invoke-LoginWithFallbackRegister -BaseUrl $script:Context.base_url -Role "freelancer" -Email $script:Context.freelancer_email -Password $script:Context.test_password

$clientClaims = Decode-JwtPayload -Token $script:Context.client_token
$freelancerClaims = Decode-JwtPayload -Token $script:Context.freelancer_token
$script:Context.client_user_id = [string]$clientClaims.sub
$script:Context.freelancer_user_id = [string]$freelancerClaims.sub

Write-Ok "Client and freelancer tokens acquired"
Write-Info "Client user id: $($script:Context.client_user_id)"
Write-Info "Freelancer user id: $($script:Context.freelancer_user_id)"

Write-Step "Fixture Bootstrap"
$pairCreate = New-JobProposalPair -Kind "fixed" -Tag "create" -Ctx $script:Context
$pairAccept = New-JobProposalPair -Kind "fixed" -Tag "accept" -Ctx $script:Context
$pairDecline = New-JobProposalPair -Kind "fixed" -Tag "decline" -Ctx $script:Context
$pairRevoke = New-JobProposalPair -Kind "fixed" -Tag "revoke" -Ctx $script:Context
$pairWorkflow = New-JobProposalPair -Kind "fixed" -Tag "workflow" -Ctx $script:Context
$pairHourly = New-JobProposalPair -Kind "hourly" -Tag "hourly" -Ctx $script:Context

$script:Context.create_job_id = $pairCreate.job_id
$script:Context.create_proposal_id = $pairCreate.proposal_id

$acceptOffer = New-ContractOffer -JobId $pairAccept.job_id -ProposalId $pairAccept.proposal_id -Kind "fixed" -MilestoneCount 1 -Ctx $script:Context
$declineOffer = New-ContractOffer -JobId $pairDecline.job_id -ProposalId $pairDecline.proposal_id -Kind "fixed" -MilestoneCount 1 -Ctx $script:Context
$revokeOffer = New-ContractOffer -JobId $pairRevoke.job_id -ProposalId $pairRevoke.proposal_id -Kind "fixed" -MilestoneCount 1 -Ctx $script:Context
$workflowOffer = New-ContractOffer -JobId $pairWorkflow.job_id -ProposalId $pairWorkflow.proposal_id -Kind "fixed" -MilestoneCount 2 -Ctx $script:Context
$hourlyOffer = New-ContractOffer -JobId $pairHourly.job_id -ProposalId $pairHourly.proposal_id -Kind "hourly" -MilestoneCount 0 -Ctx $script:Context

Accept-ContractFixture -ContractId $workflowOffer.contract_id -Ctx $script:Context
Accept-ContractFixture -ContractId $hourlyOffer.contract_id -Ctx $script:Context

$script:Context.accept_job_id = $pairAccept.job_id
$script:Context.accept_proposal_id = $pairAccept.proposal_id
$script:Context.accept_contract_id = $acceptOffer.contract_id
$script:Context.decline_contract_id = $declineOffer.contract_id
$script:Context.revoke_contract_id = $revokeOffer.contract_id
$script:Context.revoke_job_id = $pairRevoke.job_id
$script:Context.workflow_contract_id = $workflowOffer.contract_id
$script:Context.hourly_contract_id = $hourlyOffer.contract_id

$workflowMilestones = @($workflowOffer.milestone_ids)
if ($workflowMilestones.Count -lt 2) {
    throw "Workflow contract did not return two milestones."
}
$script:Context.workflow_milestone_1_id = [int64]$workflowMilestones[0]
$script:Context.workflow_milestone_2_id = [int64]$workflowMilestones[1]

$nowUnix = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$script:Context.hourly_log_start_unix = $nowUnix - 7200
$script:Context.hourly_log_end_unix = $nowUnix - 3600
$script:Context.hourly_log_update_start_unix = $nowUnix - 7500
$script:Context.hourly_log_update_end_unix = $nowUnix - 3900

$prepFundMilestone1 = Invoke-RequiredFixtureRpc -Method "contract.v1.ContractService/InternalMarkMilestoneFunded" -Request @{
    contract_id = $script:Context.workflow_contract_id
    milestone_id = $script:Context.workflow_milestone_1_id
} -Headers @{
    "x-jobconnect-internal" = "payment-service"
    "x-jobconnect-internal-secret" = $script:Context.internal_secret
} -ExpectedCodes @("OK") -Label "Prep fund workflow milestone1" -Ctx $script:Context

$prepFundMilestone2 = Invoke-RequiredFixtureRpc -Method "contract.v1.ContractService/InternalMarkMilestoneFunded" -Request @{
    contract_id = $script:Context.workflow_contract_id
    milestone_id = $script:Context.workflow_milestone_2_id
} -Headers @{
    "x-jobconnect-internal" = "payment-service"
    "x-jobconnect-internal-secret" = $script:Context.internal_secret
} -ExpectedCodes @("OK") -Label "Prep fund workflow milestone2" -Ctx $script:Context

$prepSubmitMilestone2 = Invoke-RequiredFixtureRpc -Method "contract.v1.ContractService/SubmitMilestoneWork" -Request @{
    contract_id = $script:Context.workflow_contract_id
    milestone_id = $script:Context.workflow_milestone_2_id
    note = "Prep submission for request-changes"
    attachments = @("https://example.test/prep-m2")
} -Headers @{ authorization = "Bearer $($script:Context.freelancer_token)" } -ExpectedCodes @("OK") -Label "Prep SubmitMilestoneWork for milestone2" -Ctx $script:Context

$prepDeleteLog = Invoke-RequiredFixtureRpc -Method "contract.v1.ContractService/LogHourlyWork" -Request @{
    contract_id = $script:Context.hourly_contract_id
    start_at_unix_seconds = $nowUnix - 10800
    end_at_unix_seconds = $nowUnix - 9000
    note = "Prep deletable hourly log"
    evidence_urls = @("https://example.test/prep-delete-log")
} -Headers @{ authorization = "Bearer $($script:Context.freelancer_token)" } -ExpectedCodes @("OK") -Label "Prep deletable hourly log" -Ctx $script:Context

$deleteLogId = Get-Int64 (Get-PathValue -Object $prepDeleteLog.response -Path "hourly_log.id")
if ($deleteLogId -le 0) {
    throw "Unable to get prep delete hourly log id"
}
$script:Context.hourly_log_delete_id = $deleteLogId
$script:Context.hourly_invoice_id = [int64]0
$script:Context.contract_bonus_id = [int64]0
$script:Context.amendment_id = [int64]0

Write-Ok "Fixtures created successfully"

Write-Step "Load Case Matrix"
$cases = Get-Content $CasesFile -Raw | ConvertFrom-Json
$cases = @($cases)
if (-not $cases -or $cases.Count -eq 0) {
    throw "No cases found in $CasesFile"
}
Write-Info "Loaded $($cases.Count) RPC case definitions"

Write-Step "Execute RPC Matrix"
foreach ($case in $cases) {
    $rpc = [string]$case.rpc
    if (-not $rpc) {
        continue
    }
    $extraHeaders = Get-ObjectPropertyValue -Object $case -Name "extra_headers"
    $wrongRoleActor = Get-ObjectPropertyValue -Object $case -Name "wrong_role_actor"
    $wrongRoleExpected = Get-ObjectPropertyValue -Object $case -Name "wrong_role_expect_codes"
    $invalidRequest = Get-ObjectPropertyValue -Object $case -Name "invalid_request"
    $stateViolationRequest = Get-ObjectPropertyValue -Object $case -Name "state_violation_request"
    $stateViolationExpected = Get-ObjectPropertyValue -Object $case -Name "state_violation_expect_codes"
    $saveFromResponse = Get-ObjectPropertyValue -Object $case -Name "save_from_response"

    $positiveExpected = @()
    foreach ($c in $case.expect_success_codes) {
        $positiveExpected += [string]$c
    }

    $positive = Invoke-CaseScenario -Rpc $rpc -Scenario "positive" -Actor ([string]$case.actor) -Request $case.request -ExtraHeaders $extraHeaders -ExpectedCodes $positiveExpected -Ctx $script:Context

    if ($positive.ok -and $saveFromResponse) {
        $map = Resolve-Object -Value $saveFromResponse -Context $script:Context
        foreach ($targetKey in $map.Keys) {
            $sourcePath = [string]$map[$targetKey]
            $value = Get-PathValue -Object $positive.call.response -Path $sourcePath
            if ($null -ne $value -and "$value" -ne "") {
                $script:Context[$targetKey] = $value
                Write-Info "Captured context: $targetKey = $value"
            }
        }
    }

    $unauthExpected = if ([string]$case.actor -like "internal:*") { @("PermissionDenied") } else { @("Unauthenticated", "PermissionDenied") }
    $null = Invoke-CaseScenario -Rpc $rpc -Scenario "negative:unauthenticated" -Actor "none" -Request $case.request -ExtraHeaders @{} -ExpectedCodes $unauthExpected -Ctx $script:Context

    if ($wrongRoleActor) {
        $wrongRoleCodes = @("PermissionDenied")
        if ($wrongRoleExpected) {
            $wrongRoleCodes = @()
            foreach ($c in $wrongRoleExpected) {
                $wrongRoleCodes += [string]$c
            }
        }
        $null = Invoke-CaseScenario -Rpc $rpc -Scenario "negative:wrong-role" -Actor ([string]$wrongRoleActor) -Request $case.request -ExtraHeaders $extraHeaders -ExpectedCodes $wrongRoleCodes -Ctx $script:Context
    }

    if ([string]$case.actor -like "internal:*") {
        $caller = ([string]$case.actor).Split(':')[1]
        $missingSecretHeaders = @{ "x-jobconnect-internal" = $caller }
        $null = Invoke-CaseScenario -Rpc $rpc -Scenario "negative:internal-missing-secret" -Actor "none" -Request $case.request -ExtraHeaders $missingSecretHeaders -ExpectedCodes @("PermissionDenied") -Ctx $script:Context

        $wrongSecretHeaders = @{ "x-jobconnect-internal" = $caller; "x-jobconnect-internal-secret" = "wrong-secret" }
        if ($extraHeaders) {
            $resolvedExtra = Resolve-Object -Value $extraHeaders -Context $script:Context
            $wrongSecretHeaders = Merge-Headers -A $wrongSecretHeaders -B $resolvedExtra
        }
        $null = Invoke-CaseScenario -Rpc $rpc -Scenario "negative:internal-wrong-secret" -Actor "none" -Request $case.request -ExtraHeaders $wrongSecretHeaders -ExpectedCodes @("PermissionDenied") -Ctx $script:Context
    }

    if ($invalidRequest) {
        $null = Invoke-CaseScenario -Rpc $rpc -Scenario "negative:invalid-request" -Actor ([string]$case.actor) -Request $invalidRequest -ExtraHeaders $extraHeaders -ExpectedCodes @("InvalidArgument") -Ctx $script:Context
    }

    if ($stateViolationRequest) {
        $stateExpected = @("FailedPrecondition")
        if ($stateViolationExpected) {
            $stateExpected = @()
            foreach ($c in $stateViolationExpected) { $stateExpected += [string]$c }
        }
        $null = Invoke-CaseScenario -Rpc $rpc -Scenario "negative:state-violation" -Actor ([string]$case.actor) -Request $stateViolationRequest -ExtraHeaders $extraHeaders -ExpectedCodes $stateExpected -Ctx $script:Context
    }
}

Write-Step "Build Summary"
$failed = @($script:Results | Where-Object { -not $_.success })
$passed = @($script:Results | Where-Object { $_.success })

$summary = [pscustomobject]@{
    generated_at_utc = [DateTime]::UtcNow.ToString("o")
    run_id           = $script:Context.run_id
    totals           = [pscustomobject]@{
        total  = $script:Results.Count
        passed = $passed.Count
        failed = $failed.Count
    }
    context_snapshot = [pscustomobject]@{
        create_job_id        = $script:Context.create_job_id
        create_proposal_id   = $script:Context.create_proposal_id
        accept_contract_id   = $script:Context.accept_contract_id
        decline_contract_id  = $script:Context.decline_contract_id
        revoke_contract_id   = $script:Context.revoke_contract_id
        workflow_contract_id = $script:Context.workflow_contract_id
        hourly_contract_id   = $script:Context.hourly_contract_id
        hourly_log_id        = $script:Context.hourly_log_id
        hourly_invoice_id    = $script:Context.hourly_invoice_id
        contract_bonus_id    = $script:Context.contract_bonus_id
        amendment_id         = $script:Context.amendment_id
    }
    results = $script:Results
}

$summary | ConvertTo-Json -Depth 80 | Set-Content -Encoding UTF8 $SummaryOut

$table = $script:Results |
    Select-Object rpc, scenario, actor, expected_codes, actual_code, success, elapsed_ms, request_id |
    Format-Table -AutoSize | Out-String
Write-Host $table

Write-Info "Summary JSON written to: $SummaryOut"
if ($failed.Count -gt 0) {
    Write-Warn "Completed with failures: $($failed.Count) of $($script:Results.Count) scenarios did not match expected gRPC code."
    exit 1
}

Write-Ok "All scenarios matched expected outcomes."
exit 0
