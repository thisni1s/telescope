# Cloud-based Internet Telescope

This is all the code needed for my masters thesis on cloud-based internet telescopes.

It is divided into 4 different directories:

- caida-scripts: contains the scripts and code needed for the filtering and analysis of the CAIDA data. This all had to be performed on their machines because a download of the data is not allowed.
- notebooks: contains all the jupyter notebooks used for the data analysis. All the output had to be cleared, because they contain large amounts of caida data, which is not allowed to be uplaoded to the public.
- telescope: contains all the Go code and scripts needed for deploying the internet telescope
- misc: contains other shell scripts and files used for this work
