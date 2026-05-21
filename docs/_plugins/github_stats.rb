require 'json'
require 'net/http'
require 'uri'

Jekyll::Hooks.register :site, :pre_render do |site|
  begin
    uri = URI('https://api.github.com/repos/Parallels/prl-devops-service')
    req = Net::HTTP::Get.new(uri)
    req['Accept'] = 'application/vnd.github.v3+json'
    req['User-Agent'] = 'Jekyll-PrlDevOps-Docs'

    https = Net::HTTP.new(uri.host, uri.port)
    https.use_ssl = true
    https.open_timeout = 5
    https.read_timeout = 5

    response = https.request(req)

    if response.code == '200'
      data = JSON.parse(response.body)
      site.config['github_stars'] = data['stargazers_count'].to_i
      site.config['github_forks'] = data['forks_count'].to_i
    else
      puts "[GitHub Plugin] Failed to fetch repo stats (HTTP #{response.code}). Using static values."
    end
  rescue => e
    puts "[GitHub Plugin] Error fetching GitHub stats: #{e.message}. Using static values."
  end
end
