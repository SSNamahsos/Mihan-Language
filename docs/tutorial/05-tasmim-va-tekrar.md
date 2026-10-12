# Lesson 5 — Decisions and loops

> **خلاصه:** `agar` یعنی «اگر»، `warna` یعنی «وگرنه»، `ta` یعنی «تا وقتی که»، `har` یعنی «برای هر».
> بلوک‌ها با تورفتگی (۴ فاصله) مشخص می‌شوند، مثل پایتون.

## 5.1 agar / warna

A block is the indented lines under a line. Use 4 spaces. The colon at the end
of the line is optional.

```mihan
nomre = 17
agar nomre >= 18
    chap_khat "aali"
warna agar nomre >= 12
    chap_khat "ghabool"
warna
    chap_khat "mardood"
```

```output
ghabool
```

## 5.2 har — counting

`har i az a ta b` counts from `a` up to but **not including** `b`. Write `ta b+`
to include `b`.

```mihan
har i az 0 ta 3
    chap_khat i
har i az 1 ta 3+
    chap_khat $"dor {i}"
```

```output
0
1
2
dor 1
dor 2
dor 3
```

## 5.3 har — walking a list

```mihan
rangha = ["ghermez", "sabz", "abi"]
har r dar rangha
    chap_khat r.gooshband()
har jofte dar doreh(rangha)
    chap_khat $"{jofte[0]}: {jofte[1]}"
```

```output
GHERMEZ
SABZ
ABI
0: ghermez
1: sabz
2: abi
```

`ke` keeps only the items that pass a test:

```mihan
har n az 1 ta 10 ke n % 3 == 0
    chap_khat n
```

```output
3
6
9
```

## 5.4 ta — repeat while

```mihan
pool = 100
sal = 0
ta pool < 200
    pool = pool * 1.25
    sal += 1
chap_khat $"{sal} sal, pool: {pool}"
```

```output
4 sal, pool: 244.140625
```

## 5.5 shek and edame

`shek` (break) leaves the loop. `edame` (continue) jumps to the next turn.

```mihan
har i az 1 ta 10
    agar i == 3
        edame
    agar i == 6
        shek
    chap_khat i
```

```output
1
2
4
5
```

## 5.6 Names made inside a block are still there afterwards

Like Python, a name made inside an indented block (`agar`, `har`, `ta`,
`koshish`) belongs to the function (or file) around it, so you can use it
after the block:

```mihan
har i az 0 ta 3
    akhar = i
chap_khat akhar

agar dorost
    payam = "salam"
chap_khat payam
```

```output
2
salam
```

Names made inside a `funct` still stay inside that function.

## 5.7 A small program: FizzBuzz

```mihan
har i az 1 ta 15+
    agar i % 15 == 0
        chap_khat "FizzBuzz"
    warna agar i % 3 == 0
        chap_khat "Fizz"
    warna agar i % 5 == 0
        chap_khat "Buzz"
    warna
        chap_khat i
```

```output
1
2
Fizz
4
Buzz
Fizz
7
8
Fizz
Buzz
11
Fizz
13
14
FizzBuzz
```

## Exercises

1. Print the multiplication table of 7 (7 x 1 … 7 x 10).
2. Add up all numbers from 1 to 100.
3. Find the first number above 1000 that is divisible by 37.

<details><summary>Solutions</summary>

```mihan
har i az 1 ta 10+
    chap_khat $"7 x {i} = {7 * i}"
jam = 0
har i az 1 ta 100+
    jam += i
chap_khat jam
n = 1001
ta n % 37 != 0
    n += 1
chap_khat n
```

```output
7 x 1 = 7
7 x 2 = 14
7 x 3 = 21
7 x 4 = 28
7 x 5 = 35
7 x 6 = 42
7 x 7 = 49
7 x 8 = 56
7 x 9 = 63
7 x 10 = 70
5050
1036
```

</details>

**Next:** [Lesson 6 — Functions](06-tabe-ha.md)
