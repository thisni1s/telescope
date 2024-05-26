import pandas as pd
import numpy as np

# Read the notice.log file into a DataFrame
scan_log_file = 'concat-notice.log'
scan_log_df = pd.read_csv(scan_log_file, sep='\t')
scan_log_df.replace('-', np.nan, inplace=True)
scan_log_clean = scan_log_df.dropna(axis=1, how='all')

# Manually copied
ips_to_check = ["221.229.215.71", "164.90.174.244", "63.251.238.12", "79.110.62.185", "185.242.226.39", "188.166.248.56", "45.14.226.132", "185.242.226.3", "185.242.226.5", "104.156.155.10", "91.218.114.197", "188.166.247.242", "185.242.226.40", "104.156.155.8", "141.105.67.7", "104.156.155.11", "158.255.7.153", "104.156.155.37", "165.154.230.251", "185.242.226.6", "165.154.240.27", "165.154.225.190", "185.242.226.47", "45.14.226.152", "92.118.39.34", "39.152.141.101", "91.92.245.242", "185.242.226.2"]


# Get all logs for ips marked as port scanners
port_scan_logs = scan_log_clean[scan_log_clean['note'] == 'Scan::Port_Scan']

# Extract flagged IPs
flagged_ips = set(port_scan_logs['src'])
ips_to_check_set = set(ips_to_check)

# Find IPs that are common in both sets
common_ips = flagged_ips.intersection(ips_to_check_set)

# Find IPs only in one of the sets
ips_only_in_caida = flagged_ips - ips_to_check_set
ips_only_in_mine = ips_to_check_set - flagged_ips

# Print common IPs
print(f"Common IPs: {len(common_ips)}")
for ip in common_ips:
    print(ip)

# Print IPs only in flagged set
print(f"\nIPs only in caida set (should be none): {len(ips_only_in_caida)}")
for ip in ips_only_in_caida:
    print(ip)

# Print IPs only in check set
print(f"\nIPs only in my set: {len(ips_only_in_mine)}")
for ip in ips_only_in_mine:
    print(ip)

