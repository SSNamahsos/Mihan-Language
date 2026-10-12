parts = []
for i in range(300000):
    parts.append("x" + str(i))
s = ",".join(parts)
t = ""
for i in range(100000):
    t += "ab"
print(len(s), len(t))
