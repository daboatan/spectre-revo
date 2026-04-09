#!/bin/bash
VERSION=1.0

function usage() {
	prog=$(basename "$0")
	echo "Syntax: $prog [-p] <filename> [language]" >&2
	echo "        $prog -u <paste> <filename> [language]	- Update <paste>" >&2
	echo "        $prog -e <paste> [language]			- Edit <paste> in \$EDITOR (or vi.)" >&2
	echo "        $prog -d <paste>				- Delete <paste>" >&2
	echo "        $prog -s <paste>				- Show <paste>" >&2
	echo "        $prog -l					- List pastes" >&2
	echo "        $prog -L					- Request Login" >&2
	echo "Options:" >&2
	echo "        -x <expiry>					- Expiration for paste (with units: ns/us/ms/s/m/h)" >&2
	echo "        -p						- Prompt for password" >&2
	echo "        -S <server>					- Override server" >&2
	echo "        -i						- Use http" >&2
	echo "        -I						- Use https, but disable certificate validation" >&2
}

if [[ -z $1 ]]; then
	usage
	exit 1
fi

rcdir="${HOME}/.spectre-updated"
if [[ ! -d "${rcdir}" ]]; then
	mkdir -p "${rcdir}"
fi

export -a curl_opts=(
	"-c" "${rcdir}/cookie.jar"
	"-b" "${rcdir}/cookie.jar"
	"-A" "ghost.sh/${VERSION}"
	"-f"
	"-s"
)

passworded=0

while getopts "d:e:hIiLlpS:s:u:x:" o; do
	case $o in
		d)
			mode="delete"
			paste=$OPTARG
			;;
		e)
			mode="edit"
			paste=$OPTARG
			;;
		h)
			usage
			exit 0
			;;
		I)
			curl_opts+=("-k")
			;;
		i)
			proto="http"
			;;
		L)
			mode="login"
			;;
		l)
			mode="list"
			;;
		p)
			passworded=1
			;;
		S)
			server="$OPTARG"
			;;
		s)
			mode="show"
			paste=$OPTARG
			;;
		u)
			mode="update"
			paste=$OPTARG
			;;
		x)
			expiry=$OPTARG
			;;
		?)
			usage
			exit 1
			;;
	esac
done

server=${server:-${SPECTRE_SERVER:-${SERVER:-${proto:-http}://127.0.0.1:9111}}}

function _password() {
	read -p "Password:" -r -s "$1" < /dev/tty
}

shift $((OPTIND-1))

filename="$1"
lang="text"
if [[ -n $2 ]]; then
	lang=$2
fi

if [[ "${mode}" == "delete" ]]; then
	IFS='|' read -r code < <(curl "${curl_opts[@]}" -w '%{http_code}' --data-urlencode "(no body)" "${server}/paste/${paste}/delete")
	if [[ $code -ne 200 && $code -ne 303 && $code -ne 302 ]]; then
		echo "Rejected: $code" >&2
		exit 1
	fi
	echo "Deleted $paste."
	exit 0

elif [[ "${mode}" == "edit" ]]; then
	filename=$(mktemp /tmp/ghost.XXXXXX)
	lang=$1
	curl "${curl_opts[@]}" -o "${filename}" "${server}/paste/${paste}/raw"
	${EDITOR:-vi} "${filename}"

elif [[ "${mode}" == "show" ]]; then
	curl "${curl_opts[@]}" "${server}/paste/${paste}/raw"
	exit 0

elif [[ "${mode}" == "list" ]]; then
	IFS=' ' read -r -a pastes < <(curl "${curl_opts[@]}" "${server}/session/raw")
	for i in "${pastes[@]}"; do
		echo "$i: ${server}/paste/$i"
	done
	exit 0

elif [[ "${mode}" == "login" ]]; then
	url="${server}/auth/token"
	IFS='|' read -r code url < <(curl "${curl_opts[@]}" -w '%{http_code}|%{redirect_url}' "${url}" | sed -e 's/HTTP/http/g')
	if [[ $code -ne 200 && $code -ne 303 && $code -ne 302 ]]; then
		echo "Rejected: $code" >&2
		exit 1
	fi

	token=${url##*/}

	echo "To log in, please visit $url" >&2

	type open &>/dev/null && open "$url"
	type xdg-open &>/dev/null && xdg-open "$url"

	echo "" >&2
	echo "(waiting for login)" >&2
	{
		l=0
		trap "l=-1" 2
		filename=$(mktemp /tmp/ghost.XXXXXX)
		while [[ $l -eq 0 ]]; do
			sleep 2
			IFS='|' read -r code < <(curl -w '%{http_code}' -o "${filename}" "${curl_opts[@]}" --data-urlencode "type=token" --data-urlencode "token=${token}" "${server}/auth/login")
			if [[ $code -ne 200 && $code -ne 418 ]]; then
				echo "" >&2
				echo "Login Rejected. Detailed response follows." >&2
				cat "${filename}" >&2
				echo "" >&2
				l=-1
			elif [[ $code -eq 200 ]]; then
				echo "" >&2
				echo "Success!" >&2
				l=1
			else
				printf "."
			fi
		done
		rm -f "${filename}"
	}
	exit 0
fi

if [[ -z "${filename}" ]]; then
	usage
	exit 1
fi

pboard=
[[ -z "${pboard}" ]] && type pbcopy &>/dev/null && pboard=pbcopy
[[ -z "${pboard}" ]] && type xclip &>/dev/null && [[ -n "${DISPLAY}" ]] && pboard=xclip

url="${server}/paste/new"
[[ "${mode}" == "edit" || "${mode}" == "update" ]] && url="${server}/paste/${paste}/edit"

[[ $passworded -eq 1 ]] && _password pw && echo

export -a curl_formargs=("--data-urlencode" "text@$filename")
[[ -n "${lang}" ]] && curl_formargs+=("--data-urlencode" "lang=${lang}")
[[ -n "${pw}" ]] && curl_formargs+=("--data-urlencode" "password=${pw}")
[[ -n "${expiry}" ]] && curl_formargs+=("--data-urlencode" "expire=${expiry}")

IFS='|' read -r code url < <(curl "${curl_opts[@]}" -w '%{http_code}|%{redirect_url}' "${curl_formargs[@]}" "${url}" | sed -e 's/HTTP/http/g')

[[ "${mode}" == "edit" ]] && rm -f "${filename}"

if [[ $code -ne 200 && $code -ne 303 && $code -ne 302 ]]; then
	echo "Rejected: $code" >&2
	exit 1
fi

echo "$url"
[[ -n "${pboard}" ]] && (echo -n "$url" | "$pboard"; echo "Paste URL copied to clipboard." >&2)