#!/bin/sh
owner=osamsam321
repo=sbot
extension_type=zip
output_file=$repo.$extension_type
base_dir=~/.$repo

release_data=$(curl -s https://api.github.com/repos/$owner/$repo/releases/latest)
latest_tag=$(echo "$release_data" | sed -n 's/.*"tag_name": *"\(v[^"]*\)".*/\1/p')
echo "Using version $latest_tag"
asset_url=$(echo "$release_data" | grep '"browser_download_url":' | grep '.zip"' | head -n1 | sed -E 's/.*"browser_download_url": *"([^"]+)".*/\1/')

if [ -z "$asset_url" ]; then
    echo "Error: No suitable asset found in the latest release."
    exit 1
fi

echo "Downloading asset from $asset_url"
curl -L -o $output_file "$asset_url"
unzip -o $output_file
extracted_dir=$(unzip -Z -1 $output_file | head -n1 | cut -d/ -f1)
rm -rf $base_dir && mv -f "$extracted_dir" $base_dir
rm -f $output_file
echo "Extracted to $base_dir"

