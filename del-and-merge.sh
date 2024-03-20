#!/bin/bash

# Loop through each folder
for folder in */; do
    # Go into the folder
    cd "$folder" || continue
    
    # Check if there are .pcap files
    pcap_files=$(ls *.pcap 2>/dev/null)
    
    # If there are .pcap files, merge them
    if [ -n "$pcap_files" ]; then
        echo "Merging .pcap files in $folder"
        mergecap -w merged.pcap.keep *.pcap
        echo "Merged .pcap files saved as merged.pcap.keep in $folder"
        
        # Delete original .pcap files
        echo "Deleting original .pcap files in $folder"
        rm *.pcap
        mv merged.pcap.keep merged.pcap
    else
        echo "No .pcap files found in $folder"
    fi
    
    # Go back to the parent directory
    cd ..
done

