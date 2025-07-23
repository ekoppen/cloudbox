# Backup & Restore

CloudBox provides comprehensive backup and restore functionality to protect your project data and enable easy migration.

## Overview

The backup system creates complete snapshots of your project data, including:
- Collections and documents
- Files and storage buckets
- Functions and deployments
- SSH keys and server configurations
- API keys and CORS settings
- Audit logs (last 3 months)
- User data and authentication settings

## Features

- **Complete Data Coverage** - Backs up all project-related data
- **Compressed Archives** - Uses tar.gz format with SHA256 checksums
- **Atomic Operations** - Restore operations are transactional (all-or-nothing)
- **Flexible Restore** - Can restore to original or different project
- **Background Processing** - Backup creation runs asynchronously
- **Status Tracking** - Real-time backup progress monitoring

## Creating Backups

### Via API

```bash
# Create a backup
curl -X POST http://localhost:8080/api/projects/1/backups \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Weekly Backup",
    "description": "Automated weekly backup",
    "type": "manual"
  }'
```

### Via JavaScript SDK

```javascript
// Create a backup
const backup = await cloudbox.backups.create({
  name: 'Pre-deployment backup',
  description: 'Backup before major update',
  type: 'manual'
});

console.log('Backup created:', backup.id);
console.log('Status:', backup.status); // "creating"
```

### Backup Types

- **manual** - User-initiated backup
- **automatic** - System-scheduled backup (future feature)

## Listing Backups

```bash
# List all backups for a project
curl -X GET http://localhost:8080/api/projects/1/backups \
  -H "Authorization: Bearer YOUR_TOKEN"
```

```javascript
// List backups
const backups = await cloudbox.backups.list();
console.log(`Found ${backups.length} backups`);

backups.forEach(backup => {
  console.log(`${backup.name} - ${backup.status} (${backup.size} bytes)`);
});
```

## Monitoring Backup Progress

```javascript
// Get backup status
const backup = await cloudbox.backups.get(backupId);

switch (backup.status) {
  case 'creating':
    console.log('Backup is being created...');
    break;
  case 'completed':
    console.log(`Backup completed: ${backup.size} bytes`);
    console.log(`Checksum: ${backup.checksum}`);
    break;
  case 'failed':
    console.error('Backup failed!');
    break;
}
```

## Restoring from Backup

### Same Project Restore

```bash
# Restore to original project
curl -X POST http://localhost:8080/api/backups/123/restore \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

### Cross-Project Restore

```bash
# Restore to different project
curl -X POST http://localhost:8080/api/backups/123/restore \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "target_project_id": 2
  }'
```

```javascript
// Restore backup to same project
await cloudbox.backups.restore(backupId);

// Restore backup to different project
await cloudbox.backups.restore(backupId, {
  targetProjectId: 2
});
```

## Backup Structure

Each backup contains:

```json
{
  "metadata": {
    "project_id": 1,
    "project_name": "My Project",
    "created_at": "2024-01-15T10:30:00Z",
    "backup_type": "manual",
    "cloudbox_version": "1.0.0"
  },
  "collections": [...],
  "documents": [...],
  "files": [...],
  "functions": [...],
  "deployments": [...],
  "github_repositories": [...],
  "web_servers": [...],
  "ssh_keys": [...],
  "api_keys": [...],
  "cors_configs": [...],
  "function_domains": [...],
  "audit_logs": [...],
  "backup_version": "1.0"
}
```

## Security Considerations

- **Encrypted Storage** - SSH private keys remain encrypted in backups
- **Checksum Verification** - SHA256 checksums ensure data integrity
- **Access Control** - Only project members can create/restore backups
- **Audit Trail** - All backup operations are logged

## Best Practices

### Regular Backups
- Create backups before major deployments
- Schedule regular automated backups (when available)
- Test restore procedures periodically

### Backup Naming
- Use descriptive names: "pre-v2.0-deployment"
- Include dates: "daily-backup-2024-01-15"
- Note significant changes: "before-schema-migration"

### Storage Management
- Monitor backup storage usage
- Delete old backups when no longer needed
- Keep critical milestone backups longer

## Configuration

### Environment Variables

```bash
# Backup storage directory
BACKUP_DIR=/var/lib/cloudbox/backups

# Master encryption key for sensitive data
MASTER_KEY=your-secure-master-key
```

### Storage Requirements

- **Disk Space** - Backups are compressed but can be large
- **Permissions** - Backup directory must be writable
- **Network** - Sufficient bandwidth for large backup transfers

## Troubleshooting

### Common Issues

**Backup Creation Fails**
```bash
# Check disk space
df -h /var/lib/cloudbox/backups

# Check permissions
ls -la /var/lib/cloudbox/backups

# Check logs
docker logs cloudbox-backend
```

**Restore Fails**
- Verify backup file integrity
- Check target project permissions
- Ensure sufficient database resources
- Review backup version compatibility

### Error Codes

- `BACKUP_NOT_FOUND` - Backup ID doesn't exist
- `BACKUP_INCOMPLETE` - Backup is still being created
- `RESTORE_FAILED` - Restore operation failed (check logs)
- `INSUFFICIENT_PERMISSIONS` - User lacks backup permissions
- `STORAGE_FULL` - Not enough disk space for backup

## API Reference

### Endpoints

```
GET    /api/projects/{id}/backups        List backups
POST   /api/projects/{id}/backups        Create backup
GET    /api/backups/{id}                 Get backup details
DELETE /api/backups/{id}                 Delete backup
POST   /api/backups/{id}/restore         Restore backup
```

### Response Examples

**Backup Object**
```json
{
  "id": 123,
  "name": "Weekly Backup",
  "description": "Automated weekly backup",
  "type": "manual",
  "status": "completed",
  "size": 1048576,
  "file_path": "/backups/backup-1-123.tar.gz",
  "checksum": "sha256:abc123...",
  "created_at": "2024-01-15T10:30:00Z",
  "completed_at": "2024-01-15T10:32:15Z",
  "project_id": 1
}
```

**Restore Response**
```json
{
  "message": "Backup restore completed successfully",
  "target_project_id": 1
}
```