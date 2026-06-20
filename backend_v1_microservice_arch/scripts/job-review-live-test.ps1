$ErrorActionPreference='Stop'

function Get-JwtSub([string]$token){
  $parts=$token.Split('.')
  if($parts.Length -lt 2){ return '' }
  $p=$parts[1].Replace('-','+').Replace('_','/')
  switch($p.Length % 4){2{$p+='=='};3{$p+='='}}
  $json=[Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($p)) | ConvertFrom-Json
  if($json.sub){ return [string]$json.sub }
  if($json.user_id){ return [string]$json.user_id }
  return ''
}

function Invoke-HttpJson([string]$method,[string]$url,$body){
  try {
    $resp = Invoke-RestMethod -Method $method -Uri $url -Body ($body|ConvertTo-Json -Depth 20) -ContentType 'application/json' -TimeoutSec 30
    return @{ok=$true; body=$resp}
  } catch {
    return @{ok=$false; err=$_.Exception.Message}
  }
}

function Login-Or-Register([string]$base,[string]$email,[string]$password,[string]$role){
  $login=Invoke-HttpJson 'POST' "$base/api/v1/auth/login" @{email=$email; password=$password}
  if($login.ok -and $login.body.access_token){ return [string]$login.body.access_token }
  $null=Invoke-HttpJson 'POST' "$base/api/v1/auth/register" @{email=$email; password=$password; first_name='Live'; last_name='Tester'; role=$role; accept_terms=$true}
  $retry=Invoke-HttpJson 'POST' "$base/api/v1/auth/login" @{email=$email; password=$password}
  if($retry.ok -and $retry.body.access_token){ return [string]$retry.body.access_token }
  throw "login/register failed for $role ($email)"
}

function Invoke-GrpcCall([string]$addr,[string]$method,[hashtable]$headers,[string]$json,[switch]$UseReviewProto){
  if(-not $json){ $json='{}' }
  $args=@('-plaintext')
  if($UseReviewProto){
    $args += @('-import-path','backend_v1_microservice_arch/api/proto','-proto','review/v1/review.proto')
  }
  foreach($k in $headers.Keys){ $args += @('-H',("{0}: {1}" -f $k,$headers[$k])) }
  $args += @('-d','@',$addr,$method)
  $hadNativePref = $false
  $prevNativePref = $false
  if (Get-Variable -Name PSNativeCommandUseErrorActionPreference -Scope Global -ErrorAction SilentlyContinue) {
    $hadNativePref = $true
    $prevNativePref = $global:PSNativeCommandUseErrorActionPreference
    $global:PSNativeCommandUseErrorActionPreference = $false
  }
  $prevErr = $ErrorActionPreference
  $ErrorActionPreference = 'Continue'
  try {
    $output = ($json | & grpcurl @args 2>&1)
    $exit=$LASTEXITCODE
  } finally {
    $ErrorActionPreference = $prevErr
    if ($hadNativePref) {
      $global:PSNativeCommandUseErrorActionPreference = $prevNativePref
    }
  }
  $text = ($output|Out-String)
  $code='OK'
  if($exit -ne 0){
    $m=[regex]::Match($text,'Code:\s*([A-Za-z]+)')
    if($m.Success){ $code=$m.Groups[1].Value } else { $code='Unknown' }
  }
  return @{code=$code; text=$text; exit=$exit}
}

$base='http://localhost:8080'
$run=(Get-Date).ToString('yyyyMMddHHmmss')
$pw='Passw0rd!123'
$clientEmail="itest.job.client.$run@jobconnect.test"
$freelEmail="itest.job.freel.$run@jobconnect.test"
$clientToken=Login-Or-Register $base $clientEmail $pw 'client'
$freelToken=Login-Or-Register $base $freelEmail $pw 'freelancer'
$clientId=Get-JwtSub $clientToken
$freelId=Get-JwtSub $freelToken

$jobCreateReq = (@{title="Live Job $run"; description='Live job rpc test'; required_skills=@('go','grpc'); budget_fixed=1200; deadline_unix_seconds=[DateTimeOffset]::UtcNow.AddDays(10).ToUnixTimeSeconds(); job_type_enum='JOB_TYPE_FIXED'} | ConvertTo-Json -Compress)
$create = Invoke-GrpcCall 'localhost:50053' 'job.v1.JobService/CreateJob' @{authorization="Bearer $clientToken"} $jobCreateReq
if($create.code -ne 'OK'){ throw "CreateJob failed: $($create.text)" }
$createObj = $create.text | ConvertFrom-Json
$jobId = [int64]$createObj.job.id

$results = New-Object System.Collections.Generic.List[object]

function Add-Res($svc,$method,$scenario,$code,[string]$detail){
  $results.Add([pscustomobject]@{service=$svc; method=$method; scenario=$scenario; code=$code; detail=$detail})
}

