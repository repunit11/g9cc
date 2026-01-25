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

assert 0 '0;'
assert 42 '42;'
assert 21 '5+20-4;'
assert 30 '17+130-117;'
assert 15 '1 + 3+10+4 -3 ;'
assert 20 ' 10 + 13-3;'
assert 3 ' 5*3/5;'
assert 20 '5+3*3+6;'
assert 20 '(4+1)*4;'
assert 3 '1--2;'
assert 4 '1*-2+6;'
assert 1 '0==0;'
assert 0 '0!=0;'
assert 1 '1==1;'
assert 0 '1==2;'
assert 1 '1!=2;'
assert 1 '0<1;'
assert 0 '1<0;'
assert 1 '1<=1;'
assert 0 '2<=1;'
assert 1 '2>1;'
assert 1 '2>=2;'
assert 0 '1>=2;'
assert 1 '1+2==3;'
assert 1 '1+2<4;'
assert 1 '-1<0;'
assert 3 '1;2; 3;'
assert 2 'a=2; a;'
assert 10 'a=2; b=3+2; a*b;'
assert 5 'foo=2; bar=3; foo+bar;'
assert 5 'var1=2; var2=3; var1+var2;'
assert 9 'ab=4; abc=5; ab+abc;'
assert 2 'return 2;'
assert 3 'a=1; return a+2;'
assert 7 'a=5; b=2; return a+b; 9;'
assert 2 'if (1) 2; else 3;'
assert 3 'if (0) 2; else 3;'
assert 7 'if (2-1) 7;'
assert 9 'if (0) 7; 9;'
assert 3 'i=0; while(i<3) i=i+1; i;'
assert 0 'i=0; while(i<0) i=i+1; i;'
assert 8 'i=0; sum=0; while((i=i+1)<5) sum=sum+2; sum;'
assert 3 'i=0; for(i=0;i<3;i=i+1) i; return i;'
assert 3 'i=0; for(;i<3;i=i+1) i; return i;'
assert 3 'for(;;) return 3; return 5;'

echo OK
