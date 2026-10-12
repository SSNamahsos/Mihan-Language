# Benchmarks

Numbers, not wishes. Run them with:

```
.\build.ps1 bench
```

## What each one measures

| Name | What it does | Why it is here |
|---|---|---|
| `fib` | `fib(22)`, recursive | call overhead, which every language pays most for |
| `loop` | three million counted turns of arithmetic | the commonest thing a program does |
| `list` | build 200,000 items, then walk them | list growth and item access |
| `text` | read one character 400,000 times out of a 4,000-character text | character work, and the one that finds a whole-text copy |
| `class` | a method call a million times through an object | field lookup and method binding |
| `closure` | call a function value a million times | the cost of a function being a value |
| `MihanCompile` | read and parse, no running | what a build system pays once and then reuses |

`BenchmarkGoDirect` runs the same `fib` and `loop` in plain Go with no
abstraction. It is the far end of the scale, not a target for a tree-walking
interpreter.

## The baseline, on an i5-9500 at 3.00 GHz, Windows

Taken with `-benchtime=1x`, so one whole program per row. These are the numbers
the bytecode virtual machine has to beat.

```
BenchmarkMihan/class-6          799414600 ns/op    704015160 B/op   10999844 allocs/op
BenchmarkMihan/closure-6        660622400 ns/op    472005000 B/op    8999669 allocs/op
BenchmarkMihan/fib-6            61558000 ns/op      28893944 B/op     401306 allocs/op
BenchmarkMihan/list-6          155862500 ns/op      57947624 B/op    1399795 allocs/op
BenchmarkMihan/loop-6          664590400 ns/op      72000912 B/op    8999639 allocs/op
BenchmarkMihan/text-6         1445615500 ns/op      30199144 B/op    2493935 allocs/op
BenchmarkMihanCompile/class-6       78500 ns/op          17584 B/op       130 allocs/op
BenchmarkMihanCompile/closure-6     42600 ns/op          15088 B/op        78 allocs/op
BenchmarkMihanCompile/fib-6         39800 ns/op          15296 B/op        83 allocs/op
BenchmarkMihanCompile/list-6        41400 ns/op          15576 B/op        82 allocs/op
```

## What was found by measuring, and what it changed

The first run of this suite was three times worse than the one above, and none of
it was guesswork. A memory profile of a plain counting loop said:

```
    47.59%  runtime.NewEnv        a new scope for every turn of the loop
    27.83%  Env.Declare           a new binding for the loop variable
    13.82%  eval.arith            boxing a number into an interface
     9.21%  execFor               the rest of the loop
```

Three quarters of every allocation was bookkeeping around the loop rather than
anything the program did. Two changes removed it:

- A loop that cannot capture its scope now runs every turn over the same scope,
  emptied by counting turns rather than by clearing a map. A funct or a lambda in
  the body is the one thing that can hold a scope alive, so a body that makes
  neither gets a scope built once instead of once per turn.
- A name given a value inside a funct belongs to that funct, so a name is a slot
  in a frame rather than a search up a chain of scopes.

After that, one hundred per cent of what was left was `interface{}` boxing:

```
    74.96%  eval.arith
    24.99%  execFor
```

That is the thing the virtual machine exists to fix, and it is why
`internal/ir` was written next. A value in a machine is a slot on a stack, so a
number is a number and boxing it is not a thing that can happen.

| | before | now | |
|---|---|---|---|
| `loop` | 1 698 303 600 ns, 1 512 016 280 B, 26 999 742 allocs | 664 590 400 ns, 72 000 912 B, 8 999 639 allocs | **2.6x faster, 21x less memory** |
| `list` | 284 690 900 ns, 249 946 688 B, 3 799 776 allocs | 155 862 500 ns, 57 947 624 B, 1 399 795 allocs | 1.8x faster |
| `class` | 998 935 100 ns, 1 152 013 560 B, 16 999 859 allocs | 799 414 600 ns, 704 015 160 B, 10 999 844 allocs | 1.2x faster |
| `text` | 193 846 112 700 ns, 321 345 261 040 B | 1 445 615 500 ns, 30 199 144 B | **134x faster** |

## What the numbers say, plainly

**Front-end speed is already good.** Reading and parsing a whole program takes
40 to 80 microseconds. A build system pays that once per file, so compilation
was never going to be the bottleneck.

**`text` was 130x too slow and now is not.** The first run took 193 seconds and
allocated 321 GB. Reading one character turned the whole string into a slice of
characters first, so a loop over long text copied the text over and over.
`runtime.RuneAt` now finds the character in place, and
`TestReadingTextDoesNotCopyTheWholeThing` fails if that ever comes back.

**`fib` barely moved, and that is the honest shape of the remaining work.** A
recursive call is not spent on scopes or on boxing numbers, it is spent on making
a call, so it is the one thing the scope work could not touch. Making a call
cheap is the machine's job, and it is the next number that will move.

## The plan from here

1. ~~Benchmarks.~~ Done, and they found two hundred times the time.
2. ~~Loop scopes and name resolution.~~ Done; the profile above is the receipt.
3. **The bytecode virtual machine** over `internal/ir`, with a value that is a
   stack slot rather than a boxed interface. `mihan compile` prints the
   instructions a program becomes, so what the machine is told can be read.
4. Inline caches on field and method lookup, and packed small integers.
5. Windows programs on WebView2, then the web.


Every one of those is a swap of one phase. Nothing a program can see changes.
