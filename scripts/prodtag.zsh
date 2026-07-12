# ProdTag zsh integration.

autoload -Uz add-zsh-hook
zmodload zsh/datetime 2>/dev/null

: "${PRODTAG_ZSH_ENABLED:=1}"
: "${PRODTAG_ZSH_DEBUG:=0}"
_prodtag_script_dir="${${(%):-%N}:A:h}"
: "${PRODTAG_HELPER:=${_prodtag_script_dir}/../bin/prodtag-helper}"
: "${PRODTAG_MATCHER_CACHE:=${_prodtag_script_dir}/../matcher-cache.json}"
: "${PRODTAG_CONFIG:=${_prodtag_script_dir}/../config.json}"
: "${PRODTAG_ZSH_DEBUG_LOG:=${TMPDIR:-/tmp}/prodtag-zsh-debug.log}"

_prodtag_last_command=""
_prodtag_started_ms=""
_prodtag_cache_state="unchecked"
_prodtag_cache_reason=""

_prodtag_now_ms() {
  if [[ -n "${EPOCHREALTIME:-}" ]]; then
    printf '%.0f\n' "$(( EPOCHREALTIME * 1000 ))"
  else
    printf '%s000\n' "$(date +%s)"
  fi
}

_prodtag_debug() {
  [[ "$PRODTAG_ZSH_DEBUG" == "1" ]] || return 0
  : >> "$PRODTAG_ZSH_DEBUG_LOG" 2>/dev/null || return 0
  print -r -- "$(date -u +%Y-%m-%dT%H:%M:%SZ) $*" >> "$PRODTAG_ZSH_DEBUG_LOG"
  local line_count="$(wc -l < "$PRODTAG_ZSH_DEBUG_LOG" 2>/dev/null)"
  if [[ "$line_count" == <-> ]] && (( line_count > 200 )); then
    tail -n 200 "$PRODTAG_ZSH_DEBUG_LOG" > "${PRODTAG_ZSH_DEBUG_LOG}.tmp" 2>/dev/null && mv "${PRODTAG_ZSH_DEBUG_LOG}.tmp" "$PRODTAG_ZSH_DEBUG_LOG"
  fi
}

_prodtag_should_ignore_command() {
  local command_text="${1:-}"
  [[ -z "${command_text//[[:space:]]/}" ]] && return 0
  [[ "$command_text" == prodtag-helper* || "$command_text" == *"/prodtag-helper"* ]] && return 0
  [[ "$command_text" == *"prodtag.zsh"* || "$command_text" == *"PRODTAG_ZSH_"* ]] && return 0
  return 1
}

_prodtag_infer_event_type() {
  local command_text="${(L)1}" exit_code="${2:-0}" suffix="success"
  (( exit_code != 0 )) && suffix="failure"
  case "$command_text" in
    git\ commit*) echo "git_commit_${suffix}" ;;
    git\ push*) echo "git_push_${suffix}" ;;
    npm\ test*|npm\ run\ test*|pnpm\ test*|pnpm\ run\ test*|yarn\ test*|pytest*|go\ test*|cargo\ test*|flutter\ test*) echo "test_${suffix}" ;;
    npm\ run\ build*|pnpm\ build*|pnpm\ run\ build*|yarn\ build*|go\ build*|cargo\ build*|flutter\ build*) echo "build_${suffix}" ;;
    *) echo "command_${suffix}" ;;
  esac
}

_prodtag_cache_may_match() {
  local event_type="$1" cache="$PRODTAG_MATCHER_CACHE" cache_text
  _prodtag_cache_state="missing"; _prodtag_cache_reason="cache missing; helper fallback"
  [[ -r "$cache" ]] || return 0
  if [[ -r "$PRODTAG_CONFIG" && "$cache" -ot "$PRODTAG_CONFIG" ]]; then
    _prodtag_cache_state="stale"; _prodtag_cache_reason="cache older than config; helper fallback"; return 0
  fi
  cache_text="$(<"$cache")" 2>/dev/null || { _prodtag_cache_state="invalid"; _prodtag_cache_reason="cache unreadable; helper fallback"; return 0; }
  if [[ "$cache_text" != *'"version": 2'* || "$cache_text" != *'"complete": true'* || "$cache_text" != *'"enabledEventTypes"'* ]]; then
    _prodtag_cache_state="invalid"; _prodtag_cache_reason="cache validation failed; helper fallback"; return 0
  fi
  _prodtag_cache_state="valid"
  if [[ "$cache_text" == *'"enabledEventTypes": ['*'"'"$event_type"'"'* ]]; then
    _prodtag_cache_reason="enabled event type present; helper allowed"; return 0
  fi
  _prodtag_cache_reason="no enabled rule for event type; helper skipped"; return 1
}

_prodtag_preexec() {
  [[ "$PRODTAG_ZSH_ENABLED" == "0" ]] && return 0
  _prodtag_last_command="$1"
  _prodtag_started_ms="$(_prodtag_now_ms)"
}

_prodtag_precmd() {
  local exit_code="$?"
  [[ "$PRODTAG_ZSH_ENABLED" == "0" ]] && return "$exit_code"
  [[ -n "$_prodtag_last_command" ]] || return "$exit_code"
  local command_text="$_prodtag_last_command" started_ms="$_prodtag_started_ms"
  _prodtag_last_command=""; _prodtag_started_ms=""
  _prodtag_should_ignore_command "$command_text" && return "$exit_code"

  local duration_ms="" finished_ms="$(_prodtag_now_ms)"
  if [[ -n "$started_ms" ]]; then duration_ms="$(( finished_ms - started_ms ))"; (( duration_ms < 0 )) && duration_ms=0; fi
  local event_type="$(_prodtag_infer_event_type "$command_text" "$exit_code")"

  _prodtag_debug "captured command=$command_text exit_code=$exit_code event_type=$event_type helper=$PRODTAG_HELPER cache=$PRODTAG_MATCHER_CACHE"
  if ! _prodtag_cache_may_match "$event_type"; then
    _prodtag_debug "cache_state=$_prodtag_cache_state launch=skipped reason=$_prodtag_cache_reason"
    return "$exit_code"
  fi
  _prodtag_debug "cache_state=$_prodtag_cache_state launch=allowed reason=$_prodtag_cache_reason"
  if [[ ! -x "$PRODTAG_HELPER" ]]; then _prodtag_debug "helper_result=error reason=helper missing or not executable path=$PRODTAG_HELPER"; return "$exit_code"; fi

  if [[ "$PRODTAG_ZSH_DEBUG" == "1" ]]; then
    ( local result code; result="$("$PRODTAG_HELPER" emit --event-type "$event_type" --command "$command_text" --exit-code "$exit_code" --cwd "$PWD" --duration-ms "$duration_ms" 2>&1)"; code=$?; _prodtag_debug "helper_result_code=$code helper_result=$result" ) &!
  else
    "$PRODTAG_HELPER" emit --event-type "$event_type" --command "$command_text" --exit-code "$exit_code" --cwd "$PWD" --duration-ms "$duration_ms" >/dev/null 2>&1 &!
  fi
  return "$exit_code"
}

add-zsh-hook -d preexec _prodtag_preexec 2>/dev/null
add-zsh-hook -d precmd _prodtag_precmd 2>/dev/null
add-zsh-hook preexec _prodtag_preexec
add-zsh-hook precmd _prodtag_precmd
