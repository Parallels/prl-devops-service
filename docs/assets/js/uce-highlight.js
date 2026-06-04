/**
 * UCE Syntax Highlighter — zero-dependency
 * Layered approach: highlighted <pre> behind a transparent <textarea>
 */
(function() {
  'use strict';

  var ESC = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' };
  function esc(s) { return s.replace(/[&<>"']/g, function(c) { return ESC[c]; }); }

  /* ── Language regexes ──────────────────────────────────────────── */

  function highlightLine(lang, line) {
    var tokens = [];
    var i = 0, len = line.length;

    while (i < len) {
      // Comment
      if ((lang==='yaml'||lang==='yml'||lang==='shell'||lang==='bash'||lang==='python'||lang==='docker') && line[i]==='#') {
        return '<span class="hl-comment">'+esc(line.substring(i))+'</span>';
      }

      // String (double-quoted)
      if (line[i]==='"') {
        var j=i+1, s='';
        while (j<len && line[j]!=='"') {
          if (line[j]==='\\') { s+=esc(line.substr(j,2)); j+=2; }
          else { s+=esc(line[j]); j++; }
        }
        if (j<len) { s+='"'; j++; }
        return '<span class="hl-string">'+esc('"'+s+'"')+'</span>'+highlightLine(lang, line.substring(j));
      }

      // String (single-quoted)
      if (line[i]==="'") {
        var j=i+1, s='';
        while (j<len && line[j]!=="'") { s+=esc(line[j]); j++; }
        if (j<len) { s+="'"; j++; }
        return '<span class="hl-string">'+esc("'"+s+"'")+'</span>'+highlightLine(lang, line.substring(j));
      }

      // Number
      if (/\d/.test(line[i]) && (i===0||!/\w/.test(line[i-1]))) {
        var j=i;
        while (j<len && /[\d.xXa-fA-F]/.test(line[j])) j++;
        return '<span class="hl-number">'+esc(line.substring(i,j))+'</span>'+highlightLine(lang, line.substring(j));
      }

      // Word
      if (/[a-zA-Z_$]/.test(line[i])) {
        var j=i;
        while (j<len && /[\w.$]/.test(line[j])) j++;
        var word=line.substring(i,j);
        var w=word.toLowerCase();
        // Keywords
        if (YAML_KW.indexOf(w)!==-1) return '<span class="hl-keyword">'+esc(word)+'</span>'+highlightLine(lang, line.substring(j));
        if (JS_KW.indexOf(w)!==-1) return '<span class="hl-keyword">'+esc(word)+'</span>'+highlightLine(lang, line.substring(j));
        if (PY_KW.indexOf(w)!==-1) return '<span class="hl-keyword">'+esc(word)+'</span>'+highlightLine(lang, line.substring(j));
        if (JAVA_KW.indexOf(w)!==-1) return '<span class="hl-keyword">'+esc(word)+'</span>'+highlightLine(lang, line.substring(j));
        if (GO_KW.indexOf(w)!==-1) return '<span class="hl-keyword">'+esc(word)+'</span>'+highlightLine(lang, line.substring(j));
        if (RUST_KW.indexOf(w)!==-1) return '<span class="hl-keyword">'+esc(word)+'</span>'+highlightLine(lang, line.substring(j));
        // Builtins
        if ((lang==='javascript'||lang==='typescript') && JS_BI.indexOf(word)!==-1) return '<span class="hl-builtin">'+esc(word)+'</span>'+highlightLine(lang, line.substring(j));
        if (lang==='python' && PY_BI.indexOf(word)!==-1) return '<span class="hl-builtin">'+esc(word)+'</span>'+highlightLine(lang, line.substring(j));
        if (lang==='go' && GO_BI.indexOf(word)!==-1) return '<span class="hl-builtin">'+esc(word)+'</span>'+highlightLine(lang, line.substring(j));
        // Shell commands
        if ((lang==='shell'||lang==='bash') && SH_CMD.indexOf(word)!==-1) return '<span class="hl-command">'+esc(word)+'</span>'+highlightLine(lang, line.substring(j));
        // Variable ($var or ${var})
        if (lang==='shell'||lang==='bash') {
          if (word==='$(' || word==='${' || (word==='$' && j<len && /[a-zA-Z_{(]/.test(line[j]))) {
            return '<span class="hl-variable">'+esc(word)+'</span>'+highlightLine(lang, line.substring(j));
          }
        }
        // Docker keywords
        if (lang==='docker' && line.substring(i,i+4).toUpperCase()==='FROM') {
          return '<span class="hl-keyword">FROM</span>'+highlightLine(lang, line.substring(j));
        }
        return '<span class="hl-ident">'+esc(word)+'</span>'+highlightLine(lang, line.substring(j));
      }

      // YAML list marker
      if ((lang==='yaml'||lang==='yml') && line[i]==='-') {
        return '<span class="hl-punct">-</span>'+highlightLine(lang, line.substring(i+1));
      }

      // Punctuation / operators
      if ('+-*/%=<>!&|^~:;,.?(){}[]'.indexOf(line[i])!==-1) {
        return '<span class="hl-punct">'+esc(line[i])+'</span>'+highlightLine(lang, line.substring(i+1));
      }

      // Fallback
      return esc(line[i])+highlightLine(lang, line.substring(i+1));
    }
    return '';
  }

  /* ── Keyword sets ──────────────────────────────────────────────── */

  var YAML_KW=['true','false','null','yes','no','on','off','True','False','Null'].join(',').split(',');
  var JS_KW=['function','var','let','const','if','else','for','while','do','switch','case','break','continue','return','try','catch','finally','throw','new','delete','typeof','instanceof','in','of','class','extends','super','import','export','from','default','async','await','yield','this','null','undefined','true','false','any','string','number','boolean','void','never','unknown','type','interface','enum','implements','declare','as','readonly','abstract','private','protected','public','static'].join(',').split(',');
  var JS_BI=['console','Math','JSON','Array','Object','String','Number','Boolean','Promise','Map','Set','RegExp','Date','Error','parseInt','parseFloat','setTimeout','setInterval','fetch','require','module','process','window','document'].join(',').split(',');
  var PY_KW=['def','class','if','elif','else','for','while','return','import','from','as','try','except','finally','raise','with','yield','lambda','pass','break','continue','and','or','not','is','in','True','False','None','assert','del','global','nonlocal','print'].join(',').split(',');
  var PY_BI=['len','range','enumerate','zip','map','filter','sorted','reversed','sum','min','max','abs','round','int','float','str','list','dict','set'].join(',').split(',');
  var JAVA_KW=['public','private','protected','static','final','abstract','class','interface','extends','implements','new','return','if','else','for','while','do','switch','case','break','continue','try','catch','finally','throw','throws','void','int','long','float','double','boolean','char','byte','short','String','this','super','null','true','false','instanceof','synchronized','volatile','transient','native'].join(',').split(',');
  var GO_KW=['package','import','func','var','const','type','struct','interface','map','chan','go','defer','return','if','else','for','range','switch','case','select','break','continue','break','goto','default','nil','true','false','make','new','append','len','cap','close','delete','copy','panic','recover','iota','string','int','int8','int16','int32','int64','uint','uint8','uint16','uint32','uint64','float32','float64','bool','byte','rune','error'].join(',').split(',');
  var GO_BI=['make','new','append','len','cap','close','delete','copy','panic','recover'].join(',').split(',');
  var RUST_KW=['fn','let','const','static','mut','ref','move','self','Self','super','crate','mod','use','pub','struct','enum','impl','trait','type','where','async','await','for','while','loop','if','else','match','return','break','continue','box','dyn','unsafe','extern','in','true','false','None','Some','Ok','Err','Vec','String','str','i8','i16','i32','i64','u8','u16','u32','u64','f32','f64','bool','char','Option','Result'].join(',').split(',');
  var SH_CMD=['echo','cat','ls','cd','mkdir','rm','cp','mv','chmod','chown','git','docker','docker-compose','kubectl','curl','wget','npm','pip','make','tar','zip','unzip','grep','sed','awk','find','xargs','tee','tr','sort','uniq','wc','head','tail','cut','ssh','scp','rsync','sudo','apt','yum','brew','python','node','java','javac','gcc','g++','clang','rustc','go','cargo'].join(',').split(',');

  /* ── Public API ────────────────────────────────────────────────── */

  window.UCEHighlight = {
    highlight: function(code, lang) {
      var lines = code.split('\n');
      var out = [];
      for (var i = 0; i < lines.length; i++) {
        out.push(highlightLine(lang || 'text', lines[i]));
      }
      return out.join('\n');
    },
    languages: ['yaml','yml','json','shell','bash','python','java','javascript','typescript','go','rust','ruby','nginx','docker','markdown','toml','xml','html','css','sql','ini','properties','makefile','text']
  };
})();