$core = @(
  @{m='job.v1.JobService/GetJob'; h=@{authorization="Bearer $clientToken"}; d=(@{job_id=$jobId}|ConvertTo-Json -Compress)},
  @{m='job.v1.JobService/UpdateJob'; h=@{authorization="Bearer $clientToken"}; d=(@{job_id=$jobId; title="Live Job $run updated"}|ConvertTo-Json -Compress)},
  @{m='job.v1.JobService/ListMyJobs'; h=@{authorization="Bearer $clientToken"}; d=(@{page_size=10}|ConvertTo-Json -Compress)},
  @{m='job.v1.JobService/SetJobVisibility'; h=@{authorization="Bearer $clientToken"}; d=(@{job_id=$jobId; visibility='VISIBILITY_PUBLIC'}|ConvertTo-Json -Compress)},
  @{m='job.v1.JobService/PauseJob'; h=@{authorization="Bearer $clientToken"}; d=(@{job_id=$jobId}|ConvertTo-Json -Compress)},
  @{m='job.v1.JobService/ReopenJob'; h=@{authorization="Bearer $clientToken"}; d=(@{job_id=$jobId}|ConvertTo-Json -Compress)},
  @{m='job.v1.JobService/SaveJob'; h=@{authorization="Bearer $freelToken"}; d=(@{job_id=$jobId}|ConvertTo-Json -Compress)},
  @{m='job.v1.JobService/ListSavedJobs'; h=@{authorization="Bearer $freelToken"}; d=(@{page_size=10}|ConvertTo-Json -Compress)},
  @{m='job.v1.JobService/UnsaveJob'; h=@{authorization="Bearer $freelToken"}; d=(@{job_id=$jobId}|ConvertTo-Json -Compress)},
  @{m='job.v1.JobService/GetPublicJobDetail'; h=@{}; d=(@{job_id=$jobId}|ConvertTo-Json -Compress)},
  @{m='job.v1.JobService/ListOpenJobs'; h=@{}; d=(@{page_size=10}|ConvertTo-Json -Compress)}
)
foreach($c in $core){
  $r=Invoke-GrpcCall 'localhost:50053' $c.m $c.h $c.d
  Add-Res 'job' $c.m 'core-positive' $r.code $r.text
}

$jobMethods = & grpcurl -plaintext localhost:50053 list job.v1.JobService
foreach($m in $jobMethods){
  $r1=Invoke-GrpcCall 'localhost:50053' $m @{} '{}'
  Add-Res 'job' $m 'sweep-unauth-empty' $r1.code $r1.text
  $r2=Invoke-GrpcCall 'localhost:50053' $m @{authorization="Bearer $clientToken"} '{}'
  Add-Res 'job' $m 'sweep-auth-empty' $r2.code $r2.text
}

$reviewMethods=@(
  'review.v1.ReviewService/CreateReview',
  'review.v1.ReviewService/GetReview',
  'review.v1.ReviewService/ListReviewsByUser',
  'review.v1.ReviewService/ListReviewsByContract',
  'review.v1.ReviewService/GetUserRatingSummary',
  'review.v1.ReviewService/UpdateReview',
  'review.v1.ReviewService/DeleteReview',
  'review.v1.ReviewService/ReplyToReview'
)
foreach($m in $reviewMethods){
  $r1=Invoke-GrpcCall 'localhost:50056' $m @{} '{}' -UseReviewProto
  Add-Res 'review' $m 'sweep-unauth-empty' $r1.code $r1.text
  $r2=Invoke-GrpcCall 'localhost:50056' $m @{authorization="Bearer $clientToken"} '{}' -UseReviewProto
  Add-Res 'review' $m 'sweep-auth-empty' $r2.code $r2.text
}

$validReviewCalls=@(
  @{m='review.v1.ReviewService/ListReviewsByUser'; h=@{}; d=(@{user_id=$freelId; page_size=10}|ConvertTo-Json -Compress)},
  @{m='review.v1.ReviewService/GetUserRatingSummary'; h=@{}; d=(@{user_id=$freelId}|ConvertTo-Json -Compress)},
  @{m='review.v1.ReviewService/CreateReview'; h=@{authorization="Bearer $clientToken"}; d=(@{contract_id=1; reviewee_id=$freelId; rating=5; title='Great'; comment='Solid work'}|ConvertTo-Json -Compress)}
)
foreach($c in $validReviewCalls){
  $r=Invoke-GrpcCall 'localhost:50056' $c.m $c.h $c.d -UseReviewProto
  Add-Res 'review' $c.m 'targeted' $r.code $r.text
}

$jobSummary = $results | Where-Object {$_.service -eq 'job'} | Group-Object scenario,code | Sort-Object Name | Select-Object Name,Count
$reviewSummary = $results | Where-Object {$_.service -eq 'review'} | Group-Object scenario,code | Sort-Object Name | Select-Object Name,Count

Write-Host 'JOB_LIVE_TEST_SUMMARY'
$jobSummary | Format-Table -AutoSize
Write-Host 'REVIEW_LIVE_TEST_SUMMARY'
$reviewSummary | Format-Table -AutoSize
Write-Host 'JOB_CORE_RESULTS'
$results | Where-Object {$_.service -eq 'job' -and $_.scenario -eq 'core-positive'} | Select-Object service,method,scenario,code | Format-Table -AutoSize
Write-Host 'NON_OK_DETAILS'
$results | Where-Object {$_.code -ne 'OK'} | Select-Object service,method,scenario,code,detail | Format-List
Write-Host 'CONTEXT'
Write-Host "job_id=$jobId"
Write-Host "client_id=$clientId"
Write-Host "freelancer_id=$freelId"

$results | ConvertTo-Json -Depth 5 | Set-Content 'backend_v1_microservice_arch/scripts/job-review-live-test-summary.json'
Write-Host 'summary_file=backend_v1_microservice_arch/scripts/job-review-live-test-summary.json'
