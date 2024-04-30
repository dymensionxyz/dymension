#!/bin/bash

# Target URL
URL="http://hub:36657/status"

# Wait for the /status endpoint to be available and the latest_block_height to be present
while true; do
    # Use curl to fetch the status endpoint and jq to parse the JSON response
    HEIGHT=$(curl -s $URL | jq -r '.result.sync_info.latest_block_height')

    # Check if HEIGHT is a number and greater than 0 (modify this check as needed)
    if [[ "$HEIGHT" =~ ^[0-9]+$ ]] && [ "$HEIGHT" -gt 0 ]; then
        echo "hub is ready with latest_block_height: $HEIGHT"
        break
    else
        echo "Waiting for hub to be ready..."
    fi

    sleep 5 # Wait for 5 seconds before trying again
done

# Proceed with the rest of your script or command to start the service
exec "$@"