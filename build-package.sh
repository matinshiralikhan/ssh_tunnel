# Navigate to the directory
cd /tools/ssh_tunnel
# Build for Windows (.exe file)
cd /tools/ssh_tunnel
export GOOS=windows
export GOARCH=amd64
go build -o /tools/ssh_tunnel/ssh-tunnel.exe -ldflags="-s -w" ssh-tunnel.go


# Build for Linux (.bin file)
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0
go build -o /tools/ssh_tunnel/ssh-tunnel.bin -ldflags="-s -w" ssh-tunnel.go

