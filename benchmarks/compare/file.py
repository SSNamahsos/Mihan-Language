lines = []
for i in range(200000):
    lines.append("khat " + str(i))
with open("/tmp/mihan-bench.txt", "w") as f:
    f.write("\n".join(lines))
with open("/tmp/mihan-bench.txt") as f:
    back = f.read().split("\n")
print(len(back))
