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
        MACHINE_NAME OWNER DESTINATION START_AFTER_PULL DO MINIMUM_REQUIREMENT COMPRESS_PACK COMPRESS_PACK_LEVEL
        VM_REMOTE_PATH FORCE VM_SIZE VM_TYPE IS_COMPRESSED EXECUTE CLONE RUN
      ).join('|')

      AUTHENTICATE_SUBCOMMANDS = %w(
        USERNAME PASSWORD NAME BUCKET REGION ACCESS_KEY SECRET_KEY
      ).join('|')

      PROVIDER_SUBCOMMANDS = %w(
        USERNAME PASSWORD NAME BUCKET REGION ACCESS_KEY SECRET_KEY
      ).join('|')

      MINIMUM_REQUIREMENT_SUBCOMMANDS = %w(
        CPU RAM DISK MEMORY
      ).join('|')

      FORCE_SUBCOMMANDS = %w(
        true false
      ).join('|')

      COMPRESS_PACK_LEVEL_SUBCOMMANDS = %w(
        default balanced best_speed best_compression no_compression
      ).join('|')

      COMPRESS_PACK_SUBCOMMANDS = %w(
        true false
      ).join('|')

      IS_COMPRESSED_SUBCOMMANDS = %w(
        true false
      ).join('|')

      VM_TYPE_SUBCOMMANDS = %w(
        pvm macvm
      ).join('|')

      state :root do
        rule %r/\s+/, Text
        rule %r/#.*?$/, Comment

        rule %r/^((?:ONBUILD\s+)?RUN)(\s+)([^\n#]*)(\s*)(#.*)?$/io do |m|
          token Keyword, m[1]
          token Text::Whitespace, m[2]
          command = m[3] || ''
          unless command.empty?
            stripped = command.rstrip
            trailing = command[stripped.length..] || ''
            token Name::Function, stripped unless stripped.empty?
            token Text::Whitespace, trailing unless trailing.empty?
          end
          token Text::Whitespace, m[4] if m[4] && !m[4].empty?
          token Comment, m[5] if m[5]
        end

        rule %r/^(FROM)(\s+)(.*)(\s+)(AS)(\s+)(.*)/io do
          groups Keyword, Text::Whitespace, Str, Text::Whitespace, Keyword, Text::Whitespace, Str
        end

        rule %r/^(AUTHENTICATE)(\s+)(#{AUTHENTICATE_SUBCOMMANDS})(\s+)(.*)/io do
            groups Keyword, Text::Whitespace, Name::Variable, Text::Whitespace, Str
        end

        rule %r/^(AUTHENTICATE)(\s+)(#{AUTHENTICATE_SUBCOMMANDS})(.*)/io do
            groups Keyword, Text::Whitespace, Error, Error
        end

        rule %r/^(PROVIDER)(\s+)(?=[^#]*=)/io do
            groups Keyword, Text::Whitespace
            push :provider
        end

        rule %r/^(PROVIDER)(\s+)(#{PROVIDER_SUBCOMMANDS})(\s+)(.*)/io do
            groups Keyword, Text::Whitespace, Name::Variable, Text::Whitespace, Str
        end

        rule %r/^(MINIMUM_REQUIREMENT)(\s+)(#{MINIMUM_REQUIREMENT_SUBCOMMANDS})(\s+)(.*)/io do
            groups Keyword, Text::Whitespace, Name::Variable, Text::Whitespace, Str
        end

        rule %r/^(FORCE)(\s+)(#{FORCE_SUBCOMMANDS})(?=\s*(#|$))/io do
            groups Keyword, Text::Whitespace, Name::Constant
        end

        rule %r/^(FORCE)(\s+)(\S.*)/io do
            groups Keyword, Text::Whitespace, Error
        end

        rule %r/^(COMPRESS_PACK_LEVEL)(\s+)(#{COMPRESS_PACK_LEVEL_SUBCOMMANDS})(?=\s*(#|$))/io do
            groups Keyword, Text::Whitespace, Name::Constant
        end

        rule %r/^(COMPRESS_PACK_LEVEL)(\s+)(\S.*)/io do
            groups Keyword, Text::Whitespace, Error
        end

        rule %r/^(COMPRESS_PACK)(\s+)(#{COMPRESS_PACK_SUBCOMMANDS})(?=\s*(#|$))/io do
            groups Keyword, Text::Whitespace, Name::Constant
        end

        rule %r/^(COMPRESS_PACK)(\s+)(\S.*)/io do
            groups Keyword, Text::Whitespace, Error
        end

        rule %r/^(IS_COMPRESSED)(\s+)(#{IS_COMPRESSED_SUBCOMMANDS})(?=\s*(#|$))/io do
            groups Keyword, Text::Whitespace, Name::Constant
        end

        rule %r/^(IS_COMPRESSED)(\s+)(\S.*)/io do
            groups Keyword, Text::Whitespace, Error
        end

        rule %r/^(VM_TYPE)(\s+)(#{VM_TYPE_SUBCOMMANDS})(?=\s*(#|$))/io do
            groups Keyword, Text::Whitespace, Name::Constant
        end

        rule %r/^(VM_TYPE)(\s+)(\S.*)/io do
            groups Keyword, Text::Whitespace, Error
        end

        rule %r/^(PROVIDER)(\s+)(#{PROVIDER_SUBCOMMANDS})(.*)/io do
            groups Keyword, Text::Whitespace, Error, Error
        end

        rule %r/^(TAG)(\s+)/io do
            groups Keyword, Text::Whitespace
            push :csv
        end

        rule %r/^(ROLE)(\s+)/io do
            groups Keyword, Text::Whitespace
            push :csv
        end

        rule %r/^(CLAIM)(\s+)/io do
            groups Keyword, Text::Whitespace
            push :csv
        end

        rule %r/^(#{KEYWORDS})\b(.*)/io do
          groups Keyword, Str
        end

        rule %r/\w+/, Text
        rule %r/[^\w]+/, Text
        rule %r/./, Text
      end

      state :provider do
        rule %r/#.*?$/, Comment, :pop!
        rule %r/\n/, Text, :pop!
        rule %r/\s+/, Text

        rule %r/([A-Z0-9_-]+)(=)([^;#\s]+)(;)/io do
          groups Name::Attribute, Operator, Str, Punctuation
        end

        rule %r/([A-Z0-9_-]+)(=)([^;#\s]+)/io do
          groups Name::Attribute, Operator, Str
        end

        rule %r/./, Text
      end

      state :csv do
        rule %r/#.*?$/, Comment, :pop!
        rule %r/\n/, Text, :pop!
        rule %r/\s+/, Text

        rule %r/([^,#\s]+)(,)/ do
          groups Name::Variable, Punctuation
        end

        rule %r/[^,#\s]+/ do |m|
          token Name::Variable, m[0]
        end

        rule %r/./, Text
      end

    end
end
