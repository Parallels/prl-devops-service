---
layout: page
title: List Snapshots
subtitle: Guide to listing Virtual Machine Snapshots
menubar: docs_devops_menu
show_sidebar: false
---

# Listing Snapshots

To view all the snapshots currently available for a specific virtual machine, you can use the List Snapshots endpoint.

**Endpoint:**
`GET /v1/machines/{id}/snapshots`

**Description:**
Retrieves a complete list of snapshots associated with the specified machine ID.

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
