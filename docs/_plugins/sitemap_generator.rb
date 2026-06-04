# Sitemap Generator Plugin
# Generates a rich sitemap.xml with use case metadata, lastmod dates, and priority scores
# Uses post_write hook so _site/ directory exists

require 'yaml'
require 'rexml/document'
require 'set'

module SitemapGeneratorPlugin
  def self.compute_priority(page, site)
    path = page.path

    # Highest priority: index, use cases index
    if path == 'index.md' || path == 'use-cases/index.md'
      return '1.0'
    end

    # Individual use cases
    if path.start_with?('use-cases/')
      return '0.9'
    end

    # Docs
    if path.start_with?('docs/')
      return '0.8'
    end

    # Features and solutions
    if path.start_with?('features/', 'solutions/')
      return '0.7'
    end

    # Blog posts
    if path.start_with?('_posts/')
      return '0.6'
    end

    # Everything else
    return '0.5'
  end

  def self.escape_xml(text)
    text.to_s
      .gsub('&', '&amp;')
      .gsub('<', '&lt;')
      .gsub('>', '&gt;')
      .gsub('"', '&quot;')
      .gsub("'", '&apos;')
  end
end

Jekyll::Hooks.register :site, :post_write do |site|
  sitemap_path = File.join(site.dest, 'sitemap.xml')

  # Collect all pages that should appear in sitemap
  pages_to_index = []

  # 1. Regular pages (non-use-case, non-docs)
  site.pages.each do |page|
    next if page.path.include?('_site/')
    next if page.path.include?('404.html')
    next if page.path.include?('test.html')
    next if page.data && page.data['layout'] == 'none'
    next if page.data && page.data['is_home']
    next if page.path.end_with?('.html') && !page.path.include?('/use-cases/')
    next unless page.path.end_with?('.md')

    pages_to_index << page
  end

  # 2. Use case pages
  site.pages.each do |page|
    next if page.path.include?('_site/')
    next unless page.path.start_with?('use-cases/')
    next unless page.path.end_with?('.md')
    next if page.data && page.data['layout'] == 'none'

    pages_to_index << page
  end

  # 3. Solution pages
  site.pages.each do |page|
    next if page.path.include?('_site/')
    next unless page.path.start_with?('solutions/')
    next unless page.path.end_with?('.md')
    next if page.data && page.data['layout'] == 'none'

    pages_to_index << page
  end

  # 4. Feature pages
  site.pages.each do |page|
    next if page.path.include?('_site/')
    next unless page.path.start_with?('features/')
    next unless page.path.end_with?('.md')
    next if page.data && page.data['layout'] == 'none'

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

  # Sort by path for deterministic output
  pages_to_index.sort_by! { |p| p.path }

  base_url = site.config['url'].to_s.chomp('/')
  baseurl = site.config['baseurl'].to_s.chomp('/')

  # Build sitemap XML as a string (avoids REXML formatter issues)
  lines = []
  lines << '<?xml version="1.0" encoding="UTF-8"?>'
  lines << '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"'
  lines << '        xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"'
  lines << '        xsi:schemaLocation="http://www.sitemaps.org/schemas/sitemap/0.9'
  lines << '                           http://www.sitemaps.org/schemas/sitemap/0.9/sitemap.xsd">'

  pages_to_index.each do |page|
    # Build the full URL
    rel_url = page.url.to_s
    rel_url = '/' + rel_url unless rel_url.start_with?('/')
    full_url = base_url + baseurl + rel_url

    # Determine lastmod from page data or file mtime
    if page.data && page.data['date']
      lastmod = page.data['date'].strftime('%Y-%m-%d')
    elsif page.data && page.data['last_modified_at']
      lastmod = page.data['last_modified_at'].strftime('%Y-%m-%d')
    else
      src_file = File.join(site.source, page.path)
      if File.exist?(src_file)
        lastmod = File.mtime(src_file).strftime('%Y-%m-%d')
      else
        lastmod = Time.now.strftime('%Y-%m-%d')
      end
    end

    # Priority based on page type
    priority = SitemapGeneratorPlugin.compute_priority(page, site)

    # Escape all text content
    escaped_loc = SitemapGeneratorPlugin.escape_xml(full_url)
    escaped_lastmod = SitemapGeneratorPlugin.escape_xml(lastmod)
    escaped_priority = SitemapGeneratorPlugin.escape_xml(priority)

    lines << '  <url>'
    lines << "    <loc>#{escaped_loc}</loc>"
    lines << "    <lastmod>#{escaped_lastmod}</lastmod>"
    lines << "    <priority>#{escaped_priority}</priority>"

    # Use case pages get extra metadata as XML comments for crawlers
    if page.path.start_with?('use-cases/')
      uce_id = page.data && page.data['uce_data']
      if uce_id && site.data[uce_id]
        uc = site.data[uce_id]
        comment_parts = [
          "Use Case: #{uc['title'] || ''}",
          "Level: #{uc['level'] || ''}",
          "Duration: #{uc['duration'] || ''}",
          "Tags: #{(uc['tags'] || []).join(', ')}"
        ].compact.join(' | ')
        escaped_comment = SitemapGeneratorPlugin.escape_xml(comment_parts)
        lines << "    <!-- #{escaped_comment} -->"
      end
    end

    lines << '  </url>'
  end

  lines << '</urlset>'

  # Write sitemap.xml
  File.write(sitemap_path, lines.join("\n") + "\n")

  puts "[Sitemap] Generated sitemap.xml with #{pages_to_index.size} URLs"
end
