# ProdTag macOS zsh integration MVP.
# Source this file manually; it does not edit shell startup files.

autoload -Uz add-zsh-hook
zmodload zsh/datetime 2>/dev/null

: "${PRODTAG_ZSH_ENABLED:=1}"

_prodtag_script_dir="${${(%):-%N}:A:h}"
: "${PRODTAG_HELPER:=${_prodtag_script_dir}/../build/bin/prodtag-helper}"

_prodtag_last_command=""
_prodtag_started_ms=""

_prodtag_now_ms() {
  if [[ -n "${EPOCHREALTIME:-}" ]]; then
    printf '%.0f\n' "$(( EPOCHREALTIME * 1000 ))"
  else
    printf '%s000\n' "$(date +%s)"
  fi
}

_prodtag_should_ignore_command() {
  local command_text="${1:-}"
  local trimmed="${command_text##[[:space:]]}"
  trimmed="${trimmed%%[[:space:]]}"

  [[ -z "${command_text//[[:space:]]/}" ]] && return 0
  [[ "$command_text" == prodtag-helper* ]] && return 0
  [[ "$command_text" == *"/prodtag-helper"* ]] && return 0
  [[ "$command_text" == *"prodtag.zsh"* ]] && return 0

  return 1
}

_prodtag_infer_event_type() {
  local command_text="${(L)1}"
  local exit_code="${2:-0}"
  local suffix="success"
  if (( exit_code != 0 )); then
    suffix="failure"
  fi

  case "$command_text" in
    git\ commit*) echo "git_commit_${suffix}" ;;
    git\ push*) echo "git_push_${suffix}" ;;
    npm\ test*|npm\ run\ test*|pnpm\ test*|pnpm\ run\ test*|yarn\ test*|pytest*|go\ test*|cargo\ test*|flutter\ test*) echo "test_${suffix}" ;;
    npm\ run\ build*|pnpm\ build*|pnpm\ run\ build*|yarn\ build*|go\ build*|cargo\ build*|flutter\ build*) echo "build_${suffix}" ;;
    *) echo "command_${suffix}" ;;
  esac
}

_prodtag_preexec() {
  [[ "${PRODTAG_ZSH_ENABLED}" == "0" ]] && return 0

  _prodtag_last_command="$1"
  _prodtag_started_ms="$(_prodtag_now_ms)"
}

_prodtag_precmd() {
  local exit_code="$?"
  [[ "${PRODTAG_ZSH_ENABLED}" == "0" ]] && return 0
  [[ -x "$PRODTAG_HELPER" ]] || return 0
  [[ -n "$_prodtag_last_command" ]] || return 0

  local command_text="$_prodtag_last_command"
  local started_ms="$_prodtag_started_ms"
  _prodtag_last_command=""
  _prodtag_started_ms=""

  _prodtag_should_ignore_command "$command_text" && return 0

  local finished_ms="$(_prodtag_now_ms)"
  local duration_ms=""
  if [[ -n "$started_ms" ]]; then
    duration_ms="$(( finished_ms - started_ms ))"
    (( duration_ms < 0 )) && duration_ms=0
  fi

  local event_type="$(_prodtag_infer_event_type "$command_text" "$exit_code")"

  "$PRODTAG_HELPER" emit \
    --event-type "$event_type" \
    --command "$command_text" \
    --exit-code "$exit_code" \
    --cwd "$PWD" \
    --duration-ms "$duration_ms" >/dev/null 2>&1 &

  return 0
}

add-zsh-hook preexec _prodtag_preexec
add-zsh-hook precmd _prodtag_precmd
