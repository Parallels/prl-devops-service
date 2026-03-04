---
layout: page
title: Snapshots
subtitle: Managing Virtual Machine Snapshots
menubar: docs_devops_menu
show_sidebar: false
toc: true
---

# Managing Virtual Machine Snapshots

A **Snapshot** is a saved state of a virtual machine (VM) at a specific point in time. It captures the VM's disk data, configuration, and sometimes its memory state. Snapshots act as a safety net, allowing you to easily roll back a VM to a known good state if something goes wrong during development, testing, or system upgrades.

With the Parallels DevOps Service, you can manage your VM snapshots remotely using our REST API. This guide explains how to list, create, revert, and delete snapshots easily.

---

##  Prerequisites

Before managing snapshots via the API, make sure you have:
1. Installed and configured the service (see [Getting Started](/prl-devops-service/quick-start/)).
2. Obtained an API token for authentication. All examples below use `<YOUR_TOKEN>` as a placeholder for your Bearer token (see [How to get API token]({{ site.url }}{{ site.baseurl }}/docs/devops/restapi/reference/api_keys))

---

## 1. Creating a Snapshot

Before making major changes to a VM, it is always a good practice to create a snapshot.

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
---

## 2. Listing Snapshots

To view all the snapshots currently available for a specific virtual machine, you can use the List Snapshots endpoint.

**Endpoint:**
`GET /v1/machines/{id}/snapshots`

**Description:**
Restrieves a complete list of snapshots associated with the specified machine ID.

**Example Request:**
```bash
curl -X GET "https://api.example.com/v1/machines/123e4567-e89b-12d3-a456-426614174000/snapshots" \
     -H "Authorization: Bearer <YOUR_TOKEN>"
```

**Output Reference:**
```json
{
  "snapshots": [
    {
      "id": "{snapshot_id}",
      "name": "Snapshot name",
      "date": "Creation date",
      "state": "Snapshot state",
      "current": true,
      "parent": "{parent_snapshot_id}"
    }
  ]
}
```


---

## 3. Reverting to a Snapshot

If an upgrade failed or you need to restore the VM to a previous state, you can revert it to any existing snapshot.

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

---

## 4. Deleting a Single Snapshot

To save storage space or remove snapshots that are no longer needed, you can delete them individually.

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

## 5. Deleting All Snapshots

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

## Best Practices

- **Don't keep snapshots indefinitely**: Snapshots are meant for short-term use (e.g., during upgrades). Keeping them for long periods consumes disk space and can degrade VM performance.
- **Limit the number of snapshots**: Having a deep tree of snapshots can slow down disk operations. Try to keep only the snapshots you actively need.
- **Snapshots are not backups**: A snapshot depends on the base virtual disk. If the underlying virtual disk is corrupted or deleted, the snapshot is also lost. Always use proper backup solutions for long-term data retention.

## Troubleshooting & FAQ

**Q: Why is my virtual machine running slowly?**
**A:** You might have too many snapshots or snapshots that have been kept for a long time. Consider deleting old snapshots to consolidate the virtual disk files.

**Q: Can I restore a snapshot that I accidentally deleted?**
**A:** No, snapshot deletion is permanent and cannot be undone.

**Q: What happens to my current work if I revert to an older snapshot?**
**A:** Any changes made since that older snapshot was taken will be completely lost. If you want to save your current state, create a new snapshot before reverting.

**Q: Why does deleting a snapshot take a long time?**
**A:** When you delete a snapshot, the system has to merge the snapshot data back into the parent disk. If the snapshot is old and contains many changes, this merge process can take some time.

---

## Detailed API Reference

For full schemas, HTTP response codes, and advanced options, please refer to the [API Reference](/prl-devops-service/docs/devops/restapi/reference/).
