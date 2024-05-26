# Authenticate
#source /home/limbo/.limbo_creds
#eval $(swift auth)

# get all available caps
#swift list telescope-ucsdnt-pcap-live > caps.txt

# filter caps
filter="$1"
filename="$2"
#while IFS= read -r p; do
while read p; do
  echo "Working on file: $p"
  outfile=$(echo "$p" | sed 's/[\/=]/-/g')
  foldername=$(echo "fol-${outfile%%.[^.]*}")
  mkdir ${foldername}
  cd ${foldername}
  tracesplit -m 1 --filter="${filter}" -c 1000000 pcapfile:swift://telescope-ucsdnt-pcap-live/${p} pcapfile:out-${outfile}
  cd ..
  sudo docker run -i --rm -v ./cfg:/cfg -v ./${foldername}:/mnt zeek/zeek:latest zeek -C -r /mnt/out-${outfile} /cfg/cfg.zeek Log::default_logdir=/mnt
  mv ${foldername}/conn.log ${foldername}-conn.log
  mv ${foldername}/notice.log ${foldername}-notice.log
  rm ${foldername}/out-${outfile}
  rmdir ${foldername}
  echo "Finished working on file: $p"

done < "$filename"

echo "Finished with file $filename"
