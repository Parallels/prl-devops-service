---
layout: page
title: Create Snapshots
subtitle: Guide to creating Virtual Machine Snapshots
menubar: docs_devops_menu
show_sidebar: false
---

# Creating a Snapshot

Before making major changes to a VM, it is always a good practice to create a snapshot.

Check [Prerequisites](/prl-devops-service/docs/devops/snapshots/#prerequisites) before creating a snapshot.

**Endpoint:**
`POST /v1/machines/{id}/snapshots`

**Description:**
Captures the current state of the virtual machine and saves it as a new snapshot. You can provide a name and description to easily identify the snapshot later.

**Example Request:**
```bash
curl -X POST "https://api.example.com/v1/machines/123e4567-e89b-12d3-a456-426614174000/snapshots" \
     -H "Authorization: Bearer <YOUR_TOKEN>" \
     -H "Content-Type: application/json" \
     -d '{
           "snapshot_name": "Before big upgrade",
           "snapshot_description": "Snapshot taken before applying OS updates."
         }'
```

**Input Reference (`CreateSnapShotRequest`):**
```json
{
  "snapshot_name": "string (optional)",
  "snapshot_description": "string (optional)"
}
```

**Output Reference (`CreateSnapShotResponse`):**
```json
{
  "snapshot_name": "string",
  "snapshot_id": "string"
}
```
