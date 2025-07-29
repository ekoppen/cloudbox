#!/bin/bash

# Test GitHub webhook to see if notifications are created
echo "Testing GitHub webhook..."

curl -X POST "https://cloudbox.doorkoppen.nl/api/v1/deploy/webhook/2" \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -d '{
    "ref": "refs/heads/main",
    "head_commit": {
      "id": "abc123456789",
      "message": "Test commit for webhook notification debugging",
      "author": {
        "name": "Test User",
        "email": "test@example.com"
      }
    },
    "repository": {
      "name": "test-repo",
      "full_name": "user/test-repo"
    }
  }' \
  -v

echo -e "\n\nChecking if notification was created..."

# Test messaging endpoints to see if notification appears
curl -s "https://cloudbox.doorkoppen.nl/api/v1/projects/1/messaging/messages" \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" | head -10