import os
import pandas as pd
import numpy as np

# Read the notice.log file into a DataFrame
conn_log_file = 'concat-conn.log'
columns=['ts', 'uid', 'id.orig_h', 'id.orig_p', 'id.resp_h', 'id.resp_p']
conn_log_df = pd.read_csv(conn_log_file, sep='\t', usecols=columns)
conn_log_df.replace('-', np.nan, inplace=True)
conn_log_clean = conn_log_df.dropna(axis=1, how='all')

# Manually copied
iplist = ['163.47.36.34',
          '103.98.236.58',
          '185.73.23.133',
          '169.228.66.212',
          '195.37.190.88',
          '145.90.8.10',
          '58.205.215.76',
          '218.197.36.118',
          '141.22.28.227',
          '171.67.71.209',
          '138.246.253.24',
          '219.243.212.85',
          '82.179.36.254',
          '163.47.36.33',
          '203.91.121.252',
          '134.75.30.73',
          '147.52.44.55',
          '140.116.96.199',
          '128.91.61.8']

ipnames = {'163.47.36.34': 'Bangladesh Research and Education Network',
           '103.98.236.58': "The Council of the King's School",
           '185.73.23.133': 'RUB LIR INET',
           '169.228.66.212': 'University of California San Diego',
           '195.37.190.88': 'Landwirtschaftsschule Passau',
           '145.90.8.10': 'Universiteit Twente',
           '58.205.215.76': 'China Education and Research Network',
           '218.197.36.118': 'China Education and Research Network',
           '141.22.28.227': 'Hochschule fuer angewandte Wissenschaften',
           '171.67.71.209': 'Stanford University',
           '138.246.253.24': 'Ludwig-Maximilians-Universitaet Muenchen',
           '219.243.212.85': 'China Education and Research Network',
           '82.179.36.254': 'Russian Academy of Sciences',
           '163.47.36.33': 'Bangladesh Research and Education Network',
           '203.91.121.252': 'Network Technology Experiment Validation and Demonstration Center',
           '134.75.30.73': 'KISTI',
           '147.52.44.55': 'University of Crete',
           '140.116.96.199': 'Taiwan Academic Network',
           '128.91.61.8': 'University of Pennsylvania'}

targetips = len(conn_log_clean['id.resp_h'].unique())
print("targetips: ", targetips)

total_hours = 48
ip_len = {}
ip_targets = {}
aph_sum = 0

for ip in iplist:
    print(f"Creating data entry for ip: {ip}")
    data = conn_log_clean[conn_log_clean['id.orig_h'] == ip]
    ip_len[ip] = len(data)
    ip_targets[ip] = len(data['id.resp_h'].unique())
    print(f"It is {len(data)} long and goes to {ip_targets[ip]} ips")

for ip in iplist:
    if ip_targets[ip] != 0:
        aph = round((ip_len[ip] / total_hours) / ip_targets[ip], 2)
        aph_sum += aph
        print(f"{ipnames[ip]} & {aph} \\\ ")

print(aph_sum)
