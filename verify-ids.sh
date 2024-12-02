#!/bin/bash

# KYC Allow List Manager
# Manages a list of trusted IDs for KYC verification stored in .app.env file.
#
# Features:
# - Add IDs to the allow list
# - Remove IDs from the allow list
# - Display current IDs
# - Fix formatting issues (spaces, multiple commas)
# - Validate ID format (48 alphanumeric characters)
#
# File Format:
# The script manages the VERIFICATION_ALWAYS_VERIFIED_IDS key in .app.env file.
# Example: VERIFICATION_ALWAYS_VERIFIED_IDS=ID1,ID2,ID3
#
# Usage Examples:
# 1. Display current IDs:
#    ./kyc-allow-list.sh
#
# 2. Add one or more IDs:
#    ./kyc-allow-list.sh 5CnfXxJ2d2vRwCvpXuGykfPtSbSXru4EFPNMRWyW2h4wCKQd
#    ./kyc-allow-list.sh ID1 ID2 ID3
#
# 3. Remove one or more IDs:
#    ./kyc-allow-list.sh --remove 5CnfXxJ2d2vRwCvpXuGykfPtSbSXru4EFPNMRWyW2h4wCKQd
#    ./kyc-allow-list.sh --remove ID1 ID2
#
# 4. Fix formatting issues:
#    ./kyc-allow-list.sh --fix
#
# 5. Check ID validity:
#    ./kyc-allow-list.sh --check
#
# 6. Combine operations:
#    ./kyc-allow-list.sh --fix ID1 ID2  # Fix format and add IDs
#
# Validation:
# - IDs must be exactly 48 alphanumeric characters
# - No spaces or empty values allowed in the list
# - Duplicate IDs are not allowed
#
# Exit Codes:
# 0 - Success
# 1 - Error (invalid format, missing file, etc.)

ENV_FILE=".app.env"
FIX_SPACES=false
CHECK_ONLY=false
REMOVE_MODE=false
NEW_IDS=()

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --fix)
            FIX_SPACES=true
            shift
            ;;
        --check)
            CHECK_ONLY=true
            shift
            ;;
        --remove)
            REMOVE_MODE=true
            shift
            ;;
        *)
            NEW_IDS+=("$1")
            shift
            ;;
    esac
done

# Function to clean IDs string
clean_ids() {
    local ids=$1
    # Remove spaces and multiple commas, then trim leading/trailing commas
    cleaned=$(echo "$ids" | tr -d '[:space:]' | tr -s ',' | sed 's/^,//;s/,$//')
    echo "$cleaned"
}

# Function to count IDs
count_ids() {
    local ids=$1
    if [ -z "$ids" ]; then
        echo 0
        return
    fi
    IFS=',' read -ra ID_ARRAY <<< "$ids"
    echo "${#ID_ARRAY[@]}"
}

# Function to validate ID
validate_id() {
    local id=$1
    if [[ ! "$id" =~ ^[a-zA-Z0-9]{48}$ ]]; then
        echo "Error: ID '$id' is invalid. It must be exactly 48 alphanumeric characters."
        exit 1
    fi
}

# Function to check IDs
check_ids() {
    local ids=$1
    if [ -z "$ids" ]; then
        echo "No IDs present"
        return
    fi

    # Check for spaces or multiple commas in the line
    if [[ "$ids" =~ [[:space:]] ]] || [[ "$ids" =~ ,, ]]; then
        echo "Error: Invalid format - line contains spaces or empty values (multiple commas)"
        exit 1
    fi

    # Convert comma-separated string to array and validate each ID
    IFS=',' read -ra ID_ARRAY <<< "$ids"
    for id in "${ID_ARRAY[@]}"; do
        validate_id "$id"
    done
    echo "All IDs are valid."
}

# Function to display IDs
display_ids() {
    local ids=$1
    if [ -z "$ids" ]; then
        echo "No IDs present"
        return
    fi

    # Convert comma-separated string to array
    IFS=',' read -ra ID_ARRAY <<< "$ids"
    echo "Total IDs: ${#ID_ARRAY[@]}"
    echo "IDs:"
    printf '%s\n' "${ID_ARRAY[@]}"
}

# Function to remove ID from comma-separated list
remove_id() {
    local ids=$1
    local id_to_remove=$2
    
    # Convert to array
    IFS=',' read -ra ID_ARRAY <<< "$ids"
    local new_ids=()
    local found=false
    
    # Rebuild array without the removed ID
    for id in "${ID_ARRAY[@]}"; do
        if [ "$id" != "$id_to_remove" ]; then
            new_ids+=("$id")
        else
            found=true
        fi
    done
    
    # Join array back to string
    local result=$(IFS=,; echo "${new_ids[*]}")
    
    if $found; then
        echo "$result"
        return 0
    else
        echo "$ids"
        return 1
    fi
}

