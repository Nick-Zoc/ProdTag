# ProdTag Bash integration. Source from .bashrc/.bash_profile or the current shell.

: "${PRODTAG_BASH_ENABLED:=1}"
: "${PRODTAG_BASH_DEBUG:=0}"
_prodtag_bash_script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
: "${PRODTAG_HELPER:=${_prodtag_bash_script_dir}/../bin/prodtag-helper}"
: "${PRODTAG_MATCHER_CACHE:=${_prodtag_bash_script_dir}/../matcher-cache.json}"
: "${PRODTAG_CONFIG:=${_prodtag_bash_script_dir}/../config.json}"
: "${PRODTAG_BASH_DEBUG_LOG:=${TMPDIR:-/tmp}/prodtag-bash-debug.log}"

_prodtag_bash_debug() {
  [[ "$PRODTAG_BASH_DEBUG" == "1" ]] || return 0
  : >> "$PRODTAG_BASH_DEBUG_LOG" 2>/dev/null || return 0
  printf '%s %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" >> "$PRODTAG_BASH_DEBUG_LOG"
  local count; count="$(wc -l < "$PRODTAG_BASH_DEBUG_LOG" 2>/dev/null)"
  if [[ "$count" =~ ^[0-9]+$ ]] && (( count > 200 )); then tail -n 200 "$PRODTAG_BASH_DEBUG_LOG" > "${PRODTAG_BASH_DEBUG_LOG}.tmp" && mv "${PRODTAG_BASH_DEBUG_LOG}.tmp" "$PRODTAG_BASH_DEBUG_LOG"; fi
}

_prodtag_bash_infer() {
  local command_text exit_code="${2:-0}" suffix=success
  command_text="$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')"
  (( exit_code != 0 )) && suffix=failure
  case "$command_text" in
    git\ commit*) echo "git_commit_${suffix}";; git\ push*) echo "git_push_${suffix}";;
    npm\ test*|npm\ run\ test*|pnpm\ test*|pnpm\ run\ test*|yarn\ test*|pytest*|go\ test*|cargo\ test*|flutter\ test*) echo "test_${suffix}";;
    npm\ run\ build*|pnpm\ build*|pnpm\ run\ build*|yarn\ build*|go\ build*|cargo\ build*|flutter\ build*) echo "build_${suffix}";;
    *) echo "command_${suffix}";;
  esac
}

_prodtag_bash_cache_allows() {
  local event_type="$1"
  _prodtag_bash_cache_state=missing; _prodtag_bash_cache_reason="cache missing; helper fallback"
  [[ -r "$PRODTAG_MATCHER_CACHE" ]] || return 0
  if [[ -r "$PRODTAG_CONFIG" && "$PRODTAG_MATCHER_CACHE" -ot "$PRODTAG_CONFIG" ]]; then _prodtag_bash_cache_state=stale; _prodtag_bash_cache_reason="cache older than config; helper fallback"; return 0; fi
  if ! grep -Fq '"version": 2' "$PRODTAG_MATCHER_CACHE" || ! grep -Fq '"complete": true' "$PRODTAG_MATCHER_CACHE" || ! grep -Fq '"enabledEventTypes"' "$PRODTAG_MATCHER_CACHE"; then _prodtag_bash_cache_state=invalid; _prodtag_bash_cache_reason="cache validation failed; helper fallback"; return 0; fi
  _prodtag_bash_cache_state=valid
  if grep -Fq "\"$event_type\"" "$PRODTAG_MATCHER_CACHE"; then _prodtag_bash_cache_reason="enabled event type present; helper allowed"; return 0; fi
  _prodtag_bash_cache_reason="no enabled rule for event type; helper skipped"; return 1
}

_prodtag_bash_prompt() {
  local exit_code="$1"
  [[ "$PRODTAG_BASH_ENABLED" == "0" ]] && return "$exit_code"
  local command_text; command_text="$(HISTTIMEFORMAT= history 1 2>/dev/null | sed -E 's/^[[:space:]]*[0-9]+[[:space:]]+//')"
  [[ -z "${command_text//[[:space:]]/}" ]] && return "$exit_code"
  [[ "$command_text" == *prodtag-helper* || "$command_text" == *prodtag.bash* || "$command_text" == *PRODTAG_BASH_* ]] && return "$exit_code"
  local event_type; event_type="$(_prodtag_bash_infer "$command_text" "$exit_code")"
  _prodtag_bash_debug "captured command=$command_text exit_code=$exit_code event_type=$event_type helper=$PRODTAG_HELPER cache=$PRODTAG_MATCHER_CACHE"
  if ! _prodtag_bash_cache_allows "$event_type"; then _prodtag_bash_debug "cache_state=$_prodtag_bash_cache_state launch=skipped reason=$_prodtag_bash_cache_reason"; return "$exit_code"; fi
  _prodtag_bash_debug "cache_state=$_prodtag_bash_cache_state launch=allowed reason=$_prodtag_bash_cache_reason"
  if [[ ! -x "$PRODTAG_HELPER" ]]; then _prodtag_bash_debug "helper_result=error reason=helper missing or not executable path=$PRODTAG_HELPER"; return "$exit_code"; fi
  if [[ "$PRODTAG_BASH_DEBUG" == "1" ]]; then
    { local result code; result="$("$PRODTAG_HELPER" emit --event-type "$event_type" --command "$command_text" --exit-code "$exit_code" --cwd "$PWD" 2>&1)"; code=$?; _prodtag_bash_debug "helper_result_code=$code helper_result=$result"; } & disown 2>/dev/null
  else
    { "$PRODTAG_HELPER" emit --event-type "$event_type" --command "$command_text" --exit-code "$exit_code" --cwd "$PWD" >/dev/null 2>&1 & disown; } 2>/dev/null
  fi
  return "$exit_code"
}

if [[ ";${PROMPT_COMMAND:-};" != *";_prodtag_bash_prompt"* ]]; then
  PROMPT_COMMAND="_prodtag_bash_prompt \"\$?\"${PROMPT_COMMAND:+;$PROMPT_COMMAND}"
fi
