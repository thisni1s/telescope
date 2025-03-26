#!/bin/bash

# Directory containing .pcap files
directory="/var/spool/gotrace"
bucket=$(cat /root/config/bucket.txt)
provider=$(cat /root/config/provider.txt)
region=$(cat /root/config/region.txt)
ip=$(cat /root/config/ip4.txt | sed -r 's/\./-/g' )

# Do we need to rewrite the pcaps?
providers=("aws" "gcp" "azure")

# Check if the file content matches any of the strings in the array
for prov in "${providers[@]}"; do
  if [[ "$provider" == "$prov" ]]; then
    /var/scripts/rewrite.sh
    break
  fi
done


# Check if the directory exists
if [ ! -d "$directory" ]; then
    echo "Directory not found: $directory"
    exit 1
fi

cd $directory
for file in *.pcap.gz; do
    timestamp=$(echo "$file" | cut -d '.' -f 1 | cut -d '-' -f 2)
    year=$(date -u -d @"$timestamp" +"%Y")
    month=$(date -u -d @"$timestamp" +"%m")
    day=$(date -u -d @"$timestamp" +"%d")
    hour=$(date -u -d @"$timestamp" +"%H")

    target="tupload/${bucket}/provider=${provider}/region=${region}/ip=${ip}/year=${year}/month=${month}/day=${day}/hour=${hour}"

    if mc cp "${file}" "$target/"; then
        rm "$file" 
    fi
done

