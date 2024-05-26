#!/bin/sh
#
output_file="concat-conn.log"

# Remove the output file if it already exists
[ -f "$output_file" ] && rm "$output_file"

head -n 1 fol-datasource-ucsd-nt-year-2024-month-05-day-10-hour-04-ucsd-nt-conn.log >> "$output_file"

# Loop through all CSV files in the current directory
for file in *conn.log; do
    echo "Working on $file"
    # Skip the loop iteration if the file is the output file itself
    [ "$file" = "$output_file" ] && continue

    # Skip the loop iteration if the file is empty
    [ ! -s "$file" ] && continue

    # Append the contents of the file without the header to the output file
    tail -n +2 "$file" >> "$output_file"
done

echo "Concatenation complete. Output saved to $output_file"

