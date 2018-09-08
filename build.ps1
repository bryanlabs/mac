Write-Host "### Building Linux Binary."
#Build the Linux Binary.
$env:GOOS = "linux"
go build -o .\releases\mac .

Write-Host "### Building and installing Windows exe."
#Build and install the Windows executable.
$env:GOOS = "windows"
go build -o .\releases\mac.exe .
go install .



