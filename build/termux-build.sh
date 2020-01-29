#!/bin/bash

PATH=$PATH:~/.local/bin
export PATH

DEB_VERSION=$(echo $GIT_DESCRIBE | cut -c2-)

APP_NAME=go-mud

sed -i 's/VAR-VERSION/'$DEB_VERSION'/' build/termux-armv6.json
termux-create-package build/termux-armv6.json
test -f ${APP_NAME}_${DEB_VERSION}_arm.deb || exit 1
mv ${APP_NAME}_*_arm.deb dist/${APP_NAME}_v${DEB_VERSION}_Termux_ARMv6.deb

sed -i 's/VAR-VERSION/'$DEB_VERSION'/' build/termux-armv7.json
termux-create-package build/termux-armv7.json
test -f ${APP_NAME}_${DEB_VERSION}_arm.deb || exit 1
mv ${APP_NAME}_*_arm.deb dist/${APP_NAME}_v${DEB_VERSION}_Termux_ARMv7.deb

sed -i 's/VAR-VERSION/'$DEB_VERSION'/' build/termux-armv8.json
termux-create-package build/termux-armv8.json
test -f ${APP_NAME}_${DEB_VERSION}_aarch64.deb || exit 1
mv ${APP_NAME}_*_aarch64.deb dist/${APP_NAME}_v${DEB_VERSION}_Termux_ARMv8.deb

(
    cd dist;
    rm -f checksums.txt
    (
        echo SHA256SUM:
        sha256sum ${APP_NAME}_*.{tar.gz,zip,deb}
        echo; echo MD5SUM:
        md5sum ${APP_NAME}_*.{tar.gz,zip,deb}
    ) > checksums.txt
)

GH_TAGS=https://api.github.com/repos/$GITHUB_REPOSITORY/releases/tags
curl -s $GH_TAGS/$GIT_DESCRIBE > out.json
URL=$(jq -r '.upload_url' out.json | sed 's/{.*}//')
RELEASE_ID=$(jq -r '.id' out.json)

if [[ $URL != http* ]]; then
    echo Cant get URL for $GIT_DESCRIBE
    exit 1
fi

echo "Delete checksums.txt ..."

CHECKSUMS_URL=$(jq -r '.assets[] | select(.name == "checksums.txt") | .url' out.json)
if [[ $CHECKSUMS_URL != http* ]]; then
    echo Cant get URL for checksums.txt
    exit 1
fi

curl --silent -X DELETE -H "Authorization: token $GITHUB_TOKEN" $CHECKSUMS_URL

echo "Uploading asset ..."

for file in dist/checksums.txt dist/${APP_NAME}_*.deb; do
    echo "    Uploading $file"
    GH_ASSET="$URL?name=$(basename $file)"
    HTTP_CODE=$(curl --silent --output out.json -w '%{http_code}' \
            -H "Authorization: token $GITHUB_TOKEN" \
            -H "Content-Type: application/octet-stream" \
            --data-binary @"$file" \
            $GH_ASSET
        )
    if [[ $HTTP_CODE != 2* ]]; then
        jq --monochrome-output '' out.json
        URL=https://api.github.com/repos/$GITHUB_REPOSITORY/releases/$RELEASE_ID
        curl --silent -X DELETE -H "Authorization: token $GITHUB_TOKEN" $URL
        exit 1
    fi
done
