#!/bin/bash

# Test script for the new public file access system

CLOUDBOX_URL="http://localhost:8080"
PROJECT_ID="2"
PROJECT_SLUG="dsqewdq"  # Replace with actual project slug
BUCKET_NAME="images"    # Test bucket
JWT_TOKEN="YOUR_JWT_TOKEN_HERE"  # Replace with actual JWT token
API_KEY="YOUR_API_KEY_HERE"      # Replace with actual API key

echo "🧪 Testing CloudBox Public File Access System"
echo "=============================================="
echo ""

# Test 1: Make bucket public via admin interface
echo "1️⃣  Making bucket public via admin interface..."
echo "URL: $CLOUDBOX_URL/api/v1/projects/$PROJECT_ID/storage/buckets/$BUCKET_NAME/visibility"
curl -s -X PUT \
     -H "Authorization: Bearer $JWT_TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"is_public": true}' \
     "$CLOUDBOX_URL/api/v1/projects/$PROJECT_ID/storage/buckets/$BUCKET_NAME/visibility" | head -c 200
echo ""
echo ""

# Test 2: List public buckets
echo "2️⃣  Listing public buckets..."
echo "URL: $CLOUDBOX_URL/api/v1/projects/$PROJECT_ID/storage/public-buckets"
curl -s -H "Authorization: Bearer $JWT_TOKEN" \
     "$CLOUDBOX_URL/api/v1/projects/$PROJECT_ID/storage/public-buckets" | head -c 300
echo ""
echo ""

# Test 3: Get public URL for a file (admin)
echo "3️⃣  Getting public URL for file via admin..."
echo "URL: $CLOUDBOX_URL/api/v1/projects/$PROJECT_ID/storage/buckets/$BUCKET_NAME/files/FILE_ID/public-url"
echo "(Replace FILE_ID with actual file ID)"
echo ""

# Test 4: Get public URL for file via project API
echo "4️⃣  Getting public URL for file via project API..."
echo "URL: $CLOUDBOX_URL/p/$PROJECT_SLUG/api/storage/$BUCKET_NAME/files/FILE_ID/public-url"
curl -s -H "X-API-Key: $API_KEY" \
     "$CLOUDBOX_URL/p/$PROJECT_SLUG/api/storage/$BUCKET_NAME/files/EXAMPLE_FILE_ID/public-url" | head -c 200
echo ""
echo ""

# Test 5: Batch public URLs
echo "5️⃣  Getting batch public URLs..."
echo "URL: $CLOUDBOX_URL/p/$PROJECT_SLUG/api/storage/$BUCKET_NAME/files/batch-public-urls"
curl -s -X POST \
     -H "X-API-Key: $API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"file_ids": ["file1", "file2"]}' \
     "$CLOUDBOX_URL/p/$PROJECT_SLUG/api/storage/$BUCKET_NAME/files/batch-public-urls" | head -c 300
echo ""
echo ""

# Test 6: Access public file (no authentication)
echo "6️⃣  Accessing public file directly..."
echo "URL: $CLOUDBOX_URL/public/$PROJECT_SLUG/$BUCKET_NAME/example-image.jpg"
curl -s -I "$CLOUDBOX_URL/public/$PROJECT_SLUG/$BUCKET_NAME/example-image.jpg" | head -10
echo ""

echo "🏁 Public file access testing complete!"
echo ""
echo "📝 Expected results:"
echo "- Test 1: Bucket marked as public (200 OK with JSON response)"  
echo "- Test 2: List of public buckets (200 OK with JSON array)"
echo "- Test 3-5: Public URLs generated (200 OK with URLs in response)"
echo "- Test 6: Public file served OR 404 if file doesn't exist"
echo ""
echo "🔧 Setup needed:"
echo "1. Replace PROJECT_SLUG with actual project slug"
echo "2. Replace JWT_TOKEN with valid admin JWT token"  
echo "3. Replace API_KEY with valid project API key"
echo "4. Upload some test files to the bucket first"
echo "5. Use actual file IDs in tests 3-5"
echo ""
echo "⚠️  Security notes:"
echo "- Only files in PUBLIC buckets are accessible"
echo "- Project must be active for public access"
echo "- Files are served with proper caching headers"
echo "- Path traversal attacks are prevented"