#!/bin/bash

# Script to fix ALL hardcoded localhost URLs in CloudBox frontend

echo "🔧 Fixing ALL hardcoded localhost URLs in CloudBox frontend..."

cd frontend/src

# Find all .svelte files with localhost:8080
echo "📋 Files to update:"
find . -name "*.svelte" -exec grep -l "localhost:8080" {} \; | wc -l | xargs echo "Found files:"

# Function to add API config import if not present
add_api_import() {
    local file="$1"
    if ! grep -q "API_BASE_URL.*config" "$file"; then
        echo "  Adding API import to ${file#./}"
        # Add import after existing imports in script tag
        if grep -q "<script" "$file"; then
            # Find the last import line and add our import after it
            sed -i '' '/import.*from.*$/a\
  import { API_BASE_URL, createApiRequest } from '\''$lib/config'\'';
' "$file"
        fi
    fi
}

# Function to replace localhost URLs
fix_localhost_urls() {
    local file="$1"
    echo "  Fixing URLs in ${file#./}"
    
    # Add the import
    add_api_import "$file"
    
    # Replace hardcoded localhost URLs with API_BASE_URL
    sed -i '' 's|http://localhost:8080|\${API_BASE_URL}|g' "$file"
    
    # Fix fetch calls to use template literals
    sed -i '' 's|fetch('\''$|fetch(`|g' "$file"
    sed -i '' 's|'\''|`|g' "$file" 
    
    # Replace common fetch patterns with createApiRequest
    # This is more complex and needs careful replacement
    echo "    URLs replaced in $file"
}

# Counter
count=0

# Process each file
for file in $(find . -name "*.svelte" -exec grep -l "localhost:8080" {} \;); do
    echo "🔍 Processing: ${file#./}"
    
    # Create backup
    cp "$file" "$file.backup"
    
    # Simple replacement - just change localhost:8080 to use template literal
    add_api_import "$file"
    sed -i '' 's|http://localhost:8080|${API_BASE_URL}|g' "$file"
    
    count=$((count + 1))
done

echo
echo "✅ Processing complete!"
echo "📊 Total files processed: $count"
echo "🔧 Backup files created with .backup extension"
echo
echo "Next steps:"
echo "1. Check that the replacements look correct"
echo "2. Test the frontend"
echo "3. If everything works: find . -name '*.backup' -delete"
echo "4. Commit the changes to git"