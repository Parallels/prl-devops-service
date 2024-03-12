# This "hook" is executed right before the site's pages are rendered
Jekyll::Hooks.register :site, :pre_render do |site|
    puts "Adding Parallels Desktop PDFile Markdown aliases..."
    require "rouge"

    # This class defines the PDL lexer which is used to highlight "pdl" code snippets during render-time
    class ParallelsDesktopFile < Rouge::RegexLexer
      title 'pdfile'
      aliases 'parallels-file', 'pdfile'

      KEYWORDS = %w(
        TO FROM INSECURE AUTHENTICATE PROVIDER LOCAL_PATH DESCRIPTION TAG ROLE CLAIM CATALOG_ID VERSION ARCHITECTURE
        MACHINE_NAME OWNER DESTINATION START_AFTER_PULL DO RUN
      ).join('|')

      AUTHENTICATE_SUBCOMMANDS = %w(
        USERNAME PASSWORD NAME BUCKET REGION ACCESS_KEY SECRET_KEY
      ).join('|')

      PROVIDER_SUBCOMMANDS = %w(
        USERNAME PASSWORD NAME BUCKET REGION ACCESS_KEY SECRET_KEY
      ).join('|')

      state :root do
        rule %r/\s+/, Text

        rule %r/^(FROM)(\s+)(.*)(\s+)(AS)(\s+)(.*)/io do
          groups Keyword, Text::Whitespace, Str, Text::Whitespace, Keyword, Text::Whitespace, Str
        end

        rule %r/^(AUTHENTICATE)(\s+)(#{AUTHENTICATE_SUBCOMMANDS})(\s+)(.*)/io do
            groups Keyword, Text::Whitespace, Name::Variable, Text::Whitespace, Punctuation
        end

        rule %r/^(AUTHENTICATE)(\s+)(#{AUTHENTICATE_SUBCOMMANDS})(.*)/io do
            groups Keyword, Text::Whitespace, Error, Error
        end

        rule %r/^(PROVIDER)(\s+)(#{PROVIDER_SUBCOMMANDS})(\s+)(.*)/io do
            groups Keyword, Text::Whitespace, Name::Variable, Text::Whitespace, Punctuation
        end

        rule %r/^(PROVIDER)(\s+)(#{PROVIDER_SUBCOMMANDS})(.*)/io do
            groups Keyword, Text::Whitespace, Error, Error
        end

        rule %r/^(TAG)(\s+)(.*)/io do
            groups Keyword, Text::Whitespace, Punctuation
        end

        rule %r/^(ROLE)(\s+)(.*)/io do
            groups Keyword, Text::Whitespace, Punctuation
        end

        rule %r/^(CLAIM)(\s+)(.*)/io do
            groups Keyword, Text::Whitespace, Punctuation
        end

        rule %r/^(#{KEYWORDS})\b(.*)/io do
          groups Keyword, Str
        end

        rule %r/#.*?$/, Comment

        rule %r/^(ONBUILD\s+)?RUN(\s+)/i do
          token Keyword
          push :run
        end

        rule %r/\w+/, Text
        rule %r/[^\w]+/, Text
        rule %r/./, Text
      end

      state :run do
        rule %r/\n/, Text, :pop!
        rule %r/\\./m, Str::Escape
        rule(/(\\.|[^\n\\])+/) { delegate @shell }
      end
    end
end