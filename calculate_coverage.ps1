# Скрипт для подсчета общего покрытия тестами
go test -coverprofile=temp_coverage.out ./... 2>$null
if ($LASTEXITCODE -eq 0) {
    go tool cover -func=temp_coverage.out | Select-String "total" | ForEach-Object {
        $_.ToString()
    }
} else {
    Write-Host "Error running tests"
}
Remove-Item temp_coverage.out -ErrorAction SilentlyContinue 