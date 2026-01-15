# g9cc

This is a minimal self-made C compiler practice project.  
This is the [9cc](https://github.com/rui314/9cc) version.

## Usage (build -> generate -> assemble -> run)

### 1. Build the binary

```
go build -o g9cc
```

### 2. Generate assembly (redirect)

```
mkdir -p build
./g9cc 3 > build/out.s
```

### 3. Assemble + link

```
gcc -no-pie -o build/out build/out.s
```

### 4. Run

```
./build/out
```

The return value of `main` becomes the process exit code. To check it:

```
./build/out

echo $?
```

## Notes

- If no argument is provided or number conversion fails, it prints an error to stderr and exits.
- On macOS, `gcc`/`clang` options may differ.
- Go treats `.s` files in the package root as build targets, so generated files are written to `build/`.
