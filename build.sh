GOOS=linux GOARCH=amd64 go build -o rofimanager src/main.go
echo "Build complete."

echo "Creating package structure"
cp rofimanager rmanager/usr/local/bin/
echo "Packaging build to .deb"
dpkg-deb --build rmanager
echo "Package created: rmanager.deb"SS