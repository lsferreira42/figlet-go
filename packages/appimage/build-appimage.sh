#!/bin/bash
# Script to build AppImage for figlet-go

set -e

APP=figlet
VERSION=${1:-"latest"}
WORKDIR="packages/appimage/work"
APPDIR="$WORKDIR/$APP.AppDir"

echo "Building AppImage for $APP version $VERSION..."

# Clean up
rm -rf "$WORKDIR"
mkdir -p "$APPDIR/usr/bin"
mkdir -p "$APPDIR/usr/share/figlet-go/fonts"

# Build the binary
go build -o "$APPDIR/usr/bin/figlet" figlet.go

# Copy fonts
cp fonts/*.flf "$APPDIR/usr/share/figlet-go/fonts/"
cp fonts/*.flc "$APPDIR/usr/share/figlet-go/fonts/"

# Create desktop file
cat > "$APPDIR/$APP.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=FIGlet Go
Icon=figlet
Exec=figlet
Categories=Utility;
Terminal=true
EOF

# Create AppRun script
cat > "$APPDIR/AppRun" <<EOF
#!/bin/sh
SELF=\$(readlink -f "\$0")
HERE=\${SELF%/*}
export PATH="\$HERE/usr/bin:\$PATH"
export FIGLET_FONTDIR="\$HERE/usr/share/figlet-go/fonts"
exec "\$HERE/usr/bin/figlet" "\$@"
EOF
chmod +x "$APPDIR/AppRun"

# Download appimagetool if not present
if [ ! -f "packages/appimage/appimagetool" ]; then
    echo "Downloading appimagetool..."
    curl -Lo packages/appimage/appimagetool https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-x86_64.AppImage
    chmod +x packages/appimage/appimagetool
fi

# Build AppImage
./packages/appimage/appimagetool "$APPDIR" "packages/appimage/figlet-go-$VERSION-x86_64.AppImage"

echo "AppImage built: packages/appimage/figlet-go-$VERSION-x86_64.AppImage"
