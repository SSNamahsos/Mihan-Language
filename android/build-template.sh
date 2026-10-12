#!/bin/sh
set -e
cd "$(dirname "$0")"
SDK=${ANDROID_HOME:-$HOME/android}
BT=$(ls -d "$SDK"/build-tools/* | sort -V | tail -n1)
JAR=$(ls -d "$SDK"/platforms/android-* | sort -V | tail -n1)/android.jar
W=$(mktemp -d)
javac -source 8 -target 8 -Xlint:-options -bootclasspath "$JAR" -d "$W/classes" src/org/mihan/runner/*.java
"$BT/d8" --release --min-api 24 --lib "$JAR" --output "$W" $(find "$W/classes" -name '*.class')
"$BT/aapt2" link -I "$JAR" --manifest AndroidManifest.xml -o "$W/base.apk"
python3 -c "import zipfile,sys; z=zipfile.ZipFile(sys.argv[1],\"a\",zipfile.ZIP_DEFLATED); z.write(sys.argv[2],\"classes.dex\"); z.close()" "$W/base.apk" "$W/classes.dex"
mkdir -p ../cmd/mihan/android
cp "$W/base.apk" ../cmd/mihan/android/template.apk
rm -rf "$W"
echo "wrote cmd/mihan/android/template.apk"
