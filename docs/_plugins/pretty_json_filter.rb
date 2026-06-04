require 'json'

module PrettyJsonFilter
  def pretty_json(input)
    return input if input.nil?

    JSON.pretty_generate(JSON.parse(input.to_s))
  rescue JSON::ParserError, TypeError
    input
  end
end

Liquid::Template.register_filter(PrettyJsonFilter)
