#!/usr/bin/env bash

# Exit immediately if a command exits with a non-zero status, and 
# treat unset variables as an error. Exit status of a pipe is the 
# exit status of the last command that was not 0.
set -eo pipefail

# --- CONFIGURATION ---
PROTO_DIR="./proto"
BUF_TEMPLATE="buf.gen.gogo.yaml"
# --- END CONFIGURATION ---

echo "Generating gogo proto code using buf..."

# Find all proto files recursively starting from the configured directory.
# We only process files that explicitly include 'go_package' to ensure 
# they are intended for Go code generation.

# Use a subshell to execute the change directory command and ensure 
# the original working directory is restored regardless of success or failure.
(
    cd $PROTO_DIR || exit 1 # Exit if proto directory is not found

    # Find all *.proto files in the current directory and subdirectories.
    find . -type f -name '*.proto' -print0 | while IFS= read -r -d $'\0' file; do
        # Check if the file contains the 'go_package' directive.
        if grep -q "go_package" "$file"; then
            echo "Processing $file..."
            
            # Use buf generate on the specific file, which should ideally
            # write the output directly to the correct destination (e.g., ./<module>/<package>/).
            # If the buf template is correctly configured, no manual file copying is needed.
            buf generate --template "$BUF_TEMPLATE" "$file"
        fi
    done
)

# NOTE ON MANUAL COPYING:
# The original script's copy/cleanup steps are usually unnecessary if 
# the 'buf.gen.gogo.yaml' is correctly configured with 'M<proto_path>=<go_module>/v<version>/...'.
# If manual copying is unavoidable (due to external module definition), 
# the cleanup step should target only the generated temporary directory.
#
# if [ -d "github.com" ]; then
#     echo "Moving generated files and cleaning up temp directory."
#     # Assuming 'github.com/dymensionxyz/dymension/v*/*' contains all generated files
#     cp -r github.com/dymensionxyz/dymension/v*/* ./
#     rm -rf github.com
# fi


# TODO: Uncomment once ORM/Pulsar support is needed.
# Ref: https://github.com/osmosis-labs/osmosis/pull/1589
# The TODO comment remains for future reference.

echo "Proto code generation completed successfully."
