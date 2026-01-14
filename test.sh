#!/usr/bin/env bash

assert() {
    expected="$1"
    input="$2"

    tmpdir="${TMPDIR:-.tmp-work}"
    mkdir -p "$tmpdir"

    ./g9cc "$input" > "$tmpdir/tmp.s"
    gcc -o "$tmpdir/tmp" "$tmpdir/tmp.s"
    "$tmpdir/tmp"
    actual="$?"

    if [ "$actual" = "$expected" ]; then
        echo "$input => $actual"
    else
        echo "$input => $expected expected, but got $actual"
        exit 1
    fi
}

assert 0 0
assert 42 42
assert 21 '5+20-4'
assert 30 '17+130-117'
assert 15 '1 + 3+10+4 -3 '
assert 20 ' 10 + 13-3'

echo OK
