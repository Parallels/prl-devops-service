---
layout: page
title: Blog Posts
show_sidebar: false
---

{% include categories.html %}

{% for post in site.posts %}
<div class="column is-12">
    {% include post-card.html %}
</div>
{% endfor %}
