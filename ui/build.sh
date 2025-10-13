#!/bin/sh
cd "$(dirname "$0")"
npx next build
rm -rf ../embed/ui
mkdir -p ../embed/ui
cp -r out/* ../embed/ui/ 