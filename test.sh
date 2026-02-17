#!/usr/bin/env bash
tmpdir="${TMPDIR:-.tmp-work}"
cat <<EOF | gcc -xc -c -o $tmpdir/tmp2.o -
int ret3() { return 3; }
int ret5() { return 5; }
int add2(int a, int b) { return a + b; }
int add3(int a, int b, int c) { return a + b + c; }
int sub2(int a, int b) { return a - b; }
EOF

assert() {
    expected="$1"
    input="$2"

    mkdir -p "$tmpdir"

    ./g9cc "int main(){ $input }" > "$tmpdir/tmp.s"
    gcc -o "$tmpdir/tmp" "$tmpdir/tmp.s" "$tmpdir/tmp2.o"
    "$tmpdir/tmp"
    actual="$?"

    if [ "$actual" = "$expected" ]; then
        echo "$input => $actual"
    else
        echo "$input => $expected expected, but got $actual"
        exit 1
    fi
}

assert_prog() {
    expected="$1"
    input="$2"

    mkdir -p "$tmpdir"

    ./g9cc "$input" > "$tmpdir/tmp.s"
    gcc -o "$tmpdir/tmp" "$tmpdir/tmp.s" "$tmpdir/tmp2.o"
    "$tmpdir/tmp"
    actual="$?"

    if [ "$actual" = "$expected" ]; then
        echo "$input => $actual"
    else
        echo "$input => $expected expected, but got $actual"
        exit 1
    fi
}

assert_prog 3 'int main() {int x;int *y;y = &x;*y = 3;return x;}'
assert_prog 0 'int main(){ return 0; }'
assert_prog 3 'int id(int x){ return x; } int main(){ return id(3); }'
assert_prog 5 'int add(int x, int y){ return x+y; } int main(){ return add(2,3); }'
assert_prog 3 'int main(){ int x; x=3; return x; }'
assert_prog 8 'int main() { int *x; return sizeof(x); }'
assert_prog 4 'int main() { int x; x=1; return sizeof(x=2); }'
assert_prog 1 'int main() { int x;x=1; sizeof(x=2); return x; }'
assert 4 "sizeof(1);"
assert 8 "int a;a=1; sizeof(&a);"
assert 7 '{ int x; x=3; int y;y=5; *(&y+2-1)=7; return x; }'
assert 8 '{ int x; x=3; int y; y=5; x=x+y; return x; }'
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
assert 2 'int a; a=2; a;'
assert 10 'int a; int b; a=2; b=3+2; a*b;'
assert 5 'int foo; int bar; foo=2; bar=3; foo+bar;'
assert 5 'int var1; int var2; var1=2; var2=3; var1+var2;'
assert 9 'int ab; int abc; ab=4; abc=5; ab+abc;'
assert 2 'return 2;'
assert 3 'int a; a=1; return a+2;'
assert 7 'int a; int b; a=5; b=2; return a+b; 9;'
assert 2 'if (1) 2; else 3;'
assert 3 'if (0) 2; else 3;'
assert 7 'if (2-1) 7;'
assert 9 'if (0) 7; 9;'
assert 3 'int i; i=0; while(i<3) i=i+1; i;'
assert 0 'int i; i=0; while(i<0) i=i+1; i;'
assert 8 'int i; int sum; i=0; sum=0; while((i=i+1)<5) sum=sum+2; sum;'
assert 3 'int i; i=0; for(i=0;i<3;i=i+1) i; return i;'
assert 3 'int i; i=0; for(;i<3;i=i+1) i; return i;'
assert 3 'for(;;) return 3; return 5;'
assert 3 'int a; int b; { a=1; b=2; } a+b;'
assert 3 'int a; a=0; if (1) { a=3; } a;'
assert 4 'int a; a=0; if (0) { a=3; } else { a=4; } a;'
assert 6 'int i; int sum; i=0; sum=0; while(i<3) { sum=sum+2; i=i+1; } sum;'
assert 3 'int i; int sum; sum=0; for(i=0;i<3;i=i+1) { sum=sum+1; } sum;'
assert 3 'return ret3();'
assert 5 '{return ret5(); }'
assert 5 'return add2(2,3);'
assert 6 'return add3(1,2,3);'
assert 2 'return sub2(5,3);'
assert 9 'int a; int b; a=4; b=5; return add2(a,b);'
assert 7 'return add3(1+1,2,3);'
assert 6 'return add2(add2(1,2),3);'
assert 3 'int a; int b; a=3; b=&a; return *b;'
assert 5 'int x; int y; x=3; y=&x; *y=5; return x;'
assert 7 'int x; int y; x=3; y=5; *(&x-1)=7; return y;'
assert 7 'int x; int y; x=3; y=5; *(&y+1)=7; return x;'
assert_prog 3 'int foo(){ return 3; } int main(){ return foo(); }'
assert_prog 7 'int foo(){ return 2; } int bar(){ return 5; } int main(){ return foo()+bar(); }'
assert_prog 8 'int fib(int n){ if (n<=1) return n; return fib(n-1)+fib(n-2); } int main(){ return fib(6); }'
echo OK
