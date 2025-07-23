#!/bin/bash

# Script to replace all hardcoded localhost:8080 URLs with dynamic API config

echo "🔧 Fixing hardcoded API URLs in CloudBox frontend..."

# Directory to search
FRONTEND_DIR="/Users/eelko/Documents/_dev/_WEBSITES/cloudbox/frontend/src"

# Find all .svelte and .ts files (excluding config.ts)
FILES=$(find "$FRONTEND_DIR" -name "*.svelte" -o -name "*.ts" | grep -v config.ts)

echo "📝 Files to process:"
echo "$FILES" | wc -l | xargs echo "Found files:"

# Counter for replacements
TOTAL_REPLACEMENTS=0

# Process each file
for FILE in $FILES; do
    if grep -q "localhost:8080" "$FILE"; then
        echo "🔍 Processing: ${FILE#$FRONTEND_DIR/}"
        
        # Count occurrences before replacement
        OCCURRENCES=$(grep -c "localhost:8080" "$FILE")
        
        # Create backup
        cp "$FILE" "$FILE.backup"
        
        # Add import if not present and file has localhost:8080
        if ! grep -q "import.*API_ENDPOINTS.*config" "$FILE"; then
            # Check if it's a .svelte file and has script tag
            if [[ "$FILE" == *.svelte ]] && grep -q "<script" "$FILE"; then
                # Add import after existing imports in script tag
                sed -i '' '/import.*from/a\
  import { API_ENDPOINTS } from '\''$lib/config'\'';
' "$FILE"
            elif [[ "$FILE" == *.ts ]]; then
                # Add import at top of .ts file
                sed -i '' '1i\
import { API_ENDPOINTS } from '\''$lib/config'\'';
' "$FILE"
            fi
        fi
        
        # Replace specific API endpoints
        sed -i '' "s|'http://localhost:8080/api/v1/auth/login'|API_ENDPOINTS.auth.login|g" "$FILE"
        sed -i '' "s|'http://localhost:8080/api/v1/auth/logout'|API_ENDPOINTS.auth.logout|g" "$FILE"
        sed -i '' "s|'http://localhost:8080/api/v1/auth/refresh'|API_ENDPOINTS.auth.refresh|g" "$FILE"
        sed -i '' "s|'http://localhost:8080/api/v1/auth/me'|API_ENDPOINTS.auth.me|g" "$FILE"
        
        # Replace admin endpoints
        sed -i '' "s|'http://localhost:8080/api/v1/admin/stats/overview'|API_ENDPOINTS.admin.stats.overview|g" "$FILE"
        sed -i '' "s|'http://localhost:8080/api/v1/admin/stats/user-growth'|API_ENDPOINTS.admin.stats.userGrowth|g" "$FILE"
        sed -i '' "s|'http://localhost:8080/api/v1/admin/stats/project-activity'|API_ENDPOINTS.admin.stats.projectActivity|g" "$FILE"
        sed -i '' "s|'http://localhost:8080/api/v1/admin/stats/system-health'|API_ENDPOINTS.admin.stats.systemHealth|g" "$FILE"
        sed -i '' "s|'http://localhost:8080/api/v1/admin/users'|API_ENDPOINTS.admin.users.list|g" "$FILE"
        sed -i '' "s|'http://localhost:8080/api/v1/admin/projects'|API_ENDPOINTS.admin.projects.list|g" "$FILE"
        
        # Replace project endpoints
        sed -i '' "s|'http://localhost:8080/api/v1/projects'|API_ENDPOINTS.projects.list|g" "$FILE"
        sed -i '' "s|'http://localhost:8080/api/v1/users'|API_ENDPOINTS.users.list|g" "$FILE"
        
        # Replace generic localhost:8080 with template literals for dynamic URLs
        sed -i '' 's|`http://localhost:8080/api/v1/admin/users/\${[^}]*}`|API_ENDPOINTS.admin.users.get(userId)|g' "$FILE"
        sed -i '' 's|`http://localhost:8080/api/v1/projects/\${[^}]*}`|API_ENDPOINTS.projects.get(projectId)|g' "$FILE"
        
        # Generic replacements for any remaining localhost:8080
        sed -i '' 's|http://localhost:8080|${API_BASE_URL}|g' "$FILE"
        
        TOTAL_REPLACEMENTS=$((TOTAL_REPLACEMENTS + OCCURRENCES))
        echo "   ✅ Replaced $OCCURRENCES occurrences"
    fi
done

echo
echo "✅ Processing complete!"
echo "📊 Total replacements made: $TOTAL_REPLACEMENTS"
echo "🔧 Backup files created with .backup extension"
echo
echo "Next steps:"
echo "1. Test the frontend to ensure all API calls work"
echo "2. If everything works, remove backup files: find $FRONTEND_DIR -name '*.backup' -delete"
echo "3. Commit the changes to git"