import pickle
import os
import pandas as pd
import numpy as np

# Read the notice.log file into a DataFrame
scan_log_file = 'concat-notice.log'
scan_log_df = pd.read_csv(scan_log_file, sep='\t')
scan_log_df.replace('-', np.nan, inplace=True)
scan_log_clean = scan_log_df.dropna(axis=1, how='all')

# Manually copied
#ips_to_check = ["221.229.215.71", "164.90.174.244", "63.251.238.12", "79.110.62.185", "185.242.226.39", "188.166.248.56", "45.14.226.132", "185.242.226.3", "185.242.226.5", "104.156.155.10", "91.218.114.197", "188.166.247.242", "185.242.226.40", "104.156.155.8",
#                "141.105.67.7", "104.156.155.11", "158.255.7.153", "104.156.155.37", "165.154.230.251", "185.242.226.6", "165.154.240.27", "165.154.225.190", "185.242.226.47", "45.14.226.152", "92.118.39.34", "39.152.141.101", "91.92.245.242", "185.242.226.2"]

ips_to_check = ['221.229.215.71', '63.251.238.12', '79.110.62.185', '185.242.226.39', '45.14.226.132', '185.242.226.3', '185.242.226.5', '91.218.114.197', '104.156.155.10', '185.242.226.40', '141.105.67.7', '104.156.155.8', '104.156.155.11', '158.255.7.153', '104.156.155.37', '165.154.230.251', '165.154.240.27', '185.242.226.6', '185.242.226.47', '165.154.225.190', '45.14.226.152', '92.118.39.34', '39.152.141.101', '91.92.245.242', '185.242.226.2']


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


folder_path = '/conn-logs'
# Port scan dict: Key: IP, Value: List(scan_count, series of value counts)
# Example: "192.168.10.5": [42, [80, 443, 22]]
port_scan_dict = {}
conn_states = pd.Series(dtype='int64')

file_num = len(os.listdir(folder_path))
print(f"Working on log files: [0/{file_num}]", end='')
counter = 0

# Iterate over each file in the folder
for file_name in os.listdir(folder_path):
    file_path = os.path.join(folder_path, file_name)
    counter = counter+1
    print(
        f"Working on log files: [{counter}/{file_num}] ({file_name})", end='\r', flush=True)

    # Check if the current file is a regular file
    if os.path.isfile(file_path):
        conn_log_df = pd.read_csv(file_path, sep='\s+')
        conn_log_df['ts'] = pd.to_datetime(conn_log_df['ts'], unit='s')

        connection_states = conn_log_df['conn_state'].value_counts()
        conn_states += connection_states

        for ip in common_ips:
            ip_frame = conn_log_df[conn_log_df['id.orig_h'] == ip]
            count = ip_frame['id.resp_h'].value_counts().count()
            scanned_ports = ip_frame['id.resp_p'].value_counts()

            # Check if key exists
            if ip in port_scan_dict:
                current_value = port_scan_dict[ip]
                current_value[0] = current_value[0] + count
                current_value[1] = current_value[1] + scanned_ports
                port_scan_dict[ip] = current_value
            else:
                port_scan_dict[ip] = [count, scanned_ports]

print()


# File path to save the dictionary
file_path = 'dictionary.pkl'

# Save the dictionary to a Pickle file
with open(file_path, 'wb') as file:
    pickle.dump(port_scan_dict, file)

print(f"Dictionary saved to {file_path}")

for ip in common_ips:
    print(f"{ip} scanned {port_scan_dict[ip][0]}")
    print(port_scan_dict[ip][1].nlargest(5))


print(conn_states)