# Check if the file exists
if [ ! -f "$ENV_FILE" ]; then
    echo "Error: $ENV_FILE does not exist."
    exit 1
fi

# If --check is used, validate all IDs
if $CHECK_ONLY; then
    if grep -q "^VERIFICATION_ALWAYS_VERIFIED_IDS=" "$ENV_FILE"; then
        current_ids=$(grep "^VERIFICATION_ALWAYS_VERIFIED_IDS=" "$ENV_FILE" | cut -d '=' -f2)
        check_ids "$current_ids"
    else
        echo "VERIFICATION_ALWAYS_VERIFIED_IDS key is missing in $ENV_FILE."
    fi
    exit 0
fi

# If --fix is used, clean the IDs
if $FIX_SPACES; then
    if grep -q "^VERIFICATION_ALWAYS_VERIFIED_IDS=" "$ENV_FILE"; then
        current_ids=$(grep "^VERIFICATION_ALWAYS_VERIFIED_IDS=" "$ENV_FILE" | cut -d '=' -f2)
        cleaned_ids=$(clean_ids "$current_ids")
        
        # Only update and show message if there were changes
        if [ "$current_ids" != "$cleaned_ids" ]; then
            sed -i "s|^VERIFICATION_ALWAYS_VERIFIED_IDS=.*|VERIFICATION_ALWAYS_VERIFIED_IDS=$cleaned_ids|" "$ENV_FILE"
            echo "Fixed format issues in the file."
        fi
        
        current_ids="$cleaned_ids"
    else
        echo "VERIFICATION_ALWAYS_VERIFIED_IDS key is missing in $ENV_FILE."
    fi
    # Exit if no new IDs to add
    if [ ${#NEW_IDS[@]} -eq 0 ]; then
        exit 0
    fi
fi

# If no arguments provided, display current IDs
if [ ${#NEW_IDS[@]} -eq 0 ]; then
    if grep -q "^VERIFICATION_ALWAYS_VERIFIED_IDS=" "$ENV_FILE"; then
        current_ids=$(grep "^VERIFICATION_ALWAYS_VERIFIED_IDS=" "$ENV_FILE" | cut -d '=' -f2)
        display_ids "$current_ids"
    else
        echo "VERIFICATION_ALWAYS_VERIFIED_IDS key is missing in $ENV_FILE."
    fi
    exit 0
fi

# Process IDs (add or remove)
if grep -q "^VERIFICATION_ALWAYS_VERIFIED_IDS=" "$ENV_FILE"; then
    current_ids=$(grep "^VERIFICATION_ALWAYS_VERIFIED_IDS=" "$ENV_FILE" | cut -d '=' -f2)
    initial_count=$(count_ids "$current_ids")

    if $REMOVE_MODE; then
        # Remove IDs
        removed_count=0
        for id in "${NEW_IDS[@]}"; do
            validate_id "$id"
            new_list=$(remove_id "$current_ids" "$id")
            if [ "$new_list" != "$current_ids" ]; then
                current_ids="$new_list"
                removed_count=$((removed_count + 1))
                echo "Removed ID: $id"
            else
                echo "ID $id not found in the list."
            fi
        done

        # Update file if any IDs were removed
        if [ $removed_count -gt 0 ]; then
            sed -i "s|^VERIFICATION_ALWAYS_VERIFIED_IDS=.*|VERIFICATION_ALWAYS_VERIFIED_IDS=$current_ids|" "$ENV_FILE"
            final_count=$((initial_count - removed_count))
            echo "IDs count changed from $initial_count to $final_count"
        fi
    else
        # Add IDs
        added_count=0
        for new_id in "${NEW_IDS[@]}"; do
            validate_id "$new_id"
            if [ -z "$current_ids" ]; then
                current_ids="$new_id"
                added_count=$((added_count + 1))
                echo "Added ID: $new_id"
            else
                if [[ ",$current_ids," == *",$new_id,"* ]]; then
                    echo "ID $new_id is already present."
                else
                    current_ids="$current_ids,$new_id"
                    added_count=$((added_count + 1))
                    echo "Added ID: $new_id"
                fi
            fi
        done

        # Update file if any IDs were added
        if [ $added_count -gt 0 ]; then
            sed -i "s|^VERIFICATION_ALWAYS_VERIFIED_IDS=.*|VERIFICATION_ALWAYS_VERIFIED_IDS=$current_ids|" "$ENV_FILE"
            final_count=$((initial_count + added_count))
            echo "IDs count changed from $initial_count to $final_count"
        fi
    fi
else
    echo "VERIFICATION_ALWAYS_VERIFIED_IDS key is missing in $ENV_FILE."
fi