#!/bin/sh
app_version=$(curl https://raw.githubusercontent.com/osamsam321/sbot/main/VERSION)
curl -OL https://github.com/osamsam321/sbot/releases/download/$app_version/sbot_$app_version.zip
unzip -o sbot_$app_version.zip && rm -rf ~/.sbot && mv -f sbot ~/.sbot
