#!/bin/sh

#when times comes run this when the repo is public
#curl -LJO 'https://github.com/osamsam321/sbot/releases/download/0.1/sbot.zip'
unzip sbot.zip && rm -rf ~/.sbot && mv -f sbot ~/.sbot
