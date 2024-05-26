# Authenticate
source /home/limbo/.limbo_creds
eval $(swift auth)

# Switch to working dir
cd /home/limbo/filter-caps

# get all available caps
#swift list telescope-ucsdnt-pcap-live > caps.txt
filter="$1"

# move old log files
mv *.log logs/
rm cap_*

# build the caps file
head -n 8 caps_all.txt > caps.txt
mv caps_all.txt caps_all.txt.old
tail -n +9 caps_all.txt.old > caps_all.txt

split --number=l/8 --additional-suffix=.txt --numeric-suffixes caps.txt cap_

for i in {0..7}; do
    filename="cap_0${i}.txt"
    echo "Starting worker for file: ${filename}"
    bash worker.sh "${filter}" ${filename} &
done

wait

echo "Done!"

