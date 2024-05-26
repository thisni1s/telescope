#!/bin/bash
sudo docker run -it --rm -v ./code:/code -v /home/limbo/filter-caps/logs/conn-logs:/conn-logs  py bash
