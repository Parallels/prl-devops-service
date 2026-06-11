require 'json'
require 'net/http'
require 'uri'

Jekyll::Hooks.register :site, :pre_render do |site|
  repos = {
    'prl-devops-service' => 'service',
    'prl-devops-ui' => 'ui'
  }

  repos.each do |repo, config_key|
    begin
      uri = URI("https://api.github.com/repos/Parallels/#{repo}")
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
        site.config["github_stars_#{config_key}"] = data['stargazers_count'].to_i
        site.config["github_forks_#{config_key}"] = data['forks_count'].to_i

        if repo == 'prl-devops-service'
          site.config['github_stars'] = data['stargazers_count'].to_i
          site.config['github_forks'] = data['forks_count'].to_i
        end
      else
        puts "[GitHub Plugin] Failed to fetch #{repo} stats (HTTP #{response.code}). Using static values."
      end
    rescue => e
      puts "[GitHub Plugin] Error fetching #{repo} stats: #{e.message}. Using static values."
    end
  end
end
