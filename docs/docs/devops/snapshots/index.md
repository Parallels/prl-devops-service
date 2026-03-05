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

With the Parallels DevOps Service, you can manage your VM snapshots remotely using our REST API. Use the following guides to learn about specific snapshot operations:

- [Creating a Snapshot](/prl-devops-service/docs/devops/snapshots/create/)
- [Listing Snapshots](/prl-devops-service/docs/devops/snapshots/list/)
- [Reverting to a Snapshot](/prl-devops-service/docs/devops/snapshots/revert/)
- [Deleting Snapshots](/prl-devops-service/docs/devops/snapshots/delete/)

---             

## Prerequisites

Before managing snapshots via the API, make sure you have:
1. Installed and configured the service (see [Getting Started](/prl-devops-service/quick-start/)).
2. Obtained an API token for authentication. All examples use `<YOUR_TOKEN>` as a placeholder for your Bearer token (see [How to get API token]({{ site.url }}{{ site.baseurl }}/docs/devops/restapi/reference/api_keys))

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
