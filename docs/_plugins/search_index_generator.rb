# Search Index Generator Plugin
# Generates search.json for client-side search using post_write hook
# (post_write fires AFTER Jekyll finishes writing _site/, so the dir exists)

require 'yaml'
require 'set'
require 'json'

Jekyll::Hooks.register :site, :post_write do |site|
  # Build a map of menu labels for breadcrumb resolution
  breadcrumb_map = {}
  menu_files = [
    '_data/docs_devops_menu.yml',
    '_data/docs_github_action_menu.yml',
    '_data/docs_terraform_menu.yml',
    '_data/docs_vscode_menu.yml'
  ]

  menu_files.each do |menu_file|
    menu_path = File.join(site.source, menu_file)
    next unless File.exist?(menu_path)

    begin
      menu_data = YAML.load_file(menu_path)
      menu_data.each do |menu_group|
        next unless menu_group.is_a?(Hash) && menu_group['items']
        group_label = menu_group['label'] || ''
        menu_group['items'].each do |item|
          next unless item.is_a?(Hash)
          if item['link']
            breadcrumb_map[item['link']] = group_label
          end
          if item['items']
            item['items'].each do |subitem|
              next unless subitem.is_a?(Hash) && subitem['link']
              breadcrumb_map[subitem['link']] = "#{group_label} > #{item['name'] || ''}".strip
            end
          end
        end
      end
    rescue => e
      puts "[SearchIndex] Warning: Could not parse #{menu_file}: #{e.message}"
    end
  end

  # Collect pages to index
  pages_to_index = []

  # 1. Pages under /docs/ collection
  site.pages.each do |page|
    next unless page.path.include?('/docs/') && page.path.end_with?('.md')
    next if page.path.include?('_site/')
    next if page.data && page.data['layout'] == 'none'

    pages_to_index << page
  end

  # 2. Top-level .md pages (like quick-start.md, features.md, etc.)
  site.pages.each do |page|
    next unless page.path.end_with?('.md')
    next if page.path.include?('/docs/')
    next if page.path.include?('_site/')
    next if page.path.include?('404.html')
    next if page.path.include?('test.html')
    next if page.data && page.data['layout'] == 'none'
    next if page.data && page.data['is_home']

    pages_to_index << page
  end

  # Deduplicate by path
  seen_paths = Set.new
  pages_to_index = pages_to_index.reject do |page|
    if seen_paths.include?(page.path)
      true
    else
      seen_paths << page.path
      false
    end
  end

  # Build search index entries
  search_index = pages_to_index.map do |page|
    # Strip HTML tags and newlines for clean text content
    raw_content = page.content.to_s
    stripped_content = raw_content.gsub(/<[^>]*>/, '').gsub(/\s+/, ' ').strip

    # Take first 500 characters as excerpt
    excerpt = stripped_content.length > 500 ? stripped_content[0, 500] + '...' : stripped_content

    # Resolve breadcrumb from menu map
    rel_url = page.url.to_s
    breadcrumb = breadcrumb_map[rel_url] || ''

    # Prepend baseurl to page.url so links work from the served site
    baseurl = site.baseurl.to_s
    full_url = baseurl.empty? ? rel_url : baseurl + rel_url

    entry = {
      title:    (page.data && page.data['title']) ? page.data['title'].to_s : '',
      category: breadcrumb,
      tags:     '',
      url:      full_url,
      content:  excerpt
    }
    if page.data && page.data['date']
      entry[:date] = page.data['date'].strftime('%Y-%m-%d')
    end
    entry
  end

  # Write search.json to site destination
  search_json_path = File.join(site.dest, 'search.json')
  json_content = JSON.pretty_generate(search_index)
  File.write(search_json_path, json_content)

  puts "[SearchIndex] Generated search.json with #{search_index.size} entries"
end
