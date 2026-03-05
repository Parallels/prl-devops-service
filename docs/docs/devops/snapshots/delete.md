---
layout: page
title: Delete Snapshots
subtitle: Guide to deleting Virtual Machine Snapshots
menubar: docs_devops_menu
show_sidebar: false
---

# Deleting Snapshots

To save storage space or remove snapshots that are no longer needed, you can delete them.

## Deleting a Single Snapshot

**Endpoint:**
`DELETE /v1/machines/{id}/snapshots/{snapshot_id}`

**Description:**
Permanently removes the specified snapshot. Note that this action cannot be undone.

**Example Request:**
```bash
curl -X DELETE "https://api.example.com/v1/machines/123e4567-e89b-12d3-a456-426614174000/snapshots/snap-12345" \
     -H "Authorization: Bearer <YOUR_TOKEN>" \
     -H "Content-Type: application/json" \
     -d '{
           "delete_children": false
         }'
```

**Input Reference (`DeleteSnapshotRequest`):**
```json
{
  "delete_children": "boolean (optional)"
}
```

**Output Reference:**
*Returns HTTP 202 Accepted upon success.*

---

## Deleting All Snapshots

If you want to clear out all snapshots for a particular virtual machine in one go, you can use the Delete All endpoint. Use this with caution!

**Endpoint:**
`DELETE /v1/machines/{id}/snapshots`

**Description:**
Permanently deletes every snapshot associated with the virtual machine.

**Example Request:**
```bash
curl -X DELETE "https://api.example.com/v1/machines/123e4567-e89b-12d3-a456-426614174000/snapshots" \
     -H "Authorization: Bearer <YOUR_TOKEN>"
```

**Input Reference:** None

**Output Reference:**
*Returns HTTP 202 Accepted upon success.*
