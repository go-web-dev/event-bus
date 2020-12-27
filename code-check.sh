#!/bin/sh

# Download the following before running the script
# go get -u golang.org/x/lint
# go get -u golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow

files=$(find . -not -path './vendor*' -name '*.go')
status=0

# go fmt
fmtOut=$(echo "$files" | xargs gofmt -l)
if [ -n "$fmtOut" ]; then
	printf "some .go files aren't formatted:\n%s\n" "$fmtOut"
	status=1
fi

# go vet
shadowOut=$(go vet -vettool="$(command -v shadow)" -tags="integration" ./... 2>&1)
vetOut="${shadowOut}"$(go vet -all -tags="integration" ./... 2>&1)
if [ -n "$vetOut" ]; then
	printf "go vet issues found:\n%s\n" "$vetOut"
	status=1
fi

# go lint
lintOut=$(go list ./... | xargs golint)
if [ -n "$lintOut" ]; then
	printf "go lint issues found:\n%s\n" "$lintOut"
	status=1
fi

exit $status
