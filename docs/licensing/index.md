---
layout: page
title: Licensing
permalink: /licensing/
---
# Licensing

**Parallels DevOps is free and open source.** You can download, run, modify, and self-host it at no cost.

It is, however, a management layer built **on top of Parallels Desktop (PD)**. It doesn't replace Parallels Desktop or remove the need for it — every host it manages still runs Parallels Desktop, and that software must be properly licensed by Parallels. So while *Parallels DevOps* costs nothing, *Parallels Desktop* still requires a valid license.

Put simply: there's nothing to buy from us. The only license you need is the Parallels Desktop license you'd already need to run PD at all.

## Which Parallels Desktop license do I need?

This depends on how many seats (managed hosts) you're running:

| Number of seats | Required Parallels Desktop license |
|---|---|
| **Up to 10** | Pro, Business, **or** Enterprise — any paid edition is fine |
| **More than 10** | Business or Enterprise only |

### Up to 10 seats

For small deployments, any paid edition of Parallels Desktop works. A standard **Parallels Desktop Pro Edition** license is enough — though Business or Enterprise are equally valid if you already hold them.

### More than 10 seats

Beyond 10 seats you must use **Parallels Desktop Business Edition (PDB)** or **Parallels Desktop Enterprise Edition (PDE)**. These editions are designed for volume deployment and centralized license management, which is what fleets of this size require. A bundle of individual Pro licenses is **not** a valid substitute at this scale.

## A note on enforcement

Parallels DevOps does **not** perform license checks and does **not** enforce seat counts. There is no technical gate that prevents you from managing more than 10 hosts using Pro licenses.

**The absence of a check is not permission.** Running beyond 10 seats without a Business or Enterprise license is a licensing violation on the Parallels Desktop side, regardless of what Parallels DevOps technically allows. Compliance with Parallels' own terms is your responsibility, and we ask that you honor it.

## FAQ

**Is Parallels DevOps itself ever paid?**
No. The project is — and will remain — free and open source. The only cost involved is the underlying Parallels Desktop license, which you would need to run PD with or without this tool.

**I already own individual Pro licenses for 15 machines. Am I covered?**
There's no technical block in Parallels DevOps. However, per Parallels' licensing, deployments over 10 seats should run on Business or Enterprise. We recommend consolidating onto a volume license to stay compliant.

**Does a seat mean a user or a machine?**
A seat corresponds to a host running Parallels Desktop under management. Confirm the exact definition against your Parallels Desktop license agreement.

**Where do I buy a Business or Enterprise license?**
Directly from Parallels: *(add purchase/contact link here)*.