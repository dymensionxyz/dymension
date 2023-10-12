#!/bin/bash

for i in {1..1000}
do
   dymd q ibc channel connections connection-1008 --limit 1000000 &
done
wait

