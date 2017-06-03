#!/usr/bin/env bash

make local && MGO_DB=zanecloud  MGO_URLS=127.0.0.1  ./apiserver -l debug start