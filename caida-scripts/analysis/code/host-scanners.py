import pandas as pd
import numpy as np

# Read the notice.log file into a DataFrame
scan_log_file = 'concat-host-scan-notice.log'
scan_log_df = pd.read_csv(scan_log_file, sep='\t')
scan_log_df.replace('-', np.nan, inplace=True)
scan_log_clean = scan_log_df.dropna(axis=1, how='all')

# Manually copied
ips_to_check = ['185.224.128.43', '79.124.49.86', '45.227.253.130', '85.209.11.132', '79.124.60.6', '87.121.69.52', '89.248.163.200', '91.148.190.130', '45.135.232.96', '78.128.113.250', '185.224.128.17', '79.124.40.70', '94.232.41.124', '210.245.120.108', '79.124.59.202', '176.111.174.29', '194.26.135.215', '194.26.135.250', '185.169.4.105', '80.94.95.249', '79.124.62.78', '80.75.212.75', '62.204.41.128', '78.128.114.66', '62.204.41.122',
                '78.128.114.178', '91.191.209.38', '190.211.255.106', '78.128.114.90', '78.128.114.186', '78.128.114.182', '78.128.114.190', '45.128.232.84', '89.248.163.42', '194.26.135.64', '194.26.135.61', '91.191.209.26', '185.161.248.120', '194.26.135.154', '62.122.184.82', '89.248.163.41', '78.128.113.34', '109.205.213.113', '89.248.163.18', '91.148.190.146', '89.248.163.62', '79.124.58.158', '79.124.60.246', '79.124.60.194', '91.148.190.134']

print(len(ips_to_check))

# Get all logs for ips marked as port scanners
port_scan_logs = scan_log_clean[scan_log_clean['note'] == 'Scan::Address_Scan']

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

print(ips_only_in_mine)

# folder_path = '/conn-logs'
# Port scan dict: Key: IP, Value: List(scan_count, series of value counts)
# Example: "192.168.10.5": [42, [80, 443, 22]]
# port_scan_dict = {}
# conn_states = pd.Series(dtype='int64')

# file_num = len(os.listdir(folder_path))
# print(f"Working on log files: [0/{file_num}]", end='')
# counter = 0

# Iterate over each file in the folder
# for file_name in os.listdir(folder_path):
#    file_path = os.path.join(folder_path, file_name)
#    counter = counter+1
#    print(
#        f"Working on log files: [{counter}/{file_num}] ({file_name})", end='\r', flush=True)
#
#    # Check if the current file is a regular file
#    if os.path.isfile(file_path):
#        conn_log_df = pd.read_csv(file_path, sep='\s+')
#        conn_log_df['ts'] = pd.to_datetime(conn_log_df['ts'], unit='s')
#
#        connection_states = conn_log_df['conn_state'].value_counts()
#        conn_states += connection_states
#
#        for ip in common_ips:
#            ip_frame = conn_log_df[conn_log_df['id.orig_h'] == ip]
#            count = ip_frame['id.resp_h'].value_counts().count()
#           scanned_ports = ip_frame['id.resp_p'].value_counts()
#
#            # Check if key exists
#            if ip in port_scan_dict:
#               current_value = port_scan_dict[ip]
#               current_value[0] = current_value[0] + count
#               current_value[1] = current_value[1] + scanned_ports
#              port_scan_dict[ip] = current_value
#            else:
#                port_scan_dict[ip] = [count, scanned_ports]
#
# print()


# File path to save the dictionary
# file_path = 'dictionary.pkl'

# Save the dictionary to a Pickle file
# with open(file_path, 'wb') as file:
#    pickle.dump(port_scan_dict, file)

# print(f"Dictionary saved to {file_path}")
#
# for ip in common_ips:
#    print(f"{ip} scanned {port_scan_dict[ip][0]}")
#    print(port_scan_dict[ip][1].nlargest(5))


# print(conn_states)
