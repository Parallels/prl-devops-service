# This "hook" is executed right before the site's pages are rendered
Jekyll::Hooks.register :site, :pre_render do |site|
    puts "Adding prldevops Markdown aliases..."
    require "rouge"

    # This class defines the PDL lexer which is used to highlight "pdl" code snippets during render-time
    class ParallelsDevops < Rouge::RegexLexer
      title 'prldevops'
      aliases 'prldevops'

      KEYWORDS = %w(
        catalog push pull run list import import-vm api update-root-password gen-rsa
      ).join('|')

      AUTHENTICATE_SUBCOMMANDS = %w(
        USERNAME PASSWORD NAME BUCKET REGION ACCESS_KEY SECRET_KEY
      ).join('|')

      PROVIDER_SUBCOMMANDS = %w(
        USERNAME PASSWORD NAME BUCKET REGION ACCESS_KEY SECRET_KEY
      ).join('|')

      state :root do
        rule %r/\s+/, Text

        rule %r/^(prldevops)(\s+)(#{KEYWORDS})(\s+)(.*)/io do
          groups Keyword, Text::Whitespace, Name::Variable, Text::Whitespace, Str
        end

        rule %r/^(prldevops)/io do
            groups Keyword
        end

        rule %r/#.*?$/, Comment

        rule %r/\w+/, Text
        rule %r/[^\w]+/, Text
        rule %r/./, Text
      end

      state :run do
        rule %r/\n/, Text, :pop!
        rule %r/\\./m, Str::Escape
      end
    end
end