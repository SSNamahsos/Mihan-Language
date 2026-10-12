def fib(n):
  return n if n<2 else fib(n-1)+fib(n-2)
print(fib(30))
s=0
for i in range(5000000): s+=i
print(s)
