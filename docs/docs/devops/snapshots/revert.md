---
layout: page
title: Revert Snapshots
subtitle: Guide to reverting a Virtual Machine to a Snapshot
menubar: docs_devops_menu
show_sidebar: false
---

# Reverting to a Snapshot

If an upgrade failed or you need to restore the VM to a previous state, you can revert it to any existing snapshot.

Check [Prerequisites](/prl-devops-service/docs/devops/snapshots/#prerequisites) before reverting a snapshot.

**Endpoint:**
`POST /v1/machines/{id}/snapshots/{snapshot_id}/revert`

**Description:**
Restores the virtual machine to the exact state it was in when the specified snapshot (`snapshot_id`) was created. Any changes made after that snapshot will be lost unless saved in a separate snapshot.

**Example Request:**
```bash
curl -X POST "https://api.example.com/v1/machines/123e4567-e89b-12d3-a456-426614174000/snapshots/snap-12345/revert" \
     -H "Authorization: Bearer <YOUR_TOKEN>" \
     -H "Content-Type: application/json" \
     -d '{
           "skip_resume": false
         }'
```

**Input Reference (`RevertSnapshotRequest`):**
```json
{
  "skip_resume": "boolean (optional)"
}
```

**Output Reference:**
*Returns HTTP 202 Accepted upon success.*
