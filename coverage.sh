#!/bin/sh

workdir=".cover"
profile="$workdir/cover.out"
show_cover_report_html=0
tags=""
test_failed=0

generate_cover_data() {
    mkdir -p "$workdir"
    rm -rf "$workdir/*"

    if ! go test -cover -covermode="count" -coverprofile="$profile" -coverpkg=./... ./... -tags="$tags"
    then
        test_failed=1
    fi
}

show_cover_report() {
    go tool cover -${1}="$profile"
}

parse_cmd_flags() {
    for i in "$@"
    do
        case "$i" in
            "")
            ;;
            --html)
                show_cover_report_html=1
            ;;
          --intgr)
                tag="$tag intgr"
            ;;
        esac
    done
}

parse_cmd_flags "$@"
generate_cover_data
show_cover_report func

if [ "$show_cover_report_html" -eq 1 ]; then
    show_cover_report html
    go tool cover -html="$profile" -o="$workdir/cover.html"
fi

exit $test_failed
