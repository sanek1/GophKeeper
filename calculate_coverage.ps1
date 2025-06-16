# Скрипт для подсчета общего покрытия тестами (только internal модули)
go test -coverprofile=temp_coverage.out ./internal/... 2>$null
if ($LASTEXITCODE -eq 0) {
    $totalOutput = go tool cover -func=temp_coverage.out | Select-String "total"
    if ($totalOutput) {
        Write-Host $totalOutput.ToString()
        
        # Извлекаем процент покрытия
        $coverageMatch = $totalOutput.ToString() -match '(\d+\.\d+)%'
        if ($coverageMatch) {
            $coverage = [float]$matches[1]
            Write-Host "Coverage: $coverage%"
            if ($coverage -ge 70) {
                Write-Host "✅ Coverage check passed: $coverage% >= 70%" -ForegroundColor Green
            } else {
                Write-Host "❌ Coverage check failed: $coverage% < 70%" -ForegroundColor Red
            }
        }
    }
} else {
    Write-Host "Error running tests" -ForegroundColor Red
}
Remove-Item temp_coverage.out -ErrorAction SilentlyContinue 